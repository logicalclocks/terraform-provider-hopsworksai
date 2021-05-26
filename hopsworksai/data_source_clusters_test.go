package hopsworksai

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestAccClustersDataSourceAWS_basic(t *testing.T) {
	testSkipAWS(t)
	testAccClustersDataSource_basic(t, api.AWS)
}

func TestAccClustersDataSourceAZURE_basic(t *testing.T) {
	testSkipAZURE(t)
	testAccClustersDataSource_basic(t, api.AZURE)
}

func testAccClustersDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	dataSourceName := fmt.Sprintf("data.hopsworksai_clusters.%s", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  testAccPreCheck(t),
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(cloud, rName, suffix),
				Check:  testAccClustersDataSourceCheckAllAttributes(cloud, resourceName, dataSourceName),
			},
		},
	})
}

func testAccClustersDataSourceCheckAllAttributes(cloud api.CloudProvider, resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var index string = ""
		listClustersTagPattern := regexp.MustCompile(`^clusters\.([0-9]*)\.tags.ListClusters$`)
		for k, v := range s.RootModule().Resources[dataSourceName].Primary.Attributes {
			submatches := listClustersTagPattern.FindStringSubmatch(k)
			if len(submatches) == 2 && v == cloud.String() {
				index = submatches[1]
			}
		}

		if index == "" {
			return fmt.Errorf("no clusters returned")
		}

		for k := range s.RootModule().Resources[resourceName].Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			dataSourceKey := fmt.Sprintf("clusters.%s.%s", index, k)
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, dataSourceKey)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClustersDataSourceConfig(cloud api.CloudProvider, rName string, suffix string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"
		head {
		}

		%s


		tags = {
		  "ListClusters" = "%s"
		  "Purpose" = "acceptance-test"
		}
	  }

	  data "hopsworksai_clusters" "%s" {
		  depends_on = [
			hopsworksai_cluster.%s
		  ]
	  }
	`,
		rName,
		clusterPrefixName,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testAccClusterCloudConfigAttributes(cloud, 2),
		cloud.String(),
		rName,
		rName,
	)
}

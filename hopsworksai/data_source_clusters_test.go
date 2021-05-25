package hopsworksai

import (
	"fmt"
	"strings"
	"testing"

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
	testAccClusterDataSource_basic(t, api.AZURE)
}

func testAccClustersDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	resourceName := "hopsworksai_cluster.test"
	dataSourceName := "data.hopsworksai_clusters.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  testAccPreCheck(t),
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(cloud),
				Check:  testAccClustersDataSourceCheckAllAttributes(resourceName, dataSourceName),
			},
		},
	})
}

func testAccClustersDataSourceCheckAllAttributes(resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for k := range s.RootModule().Resources[resourceName].Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			dataSourceKey := fmt.Sprintf("clusters.0.%s", k)
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, dataSourceKey)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClustersDataSourceConfig(cloud api.CloudProvider) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "test" {
		name    = "%sds%s"
		ssh_key = "%s"	  
		head {
		}
		
		%s
		

		tags = {
		  "Purpose" = "acceptance-test"
		}
	  }

	  data "hopsworksai_clusters" "test" {

	  }
	`, clusterPrefixName, strings.ToLower(cloud.String()), testAccClusterCloudSSHKeyAttribute(cloud), testAccClusterCloudConfigAttributes(cloud))
}

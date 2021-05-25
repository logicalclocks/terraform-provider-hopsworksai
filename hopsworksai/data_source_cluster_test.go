package hopsworksai

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestAccClusterDataSourceAWS_basic(t *testing.T) {
	testSkipAWS(t)
	testAccClusterDataSource_basic(t, api.AWS)
}

func TestAccClusterDataSourceAZURE_basic(t *testing.T) {
	testSkipAZURE(t)
	testAccClusterDataSource_basic(t, api.AZURE)
}

func testAccClusterDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	resourceName := "hopsworksai_cluster.test"
	dataSourceName := "data.hopsworksai_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  testAccPreCheck(t),
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig(cloud),
				Check:  testAccClusterDataSourceCheckAllAttributes(resourceName, dataSourceName),
			},
		},
	})
}

func testAccClusterDataSourceCheckAllAttributes(resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for k := range s.RootModule().Resources[resourceName].Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, k)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClusterDataSourceConfig(cloud api.CloudProvider) string {
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

	  data "hopsworksai_cluster" "test" {
		  cluster_id = hopsworksai_cluster.test.id
	  }
	`, clusterPrefixName, strings.ToLower(cloud.String()), testAccClusterCloudSSHKeyAttribute(cloud), testAccClusterCloudConfigAttributes(cloud))
}

package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func TestAccInstanceTypesDataSource_AWS_head(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AWS, api.HeadNode)
}

func TestAccInstanceTypesDataSource_AWS_worker(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AWS, api.WorkerNode)
}

func TestAccInstanceTypesDataSource_AWS_RonDBManagement(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AWS, api.RonDBManagementNode)
}

func TestAccInstanceTypesDataSource_AWS_RonDBData(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AWS, api.RonDBDataNode)
}

func TestAccInstanceTypesDataSource_AWS_RonDBMySQL(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AWS, api.RonDBMySQLNode)
}

func TestAccInstanceTypesDataSource_AWS_RonDBAPI(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AWS, api.RonDBMySQLNode)
}

func TestAccInstanceTypesDataSource_AZURE_head(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AZURE, api.HeadNode)
}

func TestAccInstanceTypesDataSource_AZURE_worker(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AZURE, api.WorkerNode)
}

func TestAccInstanceTypesDataSource_AZURE_RonDBManagement(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AZURE, api.RonDBManagementNode)
}

func TestAccInstanceTypesDataSource_AZURE_RonDBData(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AZURE, api.RonDBDataNode)
}

func TestAccInstanceTypesDataSource_AZURE_RonDBMySQL(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AZURE, api.RonDBMySQLNode)
}

func TestAccInstanceTypesDataSource_AZURE_RonDBAPI(t *testing.T) {
	testAccInstanceTypesDataSource_basic(t, api.AZURE, api.RonDBMySQLNode)
}

func testAccInstanceTypesDataSource_basic(t *testing.T, cloud api.CloudProvider, nodeType api.NodeType) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	dataSourceName := fmt.Sprintf("data.hopsworksai_instance_types.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypesDataSourceConfig(cloud, rName, nodeType),
				Check:  testAccInstanceTypesDataSourceCheckSupportedTypes(dataSourceName),
			},
		},
	})
}

func testAccInstanceTypesDataSourceCheckSupportedTypes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found", dataSourceName)
		}
		for k, v := range ds.Primary.Attributes {
			if k == "supported_types.#" {
				if intV, err := strconv.Atoi(v); err != nil {
					return err
				} else {
					if intV > 0 {
						break
					} else {
						return fmt.Errorf("expected list of supported types but got %#v", v)
					}
				}
			}
		}
		return nil
	}
}

func testAccInstanceTypesDataSourceConfig(cloud api.CloudProvider, rName string, nodeType api.NodeType) string {
	return fmt.Sprintf(`
		data "hopsworksai_instance_types" "%s"{
			cloud_provider = "%s"
			node_type = "%s"
		}
	`, rName, cloud.String(), nodeType.String())
}

// Unit test
func TestInstanceTypesDataSourceRead(t *testing.T) {
	for _, c := range []api.CloudProvider{api.AWS, api.AZURE} {
		testInstanceTypesDataSource(t, c, api.HeadNode, []interface{}{
			map[string]interface{}{
				"id":     "head-type-1",
				"memory": 20.0,
				"cpus":   10,
				"gpus":   0,
			},
			map[string]interface{}{
				"id":     "head-type-2",
				"memory": 50.0,
				"cpus":   20,
				"gpus":   1,
			},
		})

		testInstanceTypesDataSource(t, c, api.WorkerNode, []interface{}{
			map[string]interface{}{
				"id":     "worker-type-1",
				"memory": 20.0,
				"cpus":   10,
				"gpus":   0,
			},
			map[string]interface{}{
				"id":     "worker-type-2",
				"memory": 50.0,
				"cpus":   20,
				"gpus":   1,
			},
		})

		testInstanceTypesDataSource(t, c, api.RonDBManagementNode, []interface{}{
			map[string]interface{}{
				"id":     "mgm-type-2",
				"memory": 20.0,
				"cpus":   2,
				"gpus":   0,
			},
			map[string]interface{}{
				"id":     "mgm-type-1",
				"memory": 30.0,
				"cpus":   16,
				"gpus":   0,
			},
		})

		testInstanceTypesDataSource(t, c, api.RonDBDataNode, []interface{}{
			map[string]interface{}{
				"id":     "ndbd-type-2",
				"memory": 50.0,
				"cpus":   8,
				"gpus":   0,
			},
			map[string]interface{}{
				"id":     "ndbd-type-1",
				"memory": 100.0,
				"cpus":   16,
				"gpus":   0,
			},
		})

		testInstanceTypesDataSource(t, c, api.RonDBMySQLNode, []interface{}{
			map[string]interface{}{
				"id":     "mysql-type-2",
				"memory": 50.0,
				"cpus":   8,
				"gpus":   0,
			},
			map[string]interface{}{
				"id":     "mysql-type-1",
				"memory": 100.0,
				"cpus":   16,
				"gpus":   0,
			},
		})

		testInstanceTypesDataSource(t, c, api.RonDBAPINode, []interface{}{
			map[string]interface{}{
				"id":     "api-type-2",
				"memory": 50.0,
				"cpus":   8,
				"gpus":   0,
			},
			map[string]interface{}{
				"id":     "api-type-1",
				"memory": 100.0,
				"cpus":   16,
				"gpus":   0,
			},
		})
	}
}

func testInstanceTypesDataSource(t *testing.T, cloud api.CloudProvider, nodeType api.NodeType, expectedTypes []interface{}) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/nodes/supported-types",
				Response: fmt.Sprintf(`{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload": {
						"%s": {
							"head": [
								{
									"id": "head-type-1",
									"memory": 20,
									"cpus": 10,
									"gpus": 0
								},
								{
									"id": "head-type-2",
									"memory": 50,
									"cpus": 20,
									"gpus": 1
								}
							],
							"worker": [
								{
									"id": "worker-type-1",
									"memory": 20,
									"cpus": 10,
									"gpus": 0
								},
								{
									"id": "worker-type-2",
									"memory": 50,
									"cpus": 20,
									"gpus": 1
								}
							],
							"ronDB": {
								"mgmd": [
									{
										"id": "mgm-type-1",
										"memory": 30,
										"cpus": 16,
										"gpus": 0
									},
									{
										"id": "mgm-type-2",
										"memory": 20,
										"cpus": 2,
										"gpus": 0
									}
								],
								"ndbd": [
									{
										"id": "ndbd-type-1",
										"memory": 100,
										"cpus": 16,
										"gpus": 0
									},
									{
										"id": "ndbd-type-2",
										"memory": 50,
										"cpus": 8,
										"gpus": 0
									}
								],
								"mysqld": [
									{
										"id": "mysql-type-1",
										"memory": 100,
										"cpus": 16,
										"gpus": 0
									},
									{
										"id": "mysql-type-2",
										"memory": 50,
										"cpus": 8,
										"gpus": 0
									}
								],
								"api": [
									{
										"id": "api-type-1",
										"memory": 100,
										"cpus": 16,
										"gpus": 0
									},
									{
										"id": "api-type-2",
										"memory": 50,
										"cpus": 8,
										"gpus": 0
									}
								]
							}
						}
					}
				}`, strings.ToLower(cloud.String())),
			},
		},
		Resource:             dataSourceInstanceTypes(),
		OperationContextFunc: dataSourceInstanceTypes().ReadContext,
		State: map[string]interface{}{
			"node_type":      nodeType.String(),
			"cloud_provider": cloud.String(),
		},
		ExpectState: map[string]interface{}{
			"supported_types": expectedTypes,
		},
		ExpectId: cloud.String() + "-" + nodeType.String(),
	}
	r.Apply(t, context.TODO())
}

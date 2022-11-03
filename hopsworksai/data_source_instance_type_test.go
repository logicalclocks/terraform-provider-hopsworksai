package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func TestInstanceTypeDataSourceRead(t *testing.T) {
	for _, c := range []api.CloudProvider{api.AWS, api.AZURE} {
		testInstanceTypeDataSource(t, c, api.HeadNode, "head-type-1")
		testInstanceTypeDataSource(t, c, api.WorkerNode, "worker-type-1")
		testInstanceTypeDataSource(t, c, api.RonDBManagementNode, "mgm-type-2")
		testInstanceTypeDataSource(t, c, api.RonDBDataNode, "ndbd-type-2")
		testInstanceTypeDataSource(t, c, api.RonDBMySQLNode, "mysql-type-2")
		testInstanceTypeDataSource(t, c, api.RonDBAPINode, "api-type-2")
	}
}

func TestInstanceTypeDataSourceRead_filtered(t *testing.T) {
	for _, c := range []api.CloudProvider{api.AWS, api.AZURE} {
		testInstanceTypeDataSourceBase(t, c, api.HeadNode, 20, 0, 0, false, "head-type-1")
		testInstanceTypeDataSourceBase(t, c, api.HeadNode, 21, 0, 0, false, "head-type-2")
		testInstanceTypeDataSourceBase(t, c, api.HeadNode, 0, 10, 0, false, "head-type-1")
		testInstanceTypeDataSourceBase(t, c, api.HeadNode, 0, 11, 0, false, "head-type-2")
		testInstanceTypeDataSourceBase(t, c, api.HeadNode, 0, 0, 1, false, "head-type-2")

		testInstanceTypeDataSourceBase(t, c, api.WorkerNode, 20, 0, 0, false, "worker-type-1")
		testInstanceTypeDataSourceBase(t, c, api.WorkerNode, 21, 0, 0, false, "worker-type-2")
		testInstanceTypeDataSourceBase(t, c, api.WorkerNode, 0, 10, 0, false, "worker-type-1")
		testInstanceTypeDataSourceBase(t, c, api.WorkerNode, 0, 11, 0, false, "worker-type-2")
		testInstanceTypeDataSourceBase(t, c, api.WorkerNode, 0, 0, 1, false, "worker-type-2")
		testInstanceTypeDataSourceBase(t, c, api.WorkerNode, 0, 0, 0, true, "worker-type-3")

		testInstanceTypeDataSourceBase(t, c, api.RonDBManagementNode, 20, 0, 0, false, "mgm-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBManagementNode, 21, 0, 0, false, "mgm-type-1")
		testInstanceTypeDataSourceBase(t, c, api.RonDBManagementNode, 0, 2, 0, false, "mgm-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBManagementNode, 0, 3, 0, false, "mgm-type-1")

		testInstanceTypeDataSourceBase(t, c, api.RonDBDataNode, 50, 0, 0, false, "ndbd-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBDataNode, 51, 0, 0, false, "ndbd-type-1")
		testInstanceTypeDataSourceBase(t, c, api.RonDBDataNode, 0, 8, 0, false, "ndbd-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBDataNode, 0, 9, 0, false, "ndbd-type-1")

		testInstanceTypeDataSourceBase(t, c, api.RonDBMySQLNode, 50, 0, 0, false, "mysql-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBMySQLNode, 51, 0, 0, false, "mysql-type-1")
		testInstanceTypeDataSourceBase(t, c, api.RonDBMySQLNode, 0, 8, 0, false, "mysql-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBMySQLNode, 0, 9, 0, false, "mysql-type-1")

		testInstanceTypeDataSourceBase(t, c, api.RonDBAPINode, 50, 0, 0, false, "api-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBAPINode, 51, 0, 0, false, "api-type-1")
		testInstanceTypeDataSourceBase(t, c, api.RonDBAPINode, 0, 8, 0, false, "api-type-2")
		testInstanceTypeDataSourceBase(t, c, api.RonDBAPINode, 0, 9, 0, false, "api-type-1")
	}
}

func TestInstanceTypeDataSourceRead_error(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/nodes/supported-types",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload": {
						
					}
				}`,
			},
		},
		Resource:             dataSourceInstanceType(),
		OperationContextFunc: dataSourceInstanceType().ReadContext,
		State: map[string]interface{}{
			"node_type":      api.HeadNode.String(),
			"cloud_provider": api.AWS.String(),
		},
		ExpectError: "no instance types available for " + api.HeadNode.String(),
	}
	r.Apply(t, context.TODO())
}

func testInstanceTypeDataSource(t *testing.T, cloud api.CloudProvider, nodeType api.NodeType, expectedId string) {
	testInstanceTypeDataSourceBase(t, cloud, nodeType, 0, 0, 0, false, expectedId)
}

func testInstanceTypeDataSourceBase(t *testing.T, cloud api.CloudProvider, nodeType api.NodeType, minMemory float64, minCPU int, minGPU int, withNvme bool, expectedId string) {
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
								},
								{
									"id": "worker-type-3",
									"memory": 50,
									"cpus": 20,
									"gpus": 1,
									"withNvme": true
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
		Resource:             dataSourceInstanceType(),
		OperationContextFunc: dataSourceInstanceType().ReadContext,
		State: map[string]interface{}{
			"node_type":      nodeType.String(),
			"cloud_provider": cloud.String(),
			"min_memory_gb":  minMemory,
			"min_cpus":       minCPU,
			"min_gpus":       minGPU,
			"with_nvme":      withNvme,
		},
		ExpectId: expectedId,
	}
	r.Apply(t, context.TODO())
}

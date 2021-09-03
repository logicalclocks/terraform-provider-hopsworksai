package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func TestAccClusterFromBackup_AWS_basic(t *testing.T) {
	testAccClusterFromBackup_basic(t, api.AWS)
}

func TestAccClusterFromBackup_AZURE_basic(t *testing.T) {
	testAccClusterFromBackup_basic(t, api.AZURE)
}

func testAccClusterFromBackup_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	clusterResourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	backupResourceName := fmt.Sprintf("hopsworksai_backup.%s", rName)
	backupName := fmt.Sprintf("%s-backup", default_CLUSTER_NAME_PREFIX)
	clusterFromBackupResourceName := fmt.Sprintf("hopsworksai_cluster_from_backup.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccBackupCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterFromBackupConfig_basic(cloud, rName, suffix, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(clusterResourceName, "url"),
					resource.TestCheckResourceAttr(clusterResourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "update_state", "none"),
				),
			},
			{
				Config: testAccClusterFromBackupConfig_basic(cloud, rName, suffix, `update_state = "stop"`, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(clusterResourceName, "url"),
					resource.TestCheckResourceAttr(clusterResourceName, "state", api.Stopped.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "activation_state", api.Startable.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "update_state", "stop"),
				),
			},
			{
				Config: testAccClusterFromBackupConfig_basic(cloud, rName, suffix, `update_state = "stop"`, fmt.Sprintf(`
				resource "hopsworksai_backup" "%s"{
					cluster_id = %s.id
					backup_name = "%s"
				}
				`,
					rName,
					clusterResourceName,
					backupName,
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(backupResourceName, "state", api.BackupSucceed.String()),
					resource.TestCheckResourceAttr(backupResourceName, "backup_name", backupName),
					resource.TestCheckResourceAttr(backupResourceName, "cloud_provider", cloud.String()),
					testAccClusterFromBackup_setResourceID_asEnvVar(clusterResourceName, "TF_VAR_acc_test_cluster_id_"+cloud.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
				variable "acc_test_cluster_id_%s"{

				}

				resource "hopsworksai_backup" "%s"{
					cluster_id = var.acc_test_cluster_id_%s
					backup_name = "%s"
				}
				`,
					cloud.String(),
					rName,
					cloud.String(),
					backupName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(backupResourceName, "state", api.BackupSucceed.String()),
					resource.TestCheckResourceAttr(backupResourceName, "backup_name", backupName),
					resource.TestCheckResourceAttr(backupResourceName, "cloud_provider", cloud.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
				variable "acc_test_cluster_id_%s"{
					
				}

				resource "hopsworksai_backup" "%s"{
					cluster_id = var.acc_test_cluster_id_%s
					backup_name = "%s"
				}

				resource "hopsworksai_cluster_from_backup" "%s"{
					source_backup_id = %s.id
				}
				`,
					cloud.String(),
					rName,
					cloud.String(),
					backupName,
					rName,
					backupResourceName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(backupResourceName, "state", api.BackupSucceed.String()),
					resource.TestCheckResourceAttr(backupResourceName, "backup_name", backupName),
					resource.TestCheckResourceAttr(backupResourceName, "cloud_provider", cloud.String()),
					resource.TestCheckResourceAttrSet(clusterFromBackupResourceName, "url"),
					resource.TestCheckResourceAttr(clusterFromBackupResourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(clusterFromBackupResourceName, "activation_state", api.Stoppable.String())),
			},
		},
	})
}

func testAccClusterFromBackup_setResourceID_asEnvVar(resourceName string, resourceIdVar string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		os.Setenv(resourceIdVar, rs.Primary.ID)
		return nil
	}
}

func testAccClusterFromBackupConfig_basic(cloud api.CloudProvider, rName string, suffix string, extraConfig string, backupConfig string) string {
	return testAccBackupConfig(cloud, rName, suffix, extraConfig, backupConfig, 10, true)
}

// unit tests
func TestClusterFromBackupCreate(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/restore/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
		},
		ExpectId: "cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_updateState_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
			"update_state":     "start",
		},
		ExpectError: "you cannot update cluster state during cluster restore from backup, however, you can update it later after restoration is complete",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_openPorts_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
			"open_ports": []interface{}{
				map[string]interface{}{
					"ssh":                  true,
					"kafka":                true,
					"feature_store":        true,
					"online_feature_store": true,
				},
			},
		},
		ExpectError: "you cannot update open ports during cluster restore from backup, however, you can update it later after restoration is complete",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_workers_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
			"workers": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			},
		},
		ExpectError: "you cannot add workers during cluster restore from backup, however, you can update it later after restoration is complete",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_unknownCloud(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "wrong",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
		},
		ExpectError: "Unknown cloud provider wrong for backup backup-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_AWS_incompatibleConfig(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"network": []interface{}{
						map[string]interface{}{
							"resource_group":       "resource-group-1",
							"virtual_network_name": "virtual-network-name-1",
							"subnet_name":          "subnet-name-1",
							"security_group_name":  "security-group-name-1",
						},
					},
				},
			},
		},
		ExpectError: "incompatible cloud configuration, expected aws_attributes instead",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_AZURE_incompatibleConfig(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AZURE",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"instance_profile_arn": "profile-1",
					"network": []interface{}{
						map[string]interface{}{
							"vpc_id":            "vpc-1",
							"subnet_id":         "subnet-1",
							"security_group_id": "security-group-1",
						},
					},
				},
			},
		},
		ExpectError: "incompatible cloud configuration, expected azure_attributes instead",
	}
	r.Apply(t, context.TODO())
}

func testClusterFromBackupCreate_update(t *testing.T, cloudProvider api.CloudProvider, expectedReqBody string, state map[string]interface{}) {
	state["source_backup_id"] = "backup-id-1"
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: fmt.Sprintf(`{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "%s",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`, cloudProvider.String()),
			},
			{
				Method:            http.MethodPost,
				Path:              "/api/clusters/restore/backup-id-1",
				ExpectRequestBody: expectedReqBody,
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"id" : "cluster-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id" : "cluster-id-1",
							"state": "running"
						}
					}
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State:                state,
		ExpectId:             "cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_update_name(t *testing.T) {
	t.Parallel()
	for _, c := range []api.CloudProvider{api.AWS, api.AZURE} {
		testClusterFromBackupCreate_update(t, c, `{
			"cluster": {
				"name": "new-cluster-name"
			}
		}`, map[string]interface{}{
			"name": "new-cluster-name",
		})
	}
}

func TestClusterFromBackupCreate_AWS_update_sshKey(t *testing.T) {
	t.Parallel()
	testClusterFromBackupCreate_update(t, api.AWS, `{
		"cluster": {
			"sshKeyName": "new-ssh-key"
		}
	}`, map[string]interface{}{
		"ssh_key": "new-ssh-key",
	})
}

func TestClusterFromBackupCreate_AZURE_update_sshKey_error(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AZURE",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
			"ssh_key":          "new-ssh-key",
		},
		ExpectError: "you cannot change the ssh key when restoring azure cluster from backup",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_AWS_update(t *testing.T) {
	t.Parallel()
	testClusterFromBackupCreate_update(t, api.AWS, `{
		"cluster": {
			"name": "new-cluster-name",
			"sshKeyName": "new-ssh-key",
			"tags": [
				{
					"name": "tag1",
					"value": "tag1-value"
				}
			],
			"autoscale": {
				"nonGpu": {
					"instanceType": "non-gpu-node",
					"diskSize": 100,
					"minWorkers": 0,
					"maxWorkers": 10,
					"standbyWorkers": 0.5,
					"downscaleWaitTime": 200
				}
			},
			"instanceProfileArn": "profile-1",
			"vpcId": "vpc-1",
			"subnetId": "subnet-1",
			"securityGroupId": "security-group-1"
		}
	}`, map[string]interface{}{
		"name":    "new-cluster-name",
		"ssh_key": "new-ssh-key",
		"tags": map[string]interface{}{
			"tag1": "tag1-value",
		},
		"autoscale": []interface{}{
			map[string]interface{}{
				"non_gpu_workers": []interface{}{
					map[string]interface{}{
						"instance_type":       "non-gpu-node",
						"disk_size":           100,
						"min_workers":         0,
						"max_workers":         10,
						"standby_workers":     0.5,
						"downscale_wait_time": 200,
					},
				},
			},
		},
		"aws_attributes": []interface{}{
			map[string]interface{}{
				"instance_profile_arn": "profile-1",
				"network": []interface{}{
					map[string]interface{}{
						"vpc_id":            "vpc-1",
						"subnet_id":         "subnet-1",
						"security_group_id": "security-group-1",
					},
				},
			},
		},
	})
}

func TestClusterFromBackupCreate_AZURE_update(t *testing.T) {
	t.Parallel()
	testClusterFromBackupCreate_update(t, api.AZURE, `{
		"cluster": {
			"name": "new-cluster-name",
			"tags": [
				{
					"name": "tag1",
					"value": "tag1-value"
				}
			],
			"autoscale": {
				"nonGpu": {
					"instanceType": "non-gpu-node",
					"diskSize": 100,
					"minWorkers": 0,
					"maxWorkers": 10,
					"standbyWorkers": 0.5,
					"downscaleWaitTime": 200
				}
			},
			"networkResourceGroup": "resource-group-1",
			"virtualNetworkName": "virtual-network-name-1",
			"subnetName": "subnet-name-1",
			"securityGroupName": "security-group-name-1"
		}
	}`, map[string]interface{}{
		"name": "new-cluster-name",
		"tags": map[string]interface{}{
			"tag1": "tag1-value",
		},
		"autoscale": []interface{}{
			map[string]interface{}{
				"non_gpu_workers": []interface{}{
					map[string]interface{}{
						"instance_type":       "non-gpu-node",
						"disk_size":           100,
						"min_workers":         0,
						"max_workers":         10,
						"standby_workers":     0.5,
						"downscale_wait_time": 200,
					},
				},
			},
		},
		"azure_attributes": []interface{}{
			map[string]interface{}{
				"network": []interface{}{
					map[string]interface{}{
						"resource_group":       "resource-group-1",
						"virtual_network_name": "virtual-network-name-1",
						"subnet_name":          "subnet-name-1",
						"security_group_name":  "security-group-name-1",
					},
				},
			},
		},
	})
}

func TestClusterFromBackupCreate_APIerror(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "backup-id-1",
							"backupName": "backup-1",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 100,
							"state": "succeed",
							"stateMessage": "backup completed"
						}
					}
				}`,
			},
			{
				Method: http.MethodPost,
				Path:   "/api/clusters/restore/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failure to create cluster from backup"
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
		},
		ExpectError: "failure to create cluster from backup",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_APIerror_getBackup(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failure to read backup"
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
		},
		ExpectError: "failure to read backup",
	}
	r.Apply(t, context.TODO())
}

func TestClusterFromBackupCreate_backup_notfound(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 404,
					"message": "no backup"
				}`,
			},
		},
		Resource:             clusterFromBackupResource(),
		OperationContextFunc: clusterFromBackupResource().CreateContext,
		State: map[string]interface{}{
			"source_backup_id": "backup-id-1",
		},
		ExpectError: "backup not found",
	}
	r.Apply(t, context.TODO())
}

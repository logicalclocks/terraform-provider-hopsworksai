package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func init() {
	resource.AddTestSweepers("hopsworksai_backup", &resource.Sweeper{
		Name: "hopsworksai_backup",
		F: func(region string) error {
			client := hopsworksClient()

			ctx := context.Background()
			backups, err := api.GetBackups(ctx, client, "")
			if err != nil {
				return fmt.Errorf("Error getting backups %s", err)
			}

			for _, backup := range backups {
				if strings.HasPrefix(backup.Name, default_CLUSTER_NAME_PREFIX) {
					if err := api.DeleteBackup(ctx, client, backup.Id); err != nil {
						tflog.Info(ctx, fmt.Sprintf("error destroying %s during sweep: %s", backup.Id, err))
					}
				}
			}
			return nil
		},
	})
}

func TestAccBackup_AWS_basic(t *testing.T) {
	testAccBackup_basic(t, api.AWS)
}

func TestAccBackup_AZURE_basic(t *testing.T) {
	testAccBackup_basic(t, api.AZURE)
}

func testAccBackup_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	clusterResourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	backupResourceName := fmt.Sprintf("hopsworksai_backup.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccBackupCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig_basic(cloud, rName, suffix, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(clusterResourceName, "url"),
					resource.TestCheckResourceAttr(clusterResourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "update_state", "none"),
				),
			},
			{
				Config: testAccBackupConfig_basic(cloud, rName, suffix, `update_state = "stop"`, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(clusterResourceName, "url"),
					resource.TestCheckResourceAttr(clusterResourceName, "state", api.Stopped.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "activation_state", api.Startable.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "update_state", "stop"),
				),
			},
			{
				Config: testAccBackupConfig_basic(cloud, rName, suffix, `update_state = "stop"`, fmt.Sprintf(`
				resource "hopsworksai_backup" "%s"{
					cluster_id = %s.id
					backup_name = "%s-backup"
				}
				`,
					rName,
					clusterResourceName,
					default_CLUSTER_NAME_PREFIX,
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(backupResourceName, "state", api.BackupSucceed.String()),
					resource.TestCheckResourceAttr(backupResourceName, "backup_name", fmt.Sprintf("%s-backup", default_CLUSTER_NAME_PREFIX)),
					resource.TestCheckResourceAttr(backupResourceName, "cloud_provider", cloud.String()),
				),
			},
			{
				ResourceName:      backupResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccBackupCheckDestroy() func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client := hopsworksClient()
		for _, rs := range s.RootModule().Resources {
			if rs.Type == "hopsworksai_backup" {
				backup, err := api.GetBackup(context.Background(), client, rs.Primary.ID)
				if err != nil {
					return err
				}

				if backup != nil {
					return fmt.Errorf("found unterminated backup %s", rs.Primary.ID)
				}
			}
			if rs.Type == "hopsworksai_cluster" || rs.Type == "hopsworksai_cluster_from_backup" {
				cluster, err := api.GetCluster(context.Background(), client, rs.Primary.ID)
				if err != nil {
					return err
				}

				if cluster != nil {
					return fmt.Errorf("found unterminated cluster %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}

func testAccBackupConfig_basic(cloud api.CloudProvider, rName string, suffix string, extraConfig string, backupConfig string) string {
	return testAccBackupConfig(cloud, rName, suffix, extraConfig, backupConfig, 7, false, "TestAccBackup_basic")
}

func testAccBackupConfig(cloud api.CloudProvider, rName string, suffix string, extraConfig string, backupConfig string, bucketIndex int, setNetwork bool, test string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"	
		backup_retention_period = 14
		head {
			instance_type = "%s"
		}
		
		%s
		
		%s 

		tags = {
		  "%s" = "%s"
		  "Test" = "%s"
		}
	  }

	  %s
	`,
		rName,
		default_CLUSTER_NAME_PREFIX,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testHeadInstanceType(cloud),
		testAccClusterCloudConfigAttributes(cloud, bucketIndex, setNetwork),
		extraConfig,
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
		test,
		backupConfig,
	)
}

// unit tests
func TestBackupCreate(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backupId" : "new-backup-id-1"
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
							"id": "cluster-id-1"
						}
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "new-backup-id-1",
							"backupName": "backup-1",
							"state": "succeed"
						}
					}
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().CreateContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		ExpectId: "new-backup-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestBackupCreate_APIerror(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failed to create backup"
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().CreateContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		ExpectError: "failed to create backup",
	}
	r.Apply(t, context.TODO())
}

func TestBackupRead(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "new-backup-id-1",
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
		Resource:             backupResource(),
		OperationContextFunc: backupResource().ReadContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		Id: "new-backup-id-1",
		ExpectState: map[string]interface{}{
			"backup_id":      "new-backup-id-1",
			"cluster_id":     "cluster-id-1",
			"backup_name":    "backup-1",
			"cloud_provider": api.AWS.String(),
			"creation_date":  time.Unix(100, 0).Format(time.RFC3339),
			"state":          api.BackupSucceed.String(),
			"state_message":  "backup completed",
		},
	}
	r.Apply(t, context.TODO())
}

func TestBackupRead_APIerror(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failed to read backup"
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().ReadContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		Id:          "new-backup-id-1",
		ExpectError: "failed to obtain backup state: failed to read backup",
	}
	r.Apply(t, context.TODO())
}

func TestBackupRead_notfound(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 404,
					"message": "no backup"
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().ReadContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		Id:          "new-backup-id-1",
		ExpectError: "backup not found for backup_id new-backup-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestBackupDelete(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodDelete,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 404,
					"message": "no backup"
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().DeleteContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		Id: "new-backup-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestBackupDelete_APIerror(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodDelete,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failed to delete backup"
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().DeleteContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		Id:          "new-backup-id-1",
		ExpectError: "failed to delete backup, error: failed to delete backup",
	}
	r.Apply(t, context.TODO())
}

func TestBackupCreate_pipelineInProgress(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backupId" : "new-backup-id-1"
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
							"id": "cluster-id-1",
							"backupPipelineInProgress": true
						}
					}
				}`,
				RunOnlyOnce: true,
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
							"id": "cluster-id-1"
						}
					}
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "new-backup-id-1",
							"backupName": "backup-1",
							"state": "succeed"
						}
					}
				}`,
				RunOnlyOnce: true,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backup": {
							"backupId" : "new-backup-id-1",
							"backupName": "backup-1",
							"state": "succeed"
						}
					}
				}`,
				RunOnlyOnce: true,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().CreateContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		ExpectId: "new-backup-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestBackupCreate_noCluster(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backupId" : "new-backup-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 404
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().CreateContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		ExpectError: "cluster not found for cluster id cluster-id-1",
	}
	r.Apply(t, context.TODO())
}

func TestBackupCreate_getCluster_error(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backupId" : "new-backup-id-1"
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "failed to get cluster info"
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().CreateContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		ExpectError: "failed to get cluster info",
	}
	r.Apply(t, context.TODO())
}

func TestBackupCreate_backup_nofound(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodPost,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"backupId" : "new-backup-id-1"
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
							"id": "cluster-id-1"
						}
					}
				}`,
			},
			{
				Method: http.MethodGet,
				Path:   "/api/backups/new-backup-id-1",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 404
				}`,
			},
		},
		Resource:             backupResource(),
		OperationContextFunc: backupResource().CreateContext,
		State: map[string]interface{}{
			"cluster_id":  "cluster-id-1",
			"backup_name": "backup-1",
		},
		ExpectError: "backup not found for backup id new-backup-id-1",
	}
	r.Apply(t, context.TODO())
}

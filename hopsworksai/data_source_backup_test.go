package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func TestAccBackupDataSource_AWS_basic(t *testing.T) {
	testAccBackupDataSource_basic(t, api.AWS)
}

func TestAccBackupDataSource_AZURE_basic(t *testing.T) {
	testAccBackupDataSource_basic(t, api.AZURE)
}

func testAccBackupDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	clusterResourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	backupResourceName := fmt.Sprintf("hopsworksai_backup.%s", rName)
	backupDataSourceName := fmt.Sprintf("data.hopsworksai_backup.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccBackupCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupDataSourceConfig_basic(cloud, rName, suffix, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(clusterResourceName, "url"),
					resource.TestCheckResourceAttr(clusterResourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "update_state", "none"),
				),
			},
			{
				Config: testAccBackupDataSourceConfig_basic(cloud, rName, suffix, "", fmt.Sprintf(`
				resource "hopsworksai_backup" "%s"{
					cluster_id = %s.id
					backup_name = "%s-backup"
				}

				data "hopsworksai_backup" "%s"{
					backup_id = %s.id
				}
				`,
					rName,
					clusterResourceName,
					default_CLUSTER_NAME_PREFIX,
					rName,
					backupResourceName,
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(backupResourceName, "state", api.BackupSucceed.String()),
					resource.TestCheckResourceAttr(backupResourceName, "backup_name", fmt.Sprintf("%s-backup", default_CLUSTER_NAME_PREFIX)),
					resource.TestCheckResourceAttr(backupResourceName, "cloud_provider", cloud.String()),
					testAccResourceDataSourceCheckAllAttributes(backupResourceName, backupDataSourceName),
				),
			},
		},
	})
}

func testAccBackupDataSourceConfig_basic(cloud api.CloudProvider, rName string, suffix string, extraConfig string, backupConfig string) string {
	return testAccBackupConfig(cloud, rName, suffix, extraConfig, backupConfig, 8, false, "TestAccBackupDataSource_basic")
}

// unit tests
func TestBackupDataSourceRead(t *testing.T) {
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
		Resource:             dataSourceBackup(),
		OperationContextFunc: dataSourceBackup().ReadContext,
		State: map[string]interface{}{
			"backup_id": "new-backup-id-1",
		},
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

func TestBackupDataSourceRead_APIerror(t *testing.T) {
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
					"message": "failed to read"
				}`,
			},
		},
		Resource:             dataSourceBackup(),
		OperationContextFunc: dataSourceBackup().ReadContext,
		State: map[string]interface{}{
			"backup_id": "new-backup-id-1",
		},
		ExpectError: "failed to read",
	}
	r.Apply(t, context.TODO())
}

func TestBackupDataSourceRead_notfound(t *testing.T) {
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
		Resource:             dataSourceBackup(),
		OperationContextFunc: dataSourceBackup().ReadContext,
		State: map[string]interface{}{
			"backup_id": "new-backup-id-1",
		},
		ExpectError: "backup not found for backup_id new-backup-id-1",
	}
	r.Apply(t, context.TODO())
}

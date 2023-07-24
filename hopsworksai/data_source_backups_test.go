package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func TestAccBackupsDataSource_AWS_basic(t *testing.T) {
	testAccBackupsDataSource_basic(t, api.AWS)
}

func TestAccBackupsDataSource_AZURE_basic(t *testing.T) {
	testAccBackupsDataSource_basic(t, api.AZURE)
}

func testAccBackupsDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	clusterResourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	backupResourceName := fmt.Sprintf("hopsworksai_backup.%s", rName)
	backupsDataSourceName := fmt.Sprintf("data.hopsworksai_backups.%s", rName)
	backupName := fmt.Sprintf("%s-list-backups", default_CLUSTER_NAME_PREFIX)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccBackupCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupsDataSourceConfig_basic(cloud, rName, suffix, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(clusterResourceName, "url"),
					resource.TestCheckResourceAttr(clusterResourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(clusterResourceName, "update_state", "none"),
				),
			},
			{
				Config: testAccBackupsDataSourceConfig_basic(cloud, rName, suffix, "", fmt.Sprintf(`
				resource "hopsworksai_backup" "%s"{
					cluster_id = %s.id
					backup_name = "%s"
				}

				data "hopsworksai_backups" "%s"{
					cluster_id = %s.id
					depends_on = [
						%s
					]
				}
				`,
					rName,
					clusterResourceName,
					backupName,
					rName,
					clusterResourceName,
					backupResourceName,
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(backupResourceName, "state", api.BackupSucceed.String()),
					resource.TestCheckResourceAttr(backupResourceName, "backup_name", backupName),
					resource.TestCheckResourceAttr(backupResourceName, "cloud_provider", cloud.String()),
					resource.TestCheckResourceAttr(backupsDataSourceName, "backups.#", "1"),
					resource.TestCheckResourceAttrPair(backupResourceName, "state", backupsDataSourceName, "backups.0.state"),
					resource.TestCheckResourceAttrPair(backupResourceName, "backup_name", backupsDataSourceName, "backups.0.backup_name"),
					resource.TestCheckResourceAttrPair(backupResourceName, "cloud_provider", backupsDataSourceName, "backups.0.cloud_provider"),
					resource.TestCheckResourceAttrPair(backupResourceName, "backup_id", backupsDataSourceName, "backups.0.backup_id"),
				),
			},
		},
	})
}

func testAccBackupsDataSourceConfig_basic(cloud api.CloudProvider, rName string, suffix string, extraConfig string, backupConfig string) string {
	return testAccBackupConfig(cloud, rName, suffix, extraConfig, backupConfig, 9, false, "TestAccBackupsDataSource_basic")
}

// unit tests

func TestBackupsDataSourceRead(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload": {
						"backups": [
							{
								"backupId": "backup-id-1",
								"backupName": "backup-name",
								"clusterId": "cluster-id-1",
								"cloudProvider": "AWS",
								"createdOn": 100,
								"state": "succeed",
								"stateMessage": "message"
							},
							{
								"backupId": "backup-id-2",
								"backupName": "backup-name-2",
								"clusterId": "cluster-id-1",
								"cloudProvider": "AWS",
								"createdOn": 1000,
								"state": "failed",
								"stateMessage": "failure message"
							}
						]
					}
				}`,
			},
		},
		Resource:             dataSourceBackups(),
		OperationContextFunc: dataSourceBackups().ReadContext,
		State: map[string]interface{}{
			"cluster_id": "cluster-id-1",
		},
		ExpectState: map[string]interface{}{
			"backups": []interface{}{
				map[string]interface{}{
					"backup_id":      "backup-id-1",
					"backup_name":    "backup-name",
					"cluster_id":     "cluster-id-1",
					"cloud_provider": api.AWS.String(),
					"creation_date":  time.Unix(100, 0).Format(time.RFC3339),
					"state":          api.BackupSucceed.String(),
					"state_message":  "message",
				},
				map[string]interface{}{
					"backup_id":      "backup-id-2",
					"backup_name":    "backup-name-2",
					"cluster_id":     "cluster-id-1",
					"cloud_provider": api.AWS.String(),
					"creation_date":  time.Unix(1000, 0).Format(time.RFC3339),
					"state":          api.BackupFailed.String(),
					"state_message":  "failure message",
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestBackupsDataSourceRead_APIerror(t *testing.T) {
	t.Parallel()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/backups",
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "failure to get backups"
				}`,
			},
		},
		Resource:             dataSourceBackups(),
		OperationContextFunc: dataSourceBackups().ReadContext,
		ExpectError:          "failure to get backups",
	}
	r.Apply(t, context.TODO())
}

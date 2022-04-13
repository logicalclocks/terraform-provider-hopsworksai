package hopsworksai

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func backupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster_id": {
			Description: "The id of the cluster for which you want to create a backup.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"backup_name": {
			Description: "The name to attach to this backup.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"backup_id": {
			Description: "The backup id.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cloud_provider": {
			Description: "The backup cloud provider.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"state": {
			Description: "The backup state.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"state_message": {
			Description: "The backup state message.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"creation_date": {
			Description: "The creation date of the backup. The date is represented in RFC3339 format.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func backupResource() *schema.Resource {
	return &schema.Resource{
		Description:   "Use this resource to create a backup for your Hopsworks.ai clusters. The cluster has to be stopped to create a backup.",
		Schema:        backupSchema(),
		CreateContext: resourceBackupCreate,
		ReadContext:   resourceBackupRead,
		DeleteContext: resourceBackupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
	}
}

func resourceBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)

	clusterId := d.Get("cluster_id").(string)
	backupName := d.Get("backup_name").(string)

	backupId, err := api.NewBackup(ctx, client, clusterId, backupName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(backupId)
	if err := resourceBackupWaitForCompletion(ctx, client, d.Timeout(schema.TimeoutCreate), backupId, clusterId); err != nil {
		return diag.FromErr(err)
	}
	return resourceBackupRead(ctx, d, meta)
}

func resourceBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)
	id := d.Id()

	backup, err := api.GetBackup(ctx, client, id)
	if err != nil {
		return diag.Errorf("failed to obtain backup state: %s", err)
	}

	if backup == nil {
		return diag.Errorf("backup not found for backup_id %s", id)
	}
	if err := populateBackupStateForResource(backup, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func populateBackupStateForResource(backup *api.Backup, d *schema.ResourceData) error {
	d.SetId(backup.Id)
	for k, v := range structure.FlattenBackup(backup) {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func resourceBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)
	id := d.Id()

	if err := api.DeleteBackup(ctx, client, id); err != nil {
		return diag.Errorf("failed to delete backup, error: %s", err)
	}

	if err := resourceBackupWaitForDeleting(ctx, client, d.Timeout(schema.TimeoutDelete), id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBackupWaitForCompletion(ctx context.Context, client *api.HopsworksAIClient, timeout time.Duration, backupId string, clusterId string) error {
	waitUntilRunning := helpers.BackupStateChange(
		[]api.BackupState{
			api.PendingBackup,
			api.InitializingBackup,
			api.ProcessingBackup,
		},
		[]api.BackupState{
			api.BackupSucceed,
			api.BackupFailed,
		},
		timeout,
		func() (result interface{}, state string, err error) {
			backup, err := api.GetBackup(ctx, client, backupId)
			if err != nil {
				return nil, "", err
			}
			if backup == nil {
				cluster, err := api.GetCluster(ctx, client, clusterId)
				if err != nil {
					return nil, "", err
				}
				if cluster == nil {
					return nil, "", fmt.Errorf("cluster not found for cluster id %s", clusterId)
				}
				if cluster.BackupPipelineInProgress {
					return nil, api.PendingBackup.String(), nil
				}
				return nil, "", fmt.Errorf("backup not found for backup id %s", backupId)
			}
			tflog.Debug(ctx, fmt.Sprintf("polled backup state: %s", backup.State))
			return backup, backup.State.String(), nil
		},
	)

	resp, err := waitUntilRunning.WaitForStateContext(ctx)
	if err != nil {
		return err
	}

	backup := resp.(*api.Backup)
	if backup.State != api.BackupSucceed {
		return fmt.Errorf("failed while waiting for the backup to reach succeed state: %s", backup.StateMessage)
	}
	return nil
}

func resourceBackupWaitForDeleting(ctx context.Context, client *api.HopsworksAIClient, timeout time.Duration, backupId string) error {
	waitUntilDeleted := helpers.BackupStateChange(
		[]api.BackupState{
			api.PendingBackup,
			api.DeletingBackup,
		},
		[]api.BackupState{
			api.BackupFailed,
			api.BackupDeleted,
		},
		timeout,
		func() (result interface{}, state string, err error) {
			backup, err := api.GetBackup(ctx, client, backupId)
			if err != nil {
				return nil, "", err
			}
			if backup == nil {
				tflog.Debug(ctx, fmt.Sprintf("backup (id: %s) is not found", backupId))
				return api.Backup{Id: ""}, api.BackupDeleted.String(), nil
			}
			tflog.Debug(ctx, fmt.Sprintf("polled backup state: %s", backup.State))
			return backup, backup.State.String(), nil
		},
	)

	resp, err := waitUntilDeleted.WaitForStateContext(ctx)
	if err != nil {
		return err
	}

	if resp != nil && resp.(api.Backup).Id != "" {
		return fmt.Errorf("failed to delete backup, error: %s", resp.(*api.Backup).StateMessage)
	}
	return nil
}

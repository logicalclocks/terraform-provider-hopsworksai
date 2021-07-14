package hopsworksai

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func dataSourceBackup() *schema.Resource {
	baseSchema := helpers.GetDataSourceSchemaFromResourceSchema(backupSchema())
	baseSchema["backup_id"].Required = true
	baseSchema["backup_id"].Computed = false
	return &schema.Resource{
		Description: "Use this data source to retrieve backup information using its id.",
		Schema:      baseSchema,
		ReadContext: dataSourceBackupRead,
	}
}

func dataSourceBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)
	backupId := d.Get("backup_id").(string)
	backup, err := api.GetBackup(ctx, client, backupId)
	if err != nil {
		return diag.FromErr(err)
	}

	if backup == nil {
		return diag.Errorf("backup not found for backup_id %s", backupId)
	}

	d.SetId(backupId)
	for k, v := range structure.FlattenBackup(backup) {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

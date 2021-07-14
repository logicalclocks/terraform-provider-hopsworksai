package hopsworksai

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func dataSourceBackups() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve all your backups.",
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "The id of the cluster to retrieve its backups. If not set, all the backups are retrieved.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"backups": {
				Description: "The list of backups sorted based on creation date with latest created backup first.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: helpers.GetDataSourceSchemaFromResourceSchema(backupSchema()),
				},
			},
		},
		ReadContext: dataSourceBackupsRead,
	}
}

func dataSourceBackupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)
	clusterId := d.Get("cluster_id").(string)
	backupsArr, err := api.GetBackups(ctx, client, clusterId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	backups := structure.FlattenBackups(backupsArr)
	if err := d.Set("backups", backups); err != nil {
		return diag.Errorf("data passed %s, err: %s", backups, err)
	}

	return nil
}

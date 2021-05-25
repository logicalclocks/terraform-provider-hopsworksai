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

func dataSourceCluster() *schema.Resource {
	clusterDataSchema := helpers.GetDataSourceSchemaFromResourceSchema(clusterSchema())
	clusterDataSchema["cluster_id"].Computed = false
	clusterDataSchema["cluster_id"].Required = true

	return &schema.Resource{
		Description: "Use this data source to get information about a cluster on Hopsworks.ai.",
		Schema:      clusterDataSchema,
		ReadContext: dataSourceClusterRead,
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.HopsworksAIClient)

	clusterId := d.Get("cluster_id").(string)
	cluster, err := api.GetCluster(ctx, client, clusterId)
	if err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	for k, v := range structure.FlattenCluster(cluster) {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

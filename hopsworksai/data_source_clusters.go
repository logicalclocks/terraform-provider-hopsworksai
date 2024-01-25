package hopsworksai

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func dataSourceClusters() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about all your clusters in Hopsworks.ai",
		Schema: map[string]*schema.Schema{
			"clusters": {
				Description: "The list of clusters in the user's account.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: helpers.GetDataSourceSchemaFromResourceSchema(clusterSchema()),
				},
			},
			"filter": {
				Description: "Filter requested clusters based on cloud provider.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud": {
							Description:  "Filter based on cloud provider.",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{api.AWS.String(), api.AZURE.String(), api.GCP.String()}, false),
						},
					},
				},
			},
		},
		ReadContext: dataSourceClustersRead,
	}
}

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*api.HopsworksAIClient)

	var cloud api.CloudProvider = ""
	if v, ok := d.GetOk("filter.0.cloud"); ok {
		cloud = api.CloudProvider(v.(string))
	}

	clustersArray, err := api.GetClusters(ctx, client, cloud)
	if err != nil {
		return diag.FromErr(err)
	}

	clusters := structure.FlattenClusters(clustersArray)
	if err := d.Set("clusters", clusters); err != nil {
		return diag.Errorf("data passed %s, err: %s", clusters, err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

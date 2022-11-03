package hopsworksai

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func dataSourceInstanceTypes() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get all the supported instance types for head, worker, and RonDB nodes.",
		Schema: map[string]*schema.Schema{
			"node_type": {
				Description:  "The node type that you want to get its supported instance types.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(api.GetAllNodeTypes(), false),
			},
			"cloud_provider": {
				Description:  "The cloud provider where you plan to create your cluster.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{api.AWS.String(), api.AZURE.String()}, false),
			},
			"region": {
				Description: "The region/location where you plan to create your cluster.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"supported_types": {
				Description: "The list of supported instance types.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The instance type Id.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"memory": {
							Description: "The instance type memory size in gigabytes.",
							Type:        schema.TypeFloat,
							Computed:    true,
						},
						"cpus": {
							Description: "The instance type number of CPU cores.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"gpus": {
							Description: "The instance type number of GPUs.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"with_nvme": {
							Description: "The instance type is equipped of NVMe drives.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},
		},
		ReadContext: dataSourceInstanceTypesRead,
	}
}

func dataSourceInstanceTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)

	cloud := api.CloudProvider(d.Get("cloud_provider").(string))
	region := d.Get("region").(string)
	supportedTypes, err := api.GetSupportedInstanceTypes(ctx, client, cloud, region)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeType := d.Get("node_type").(string)
	instanceTypesArr := supportedTypes.GetByNodeType(api.NodeType(nodeType))

	if len(instanceTypesArr) == 0 {
		return diag.Errorf("no instance types available for %s", nodeType)
	}

	instanceTypesArr.Sort()

	d.SetId(fmt.Sprintf("%s-%s", cloud.String(), nodeType))
	if err := d.Set("supported_types", structure.FlattenSupportedInstanceTypes(instanceTypesArr)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

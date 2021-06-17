package hopsworksai

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func dataSourceInstanceType() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the smallest instance type for head, worker, and RonDB nodes.",
		Schema: map[string]*schema.Schema{
			"node_type": {
				Description:  fmt.Sprintf("The node type that you want to get its smallest instance type. It has to be one of these types (%s).", strings.Join(api.GetAllNodeTypes(), ", ")),
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
			"min_memory_gb": {
				Description:  "Filter based on the minimum memory in gigabytes.",
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.FloatAtLeast(0),
			},
			"min_cpus": {
				Description:  "Filter based on the minimum number of CPU cores.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"min_gpus": {
				Description:  "Filter based on the minimum number of GPUs.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
		ReadContext: dataSourceInstanceTypeRead,
	}
}

func dataSourceInstanceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)

	cloud := api.CloudProvider(d.Get("cloud_provider").(string))
	supportedTypes, err := api.GetSupportedInstanceTypes(ctx, client, cloud)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeType := d.Get("node_type").(string)
	instanceTypesArr := supportedTypes.GetByNodeType(api.NodeType(nodeType))

	if instanceTypesArr == nil || len(instanceTypesArr) == 0 {
		return diag.Errorf("no instance types available for %s", nodeType)
	}

	instanceTypesArr.Sort()

	minMemory := d.Get("min_memory_gb").(float64)
	minCPUs := d.Get("min_cpus").(int)
	minGPUs := d.Get("min_gpus").(int)

	var chosenType *api.SupportedInstanceType = nil
	for _, v := range instanceTypesArr {
		if minMemory > 0 && v.Memory < minMemory {
			continue
		}
		if minCPUs > 0 && v.CPUs < minCPUs {
			continue
		}
		if minGPUs > 0 && v.GPUs < minGPUs {
			continue
		}

		chosenType = &v
		break
	}

	if chosenType != nil {
		d.SetId(chosenType.Id)
	} else {
		d.SetId(instanceTypesArr[0].Id)
	}
	return nil
}

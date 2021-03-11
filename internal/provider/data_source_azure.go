package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func instanceBaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"instance_name": {
			Description: "Instance name.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"state": {
			Description: "Instance state.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"initialization_stage": {
			Description: "Initialization stage.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"activation_stage": {
			Description: "Activation stage.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func dataSourceHopsworksAI() *schema.Resource {
	instanceSchema := instanceBaseSchema()
	instanceSchema["instance_id"] = &schema.Schema{
		Description: "Instance ID.",
		Type:        schema.TypeString,
		Computed:    true,
	}
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Data source in the Terraform provider.",

		ReadContext: dataSourceHopsworksAIRead,

		Schema: map[string]*schema.Schema{
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: instanceSchema,
				},
			},
		},
	}
}

func dataSourceClusterHopsworksAI() *schema.Resource {
	instanceSchema := instanceBaseSchema()
	instanceSchema["cluster_id"] = &schema.Schema{
		Description: "Cluster ID.",
		Type:        schema.TypeString,
		Required:    true,
	}
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Data source in the Terraform provider.",

		ReadContext: dataSourceInstanceHopsworksAIRead,

		Schema: instanceSchema,
	}
}

func dataSourceInstanceHopsworksAIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)
	instanceID := d.Get("cluster_id").(string)
	instance, err := client.GetInstance(instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	v := instance.Payload.InstanceData

	d.Set("cluster_name", v.InstanceName)
	d.Set("state", v.State)
	d.Set("initialization_stage", v.InitializationStage)
	d.Set("activation_stage", v.ActivationStage)

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func dataSourceHopsworksAIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)
	instances, err := client.GetInstances()
	if err != nil {
		return diag.FromErr(err)
	}

	i := make([]map[string]interface{}, 0)

	for _, v := range instances.Payload.Instances {
		instance := make(map[string]interface{})

		instance["cluster_id"] = v.InstanceID
		instance["cluster_name"] = v.InstanceName
		instance["state"] = v.State
		instance["initialization_stage"] = v.InitializationStage
		instance["activation_stage"] = v.ActivationStage

		i = append(i, instance)
	}

	if err := d.Set("instances", i); err != nil {
		return diag.Errorf("data passed %s, err: %s", instances.Payload.Instances, err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

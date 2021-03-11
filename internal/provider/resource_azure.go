package provider

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func baseClusterResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster_id": {
			Description: "Unique identifier of the hopsworks cluster, automatically generated",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cluster_name": {
			Description: "Name of the hopsworks cluster, must be unique",
			Type:        schema.TypeString,
			Required:    true,
		},
		"version": {
			// This description is used by the documentation generator and the language server.
			Description: "Hopsworks version, default 2.1.0",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "2.1.0",
		},
		"head_node_instance_type": {
			// This description is used by the documentation generator and the language server.
			Description: "Instance type of the head node, default Standard_D8_v3",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Standard_D8_v3",
		},
		"head_node_local_storage": {
			// This description is used by the documentation generator and the language server.
			Description: "Disk size of the head node in units of GB, default 512 GB",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     512,
		},
		"worker_node_instance_type": {
			// This description is used by the documentation generator and the language server.
			Description: "Instance type of worker nodes, default Standard_D8_v3",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Standard_D8_v3",
		},
		"worker_node_count": {
			// This description is used by the documentation generator and the language server.
			Description: "Number of worker nodes",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
		},
		"worker_node_local_storage": {
			// This description is used by the documentation generator and the language server.
			Description: "Disk size of worker node(s) in units of GB, default 512 GB",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     512,
		},
		"ssh_key": {
			// This description is used by the documentation generator and the language server.
			Description: "SSH key resource for the instances",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func azureClusterResource() *schema.Resource {
	baseSchema := baseClusterResourceSchema()
	baseSchema["resource_group"] = &schema.Schema{
		Description: "Resource group the Hopsworks cluster will reside in",
		Type:        schema.TypeString,
		Required:    true,
	}
	baseSchema["location"] = &schema.Schema{
		Description: "Azure location the Hopsworks cluster will reside in",
		Type:        schema.TypeString,
		Required:    true,
	}
	baseSchema["storage_account"] = &schema.Schema{
		Description: "Azure storage account the Hopsworks cluster will use to store data in",
		Type:        schema.TypeString,
		Required:    true,
	}
	baseSchema["storage_container_name"] = &schema.Schema{
		Description: "Azure storage container the Hopsworks cluster will use to store data in, automatically generated if not set.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
	}
	baseSchema["managed_identity"] = &schema.Schema{
		Description: "Azure managed identity the Hopsworks instances will be started with",
		Type:        schema.TypeString,
		Required:    true,
	}

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Sample resource in the Terraform provider scaffolding.",

		CreateContext: resourceAzureClusterCreate,
		ReadContext:   resourceAzureClusterRead,
		UpdateContext: resourceAzureClusterUpdate,
		DeleteContext: resourceAzureClusterDelete,

		Schema: baseSchema,
	}
}

func resourceAzureClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	var diags diag.Diagnostics

	containerName := d.Get("storage_container_name").(string)
	if containerName == "" {
		suffix := time.Now().UnixNano() / 1e6
		containerName = fmt.Sprintf("hopsworksai-%d", suffix)
	}

	req := InstanceRequest{
		InstanceName:    d.Get("cluster_name").(string),
		InstanceProfile: d.Get("managed_identity").(string),
		ResourceGroup:   d.Get("resource_group").(string),
		Region:          d.Get("location").(string),
		StorageName:     d.Get("storage_account").(string),
		BucketName:      containerName,
		KeyName:         d.Get("instance_ssh_key").(string),
		Version:         d.Get("version").(string),
		InstanceConfiguration: map[string]InstanceConfiguration{
			"Master": {
				InstanceType: d.Get("head_node_instance_type").(string),
				VolumeSize:   d.Get("head_node_local_storage").(int),
			},
			"Worker": {
				InstanceType: d.Get("worker_node_instance_type").(string),
				VolumeSize:   d.Get("worker_node_local_storage").(int),
			},
		},
		InstanceTags: []InstanceTag{
			{
				Name:  "",
				Value: "",
			},
		},
		NBNodes:                   d.Get("worker_node_count").(int),
		ManagedUsers:              true,
		IssueLetsEncrypt:          true,
		DeleteBlocksRetentionDays: 0,
		Ports: InstancePorts{
			FeatureStore:       "",
			Kafka:              "",
			SSH:                "",
			OnlineFeatureStore: "",
		},
	}

	err := client.CreateInstance(&req, "azure")

	if err != nil {
		return diag.Errorf("failed to create instance resource, error: %s", err)
	}

	instance, err := getInstance(client, d.Get("cluster_name").(string), 0)
	if err != nil {
		return diag.Errorf("failed to get instance, error: %s", err)
	}

	instance, err = waitUntilRunning(client, instance.InstanceID)
	if err != nil {
		return diag.FromErr(err)

	}

	d.SetId(instance.InstanceID)
	d.Set("cluster_id", instance.InstanceID)

	return diags
}

func waitUntilRunning(client *apiClient, instanceID string) (*Instance, error) {
	for {
		instance, err := client.GetInstance(instanceID)
		if err != nil {
			return nil, err
		}

		log.Printf("[INFO] polled instance state: %s, stage: %s", instance.Payload.InstanceData.State, instance.Payload.InstanceData.InitializationStage)

		switch instance.Payload.InstanceData.State {
		case "error":
			return nil, fmt.Errorf("instance is in error state")
		case "running":
			return &instance.Payload.InstanceData, nil
		}

		time.Sleep(time.Second * 10)
	}
}

func waitUntilDeleted(client *apiClient, instanceID string) error {
	instance, err := client.GetInstance(instanceID)
	if err != nil {
		return err
	}

	for {
		instances, err := client.GetInstances()
		if err != nil {
			return err
		}
		_, found := getInstanceByName(instances.Payload.Instances, instance.Payload.InstanceData.InstanceName)
		if !found {
			return nil
		}

		time.Sleep(time.Second * 10)
	}
}

func getInstance(client *apiClient, instanceName string, attempts int) (*Instance, error) {
	if attempts > 4 {
		return nil, fmt.Errorf("failed to obtain instance %s, giving up after 5 attempts", instanceName)
	}
	instances, err := client.GetInstances()

	if err != nil {
		return nil, err
	}

	instance, found := getInstanceByName(instances.Payload.Instances, instanceName)
	if !found {
		time.Sleep(time.Second + time.Second*time.Duration(attempts))
		return getInstance(client, instanceName, attempts+1)
	}

	return &instance, nil
}

func resourceAzureClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)
	var diags diag.Diagnostics

	instances, err := client.GetInstances()
	if err != nil {
		return diag.Errorf("failed to obtain instance state: %s", err)
	}
	instanceName := d.Get("cluster_name").(string)
	instance, found := getInstanceByName(instances.Payload.Instances, instanceName)
	if !found {
		return diags
	}

	d.SetId(instance.InstanceID)
	return diags
}

func getInstanceByName(instances []Instance, instanceName string) (Instance, bool) {
	for _, i := range instances {
		if i.InstanceName == instanceName {
			return i, true
		}
	}

	return Instance{}, false
}

func resourceAzureClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	var diags diag.Diagnostics

	instanceID := d.Id()

	if d.HasChange("worker_node_count") {
		o, n := d.GetChange("worker_node_count")
		old, new := o.(int), n.(int)
		if new > old {
			toAdd := new - old
			ir := InstanceUpdateRequest{
				WorkerConfiguration: map[string]InstanceConfiguration{
					"Worker": {
						InstanceType: d.Get("worker_node_instance_type").(string),
						VolumeSize:   d.Get("worker_node_local_storage").(int),
					},
				},
				NBNodes: toAdd,
			}
			log.Printf("[INFO] will add %d nodes, req: %#v", toAdd, ir)

			err := client.AddNodes(&ir, instanceID)
			if err != nil {
				diag.FromErr(err)
			}
		} else if new < 1 {
			return diag.Errorf("you most provide a worker count > 0")
		}
	}

	return diags
}

func resourceAzureClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	var diags diag.Diagnostics

	instanceID := d.Id()

	err := client.DeleteInstance(instanceID)
	if err != nil {
		return diag.Errorf("failed to delete instance %s, error: %s", instanceID, err)
	}

	err = waitUntilDeleted(client, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

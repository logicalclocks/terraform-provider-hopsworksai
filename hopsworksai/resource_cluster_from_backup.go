package hopsworksai

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func clusterFromBackupResource() *schema.Resource {
	clusterResourceSchema := clusterSchema()
	baseSchema := helpers.GetDataSourceSchemaFromResourceSchema(clusterResourceSchema)
	baseSchema["source_backup_id"] = &schema.Schema{
		Description: "",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}

	// allow the user to edit the following during restoring from backup
	baseSchema["name"].Optional = true
	baseSchema["name"].ForceNew = true

	baseSchema["ssh_key"].Optional = true
	baseSchema["ssh_key"].ForceNew = true

	baseSchema["tags"].Optional = true
	baseSchema["tags"].ForceNew = true

	baseSchema["autoscale"] = clusterResourceSchema["autoscale"]

	// allow changing aws instance profile and network configuration during restore
	baseSchema["aws_attributes"].Optional = true
	baseSchema["aws_attributes"].ForceNew = true
	baseSchema["aws_attributes"].MaxItems = 1
	baseSchema["aws_attributes"].ConflictsWith = []string{"azure_attributes", "gcp_attributes"}

	clusterAWSAttributesSchema := baseSchema["aws_attributes"].Elem.(*schema.Resource).Schema
	clusterAWSAttributesSchema["instance_profile_arn"].Optional = true
	clusterAWSAttributesSchema["instance_profile_arn"].ForceNew = true
	clusterAWSAttributesSchema["head_instance_profile_arn"].Optional = true
	clusterAWSAttributesSchema["head_instance_profile_arn"].ForceNew = true
	clusterAWSAttributesSchema["network"] = awsAttributesSchema().Schema["network"]

	// allow changing azure managed identity and network configurations during restore
	baseSchema["azure_attributes"].Optional = true
	baseSchema["azure_attributes"].ForceNew = true
	baseSchema["azure_attributes"].MaxItems = 1
	baseSchema["azure_attributes"].ConflictsWith = []string{"aws_attributes", "gcp_attributes"}

	clusterAZUREAttributesSchema := baseSchema["azure_attributes"].Elem.(*schema.Resource).Schema
	clusterAZUREAttributesSchema["network"] = azureAttributesSchema().Schema["network"]

	// allow changing gcp
	baseSchema["gcp_attributes"].Optional = true
	baseSchema["gcp_attributes"].ForceNew = true
	baseSchema["gcp_attributes"].MaxItems = 1
	baseSchema["gcp_attributes"].ConflictsWith = []string{"aws_attributes", "azure_attributes"}

	clusterGCPAttributesSchema := baseSchema["gcp_attributes"].Elem.(*schema.Resource).Schema
	clusterGCPAttributesSchema["service_account_email"].Optional = true
	clusterGCPAttributesSchema["service_account_email"].ForceNew = true
	clusterGCPAttributesSchema["network"] = awsAttributesSchema().Schema["network"]

	// allow the following attributes to be updated later after creation
	baseSchema["update_state"] = clusterResourceSchema["update_state"]
	baseSchema["open_ports"] = clusterResourceSchema["open_ports"]
	baseSchema["workers"] = clusterResourceSchema["workers"]

	return &schema.Resource{
		Description:   "Use this resource to create a cluster from an existing backup.",
		Schema:        baseSchema,
		CreateContext: resourceClusterFromBackupCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Hour),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(45 * time.Minute),
		},
	}
}

func resourceClusterFromBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)

	if v, ok := d.GetOk("update_state"); ok && v != "none" {
		return diag.Errorf("you cannot update cluster state during cluster restore from backup, however, you can update it later after restoration is complete")
	}

	if _, ok := d.GetOk("open_ports"); ok {
		return diag.Errorf("you cannot update open ports during cluster restore from backup, however, you can update it later after restoration is complete")
	}

	if _, ok := d.GetOk("workers"); ok {
		return diag.Errorf("you cannot add workers during cluster restore from backup, however, you can update it later after restoration is complete")
	}

	baseRequest := api.CreateClusterFromBackup{}
	if v, ok := d.GetOk("name"); ok {
		baseRequest.Name = v.(string)
	}

	if v, ok := d.GetOk("ssh_key"); ok {
		baseRequest.SshKeyName = v.(string)
	}

	if v, ok := d.GetOk("tags"); ok {
		baseRequest.Tags = structure.ExpandTags(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("autoscale"); ok {
		baseRequest.Autoscale = structure.ExpandAutoscaleConfiguration(v.([]interface{}))
	}

	backupId := d.Get("source_backup_id").(string)
	backup, err := api.GetBackup(ctx, client, backupId)
	if err != nil {
		return diag.FromErr(err)
	}
	if backup == nil {
		return diag.Errorf("backup not found")
	}
	var restoreRequest interface{}
	switch backup.CloudProvider {
	case api.AWS:
		restoreRequest = &api.CreateAWSClusterFromBackup{
			CreateClusterFromBackup: baseRequest,
		}
	case api.AZURE:
		if baseRequest.SshKeyName != "" {
			return diag.Errorf("you cannot change the ssh key when restoring azure cluster from backup")
		}
		restoreRequest = &api.CreateAzureClusterFromBackup{
			CreateClusterFromBackup: baseRequest,
		}
	case api.GCP:
		restoreRequest = &api.CreateGCPClusterFromBackup{
			CreateClusterFromBackup: baseRequest,
		}
	default:
		return diag.Errorf("Unknown cloud provider %s for backup %s", backup.CloudProvider, backupId)
	}

	if aws, ok := d.GetOk("aws_attributes"); ok && len(aws.([]interface{})) > 0 {
		if awsRequest, okV := restoreRequest.(*api.CreateAWSClusterFromBackup); okV {
			if v, ok := d.GetOk("aws_attributes.0.instance_profile_arn"); ok {
				awsRequest.InstanceProfileArn = v.(string)
			}

			if v, ok := d.GetOk("aws_attributes.0.head_instance_profile_arn"); ok {
				awsRequest.HeadInstanceProfileArn = v.(string)
			}

			if v, ok := d.GetOk("aws_attributes.0.network.0.vpc_id"); ok {
				awsRequest.VpcId = v.(string)
			}

			if v, ok := d.GetOk("aws_attributes.0.network.0.subnet_id"); ok {
				awsRequest.SubnetId = v.(string)
			}

			if v, ok := d.GetOk("aws_attributes.0.network.0.security_group_id"); ok {
				awsRequest.SecurityGroupId = v.(string)
			}
		} else {
			return diag.Errorf("incompatible cloud configuration, expected %s_attributes instead", strings.ToLower(backup.CloudProvider.String()))
		}
	}

	if azure, ok := d.GetOk("azure_attributes"); ok && len(azure.([]interface{})) > 0 {
		if azureRequest, okV := restoreRequest.(*api.CreateAzureClusterFromBackup); okV {
			if v, ok := d.GetOk("azure_attributes.0.network.0.resource_group"); ok {
				azureRequest.NetworkResourceGroup = v.(string)
			}

			if v, ok := d.GetOk("azure_attributes.0.network.0.virtual_network_name"); ok {
				azureRequest.VirtualNetworkName = v.(string)
			}

			if v, ok := d.GetOk("azure_attributes.0.network.0.subnet_name"); ok {
				azureRequest.SubnetName = v.(string)
			}

			if v, ok := d.GetOk("azure_attributes.0.network.0.security_group_name"); ok {
				azureRequest.SecurityGroupName = v.(string)
			}
		} else {
			return diag.Errorf("incompatible cloud configuration, expected %s_attributes instead", strings.ToLower(backup.CloudProvider.String()))
		}
	}

	if gcp, ok := d.GetOk("gcp_attributes"); ok && len(gcp.([]interface{})) > 0 {
		if gcpRequest, okV := restoreRequest.(*api.CreateGCPClusterFromBackup); okV {
			if v, ok := d.GetOk("gcp_attributes.0.service_account_email"); ok {
				gcpRequest.ServiceAccountEmail = v.(string)
			}

			if v, ok := d.GetOk("gcp_attributes.0.network.0.network_name"); ok {
				gcpRequest.NetworkName = v.(string)
			}

			if v, ok := d.GetOk("gcp_attributes.0.network.0.subnetwork_name"); ok {
				gcpRequest.SubNetworkName = v.(string)
			}
		} else {
			return diag.Errorf("incompatible cloud configuration, expected %s_attributes instead", strings.ToLower(backup.CloudProvider.String()))
		}
	}

	clusterId, err := api.NewClusterFromBackup(ctx, client, backupId, restoreRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(clusterId)
	if err := resourceClusterWaitForRunning(ctx, client, d.Timeout(schema.TimeoutCreate), clusterId); err != nil {
		return diag.FromErr(err)
	}
	return resourceClusterRead(ctx, d, meta)
}

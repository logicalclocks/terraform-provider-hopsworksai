package hopsworksai

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAzureUserAssignedIdentityPermissions() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the azure user assigned identity permissions needed by Hopsworks.ai",
		Schema: map[string]*schema.Schema{
			"enable_storage": {
				Description: "Add permissions required to allow Hopsworks clusters to read and write from and to your azure storage accounts.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_backup": {
				Description: "Add permissions required to allow creating backups of your clusters.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_upgrade": {
				Description: "Add permissions required to enable upgrade to newer versions of Hopsworks.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"actions": {
				Description: "The actions permissions.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"not_actions": {
				Description: "The not actions permissions.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"data_actions": {
				Description: "The data actions permissions.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"not_data_actions": {
				Description: "The not data actions permissions.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
		},
		ReadContext: dataSourceAzureUserAssignedIdentityPermissionsRead,
	}
}

func dataSourceAzureUserAssignedIdentityPermissionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	actions := []string{}
	notActions := []string{}
	dataActions := []interface{}{}
	notDataActions := []interface{}{}

	if d.Get("enable_storage").(bool) {
		actions = append(actions, "Microsoft.Storage/storageAccounts/blobServices/containers/write",
			"Microsoft.Storage/storageAccounts/blobServices/containers/read",
			"Microsoft.Storage/storageAccounts/blobServices/read")

		dataActions = append(dataActions, "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/delete",
			"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
			"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/move/action",
			"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write")
	}

	if d.Get("enable_backup").(bool) {
		actions = append(actions, "Microsoft.Storage/storageAccounts/blobServices/write")
	}

	if d.Get("enable_upgrade").(bool) {
		actions = append(actions, "Microsoft.Compute/virtualMachines/read",
			"Microsoft.Compute/virtualMachines/write",
			"Microsoft.Compute/disks/read",
			"Microsoft.Compute/disks/write")
	}

	d.SetId(strconv.Itoa(schema.HashString(strings.Join(actions, ","))))
	if err := d.Set("actions", actions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("not_actions", notActions); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("data_actions", schema.NewSet(schema.HashString, dataActions)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("not_data_actions", schema.NewSet(schema.HashString, notDataActions)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

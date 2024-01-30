package hopsworksai

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGCPServiceAccountCustomRolePermissions() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the GCP service account custom role permissions needed by Hopsworks.ai",
		Schema: map[string]*schema.Schema{
			"enable_storage": {
				Description: "Add permissions required to allow Hopsworks clusters to read and write from and to your google storage bucket.",
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
			"enable_artifact_registry": {
				Description: "Add permissions required to enable access to the artifact registry",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"permissions": {
				Description: "The list of permissions.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		ReadContext: dataSourceGCPServiceAccountCustomRolePermissionsRead,
	}
}

func dataSourceGCPServiceAccountCustomRolePermissionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	permissions := []string{}

	if d.Get("enable_storage").(bool) {
		permissions = append(permissions, "storage.buckets.get",
			"storage.multipartUploads.abort",
			"storage.multipartUploads.create",
			"storage.multipartUploads.list",
			"storage.multipartUploads.listParts",
			"storage.objects.create",
			"storage.objects.delete",
			"storage.objects.get",
			"storage.objects.list",
			"storage.objects.update")
	}

	if d.Get("enable_backup").(bool) {
		permissions = append(permissions, "storage.buckets.update")
	}

	if d.Get("enable_artifact_registry").(bool) {
		permissions = append(permissions, "artifactregistry.repositories.create",
			"artifactregistry.repositories.get",
			"artifactregistry.repositories.uploadArtifacts",
			"artifactregistry.repositories.downloadArtifacts",
			"artifactregistry.tags.list",
			"artifactregistry.tags.delete")
	}

	d.SetId(strconv.Itoa(schema.HashString(strings.Join(permissions, ","))))
	if err := d.Set("permissions", permissions); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

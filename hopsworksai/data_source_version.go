package hopsworksai

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/structure"
)

func dataSourceVersion() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the latest supported Hopsworks version.",
		Schema: map[string]*schema.Schema{
			"cloud_provider": {
				Description:  "The cloud provider where you plan to create your cluster.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{api.AWS.String(), api.AZURE.String()}, false),
			},
			"os": {
				Description:  "Filter based on the supported os.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{api.Ubuntu.String(), api.CentOS.String()}, false),
			},
			"region": {
				Description:  "Filter based on the region.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"os"},
			},
			"default": {
				Description: "The version is the default version.",
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
			},
			"experimental": {
				Description: "The version is an experimental version.",
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
			},
			"upgradeable_from_version": {
				Description: "The version which is upgradeable to this version.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"supported_regions": {
				Description: "The list of supported operating systems per regions.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						api.Ubuntu.String(): {
							Description: "The list of regions that support Ubuntu.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        schema.TypeString,
						},
						api.CentOS.String(): {
							Description: "The list of regions that support CentOS.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        schema.TypeString,
						},
					},
				},
			},
			"release_notes_url": {
				Description: "The release notes url for this version.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		ReadContext: dataSourceVersionRead,
	}
}

func dataSourceVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.HopsworksAIClient)

	cloud := api.CloudProvider(d.Get("cloud_provider").(string))
	supportedVersions, err := api.GetSupportedVersions(ctx, client, cloud)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("upgradeable_from_version"); ok {
		upgradeableFromVersion := v.(string)
		var filteredVersions []api.SupportedVersion = make([]api.SupportedVersion, 0)
		for _, version := range supportedVersions {
			if version.UpgradableFromVersion == upgradeableFromVersion {
				filteredVersions = append(filteredVersions, version)
			}
		}
		supportedVersions = filteredVersions
	}

	var chosenVersion *api.SupportedVersion

	filterOS := api.OS(d.Get("os").(string))
	filterRegion := d.Get("region").(string)
	filterDefault, filterDefaultOk := d.GetOk("default")
	filterExperimental, filterExperimentalOk := d.GetOk("experimental")

	for i := len(supportedVersions) - 1; i >= 0; i-- {
		v := supportedVersions[i]
		switch filterOS {
		case api.Ubuntu:
			if len(v.Regions.Ubuntu) == 0 {
				continue
			}
		case api.CentOS:
			if len(v.Regions.CentOS) == 0 {
				continue
			}
		}

		if filterRegion != "" {
			switch filterOS {
			case api.Ubuntu:
				if !contains(v.Regions.Ubuntu, filterRegion) && !contains(v.Regions.Ubuntu, "ALL") {
					continue
				}
			case api.CentOS:
				if !contains(v.Regions.CentOS, filterRegion) && !contains(v.Regions.CentOS, "ALL") {
					continue
				}
			}
		}

		if filterDefaultOk {
			if v.Default != filterDefault.(bool) {
				continue
			}
		}

		if filterExperimentalOk {
			if v.Experimental != filterExperimental.(bool) {
				continue
			}
		}

		chosenVersion = &v
		break
	}

	if chosenVersion == nil {
		return diag.Errorf("no version found matching the provided filters.")
	}

	d.SetId(chosenVersion.Version)
	for k, v := range structure.FlattenVersion(chosenVersion) {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

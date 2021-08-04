package structure

import "github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"

func FlattenVersion(version *api.SupportedVersion) map[string]interface{} {
	return map[string]interface{}{
		"upgradeable_from_version": version.UpgradableFromVersion,
		"default":                  version.Default,
		"experimental":             version.Experimental,
		"supported_regions": []interface{}{
			map[string]interface{}{
				api.Ubuntu.String(): version.Regions.Ubuntu,
				api.CentOS.String(): version.Regions.CentOS,
			},
		},
		"release_notes_url": version.ReleaseNotesUrl,
	}
}

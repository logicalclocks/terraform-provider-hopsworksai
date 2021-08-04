package structure

import (
	"reflect"
	"testing"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestFlattenVersion(t *testing.T) {
	input := &api.SupportedVersion{
		Version:               "version-1",
		UpgradableFromVersion: "upgrade-version-1",
		Default:               true,
		Experimental:          true,
		Regions: api.SupportedVersionRegions{
			Ubuntu: []string{"region-1"},
			CentOS: []string{"region-2"},
		},
		ReleaseNotesUrl: "notes-url",
	}

	expected := map[string]interface{}{
		"upgradeable_from_version": "upgrade-version-1",
		"default":                  true,
		"experimental":             true,
		"supported_regions": []interface{}{
			map[string]interface{}{
				"ubuntu": []string{"region-1"},
				"centos": []string{"region-2"},
			},
		},
		"release_notes_url": "notes-url",
	}

	output := FlattenVersion(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

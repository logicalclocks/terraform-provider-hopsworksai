package structure

import (
	"reflect"
	"testing"
	"time"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestFlattenBackup(t *testing.T) {
	input := &api.Backup{
		Id:            "backup-id",
		Name:          "backup-name",
		ClusterId:     "cluster-id",
		CreatedOn:     10,
		CloudProvider: api.AWS,
		State:         api.BackupSucceed,
		StateMessage:  "message",
	}

	expected := map[string]interface{}{
		"backup_id":      "backup-id",
		"backup_name":    "backup-name",
		"cluster_id":     "cluster-id",
		"creation_date":  time.Unix(10, 0).Format(time.RFC3339),
		"cloud_provider": api.AWS,
		"state":          api.BackupSucceed,
		"state_message":  "message",
	}

	output := FlattenBackup(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenBackups(t *testing.T) {
	input := []api.Backup{
		{
			Id:            "backup-id",
			Name:          "backup-name",
			ClusterId:     "cluster-id",
			CreatedOn:     10,
			CloudProvider: api.AWS,
			State:         api.BackupSucceed,
			StateMessage:  "message",
		},
		{
			Id:            "backup-id-2",
			Name:          "backup-name-2",
			ClusterId:     "cluster-id-2",
			CreatedOn:     100,
			CloudProvider: api.AZURE,
			State:         api.BackupSucceed,
			StateMessage:  "message 2",
		},
	}

	expected := []map[string]interface{}{
		{
			"backup_id":      "backup-id",
			"backup_name":    "backup-name",
			"cluster_id":     "cluster-id",
			"creation_date":  time.Unix(10, 0).Format(time.RFC3339),
			"cloud_provider": api.AWS,
			"state":          api.BackupSucceed,
			"state_message":  "message",
		},
		{
			"backup_id":      "backup-id-2",
			"backup_name":    "backup-name-2",
			"cluster_id":     "cluster-id-2",
			"creation_date":  time.Unix(100, 0).Format(time.RFC3339),
			"cloud_provider": api.AZURE,
			"state":          api.BackupSucceed,
			"state_message":  "message 2",
		},
	}

	output := FlattenBackups(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

package structure

import (
	"time"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func FlattenBackups(backupsArray []api.Backup) []map[string]interface{} {
	backups := make([]map[string]interface{}, 0)
	for _, v := range backupsArray {
		backups = append(backups, FlattenBackup(&v))
	}
	return backups
}

func FlattenBackup(backup *api.Backup) map[string]interface{} {
	return map[string]interface{}{
		"backup_id":      backup.Id,
		"backup_name":    backup.Name,
		"cluster_id":     backup.ClusterId,
		"state":          backup.State,
		"cloud_provider": backup.CloudProvider,
		"state_message":  backup.StateMessage,
		"creation_date":  time.Unix(backup.CreatedOn, 0).Format(time.RFC3339),
	}
}

package helpers

import (
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func SuppressDiffForNonSetKeys(k, old, new string, d *schema.ResourceData) bool {
	_, ok := d.GetOk(k)
	return old != "" && new == "" && !ok
}

func convertStateArray(states interface{}) []string {
	statesArr := reflect.ValueOf(states)
	stringArr := make([]string, statesArr.Len())
	for i := 0; i < statesArr.Len(); i++ {
		stringArr[i] = statesArr.Index(i).String()
	}
	return stringArr
}

func ClusterStateChange(pending []api.ClusterState, target []api.ClusterState, timeout time.Duration, refreshFunc resource.StateRefreshFunc) *resource.StateChangeConf {
	return stateChange(convertStateArray(pending), convertStateArray(target), timeout, refreshFunc, 30*time.Second)
}

func BackupStateChange(pending []api.BackupState, target []api.BackupState, timeout time.Duration, refreshFunc resource.StateRefreshFunc) *resource.StateChangeConf {
	return stateChange(convertStateArray(pending), convertStateArray(target), timeout, refreshFunc, 30*time.Second)
}

func stateChange(pending []string, target []string, timeout time.Duration, refreshFunc resource.StateRefreshFunc, minTimeout time.Duration) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    refreshFunc,
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      minTimeout,
	}
}

package helpers

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func SuppressDiffForNonSetKeys(k, old, new string, d *schema.ResourceData) bool {
	_, ok := d.GetOk(k)
	return old != "" && new == "" && !ok
}

func convertClusterStates(states []api.ClusterState) []string {
	stringArr := make([]string, len(states))
	for i, v := range states {
		stringArr[i] = v.String()
	}
	return stringArr
}

func ClusterStateChange(pending []api.ClusterState, target []api.ClusterState, timeout time.Duration, refreshFunc resource.StateRefreshFunc) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    convertClusterStates(pending),
		Target:     convertClusterStates(target),
		Refresh:    refreshFunc,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}
}

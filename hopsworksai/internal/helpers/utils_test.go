package helpers

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestSuppressDiffForNonSetKeys(t *testing.T) {
	r := &schema.ResourceData{}

	if !SuppressDiffForNonSetKeys("test", "old", "", r) {
		t.Fatal("should suppress diff if value changes to default empty string")
	}

	if SuppressDiffForNonSetKeys("test", "", "", r) {
		t.Fatal("should not suppress diff if old value is empty string")
	}
}

func TestConvertClusterStates(t *testing.T) {
	cases := []struct {
		input    []api.ClusterState
		expected []string
	}{
		{
			input: []api.ClusterState{
				api.Initializing,
				api.Pending,
				api.Updating,
			},
			expected: []string{
				api.Initializing.String(),
				api.Pending.String(),
				api.Updating.String(),
			},
		},
		{
			input: []api.ClusterState{
				api.WorkerInitializing,
				api.WorkerDecommissioning,
				api.WorkerError,
			},
			expected: []string{
				api.WorkerInitializing.String(),
				api.WorkerDecommissioning.String(),
				api.WorkerError.String(),
			},
		},
	}

	for i, c := range cases {
		output := convertClusterStates(c.input)
		if !reflect.DeepEqual(c.expected, output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected, output)
		}
	}
}

func TestClusterStateChange(t *testing.T) {
	pending := []api.ClusterState{
		api.Pending,
		api.Initializing,
		api.Updating,
		api.WorkerInitializing,
		api.WorkerPending,
	}

	target := []api.ClusterState{
		api.Running,
		api.Error,
	}

	expected := "Expected-Output"

	bufferSize := 5
	state := make(chan string, bufferSize)

	refreshFunc := func() (interface{}, string, error) {
		value := <-state
		if value == api.Running.String() {
			return expected, value, nil
		}
		return "", value, nil
	}

	for i := 0; i < bufferSize-1; i++ {
		state <- pending[rand.Intn(len(pending))].String()
	}
	state <- api.Running.String()

	stateChange := clusterStateChange(pending, target, 20*time.Second, refreshFunc, 1*time.Second)
	output, err := stateChange.WaitForStateContext(context.TODO())

	if len(state) != 0 {
		t.Fatal("all state changes should be consumed")
	}

	if err != nil {
		t.Fatalf("unexpected error while waiting for state change %s", err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

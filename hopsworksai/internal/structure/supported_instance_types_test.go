package structure

import (
	"reflect"
	"testing"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestFlattenSupportedInstanceType(t *testing.T) {
	input := &api.SupportedInstanceType{
		Id:       "node-type",
		CPUs:     10,
		Memory:   30,
		GPUs:     1,
		WithNVMe: false,
	}

	expected := map[string]interface{}{
		"id":        "node-type",
		"cpus":      10,
		"memory":    30.0,
		"gpus":      1,
		"with_nvme": false,
	}

	output := flattenSupportedInstanceType(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenSupportedInstanceTypes(t *testing.T) {
	input := api.SupportedInstanceTypeList{
		{
			Id:       "node-type-1",
			CPUs:     10,
			Memory:   30,
			GPUs:     1,
			WithNVMe: false,
		},
		{
			Id:       "node-type-2",
			CPUs:     5,
			Memory:   20,
			GPUs:     0,
			WithNVMe: false,
		},
	}

	expected := []map[string]interface{}{
		{
			"id":        "node-type-1",
			"cpus":      10,
			"memory":    30.0,
			"gpus":      1,
			"with_nvme": false,
		},
		{
			"id":        "node-type-2",
			"cpus":      5,
			"memory":    20.0,
			"gpus":      0,
			"with_nvme": false,
		},
	}

	output := FlattenSupportedInstanceTypes(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

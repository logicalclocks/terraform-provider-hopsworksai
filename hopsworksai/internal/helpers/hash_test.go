package helpers

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestWorkerSetHash(t *testing.T) {
	worker := map[string]interface{}{
		"instance_type": "node-type-1",
		"disk_size":     512,
		"count":         1,
	}

	expected := schema.HashString(fmt.Sprintf("%s-%d-%d-", "node-type-1", 512, 1))
	output := WorkerSetHash(worker)
	if expected != output {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestSpotWorkerSetHash(t *testing.T) {
	worker := map[string]interface{}{
		"instance_type": "node-type-1",
		"disk_size":     512,
		"count":         1,
		"spot_config": []interface{}{
			map[string]interface{}{
				"max_price_percent":   100,
				"fall_back_on_demand": false,
			},
		},
	}

	expected := schema.HashString(fmt.Sprintf("%s-%d-%d-%d-%t-", "node-type-1", 512, 1, 100, false))
	output := WorkerSetHash(worker)
	if expected != output {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

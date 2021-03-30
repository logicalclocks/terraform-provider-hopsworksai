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

	expected := schema.HashString(fmt.Sprintf("%s-%d-", "node-type-1", 512))
	output := WorkerSetHash(worker)
	if expected != output {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}

	expected = schema.HashString(fmt.Sprintf("%s-%d-%d-", "node-type-1", 512, 1))

	output = WorkerSetHashIncludingCount(worker)
	if expected != output {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

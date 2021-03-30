package helpers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

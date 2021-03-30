package helpers

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func WorkerSetHash(val interface{}) int {
	return workerSetHash(val, false)
}

func WorkerSetHashIncludingCount(val interface{}) int {
	return workerSetHash(val, true)
}

func workerSetHash(val interface{}, includeCount bool) int {
	var buf bytes.Buffer
	workerConf := val.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", workerConf["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", workerConf["disk_size"].(int)))
	if includeCount {
		buf.WriteString(fmt.Sprintf("%d-", workerConf["count"].(int)))
	}
	return schema.HashString(buf.String())
}

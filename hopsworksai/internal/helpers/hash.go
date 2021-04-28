package helpers

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func WorkerSetHash(val interface{}) int {
	var buf bytes.Buffer
	workerConf := val.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", workerConf["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", workerConf["disk_size"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", workerConf["count"].(int)))
	return schema.HashString(buf.String())
}

package helpers

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func WorkerSetHash(val interface{}) int {
	var buf bytes.Buffer
	buf.WriteString(WorkerKey(val))
	workerConf := val.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", workerConf["count"].(int)))
	return schema.HashString(buf.String())
}

func WorkerKey(val interface{}) string {
	var buf bytes.Buffer
	workerConf := val.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", workerConf["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", workerConf["disk_size"].(int)))
	if _, ok := workerConf["spot_config"]; ok {
		spot_configArr := workerConf["spot_config"].([]interface{})
		if len(spot_configArr) > 0 && spot_configArr[0] != nil {
			spot_config := spot_configArr[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", spot_config["max_price_percent"].(int)))
			buf.WriteString(fmt.Sprintf("%t-", spot_config["fall_back_on_demand"].(bool)))
		}
	}
	return buf.String()
}

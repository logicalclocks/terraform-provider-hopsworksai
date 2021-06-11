package structure

import (
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func FlattenSupportedInstanceTypes(instanceTypes api.SupportedInstanceTypeList) []map[string]interface{} {
	supportedTypes := make([]map[string]interface{}, 0)
	for _, v := range instanceTypes {
		supportedTypes = append(supportedTypes, flattenSupportedInstanceType(&v))
	}
	return supportedTypes
}

func flattenSupportedInstanceType(instanceType *api.SupportedInstanceType) map[string]interface{} {
	return map[string]interface{}{
		"id":     instanceType.Id,
		"memory": instanceType.Memory,
		"cpus":   instanceType.CPUs,
		"gpus":   instanceType.GPUs,
	}
}

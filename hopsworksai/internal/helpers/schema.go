package helpers

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func GetDataSourceSchemaFromResourceSchema(resourceSchema map[string]*schema.Schema, skipList []string) map[string]*schema.Schema {
	dataSourceSchema := make(map[string]*schema.Schema, len(resourceSchema))
	for k, v := range resourceSchema {
		if inList(skipList, k) {
			continue
		}

		newSchema := &schema.Schema{
			Type:        v.Type,
			Description: v.Description,
			Computed:    true,
		}

		if v.Type == schema.TypeSet {
			newSchema.Set = v.Set
		}

		if elem, ok := v.Elem.(*schema.Resource); ok {
			newSchema.Elem = &schema.Resource{
				Schema: GetDataSourceSchemaFromResourceSchema(elem.Schema, skipList),
			}
		} else {
			newSchema.Elem = v.Elem
		}

		dataSourceSchema[k] = newSchema
	}
	return dataSourceSchema
}

func inList(list []string, elem string) bool {
	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

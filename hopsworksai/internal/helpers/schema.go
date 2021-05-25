package helpers

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func GetDataSourceSchemaFromResourceSchema(resourceSchema map[string]*schema.Schema) map[string]*schema.Schema {
	dataSourceSchema := make(map[string]*schema.Schema, len(resourceSchema))
	for k, v := range resourceSchema {
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
				Schema: GetDataSourceSchemaFromResourceSchema(elem.Schema),
			}
		} else {
			newSchema.Elem = v.Elem
		}

		dataSourceSchema[k] = newSchema
	}
	return dataSourceSchema
}

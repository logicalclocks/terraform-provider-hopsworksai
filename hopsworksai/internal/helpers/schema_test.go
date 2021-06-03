package helpers

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestGetDataSourceSchemaFromResourceSchema(t *testing.T) {
	input := map[string]*schema.Schema{
		"cluster_id": {
			Description: "The Id of the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the cluster, must be unique.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"head": {
			Description: "The configurations of the head node of the cluster.",
			Type:        schema.TypeList,
			Required:    true,
			ForceNew:    true,
			MaxItems:    1,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"instance_type": {
						Description: "The instance type of the head node.",
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
						Default:     "",
					},
					"disk_size": {
						Description: "The disk size of the head node in units of GB.",
						Type:        schema.TypeInt,
						Optional:    true,
						ForceNew:    true,
						Default:     512,
					},
				},
			},
		},
		"workers": {
			Description: "The configurations of worker nodes. You can add as many as you want of this block to create workers with different configurations.",
			Type:        schema.TypeSet,
			Optional:    true,
			Set:         WorkerSetHash,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"instance_type": {
						Description: "The instance type of the worker nodes.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "",
					},
					"disk_size": {
						Description: "The disk size of worker nodes in units of GB",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     512,
					},
					"count": {
						Description: "The number of worker nodes.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1,
					},
				},
			},
		},
		"tags": {
			Description: "The list of custom tags to be attached to the cluster.",
			Type:        schema.TypeMap,
			Optional:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"issue_lets_encrypt_certificate": {
			Description: "Enable or disable issuing let's encrypt certificates. This can be used to disable issuing certificates if port 80 can not be open.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     true,
		},
	}

	expected := map[string]*schema.Schema{
		"cluster_id": {
			Description: "The Id of the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the cluster, must be unique.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"head": {
			Description: "The configurations of the head node of the cluster.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"instance_type": {
						Description: "The instance type of the head node.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"disk_size": {
						Description: "The disk size of the head node in units of GB.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
				},
			},
		},
		"workers": {
			Description: "The configurations of worker nodes. You can add as many as you want of this block to create workers with different configurations.",
			Type:        schema.TypeSet,
			Computed:    true,
			Set:         WorkerSetHash,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"instance_type": {
						Description: "The instance type of the worker nodes.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"disk_size": {
						Description: "The disk size of worker nodes in units of GB",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"count": {
						Description: "The number of worker nodes.",
						Type:        schema.TypeInt,
						Computed:    true,
					},
				},
			},
		},
		"tags": {
			Description: "The list of custom tags to be attached to the cluster.",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"issue_lets_encrypt_certificate": {
			Description: "Enable or disable issuing let's encrypt certificates. This can be used to disable issuing certificates if port 80 can not be open.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}

	output := GetDataSourceSchemaFromResourceSchema(input)

	if len(output) != len(expected) {
		t.Fatalf("error while matching schema len:\nexpected %#v \nbut got %#v", len(expected), len(output))
	}

	for k, v := range expected {
		outputSchemaObj := reflect.ValueOf(*output[k])
		expectedSchemaObj := reflect.ValueOf(*v)

		for i := 0; i < expectedSchemaObj.NumField(); i++ {
			name := expectedSchemaObj.Type().Field(i).Name
			eField := expectedSchemaObj.Field(i).Interface()
			oField := outputSchemaObj.Field(i).Interface()
			if name == "Set" && v.Type == schema.TypeSet {
				testWorkerSchemaSet := []reflect.Value{
					reflect.ValueOf(map[string]interface{}{
						"instance_type": "node-type-1",
						"disk_size":     512,
						"count":         1,
					}),
				}
				eValue := expectedSchemaObj.Field(i).Call(testWorkerSchemaSet)[0].Interface().(int)
				oValue := outputSchemaObj.Field(i).Call(testWorkerSchemaSet)[0].Interface().(int)
				if eValue != oValue {
					t.Fatalf("error while matching %s - %s hash values:\nexpected %#v \nbut got %#v", k, name, eValue, oValue)
				}
				continue
			}
			if !reflect.DeepEqual(eField, oField) {
				t.Fatalf("error while matching %s - %s :\nexpected %#v \nbut got %#v", k, name, eField, oField)
			}
		}
	}

}

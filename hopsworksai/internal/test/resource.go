package test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api/test"
)

type httpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

type Operation struct {
	Method            string
	Path              string
	Response          string
	ExpectRequestBody string
}

func newHttpClient(t *testing.T, ops []Operation) *httpClient {
	return &httpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			for _, op := range ops {
				if req.URL.Path == op.Path && req.Method == op.Method {
					if op.ExpectRequestBody != "" {
						reqBody, err := ioutil.ReadAll(req.Body)
						if err != nil {
							return nil, err
						}
						reqBodyString := string(reqBody)
						expected := test.CompactJSONString(op.ExpectRequestBody)
						if reqBodyString != expected {
							t.Fatalf("invalid req body, expected:\n%s, but got:\n%s", expected, reqBodyString)
						}
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(op.Response)),
					}, nil
				}
			}
			return &http.Response{StatusCode: http.StatusServiceUnavailable}, nil
		},
	}
}

type ResourceFixture struct {
	HttpOps                   []Operation
	Resource                  *schema.Resource
	OperationContextFunc      func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	State                     map[string]interface{}
	Id                        string
	ExpectId                  string
	ExpectError               string
	ExpectState               map[string]interface{}
	ExpandStateCheckOnlyArray string
	Update                    bool
}

func getResourceData(ctx context.Context, t *testing.T, r *schema.Resource, currentState *terraform.InstanceState, newState map[string]interface{}, mockClient interface{}) *schema.ResourceData {
	schemaMap := schema.InternalMap(r.Schema)
	var diff = terraform.NewInstanceDiff()
	if newState != nil {
		resourceConfig := terraform.NewResourceConfigRaw(newState)
		if d, err := r.Diff(ctx, currentState, resourceConfig, mockClient); err != nil {
			t.Fatal(err)
		} else {
			diff = d
		}
	}

	if d, err := schemaMap.Data(currentState, diff); err != nil {
		t.Fatal(err)
	} else {
		return d
	}
	return nil
}

func (r *ResourceFixture) Apply(t *testing.T, ctx context.Context) {
	mockClient := &api.HopsworksAIClient{
		Client: newHttpClient(t, r.HttpOps),
	}

	var data *schema.ResourceData
	data = getResourceData(ctx, t, r.Resource, &terraform.InstanceState{}, r.State, mockClient)

	if r.Id != "" {
		data.SetId(r.Id)
	}

	if r.Update {
		if r.Resource.ReadContext != nil {
			if diag := r.Resource.ReadContext(ctx, data, mockClient); diag.HasError() {
				t.Fatalf("unexpected error %#v", diag)
			}
		} else {
			t.Fatalf("unexpected error Update is set to true on resource with no ReadContext")
		}

		data = getResourceData(ctx, t, r.Resource, data.State(), r.State, mockClient)
	}

	diag := r.OperationContextFunc(ctx, data, mockClient)
	if r.ExpectError != "" {
		var errMessage string = ""
		if diag.HasError() {
			errMessage = diag[0].Summary
		}
		if r.ExpectError != errMessage {
			t.Fatalf("expected error %s but got %#v", r.ExpectError, diag)
		}
	} else if diag.HasError() {
		t.Fatalf("unexpected error %#v", diag)
	}

	if r.ExpectId != "" && data.Id() != r.ExpectId {
		t.Fatalf("error matching resource id, expected:\n%s, but got:\n%s", r.ExpectId, data.Id())
	}

	if r.ExpectState != nil {
		if r.ExpandStateCheckOnlyArray != "" {
			expectedArr := r.ExpectState[r.ExpandStateCheckOnlyArray].([]interface{})
			for i := range expectedArr {
				checkStateEqual(t, fmt.Sprintf("%s.%d.", r.ExpandStateCheckOnlyArray, i), expectedArr[i].(map[string]interface{}), data)
			}
		} else {
			checkStateEqual(t, "", r.ExpectState, data)
		}
	}
}

func checkStateEqual(t *testing.T, keyPrefix string, expectedState map[string]interface{}, data *schema.ResourceData) {
	for k, v := range expectedState {
		o := data.Get(keyPrefix + k)
		if vs, ok := v.(*schema.Set); ok {
			if !vs.Equal(o) {
				t.Fatalf("error matching state %s, expected:\n%s, but got:\n%s", k, v, o)
			}
			continue
		}

		if !reflect.DeepEqual(v, o) {
			t.Fatalf("error matching state %s, expected:\n%s, but got:\n%s", k, v, o)
		}
	}
}

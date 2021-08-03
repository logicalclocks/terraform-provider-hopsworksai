package hopsworksai

import (
	"context"
	"net/http"
	"testing"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

// unit tests

func TestVersionDataSourceRead_AWS_latest(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{},
		"5.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1", "region-2"},
				"ubuntu": []interface{}{"region-3", "region-4"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_latest(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{},
		"5.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1", "region-2"},
				"ubuntu": []interface{}{"region-3", "region-4"},
			},
		})
}

func TestVersionDataSourceRead_AWS_default(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{
			"default": true,
		},
		"4.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1-d", "region-2-d"},
				"ubuntu": []interface{}{"ALL"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_default(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{
			"default": true,
		},
		"4.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1-d", "region-2-d"},
				"ubuntu": []interface{}{"ALL"},
			},
		})
}

func TestVersionDataSourceRead_AWS_experimental(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{
			"experimental": true,
		},
		"5.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1", "region-2"},
				"ubuntu": []interface{}{"region-3", "region-4"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_experimental(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{
			"experimental": true,
		},
		"5.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1", "region-2"},
				"ubuntu": []interface{}{"region-3", "region-4"},
			},
		})
}

func TestVersionDataSourceRead_AWS_upgradeableFrom(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{
			"upgradeable_from_version": "2.0",
		},
		"3.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1", "region-2"},
				"ubuntu": []interface{}{"region-3", "region-4"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_upgradeableFrom(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{
			"upgradeable_from_version": "2.0",
		},
		"3.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1", "region-2"},
				"ubuntu": []interface{}{"region-3", "region-4"},
			},
		})
}

func TestVersionDataSourceRead_AWS_os_region(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{
			"os":     "centos",
			"region": "region-1-d",
		},
		"4.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1-d", "region-2-d"},
				"ubuntu": []interface{}{"ALL"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_os_region(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{
			"os":     "centos",
			"region": "region-1-d",
		},
		"4.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1-d", "region-2-d"},
				"ubuntu": []interface{}{"ALL"},
			},
		})
}

func TestVersionDataSourceRead_AWS_os_region_ALL(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{
			"os":     "ubuntu",
			"region": "region-1-d",
		},
		"4.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1-d", "region-2-d"},
				"ubuntu": []interface{}{"ALL"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_os_region_ALL(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{
			"os":     "ubuntu",
			"region": "region-1-d",
		},
		"4.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{"region-1-d", "region-2-d"},
				"ubuntu": []interface{}{"ALL"},
			},
		})
}

func TestVersionDataSourceRead_AWS_multi_condition(t *testing.T) {
	testVersionDataSourceRead(t, api.AWS,
		map[string]interface{}{
			"os":           "ubuntu",
			"region":       "region-5",
			"experimental": true,
		},
		"1.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{},
				"ubuntu": []interface{}{"region-5", "region-6"},
			},
		})
}

func TestVersionDataSourceRead_AZURE_multi_condition(t *testing.T) {
	testVersionDataSourceRead(t, api.AZURE,
		map[string]interface{}{
			"os":           "ubuntu",
			"region":       "region-5",
			"experimental": true,
		},
		"1.0",
		[]interface{}{
			map[string]interface{}{
				"centos": []interface{}{},
				"ubuntu": []interface{}{"region-5", "region-6"},
			},
		})
}
func TestVersionDataSourceRead_AWS_upgradeableFrom_error(t *testing.T) {
	testVersionDataSourceRead_error(t, api.AWS,
		map[string]interface{}{
			"upgradeable_from_version": "4.0",
		},
	)
}

func TestVersionDataSourceRead_AZURE_upgradeableFrom_error(t *testing.T) {
	testVersionDataSourceRead_error(t, api.AZURE,
		map[string]interface{}{
			"upgradeable_from_version": "4.0",
		},
	)
}

func TestVersionDataSourceRead_AWS_os_region_error(t *testing.T) {
	testVersionDataSourceRead_error(t, api.AWS,
		map[string]interface{}{
			"os":     "centos",
			"region": "region-3",
		},
	)
}

func TestVersionDataSourceRead_AZURE_os_region_error(t *testing.T) {
	testVersionDataSourceRead_error(t, api.AZURE,
		map[string]interface{}{
			"os":     "centos",
			"region": "region-3",
		},
	)
}

func TestVersionDataSourceRead_AWS_API_error(t *testing.T) {
	testVersionDataSourceRead_API_error(t, api.AWS)
}

func TestVersionDataSourceRead_AZURE_API_error(t *testing.T) {
	testVersionDataSourceRead_API_error(t, api.AZURE)
}

func testVersionDataSourceRead(t *testing.T, cloud api.CloudProvider, state map[string]interface{}, expectedId string, expectedRegions []interface{}) {
	state["cloud_provider"] = cloud.String()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/hopsworks/versions/" + cloud.String(),
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload": {
						"versions":[
							{
								"version": "1.0",
								"upgradableFromVersion": "N/A",
								"default": false,
								"experimental": true,
								"regions": {
									"ubuntu": [
										"region-5",
										"region-6"
									]
								}
							},
							{
								"version": "2.0",
								"upgradableFromVersion": "1.0",
								"default": false,
								"experimental": false,
								"regions": {
									"centos": [
										"region-1",
										"region-2"
									]
								}
							},
							{
								"version": "3.0",
								"upgradableFromVersion": "2.0",
								"default": false,
								"experimental": false,
								"regions": {
									"centos": [
										"region-1",
										"region-2"
									],
									"ubuntu": [
										"region-3",
										"region-4"
									]
								}
							},
							{
								"version": "4.0",
								"upgradableFromVersion": "3.0",
								"default": true,
								"experimental": false,
								"regions": {
									"centos": [
										"region-1-d",
										"region-2-d"
									],
									"ubuntu": [
										"ALL"
									]
								}
							},
							{
								"version": "5.0",
								"upgradableFromVersion": "N/A",
								"default": false,
								"experimental": true,
								"regions": {
									"centos": [
										"region-1",
										"region-2"
									],
									"ubuntu": [
										"region-3",
										"region-4"
									]
								}
							}
						]
					}
				}`,
			},
		},
		Resource:             dataSourceVersion(),
		OperationContextFunc: dataSourceVersion().ReadContext,
		State:                state,
		ExpectState: map[string]interface{}{
			"supported_regions": expectedRegions,
		},
		ExpectId: expectedId,
	}
	r.Apply(t, context.TODO())
}

func testVersionDataSourceRead_error(t *testing.T, cloud api.CloudProvider, state map[string]interface{}) {
	state["cloud_provider"] = cloud.String()
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/hopsworks/versions/" + cloud.String(),
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload": {
						"versions":[
							{
								"version": "1.0",
								"upgradableFromVersion": "N/A",
								"default": false,
								"experimental": false,
								"regions": {
									"ubuntu": [
										"region-1",
										"region-2"
									]
								}
							},
							{
								"version": "2.0",
								"upgradableFromVersion": "1.0",
								"default": false,
								"experimental": false,
								"regions": {
									"centos": [
										"region-1",
										"region-2"
									]
								}
							},
							{
								"version": "3.0",
								"upgradableFromVersion": "2.0",
								"default": false,
								"experimental": false,
								"regions": {
									"centos": [
										"region-1",
										"region-2"
									],
									"ubuntu": [
										"region-3",
										"region-4"
									]
								}
							},
							{
								"version": "4.0",
								"upgradableFromVersion": "3.0",
								"default": true,
								"experimental": false,
								"regions": {
									"centos": [
										"region-1-d",
										"region-2-d"
									],
									"ubuntu": [
										"ALL"
									]
								}
							},
							{
								"version": "5.0",
								"upgradableFromVersion": "N/A",
								"default": false,
								"experimental": true,
								"regions": {
									"centos": [
										"region-1",
										"region-2"
									],
									"ubuntu": [
										"region-3",
										"region-4"
									]
								}
							}
						]
					}
				}`,
			},
		},
		Resource:             dataSourceVersion(),
		OperationContextFunc: dataSourceVersion().ReadContext,
		State:                state,
		ExpectError:          "no version found matching the provided filters.",
	}
	r.Apply(t, context.TODO())
}

func testVersionDataSourceRead_API_error(t *testing.T, cloud api.CloudProvider) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/hopsworks/versions/" + cloud.String(),
				Response: `{
					"apiVersion": "v1",
					"status": "error",
					"code": 400,
					"message": "cannot retrieve versions error"
				}`,
			},
		},
		Resource:             dataSourceVersion(),
		OperationContextFunc: dataSourceVersion().ReadContext,
		State: map[string]interface{}{
			"cloud_provider": cloud.String(),
		},
		ExpectError: "cannot retrieve versions error",
	}
	r.Apply(t, context.TODO())
}

package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/test"
)

func TestAccClustersDataSourceAWS_basic(t *testing.T) {
	testAccClustersDataSource_basic(t, api.AWS)
}

func TestAccClustersDataSourceAZURE_basic(t *testing.T) {
	testAccClustersDataSource_basic(t, api.AZURE)
}

func testAccClustersDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	dataSourceName := fmt.Sprintf("data.hopsworksai_clusters.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(cloud, rName, suffix),
				Check:  testAccClustersDataSourceCheckAllAttributes(cloud, resourceName, dataSourceName),
			},
		},
	})
}

func testAccClustersDataSourceCheckAllAttributes(cloud api.CloudProvider, resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found", dataSourceName)
		}

		var index string = ""
		listClustersTagPattern := regexp.MustCompile(`^clusters\.([0-9]*)\.tags.ListClusters$`)
		for k, v := range ds.Primary.Attributes {
			submatches := listClustersTagPattern.FindStringSubmatch(k)
			if len(submatches) == 2 && v == cloud.String() {
				index = submatches[1]
			}
		}

		if index == "" {
			return fmt.Errorf("no clusters returned")
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		for k := range rs.Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			dataSourceKey := fmt.Sprintf("clusters.%s.%s", index, k)
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, dataSourceKey)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClustersDataSourceConfig(cloud api.CloudProvider, rName string, suffix string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"
		head {
			instance_type = "%s"
		}

		%s


		tags = {
		  "ListClusters" = "%s"
		  "%s" = "%s"
		  "Test" = "TestAccClustersDataSource_basic"
		}
	  }

	  data "hopsworksai_clusters" "%s" {
		  depends_on = [
			hopsworksai_cluster.%s
		  ]
	  }
	`,
		rName,
		default_CLUSTER_NAME_PREFIX,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testHeadInstanceType(cloud),
		testAccClusterCloudConfigAttributes(cloud, 1, false),
		cloud.String(),
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
		rName,
		rName,
	)
}

// Unit tests

func TestClustersDataSourceRead(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"clusters":[
							{
								"id": "cluster-1",
								"name": "cluster-name-1",
								"createdOn": 1,
								"provider": "AWS"
							},
							{
								"id": "cluster-2",
								"name": "cluster-name-2",
								"createdOn": 2,
								"provider": "AWS"
							},
							{
								"id": "cluster-3",
								"name": "cluster-name-3",
								"createdOn": 3,
								"provider": "AZURE"
							}
						]
					}
				}`,
			},
		},
		Resource:                  dataSourceClusters(),
		OperationContextFunc:      dataSourceClusters().ReadContext,
		ExpandStateCheckOnlyArray: "clusters",
		ExpectState: map[string]interface{}{
			"clusters": []interface{}{
				map[string]interface{}{
					"cluster_id":       "cluster-1",
					"name":             "cluster-name-1",
					"state":            "",
					"activation_state": "",
					"creation_date":    time.Unix(1, 0).Format(time.RFC3339),
					"start_date":       time.Unix(0, 0).Format(time.RFC3339),
					"version":          "",
					"url":              "",
					"tags":             map[string]interface{}{},
					"ssh_key":          "",
					"head": []interface{}{
						map[string]interface{}{
							"instance_type": "",
							"disk_size":     0,
							"node_id":       "",
							"ha_enabled":    false,
						},
					},
					"workers":                        schema.NewSet(helpers.WorkerSetHash, []interface{}{}),
					"attach_public_ip":               false,
					"issue_lets_encrypt_certificate": false,
					"managed_users":                  false,
					"backup_retention_period":        0,
					"azure_attributes":               []interface{}{},
					"aws_attributes": []interface{}{
						map[string]interface{}{
							"region":               "",
							"instance_profile_arn": "",
							"network": []interface{}{
								map[string]interface{}{
									"vpc_id":            "",
									"subnet_id":         "",
									"security_group_id": "",
								},
							},
							"eks_cluster_name":        "",
							"ecr_registry_account_id": "",
							"bucket": []interface{}{
								map[string]interface{}{
									"name":       "",
									"encryption": []interface{}{},
									"acl":        []interface{}{},
								},
							},
							"ebs_encryption": []interface{}{},
						},
					},
					"open_ports": []interface{}{
						map[string]interface{}{
							"ssh":                  false,
							"kafka":                false,
							"feature_store":        false,
							"online_feature_store": false,
						},
					},
					"update_state": "none",
				},
				map[string]interface{}{
					"cluster_id":       "cluster-2",
					"name":             "cluster-name-2",
					"state":            "",
					"activation_state": "",
					"creation_date":    time.Unix(2, 0).Format(time.RFC3339),
					"start_date":       time.Unix(0, 0).Format(time.RFC3339),
					"version":          "",
					"url":              "",
					"tags":             map[string]interface{}{},
					"ssh_key":          "",
					"head": []interface{}{
						map[string]interface{}{
							"instance_type": "",
							"disk_size":     0,
							"node_id":       "",
							"ha_enabled":    false,
						},
					},
					"workers":                        schema.NewSet(helpers.WorkerSetHash, []interface{}{}),
					"attach_public_ip":               false,
					"issue_lets_encrypt_certificate": false,
					"managed_users":                  false,
					"backup_retention_period":        0,
					"azure_attributes":               []interface{}{},
					"aws_attributes": []interface{}{
						map[string]interface{}{
							"region":               "",
							"instance_profile_arn": "",
							"network": []interface{}{
								map[string]interface{}{
									"vpc_id":            "",
									"subnet_id":         "",
									"security_group_id": "",
								},
							},
							"eks_cluster_name":        "",
							"ecr_registry_account_id": "",
							"bucket": []interface{}{
								map[string]interface{}{
									"name":       "",
									"encryption": []interface{}{},
									"acl":        []interface{}{},
								},
							},
							"ebs_encryption": []interface{}{},
						},
					},
					"open_ports": []interface{}{
						map[string]interface{}{
							"ssh":                  false,
							"kafka":                false,
							"feature_store":        false,
							"online_feature_store": false,
						},
					},
					"update_state": "none",
				},
				map[string]interface{}{
					"cluster_id":       "cluster-3",
					"name":             "cluster-name-3",
					"state":            "",
					"activation_state": "",
					"creation_date":    time.Unix(3, 0).Format(time.RFC3339),
					"start_date":       time.Unix(0, 0).Format(time.RFC3339),
					"version":          "",
					"url":              "",
					"tags":             map[string]interface{}{},
					"ssh_key":          "",
					"head": []interface{}{
						map[string]interface{}{
							"instance_type": "",
							"disk_size":     0,
							"node_id":       "",
							"ha_enabled":    false,
						},
					},
					"workers":                        schema.NewSet(helpers.WorkerSetHash, []interface{}{}),
					"attach_public_ip":               false,
					"issue_lets_encrypt_certificate": false,
					"managed_users":                  false,
					"backup_retention_period":        0,
					"azure_attributes": []interface{}{
						map[string]interface{}{
							"location":                       "",
							"resource_group":                 "",
							"storage_account":                "",
							"user_assigned_managed_identity": "",
							"storage_container_name":         "",
							"network": []interface{}{
								map[string]interface{}{
									"resource_group":       "",
									"virtual_network_name": "",
									"subnet_name":          "",
									"security_group_name":  "",
									"search_domain":        "",
								},
							},
							"aks_cluster_name":  "",
							"acr_registry_name": "",
							"search_domain":     "",
							"container": []interface{}{
								map[string]interface{}{
									"name":            "",
									"storage_account": "",
									"encryption":      []interface{}{},
								},
							},
						},
					},
					"aws_attributes": []interface{}{},
					"open_ports": []interface{}{
						map[string]interface{}{
							"ssh":                  false,
							"kafka":                false,
							"feature_store":        false,
							"online_feature_store": false,
						},
					},
					"update_state": "none",
				},
			},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClustersDataSourceRead_filter(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 200,
					"payload":{
						"clusters":[
						]
					}
				}`,
			},
		},
		Resource:             dataSourceClusters(),
		OperationContextFunc: dataSourceClusters().ReadContext,
		State: map[string]interface{}{
			"filter": []interface{}{
				map[string]interface{}{
					"cloud": "AWS",
				},
			},
		},
		ExpectState: map[string]interface{}{
			"clusters": []interface{}{},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClustersDataSourceRead_error(t *testing.T) {
	r := test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters",
				Response: `{
					"apiVersion": "v1",
					"status": "ok",
					"code": 400,
					"message": "bad request failed to get filtered clusters"
				}`,
			},
		},
		Resource:             dataSourceClusters(),
		OperationContextFunc: dataSourceClusters().ReadContext,
		State: map[string]interface{}{
			"filter": []interface{}{
				map[string]interface{}{
					"cloud": "AWS",
				},
			},
		},
		ExpectError: "bad request failed to get filtered clusters",
	}
	r.Apply(t, context.TODO())
}

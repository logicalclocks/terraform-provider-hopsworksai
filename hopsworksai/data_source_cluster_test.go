package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
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

func TestAccClusterDataSourceAWS_basic(t *testing.T) {
	testAccClusterDataSource_basic(t, api.AWS)
}

func TestAccClusterDataSourceAZURE_basic(t *testing.T) {
	testAccClusterDataSource_basic(t, api.AZURE)
}

func testAccClusterDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	dataSourceName := fmt.Sprintf("data.hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:  testAccPreCheck(t),
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig(cloud, rName, suffix),
				Check:  testAccClusterDataSourceCheckAllAttributes(resourceName, dataSourceName),
			},
		},
	})
}

func testAccClusterDataSourceCheckAllAttributes(resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		for k := range rs.Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, k)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClusterDataSourceConfig(cloud api.CloudProvider, rName string, suffix string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"	  
		head {
		}
		
		%s
		

		tags = {
		  "%s" = "%s"
		}
	  }

	  data "hopsworksai_cluster" "%s" {
		  cluster_id = hopsworksai_cluster.%s.id
	  }
	`,
		rName,
		default_CLUSTER_NAME_PREFIX,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testAccClusterCloudConfigAttributes(cloud, 3),
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
		rName,
		rName,
	)
}

// Unit tests
func TestClusterDataSourceRead_AWS(t *testing.T) {
	r := &test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"statue": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id": "cluster-id-1",
							"name": "cluster-name-1",
							"state" : "running", 
							"activationState": "stoppable", 
							"initializationStage": "running", 
							"createdOn": 123, 
							"startedOn" : 123,
							"version": "version-1",
							"url": "https://cluster-url",
							"provider": "AWS",
							"tags": [
								{
									"name": "tag1",
									"value": "tag1-value1"
								}
							],
							"sshKeyName": "ssh-key-1",
							"clusterConfiguration": {
								"head": {
									"instanceType": "node-type-1",
									"diskSize": 512
								},
								"workers": [
									{
										"instanceType": "node-type-2",
										"diskSize": 256,
										"count": 2
									}
								]
							},
							"publicIPAttached": true,
							"letsEncryptIssued": true,
							"managedUsers": true,
							"backupRetentionPeriod": 10,
							"aws": {
								"region": "region-1",
								"instanceProfileArn": "profile-1",
								"bucketName": "bucket-1",
								"vpcId": "vpc-1",
								"subnetId": "subnet-1",
								"securityGroupId": "security-group-1"
							}
						}
					}
				}`,
			},
		},
		Resource:             dataSourceCluster(),
		OperationContextFunc: dataSourceCluster().ReadContext,
		Id:                   "cluster-id-1",
		State: map[string]interface{}{
			"cluster_id": "cluster-id-1",
		},
		ExpectState: map[string]interface{}{
			"cluster_id":       "cluster-id-1",
			"state":            "running",
			"activation_state": "stoppable",
			"creation_date":    time.Unix(123, 0).Format(time.RFC3339),
			"start_date":       time.Unix(123, 0).Format(time.RFC3339),
			"version":          "version-1",
			"url":              "https://cluster-url",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"ssh_key": "ssh-key-1",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			}),
			"attach_public_ip":               true,
			"issue_lets_encrypt_certificate": true,
			"managed_users":                  true,
			"backup_retention_period":        10,
			"aws_attributes": []interface{}{
				map[string]interface{}{
					"region":               "region-1",
					"bucket_name":          "bucket-1",
					"instance_profile_arn": "profile-1",
					"network": []interface{}{
						map[string]interface{}{
							"vpc_id":            "vpc-1",
							"subnet_id":         "subnet-1",
							"security_group_id": "security-group-1",
						},
					},
					"eks_cluster_name":        "",
					"ecr_registry_account_id": "",
				},
			},
			"azure_attributes": []interface{}{},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterDataSourceRead_AZURE(t *testing.T) {
	r := &test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"statue": "ok",
					"code": 200,
					"payload":{
						"cluster": {
							"id": "cluster-id-1",
							"name": "cluster-name-1",
							"state" : "running", 
							"activationState": "stoppable", 
							"initializationStage": "running", 
							"createdOn": 123, 
							"startedOn" : 123,
							"version": "version-1",
							"url": "https://cluster-url",
							"provider": "AZURE",
							"tags": [
								{
									"name": "tag1",
									"value": "tag1-value1"
								}
							],
							"sshKeyName": "ssh-key-1",
							"clusterConfiguration": {
								"head": {
									"instanceType": "node-type-1",
									"diskSize": 512
								},
								"workers": [
									{
										"instanceType": "node-type-2",
										"diskSize": 256,
										"count": 2
									}
								]
							},
							"publicIPAttached": true,
							"letsEncryptIssued": true,
							"managedUsers": true,
							"backupRetentionPeriod": 10,
							"azure": {
								"location": "location-1",
								"resourceGroup": "resource-group-1",
								"managedIdentity": "profile-1",
								"blobContainerName": "container-1",
								"storageAccount": "account-1",
								"virtualNetworkName": "network-name-1",
								"subnetName": "subnet-name-1",
								"securityGroupName": "security-group-name-1"
							}
						}
					}
				}`,
			},
		},
		Resource:             dataSourceCluster(),
		OperationContextFunc: dataSourceCluster().ReadContext,
		Id:                   "cluster-id-1",
		State: map[string]interface{}{
			"cluster_id": "cluster-id-1",
		},
		ExpectState: map[string]interface{}{
			"cluster_id":       "cluster-id-1",
			"state":            "running",
			"activation_state": "stoppable",
			"creation_date":    time.Unix(123, 0).Format(time.RFC3339),
			"start_date":       time.Unix(123, 0).Format(time.RFC3339),
			"version":          "version-1",
			"url":              "https://cluster-url",
			"tags": map[string]interface{}{
				"tag1": "tag1-value1",
			},
			"ssh_key": "ssh-key-1",
			"head": []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
				},
			},
			"workers": schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         2,
				},
			}),
			"attach_public_ip":               true,
			"issue_lets_encrypt_certificate": true,
			"managed_users":                  true,
			"backup_retention_period":        10,
			"azure_attributes": []interface{}{
				map[string]interface{}{
					"location":                       "location-1",
					"resource_group":                 "resource-group-1",
					"storage_account":                "account-1",
					"user_assigned_managed_identity": "profile-1",
					"storage_container_name":         "container-1",
					"network": []interface{}{
						map[string]interface{}{
							"virtual_network_name": "network-name-1",
							"subnet_name":          "subnet-name-1",
							"security_group_name":  "security-group-name-1",
						},
					},
					"aks_cluster_name":  "",
					"acr_registry_name": "",
				},
			},
			"aws_attributes": []interface{}{},
		},
	}
	r.Apply(t, context.TODO())
}

func TestClusterDataSourceRead_error(t *testing.T) {
	r := &test.ResourceFixture{
		HttpOps: []test.Operation{
			{
				Method: http.MethodGet,
				Path:   "/api/clusters/cluster-id-1",
				Response: `{
					"apiVersion": "v1",
					"statue": "ok",
					"code": 400,
					"message": "bad request get cluster failed"
				}`,
			},
		},
		Resource:             dataSourceCluster(),
		OperationContextFunc: dataSourceCluster().ReadContext,
		State: map[string]interface{}{
			"cluster_id": "cluster-id-1",
		},
		ExpectError: "bad request get cluster failed",
	}
	r.Apply(t, context.TODO())
}

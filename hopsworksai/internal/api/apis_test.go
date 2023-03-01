package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api/test"
)

func TestInvalidAPIKey(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ResponseCode: http.StatusForbidden,
			ResponseBody: "Unauthorized",
			T:            t,
		},
	}
	output, err := GetCluster(context.TODO(), apiClient, "cluster-id")
	if err == nil {
		t.Fatal("client should relay the error")
	}
	if err.Error() != "the API token provided does not have access to hopsworks.ai, verify the token you specified matches the token hopsworks.ai created:\n\tUnauthorized" {
		t.Fatalf("client should relay the error, but got %s", err)
	}

	if output != nil {
		t.Fatalf("should return nil if encountered error, but got %#v", output)
	}
}

func TestJsonErrors(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "latest",
				}`,
			T: t,
		},
	}
	output, err := GetCluster(context.TODO(), apiClient, "cluster-id")
	if err == nil {
		t.Fatal("client should relay the error")
	}
	if !strings.HasPrefix(err.Error(), "failed to decode json") {
		t.Fatalf("client should relay the json error, but got %s", err)
	}

	if output != nil {
		t.Fatalf("should return nil if encountered error, but got %#v", output)
	}
}

func TestGetClusterAWS(t *testing.T) {
	apiClient := &HopsworksAIClient{
		ApiKey:     "my-api-key",
		ApiVersion: "testV1",
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/clusters/cluster-id-1",
			ExpectHeaders: map[string]string{
				"x-api-key":          "my-api-key",
				"hopsai-api-version": "testV1",
				"Content-Type":       "application/json",
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{
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
								"diskSize": 512,
								"nodeId": "head-node-id"
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
							"securityGroupId": "security-group-1",
							"eksClusterName": "eks-cluster-1",
							"ecrRegistryAccountId": "ecr-account-1"
						}
					}
				}
			}`,
			T: t,
		},
	}

	expected := &Cluster{
		Id:                  "cluster-id-1",
		Name:                "cluster-name-1",
		State:               Running,
		ActivationState:     Stoppable,
		InitializationStage: "running",
		CreatedOn:           123,
		StartedOn:           123,
		Version:             "version-1",
		URL:                 "https://cluster-url",
		Provider:            AWS,
		Tags: []ClusterTag{
			{
				Name:  "tag1",
				Value: "tag1-value1",
			},
		},
		SshKeyName: "ssh-key-1",
		ClusterConfiguration: ClusterConfigurationStatus{
			Head: HeadConfigurationStatus{
				HeadConfiguration: HeadConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
					HAEnabled: false,
				},
				NodeId: "head-node-id",
			},
			Workers: []WorkerConfiguration{
				{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-2",
						DiskSize:     256,
					},
					Count: 2,
				},
			},
		},
		PublicIPAttached:      true,
		LetsEncryptIssued:     true,
		ManagedUsers:          true,
		BackupRetentionPeriod: 10,
		AWS: AWSCluster{
			Region:               "region-1",
			BucketName:           "bucket-1",
			InstanceProfileArn:   "profile-1",
			VpcId:                "vpc-1",
			SubnetId:             "subnet-1",
			SecurityGroupId:      "security-group-1",
			EksClusterName:       "eks-cluster-1",
			EcrRegistryAccountId: "ecr-account-1",
		},
	}

	output, err := GetCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("get cluster shouldn't throw error, got %s", err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestGetClusterAZURE(t *testing.T) {
	apiClient := &HopsworksAIClient{
		ApiKey:     "my-api-key",
		ApiVersion: "testV1",
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/clusters/cluster-id-1",
			ExpectHeaders: map[string]string{
				"x-api-key":          "my-api-key",
				"hopsai-api-version": "testV1",
				"Content-Type":       "application/json",
			},
			ResponseCode: http.StatusOK,
			ResponseBody: `{
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
								"diskSize": 512,
								"nodeId": "head-node-id-1"
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
							"securityGroupName": "security-group-name-1",
							"aksClusterName": "aks-cluster-name-1",
							"acrRegistryName": "acr-registry-name-1",
							"searchDomain": "internal.cloudapp.net"
						}
					}
				}
			}`,
			T: t,
		},
	}

	expected := &Cluster{
		Id:                  "cluster-id-1",
		Name:                "cluster-name-1",
		State:               Running,
		ActivationState:     Stoppable,
		InitializationStage: "running",
		CreatedOn:           123,
		StartedOn:           123,
		Version:             "version-1",
		URL:                 "https://cluster-url",
		Provider:            AZURE,
		Tags: []ClusterTag{
			{
				Name:  "tag1",
				Value: "tag1-value1",
			},
		},
		SshKeyName: "ssh-key-1",
		ClusterConfiguration: ClusterConfigurationStatus{
			Head: HeadConfigurationStatus{
				HeadConfiguration: HeadConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
					HAEnabled: false,
				},
				NodeId: "head-node-id-1",
			},
			Workers: []WorkerConfiguration{
				{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-2",
						DiskSize:     256,
					},
					Count: 2,
				},
			},
		},
		PublicIPAttached:      true,
		LetsEncryptIssued:     true,
		ManagedUsers:          true,
		BackupRetentionPeriod: 10,
		Azure: AzureCluster{
			Location:           "location-1",
			ResourceGroup:      "resource-group-1",
			ManagedIdentity:    "profile-1",
			BlobContainerName:  "container-1",
			StorageAccount:     "account-1",
			VirtualNetworkName: "network-name-1",
			SubnetName:         "subnet-name-1",
			SecurityGroupName:  "security-group-name-1",
			AksClusterName:     "aks-cluster-name-1",
			AcrRegistryName:    "acr-registry-name-1",
			SearchDomain:       "internal.cloudapp.net",
		},
	}

	output, err := GetCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("get cluster shouldn't throw error, got %s", err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestGetClusterNotFound(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ResponseCode: http.StatusNotFound,
			ResponseBody: `{
				"apiVersion": "latest",
				"status": "error",
				"code": 404
			}`,
			T: t,
		},
	}

	output, err := GetCluster(context.TODO(), apiClient, "cluster-id")

	if err != nil {
		t.Fatalf("get cluster shouldn't throw error if not found, got %s", err)
	}

	if output != nil {
		t.Fatalf("get cluster should return nil if cluster not found, got %#v", output)
	}
}

func TestGetClusterError(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{
				"apiVersion": "latest",
				"status": "error",
				"code": 400,
				"message": "bad request you cannot retrieve cluster info"
			}`,
			T: t,
		},
	}

	output, err := GetCluster(context.TODO(), apiClient, "cluster-id")

	if err == nil {
		t.Fatal("get cluster should relay the error sent")
	}

	if err.Error() != "bad request you cannot retrieve cluster info" {
		t.Fatalf("get cluster should relay the error message it got from Hopsworks.ai, got %s", err)
	}

	if output != nil {
		t.Fatalf("get cluster should return nil if encountered an error, got %#v", output)
	}
}

func testGetClustersWithFilter(provider string, t *testing.T) ([]Cluster, error) {
	var expectedQuery string = ""
	if provider != "" {
		expectedQuery = "cloud=" + provider
	}

	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectPath:         "/api/clusters",
			ExpectMethod:       http.MethodGet,
			ExpectRequestQuery: expectedQuery,
			ResponseCode:       http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"statue": "ok",
				"code": 200,
				"payload":{
					"clusters": []
				}
			 }`,
			T: t,
		},
	}

	return GetClusters(context.TODO(), apiClient, CloudProvider(provider))
}
func TestGetClustersSettingFilter(t *testing.T) {
	if _, err := testGetClustersWithFilter(AWS.String(), t); err != nil {
		t.Fatalf("get clusters with filter should not throw an error, but got %s", err)
	}
	if _, err := testGetClustersWithFilter(AZURE.String(), t); err != nil {
		t.Fatalf("get clusters with filter should not throw an error, but got %s", err)
	}
	if _, err := testGetClustersWithFilter("", t); err != nil {
		t.Fatalf("get clusters with filter should not throw an error, but got %s", err)
	}
}

func TestGetClusters(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ResponseCode: http.StatusOK,
			ResponseBody: `{
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
			T: t,
		},
	}

	clusters, err := GetClusters(context.TODO(), apiClient, "")
	if err != nil {
		t.Fatalf("get clusters should not throw an error, but got %s", err)
	}
	if clusters == nil {
		t.Fatal("get clusters should return a list of clusters not nil")
	}
	if len(clusters) != 3 {
		t.Fatalf("get clusters should return 3 clusters but got %d instead", len(clusters))
	}

	for i, cluster := range clusters {
		index := i + 1
		expectedId := fmt.Sprintf("cluster-%d", index)
		expectedName := fmt.Sprintf("cluster-name-%d", index)
		if cluster.Id != expectedId {
			t.Fatalf("matching error cluster id, expected %s, but got %s", expectedId, cluster.Id)
		}
		if cluster.Name != expectedName {
			t.Fatalf("matching error cluster name, expected %s, but got %s", expectedName, cluster.Name)
		}
		if cluster.CreatedOn != int64(index) {
			t.Fatalf("matching error cluster createdOn, expected %d, but got %d", index, cluster.CreatedOn)
		}
		var cloudProvider CloudProvider
		if i == 2 {
			cloudProvider = AZURE
		} else {
			cloudProvider = AWS
		}
		if cluster.Provider != cloudProvider {
			t.Fatalf("matching error cluster provider, expected %s, but got %s", cloudProvider, cluster.Provider)
		}
	}
}

func TestNewClusterAWS(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters",
			ExpectRequestBody: `{
				"cloudProvider": "AWS",
				"cluster": {
					"name": "cluster-1",
					"version": "2.0",
					"sshKeyName": "ssh-key-1",
					"clusterConfiguration": {
						"head": {
							"instanceType": "node-type-1",
							"diskSize": 512,
							"haEnabled": false
						},
						"workers": [
							{
								"instanceType": "node-type-2",
								"diskSize": 256,
								"count": 2
							}
						]
					},
					"issueLetsEncrypt": true,
					"attachPublicIP": true,
					"backupRetentionPeriod": 10,
					"managedUsers": true,
					"tags": [
						{
							"name": "tag1",
							"value": "tag1-value1"
						}
					],
					"ronDB": {
						"allInOne": false,
						"configuration": {
							"ndbdDefault": {
								"replicationFactor": 2
							},
							"general": {
								"benchmark": {
									"grantUserPrivileges": false
								}
							}
						},
						"mgmd": {
							"instanceType": "mgm-node-1",
							"diskSize": 30,
							"count": 1
						},
						"ndbd": {
							"instanceType": "data-node-1",
							"diskSize": 512,
							"count": 2
						},
						"mysqld": {
							"instanceType": "mysqld-node-1",
							"diskSize": 100,
							"count": 1
						},
						"api": {
							"instanceType": "api-node-1",
							"diskSize": 50,
							"count": 1
						}
					},
					"initScript": "",
					"runInitScriptFirst": false,
					"deactivateLogReport": false,
					"collectLogs": false,
					"clusterDomainPrefix": "my-prefix",
					"customHostedZone": "custom.zone.ai",
					"region": "region-1",
					"bucketName": "bucket-1",
					"instanceProfileArn": "profile-1",
					"headInstanceProfileArn": "profile-2",
					"vpcId": "vpc-1",
					"subnetId": "subnet-1",
					"securityGroupId": "security-group-1",
					"eksClusterName": "eks-cluster-1",
					"ecrRegistryAccountId": "ecr-account-1"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload":{
					"id" : "new-cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	input := CreateAWSCluster{
		CreateCluster: CreateCluster{
			Name:       "cluster-1",
			Version:    "2.0",
			SshKeyName: "ssh-key-1",
			ClusterConfiguration: ClusterConfiguration{
				Head: HeadConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
				},
				Workers: []WorkerConfiguration{
					{
						NodeConfiguration: NodeConfiguration{
							InstanceType: "node-type-2",
							DiskSize:     256,
						},
						Count: 2,
					},
				},
			},
			IssueLetsEncrypt:      true,
			AttachPublicIP:        true,
			BackupRetentionPeriod: 10,
			ManagedUsers:          true,
			Tags: []ClusterTag{
				{
					Name:  "tag1",
					Value: "tag1-value1",
				},
			},
			ClusterDomainPrefix: "my-prefix",
			CustomHostedZone:    "custom.zone.ai",
			RonDB: &RonDBConfiguration{
				Configuration: RonDBBaseConfiguration{
					NdbdDefault: RonDBNdbdDefaultConfiguration{
						ReplicationFactor: 2,
					},
					General: RonDBGeneralConfiguration{
						Benchmark: RonDBBenchmarkConfiguration{
							GrantUserPrivileges: false,
						},
					},
				},
				ManagementNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "mgm-node-1",
						DiskSize:     30,
					},
					Count: 1,
				},
				DataNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "data-node-1",
						DiskSize:     512,
					},
					Count: 2,
				},
				MYSQLNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "mysqld-node-1",
						DiskSize:     100,
					},
					Count: 1,
				},
				APINodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "api-node-1",
						DiskSize:     50,
					},
					Count: 1,
				},
			},
		},
		AWSCluster: AWSCluster{
			Region:                 "region-1",
			BucketName:             "bucket-1",
			InstanceProfileArn:     "profile-1",
			HeadInstanceProfileArn: "profile-2",
			VpcId:                  "vpc-1",
			SubnetId:               "subnet-1",
			SecurityGroupId:        "security-group-1",
			EksClusterName:         "eks-cluster-1",
			EcrRegistryAccountId:   "ecr-account-1",
		},
	}

	clusterId, err := NewCluster(context.TODO(), apiClient, input)
	if err != nil {
		t.Fatalf("new cluster should not throw error, but got %s", err)
	}

	if clusterId != "new-cluster-id-1" {
		t.Fatalf("new cluster should return the new cluster id, expected: new-cluster-id-1, got %s", clusterId)
	}
}

func TestNewClusterAZURE(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters",
			ExpectRequestBody: `{
				"cloudProvider": "AZURE",
				"cluster": {
					"name": "cluster-1",
					"version": "2.0",
					"sshKeyName": "ssh-key-1",
					"clusterConfiguration": {
						"head": {
							"instanceType": "node-type-1",
							"diskSize": 512,
							"haEnabled": false
						},
						"workers": [
							{
								"instanceType": "node-type-2",
								"diskSize": 256,
								"count": 2
							}
						]
					},
					"issueLetsEncrypt": true,
					"attachPublicIP": true,
					"backupRetentionPeriod": 10,
					"managedUsers": true,
					"tags": [
						{
							"name": "tag1",
							"value": "tag1-value1"
						}
					],
					"ronDB": {
						"allInOne": false,
						"configuration": {
							"ndbdDefault": {
								"replicationFactor": 2
							},
							"general": {
								"benchmark": {
									"grantUserPrivileges": false
								}
							}
						},
						"mgmd": {
							"instanceType": "mgm-node-1",
							"diskSize": 30,
							"count": 1
						},
						"ndbd": {
							"instanceType": "data-node-1",
							"diskSize": 512,
							"count": 2
						},
						"mysqld": {
							"instanceType": "mysqld-node-1",
							"diskSize": 100,
							"count": 1
						},
						"api": {
							"instanceType": "api-node-1",
							"diskSize": 50,
							"count": 1
						}
					},
					"initScript": "",
					"runInitScriptFirst": false,
					"deactivateLogReport": false,
					"collectLogs": false,
					"location": "location-1",
					"managedIdentity": "profile-1",
					"resourceGroup": "resource-group-1",
					"blobContainerName": "container-1",
					"storageAccount": "account-1",
					"networkResourceGroup": "network-resource-group-1",
					"virtualNetworkName": "network-1",
					"subnetName": "subnet-1",
					"securityGroupName": "security-group-1",
					"aksClusterName": "aks-cluster-1",
					"acrRegistryName": "acr-registry-1",
					"searchDomain": "internal.cloudapp.net"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload":{
					"id" : "new-cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	input := CreateAzureCluster{
		CreateCluster: CreateCluster{
			Name:       "cluster-1",
			Version:    "2.0",
			SshKeyName: "ssh-key-1",
			ClusterConfiguration: ClusterConfiguration{
				Head: HeadConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
				},
				Workers: []WorkerConfiguration{
					{
						NodeConfiguration: NodeConfiguration{
							InstanceType: "node-type-2",
							DiskSize:     256,
						},
						Count: 2,
					},
				},
			},
			IssueLetsEncrypt:      true,
			AttachPublicIP:        true,
			BackupRetentionPeriod: 10,
			ManagedUsers:          true,
			Tags: []ClusterTag{
				{
					Name:  "tag1",
					Value: "tag1-value1",
				},
			},
			RonDB: &RonDBConfiguration{
				Configuration: RonDBBaseConfiguration{
					NdbdDefault: RonDBNdbdDefaultConfiguration{
						ReplicationFactor: 2,
					},
					General: RonDBGeneralConfiguration{
						Benchmark: RonDBBenchmarkConfiguration{
							GrantUserPrivileges: false,
						},
					},
				},
				ManagementNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "mgm-node-1",
						DiskSize:     30,
					},
					Count: 1,
				},
				DataNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "data-node-1",
						DiskSize:     512,
					},
					Count: 2,
				},
				MYSQLNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "mysqld-node-1",
						DiskSize:     100,
					},
					Count: 1,
				},
				APINodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "api-node-1",
						DiskSize:     50,
					},
					Count: 1,
				},
			},
		},
		AzureCluster: AzureCluster{
			Location:             "location-1",
			ManagedIdentity:      "profile-1",
			ResourceGroup:        "resource-group-1",
			BlobContainerName:    "container-1",
			StorageAccount:       "account-1",
			NetworkResourceGroup: "network-resource-group-1",
			VirtualNetworkName:   "network-1",
			SubnetName:           "subnet-1",
			SecurityGroupName:    "security-group-1",
			AksClusterName:       "aks-cluster-1",
			AcrRegistryName:      "acr-registry-1",
			SearchDomain:         "internal.cloudapp.net",
		},
	}

	clusterId, err := NewCluster(context.TODO(), apiClient, input)
	if err != nil {
		t.Fatalf("new cluster should not throw error, but got %s", err)
	}

	if clusterId != "new-cluster-id-1" {
		t.Fatalf("new cluster should return the new cluster id, expected: new-cluster-id-1, got %s", clusterId)
	}
}

func TestNewClusterInvalidCloud(t *testing.T) {
	clusterId, err := NewCluster(context.TODO(), nil, struct{}{})
	if err == nil {
		t.Fatal("new cluster should throw an error if unknown request")
	}
	if err.Error() != "unknown cloud provider struct {}{}" {
		t.Fatalf("new cluster should throw an unknown cloud provider error, but got %s", err)
	}
	if clusterId != "" {
		t.Fatalf("new cluster should return empty id if encountered an error, but got %s", clusterId)
	}
}

func TestDeleteCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodDelete,
			ExpectPath:   "/api/clusters/cluster-id-1",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}
	err := DeleteCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("delete cluster should not throw an error, but got %s", err)
	}
}

func TestStopCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPut,
			ExpectPath:   "/api/clusters/cluster-id-1/stop",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}
	err := StopCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("stop cluster should not throw an error, but got %s", err)
	}
}

func TestStartCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPut,
			ExpectPath:   "/api/clusters/cluster-id-1/start",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}
	err := StartCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("start cluster should not throw an error, but got %s", err)
	}
}

func testAddWorkers(t *testing.T, expectedReqBody string, toAdd []WorkerConfiguration) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:      http.MethodPost,
			ExpectPath:        "/api/clusters/cluster-id-1/workers",
			ExpectRequestBody: expectedReqBody,
			ResponseCode:      http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	err := AddWorkers(context.TODO(), apiClient, "cluster-id-1", toAdd)
	if err != nil {
		t.Fatalf("update cluster should not throw an error, but got %s", err)
	}
}

func testRemoveWorkers(t *testing.T, expectedReqBody string, toRemove []WorkerConfiguration) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:      http.MethodDelete,
			ExpectPath:        "/api/clusters/cluster-id-1/workers",
			ExpectRequestBody: expectedReqBody,
			ResponseCode:      http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	err := RemoveWorkers(context.TODO(), apiClient, "cluster-id-1", toRemove)
	if err != nil {
		t.Fatalf("update cluster should not throw an error, but got %s", err)
	}
}

func TestUpdateClusterWorkers(t *testing.T) {
	testAddWorkers(t, `{
		"workers":[
			{
				"instanceType": "node-type-1",
				"diskSize": 256,
				"count": 2
			},
			{
				"instanceType": "node-type-1",
				"diskSize": 1024,
				"count": 3
			}
		]
	}`,
		[]WorkerConfiguration{
			{
				NodeConfiguration: NodeConfiguration{
					InstanceType: "node-type-1",
					DiskSize:     256,
				},
				Count: 2,
			},
			{
				NodeConfiguration: NodeConfiguration{
					InstanceType: "node-type-1",
					DiskSize:     1024,
				},
				Count: 3,
			},
		})

	testRemoveWorkers(t, `{
			"workers":[
				{
					"instanceType": "node-type-2",
					"diskSize": 512,
					"count": 1
				}
			]
		}`,
		[]WorkerConfiguration{
			{
				NodeConfiguration: NodeConfiguration{
					InstanceType: "node-type-2",
					DiskSize:     512,
				},
				Count: 1,
			},
		})
}

func TestUpdateClusterWorkersSkip(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			FailWithError: "update cluster should not do http request if no updates",
			T:             t,
		},
	}

	if err := AddWorkers(context.TODO(), apiClient, "cluster-id-1", []WorkerConfiguration{}); err != nil {
		t.Fatalf("update cluster should not throw an error, but got %s", err)
	}

	if err := RemoveWorkers(context.TODO(), apiClient, "cluster-id-1", []WorkerConfiguration{}); err != nil {
		t.Fatalf("update cluster should not throw an error, but got %s", err)
	}
}

func testUpdatePorts(t *testing.T, expectedReqBody string, ports *ServiceOpenPorts) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:      http.MethodPost,
			ExpectPath:        "/api/clusters/cluster-id-1/ports",
			ExpectRequestBody: expectedReqBody,
			ResponseCode:      http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	err := UpdateOpenPorts(context.TODO(), apiClient, "cluster-id-1", ports)
	if err != nil {
		t.Fatalf("update cluster should not throw an error, but got %s", err)
	}
}

func TestUpdatePorts(t *testing.T) {
	testUpdatePorts(t, `{
		"ports":{
			"featureStore": true,
			"onlineFeatureStore": false,
			"kafka": true,
			"ssh": false
		}
	}`, &ServiceOpenPorts{
		FeatureStore:       true,
		OnlineFeatureStore: false,
		Kafka:              true,
		SSH:                false,
	})

	testUpdatePorts(t, `{
		"ports":{
			"featureStore": false,
			"onlineFeatureStore": true,
			"kafka": false,
			"ssh": true
		}
	}`, &ServiceOpenPorts{
		FeatureStore:       false,
		OnlineFeatureStore: true,
		Kafka:              false,
		SSH:                true,
	})

	testUpdatePorts(t, `{
		"ports":{
			"featureStore": false,
			"onlineFeatureStore": false,
			"kafka": false,
			"ssh": false
		}
	}`, &ServiceOpenPorts{
		FeatureStore:       false,
		OnlineFeatureStore: false,
		Kafka:              false,
		SSH:                false,
	})

	testUpdatePorts(t, `{
		"ports":{
			"featureStore": true,
			"onlineFeatureStore": true,
			"kafka": true,
			"ssh": true
		}
	}`, &ServiceOpenPorts{
		FeatureStore:       true,
		OnlineFeatureStore: true,
		Kafka:              true,
		SSH:                true,
	})
}

func testGetSupportedInstanceTypes(t *testing.T, cloud CloudProvider, region string) {
	var reqQuery = "cloud=" + cloud.String()
	if region != "" {
		reqQuery = reqQuery + "&region=" + region
	}
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:       http.MethodGet,
			ExpectPath:         "/api/clusters/nodes/supported-types",
			ExpectRequestQuery: reqQuery,
			ResponseCode:       http.StatusOK,
			ResponseBody: fmt.Sprintf(`{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"%s": {
						"head": [
							{
								"id": "head-type-1",
								"memory": 20,
								"cpus": 10,
								"gpus": 0
							},
							{
								"id": "head-type-2",
								"memory": 50,
								"cpus": 20,
								"gpus": 1
							}
						],
						"worker": [
							{
								"id": "worker-type-1",
								"memory": 20,
								"cpus": 10,
								"gpus": 0
							},
							{
								"id": "worker-type-2",
								"memory": 50,
								"cpus": 20,
								"gpus": 1
							}
						],
						"ronDB": {
							"mgmd": [
								{
									"id": "mgm-type-1",
									"memory": 30,
									"cpus": 2,
									"gpus": 0
								}
							],
							"ndbd": [
								{
									"id": "ndbd-type-1",
									"memory": 100,
									"cpus": 16,
									"gpus": 0
								}
							],
							"mysqld": [
								{
									"id": "mysql-type-1",
									"memory": 100,
									"cpus": 16,
									"gpus": 0
								}
							],
							"api": [
								{
									"id": "api-type-1",
									"memory": 100,
									"cpus": 16,
									"gpus": 0
								}
							]
						}
					}
				}
			}`, strings.ToLower(cloud.String())),
			T: t,
		},
	}

	output, err := GetSupportedInstanceTypes(context.TODO(), apiClient, cloud, region)
	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	expected := &SupportedInstanceTypes{
		Head: SupportedInstanceTypeList{
			{
				Id:     "head-type-1",
				Memory: 20,
				CPUs:   10,
				GPUs:   0,
			},
			{
				Id:     "head-type-2",
				Memory: 50,
				CPUs:   20,
				GPUs:   1,
			},
		},
		Worker: SupportedInstanceTypeList{
			{
				Id:     "worker-type-1",
				Memory: 20,
				CPUs:   10,
				GPUs:   0,
			},
			{
				Id:     "worker-type-2",
				Memory: 50,
				CPUs:   20,
				GPUs:   1,
			},
		},
		RonDB: SupportedRonDBInstanceTypes{
			ManagementNode: SupportedInstanceTypeList{
				{
					Id:     "mgm-type-1",
					Memory: 30,
					CPUs:   2,
					GPUs:   0,
				},
			},
			DataNode: SupportedInstanceTypeList{
				{
					Id:     "ndbd-type-1",
					Memory: 100,
					CPUs:   16,
					GPUs:   0,
				},
			},
			MySQLNode: SupportedInstanceTypeList{
				{
					Id:     "mysql-type-1",
					Memory: 100,
					CPUs:   16,
					GPUs:   0,
				},
			},
			APINode: SupportedInstanceTypeList{
				{
					Id:     "api-type-1",
					Memory: 100,
					CPUs:   16,
					GPUs:   0,
				},
			},
		},
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching [%s] :\nexpected %#v \nbut got %#v", cloud.String(), expected, output)
	}
}

func TestGetSupportedInstanceTypes(t *testing.T) {
	testGetSupportedInstanceTypes(t, AWS, "")
	testGetSupportedInstanceTypes(t, AWS, "region1")
	testGetSupportedInstanceTypes(t, AZURE, "")
	testGetSupportedInstanceTypes(t, AZURE, "region1")
}

func TestGetSupportedInstanceTypes_unknownProvider(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/clusters/nodes/supported-types",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
				}
			}`,
			T: t,
		},
	}

	output, err := GetSupportedInstanceTypes(context.TODO(), apiClient, "test", "")
	if err == nil || err.Error() != "unknown cloud provider test" {
		t.Fatalf("should throw an error, but got %s", err)
	}
	if output != nil {
		t.Fatalf("expected a nil output, but got %#v", output)
	}
}

func testConfigureAutoscale(t *testing.T, reqBody string, config *AutoscaleConfiguration) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:      http.MethodPost,
			ExpectPath:        "/api/clusters/cluster-id-1/autoscale",
			ExpectRequestBody: reqBody,
			ResponseCode:      http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	if err := ConfigureAutoscale(context.TODO(), apiClient, "cluster-id-1", config); err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}
}

func TestConfigureAutoscale(t *testing.T) {
	testConfigureAutoscale(t, `
	{
		"autoscale": 
		{
			"nonGpu": 
			{
				"instanceType": "non-gpu-node",
				"diskSize": 256,
				"minWorkers": 0,
				"maxWorkers": 10,
				"standbyWorkers": 0.5,
				"downscaleWaitTime": 300
			},
			"gpu":
			{
				"instanceType": "gpu-node",
				"diskSize": 512,
				"minWorkers": 1,
				"maxWorkers": 5,
				"standbyWorkers": 0.4,
				"downscaleWaitTime": 200
			}
		}
	}`,
		&AutoscaleConfiguration{
			NonGPU: &AutoscaleConfigurationBase{
				InstanceType:      "non-gpu-node",
				DiskSize:          256,
				MinWorkers:        0,
				MaxWorkers:        10,
				StandbyWorkers:    0.5,
				DownscaleWaitTime: 300,
			},
			GPU: &AutoscaleConfigurationBase{
				InstanceType:      "gpu-node",
				DiskSize:          512,
				MinWorkers:        1,
				MaxWorkers:        5,
				StandbyWorkers:    0.4,
				DownscaleWaitTime: 200,
			},
		})

	testConfigureAutoscale(t, `
	{
		"autoscale":{
			"nonGpu":{
				"instanceType": "non-gpu-node",
				"diskSize": 256,
				"minWorkers": 0,
				"maxWorkers": 10,
				"standbyWorkers": 0.5,
				"downscaleWaitTime": 300
			}
		}
	}
	`,
		&AutoscaleConfiguration{
			NonGPU: &AutoscaleConfigurationBase{
				InstanceType:      "non-gpu-node",
				DiskSize:          256,
				MinWorkers:        0,
				MaxWorkers:        10,
				StandbyWorkers:    0.5,
				DownscaleWaitTime: 300,
			},
		})

	testConfigureAutoscale(t, `
		{
			"autoscale":{
				"gpu":{
					"instanceType": "gpu-node",
					"diskSize": 512,
					"minWorkers": 1,
					"maxWorkers": 5,
					"standbyWorkers": 0.4,
					"downscaleWaitTime": 200
				}
			}
		}
		`,
		&AutoscaleConfiguration{
			GPU: &AutoscaleConfigurationBase{
				InstanceType:      "gpu-node",
				DiskSize:          512,
				MinWorkers:        1,
				MaxWorkers:        5,
				StandbyWorkers:    0.4,
				DownscaleWaitTime: 200,
			},
		})
}

func TestDisableAutoscale(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodDelete,
			ExpectPath:   "/api/clusters/cluster-id-1/autoscale",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	if err := DisableAutoscale(context.TODO(), apiClient, "cluster-id-1"); err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}
}

func TestDisableAutoscale_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodDelete,
			ExpectPath:   "/api/clusters/cluster-id-1/autoscale",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "bad request",
				"code": 400,
				"message": "failed to disable"
			}`,
			T: t,
		},
	}

	if err := DisableAutoscale(context.TODO(), apiClient, "cluster-id-1"); err == nil || err.Error() != "failed to disable" {
		t.Fatalf("should throw an error, but got %s", err)
	}
}

func TestNewBackup(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/backups",
			ExpectRequestBody: `{
				"backup": {
					"clusterId": "cluster-id-1",
					"backupName": "my-new-backup"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"backupId": "new-backup-1"
				}
			}`,
			T: t,
		},
	}

	id, err := NewBackup(context.TODO(), apiClient, "cluster-id-1", "my-new-backup")

	if id != "new-backup-1" {
		t.Fatalf("expected id new-backup-1 but got %s", id)
	}

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}
}

func TestNewBackup_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/backups",
			ExpectRequestBody: `{
				"backup": {
					"clusterId": "cluster-id-1",
					"backupName": "my-new-backup"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"backupId": "new-backup-1"
				},
			}`,
			T: t,
		},
	}

	id, err := NewBackup(context.TODO(), apiClient, "cluster-id-1", "my-new-backup")

	if err == nil || !strings.HasPrefix(err.Error(), "failed to decode json") {
		t.Fatalf("should throw an error, but got %s", err)
	}

	if id != "" {
		t.Fatalf("expected empty id but got %s", id)
	}
}

func TestGetBackup(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/backups/backup-id-1",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"backup": {
						"backupId": "backup-id-1",
						"backupName": "backup-name",
						"clusterId": "cluster-id-1",
						"cloudProvider": "AWS",
						"createdOn": 1,
						"state": "succeed",
						"stateMessage": "message"
					}
				}
			}`,
			T: t,
		},
	}

	expected := &Backup{
		Id:            "backup-id-1",
		Name:          "backup-name",
		ClusterId:     "cluster-id-1",
		CreatedOn:     1,
		CloudProvider: AWS,
		State:         BackupSucceed,
		StateMessage:  "message",
	}

	output, err := GetBackup(context.TODO(), apiClient, "backup-id-1")

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestGetBackup_notfound(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/backups/backup-id-1",
			ResponseCode: http.StatusNotFound,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "Not found",
				"code": 404
			}`,
			T: t,
		},
	}

	output, err := GetBackup(context.TODO(), apiClient, "backup-id-1")

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if output != nil {
		t.Fatalf("expected nil for not found backup but got %#v", output)
	}
}

func TestGetBackup_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/backups/backup-id-1",
			ResponseCode: http.StatusNotFound,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "Not found",
				"code": 404,
			}`,
			T: t,
		},
	}

	output, err := GetBackup(context.TODO(), apiClient, "backup-id-1")

	if err == nil || !strings.HasPrefix(err.Error(), "failed to decode json") {
		t.Fatalf("should throw an error, but got %s", err)
	}

	if output != nil {
		t.Fatalf("expected nil when encountering error during get backup but got %#v", output)
	}
}

func TestDeleteBackup(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodDelete,
			ExpectPath:   "/api/backups/backup-id-1",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	if err := DeleteBackup(context.TODO(), apiClient, "backup-id-1"); err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}
}

func TestDeleteBackup_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodDelete,
			ExpectPath:   "/api/backups/backup-id-1",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
			}`,
			T: t,
		},
	}

	if err := DeleteBackup(context.TODO(), apiClient, "backup-id-1"); err == nil || !strings.HasPrefix(err.Error(), "failed to decode json") {
		t.Fatalf("should throw an error, but got %s", err)
	}
}

func TestGetBackups(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:       http.MethodGet,
			ExpectPath:         "/api/backups",
			ExpectRequestQuery: "clusterId=cluster-id-1",
			ResponseCode:       http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"backups": [
						{
							"backupId": "backup-id-1",
							"backupName": "backup-name",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 1,
							"state": "succeed",
							"stateMessage": "message"
						},
						{
							"backupId": "backup-id-2",
							"backupName": "backup-name-2",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 10,
							"state": "failed",
							"stateMessage": "failure message"
						}
					]
				}
			}`,
			T: t,
		},
	}

	expected := []Backup{
		{
			Id:            "backup-id-1",
			Name:          "backup-name",
			ClusterId:     "cluster-id-1",
			CreatedOn:     1,
			CloudProvider: AWS,
			State:         BackupSucceed,
			StateMessage:  "message",
		},
		{
			Id:            "backup-id-2",
			Name:          "backup-name-2",
			ClusterId:     "cluster-id-1",
			CreatedOn:     10,
			CloudProvider: AWS,
			State:         BackupFailed,
			StateMessage:  "failure message",
		},
	}

	output, err := GetBackups(context.TODO(), apiClient, "cluster-id-1")

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestGetBackups_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod:       http.MethodGet,
			ExpectPath:         "/api/backups",
			ExpectRequestQuery: "clusterId=cluster-id-1",
			ResponseCode:       http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"backups": [
						{
							"backupId": "backup-id-1",
							"backupName": "backup-name",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 1,
							"state": "succeed",
							"stateMessage": "message"
						},
						{
							"backupId": "backup-id-2",
							"backupName": "backup-name-2",
							"clusterId": "cluster-id-1",
							"cloudProvider": "AWS",
							"createdOn": 10,
							"state": "failed",
							"stateMessage": "failure message"
						},
					]
				}
			}`,
			T: t,
		},
	}

	output, err := GetBackups(context.TODO(), apiClient, "cluster-id-1")

	if err == nil || !strings.HasPrefix(err.Error(), "failed to decode json") {
		t.Fatalf("should throw an error, but got %s", err)
	}

	if output != nil {
		t.Fatalf("expected nil when encountering an error during get backups, but got %#v", output)
	}
}

func TestNewClusterFromBackup(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/restore/backup-id-1",
			ExpectRequestBody: `{
				"cluster":{

				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"id": "cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	id, err := NewClusterFromBackup(context.TODO(), apiClient, "backup-id-1", CreateAWSClusterFromBackup{})
	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if id != "cluster-id-1" {
		t.Fatalf("expected cluster id (cluster-id-1), but got %s", id)
	}

	id, err = NewClusterFromBackup(context.TODO(), apiClient, "backup-id-1", CreateAzureClusterFromBackup{})
	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if id != "cluster-id-1" {
		t.Fatalf("expected cluster id (cluster-id-1), but got %s", id)
	}
}

func TestNewClusterFromBackup_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/restore/backup-id-1",
			ExpectRequestBody: `{
				"cluster":{

				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"id": "cluster-id-1"
				},
			}`,
			T: t,
		},
	}

	id, err := NewClusterFromBackup(context.TODO(), apiClient, "backup-id-1", CreateAWSClusterFromBackup{})
	if err == nil || !strings.HasPrefix(err.Error(), "failed to decode json") {
		t.Fatalf("should throw an error, but got %s", err)
	}

	if id != "" {
		t.Fatalf("expected empty cluster id when encountering an error, but got %s", id)
	}
}

func TestNewClusterFromBackup_AWS_changeConfig(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/restore/backup-id-1",
			ExpectRequestBody: `{
				"cluster":{
					"name": "new-cluster-name",
					"sshKeyName": "new-ssh-key",
					"tags": [
						{
							"name": "tag1",
							"value": "tag1-value"
						}
					],
					"autoscale":{
						"nonGpu": {
							"instanceType": "new-node-type",
							"diskSize": 512,
							"minWorkers": 1,
							"maxWorkers": 10,
							"standbyWorkers": 0.7,
							"downscaleWaitTime": 500,
							"spotInfo": {
								"maxPrice": 100,
								"fallBackOnDemand": false
							}
						}
					},
					"instanceProfileArn": "new-profile",
					"vpcId": "new-vpc",
					"subnetId": "new-subnet",
					"securityGroupId": "new-security-group"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"id": "cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	id, err := NewClusterFromBackup(context.TODO(), apiClient, "backup-id-1", CreateAWSClusterFromBackup{
		CreateClusterFromBackup: CreateClusterFromBackup{
			Name:       "new-cluster-name",
			SshKeyName: "new-ssh-key",
			Tags: []ClusterTag{
				{
					Name:  "tag1",
					Value: "tag1-value",
				},
			},
			Autoscale: &AutoscaleConfiguration{
				NonGPU: &AutoscaleConfigurationBase{
					InstanceType:      "new-node-type",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        10,
					StandbyWorkers:    0.7,
					DownscaleWaitTime: 500,
					SpotInfo: &SpotConfiguration{
						MaxPrice:         100,
						FallBackOnDemand: false,
					},
				},
			},
		},
		InstanceProfileArn: "new-profile",
		VpcId:              "new-vpc",
		SubnetId:           "new-subnet",
		SecurityGroupId:    "new-security-group",
	})

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if id != "cluster-id-1" {
		t.Fatalf("expected cluster id (cluster-id-1), but got %s", id)
	}
}

func TestNewClusterFromBackup_AZURE_changeConfig(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/restore/backup-id-1",
			ExpectRequestBody: `{
				"cluster":{
					"name": "new-cluster-name",
					"sshKeyName": "new-ssh-key",
					"tags": [
						{
							"name": "tag1",
							"value": "tag1-value"
						}
					],
					"autoscale":{
						"nonGpu": {
							"instanceType": "new-node-type",
							"diskSize": 512,
							"minWorkers": 1,
							"maxWorkers": 10,
							"standbyWorkers": 0.7,
							"downscaleWaitTime": 500,
							"spotInfo": {
								"maxPrice": 100,
								"fallBackOnDemand": false
							}
						}
					},
					"networkResourceGroup": "new-resource-group",
					"virtualNetworkName": "new-virtual-network",
					"subnetName": "new-subnet",
					"securityGroupName": "new-security-group"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"id": "cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	id, err := NewClusterFromBackup(context.TODO(), apiClient, "backup-id-1", CreateAzureClusterFromBackup{
		CreateClusterFromBackup: CreateClusterFromBackup{
			Name:       "new-cluster-name",
			SshKeyName: "new-ssh-key",
			Tags: []ClusterTag{
				{
					Name:  "tag1",
					Value: "tag1-value",
				},
			},
			Autoscale: &AutoscaleConfiguration{
				NonGPU: &AutoscaleConfigurationBase{
					InstanceType:      "new-node-type",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        10,
					StandbyWorkers:    0.7,
					DownscaleWaitTime: 500,
					SpotInfo: &SpotConfiguration{
						MaxPrice:         100,
						FallBackOnDemand: false,
					},
				},
			},
		},
		NetworkResourceGroup: "new-resource-group",
		VirtualNetworkName:   "new-virtual-network",
		SubnetName:           "new-subnet",
		SecurityGroupName:    "new-security-group",
	})

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if id != "cluster-id-1" {
		t.Fatalf("expected cluster id (cluster-id-1), but got %s", id)
	}
}

func TestNewClusterFromBackup_unknownCloud(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/restore/backup-id-1",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload": {
					"id": "cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	id, err := NewClusterFromBackup(context.TODO(), apiClient, "backup-id-1", nil)

	if err == nil || err.Error() != "unknown create request #<nil>" {
		t.Fatalf("should throw unknown request error, but got %s", err)
	}

	if id != "" {
		t.Fatalf("expected empty cluster-id, but got %s", id)
	}
}

func TestGetSupportedVersions_AWS(t *testing.T) {
	testGetSupportedVersions(t, AWS)
}

func TestGetSupportedVersions_AZURE(t *testing.T) {
	testGetSupportedVersions(t, AZURE)
}

func testGetSupportedVersions(t *testing.T, cloud CloudProvider) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/clusters/hopsworks/versions/" + cloud.String(),
			ResponseCode: http.StatusOK,
			ResponseBody: `{
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
							},
							"releaseNotesUrl": "notes-1"
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
							},
							"releaseNotesUrl": "notes-2"
						},
						{
							"version": "3.0",
							"upgradableFromVersion": "2.0",
							"default": true,
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
							},
							"releaseNotesUrl": "notes-3"
						}
					]
				}
			}`,
			T: t,
		},
	}

	expected := []SupportedVersion{
		{
			Version:               "1.0",
			UpgradableFromVersion: "N/A",
			Default:               false,
			Experimental:          true,
			Regions: SupportedVersionRegions{
				Ubuntu: []string{
					"region-5",
					"region-6",
				},
			},
			ReleaseNotesUrl: "notes-1",
		},
		{
			Version:               "2.0",
			UpgradableFromVersion: "1.0",
			Default:               false,
			Experimental:          false,
			Regions: SupportedVersionRegions{
				CentOS: []string{
					"region-1",
					"region-2",
				},
			},
			ReleaseNotesUrl: "notes-2",
		},
		{
			Version:               "3.0",
			UpgradableFromVersion: "2.0",
			Default:               true,
			Experimental:          false,
			Regions: SupportedVersionRegions{
				CentOS: []string{
					"region-1",
					"region-2",
				},
				Ubuntu: []string{
					"region-3",
					"region-4",
				},
			},
			ReleaseNotesUrl: "notes-3",
		},
	}
	output, err := GetSupportedVersions(context.TODO(), apiClient, cloud)

	if err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestGetSupportedVersions_AWS_error(t *testing.T) {
	testGetSupportedVersions_error(t, AWS)
}

func TestGetSupportedVersions_AZURE_error(t *testing.T) {
	testGetSupportedVersions_error(t, AZURE)
}

func testGetSupportedVersions_error(t *testing.T, cloud CloudProvider) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodGet,
			ExpectPath:   "/api/clusters/hopsworks/versions/" + cloud.String(),
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "error",
				"code": 400,
				"message": "no versions"
			}`,
			T: t,
		},
	}
	output, err := GetSupportedVersions(context.TODO(), apiClient, cloud)

	if err == nil || err.Error() != "no versions" {
		t.Fatalf("should throw an error [no versions], but got %s", err)
	}

	if output != nil {
		t.Fatalf("error while matching:\nexpected nil \nbut got %#v", output)
	}
}

func TestUpgradeCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/cluster-id-1/upgrade",
			ExpectRequestBody: `{
				"version": "v2"
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	if err := UpgradeCluster(context.TODO(), apiClient, "cluster-id-1", "v2", ""); err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}
}

func TestUpgradeCluster_API_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters/cluster-id-1/upgrade",
			ExpectRequestBody: `{
				"version": "v2"
			}`,
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "error",
				"code": 400,
				"message": "failure to start upgrade"
			}`,
			T: t,
		},
	}

	if err := UpgradeCluster(context.TODO(), apiClient, "cluster-id-1", "v2", ""); err == nil || err.Error() != "failure to start upgrade" {
		t.Fatalf("should throw an error, but got %s", err)
	}
}

func TestRollbackUpgradeCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPut,
			ExpectPath:   "/api/clusters/cluster-id-1/upgrade/rollback",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	if err := RollbackUpgradeCluster(context.TODO(), apiClient, "cluster-id-1"); err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}
}

func TestRollbackUpgradeCluster_API_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPut,
			ExpectPath:   "/api/clusters/cluster-id-1/upgrade/rollback",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "error",
				"code": 400,
				"message": "failure to start rollback"
			}`,
			T: t,
		},
	}

	if err := RollbackUpgradeCluster(context.TODO(), apiClient, "cluster-id-1"); err == nil || err.Error() != "failure to start rollback" {
		t.Fatalf("should throw an error, but got %s", err)
	}
}

func TestModifyInstanceType(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPut,
			ExpectPath:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200
			}`,
			T: t,
		},
	}

	if err := ModifyInstanceType(context.TODO(), apiClient, "cluster-id-1", HeadNode, "type1"); err != nil {
		t.Fatalf("should not throw an error, but got %s", err)
	}

	for _, node := range []NodeType{WorkerNode, RonDBManagementNode} {
		if err := ModifyInstanceType(context.TODO(), apiClient, "cluster-id-1", node, "type1"); err == nil || err.Error() != fmt.Sprintf("modifying instance type for %s is not supported", node.String()) {
			t.Fatalf("should throw an error, but got %s", err)
		}
	}
}

func TestModifyInstanceType_error(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPut,
			ExpectPath:   "/api/clusters/cluster-id-1/nodes/modify-instance-type",
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "error",
				"code": 400,
				"message": "failed to modify instance type"
			}`,
			T: t,
		},
	}

	if err := ModifyInstanceType(context.TODO(), apiClient, "cluster-id-1", HeadNode, "type1"); err == nil || err.Error() != "failed to modify instance type" {
		t.Fatalf("should throw an error, but got %s", err)
	}
}

func TestNewClusterAWS_HA(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &test.HttpClientFixture{
			ExpectMethod: http.MethodPost,
			ExpectPath:   "/api/clusters",
			ExpectRequestBody: `{
				"cloudProvider": "AWS",
				"cluster": {
					"name": "cluster-1",
					"version": "2.0",
					"sshKeyName": "ssh-key-1",
					"clusterConfiguration": {
						"head": {
							"instanceType": "node-type-1",
							"diskSize": 512,
							"haEnabled": true
						},
						"workers": [
							{
								"instanceType": "node-type-2",
								"diskSize": 256,
								"count": 2
							}
						]
					},
					"issueLetsEncrypt": true,
					"attachPublicIP": true,
					"backupRetentionPeriod": 10,
					"managedUsers": true,
					"tags": [
						{
							"name": "tag1",
							"value": "tag1-value1"
						}
					],
					"ronDB": {
						"allInOne": false,
						"configuration": {
							"ndbdDefault": {
								"replicationFactor": 2
							},
							"general": {
								"benchmark": {
									"grantUserPrivileges": false
								}
							}
						},
						"mgmd": {
							"instanceType": "mgm-node-1",
							"diskSize": 30,
							"count": 1
						},
						"ndbd": {
							"instanceType": "data-node-1",
							"diskSize": 512,
							"count": 2
						},
						"mysqld": {
							"instanceType": "mysqld-node-1",
							"diskSize": 100,
							"count": 1
						},
						"api": {
							"instanceType": "api-node-1",
							"diskSize": 50,
							"count": 1
						}
					},
					"initScript": "",
					"runInitScriptFirst": false,
					"deactivateLogReport": false,
					"collectLogs": false,
					"region": "region-1",
					"bucketName": "bucket-1",
					"instanceProfileArn": "profile-1",
					"headInstanceProfileArn": "profile-2",
					"vpcId": "vpc-1",
					"subnetId": "subnet-1",
					"securityGroupId": "security-group-1",
					"eksClusterName": "eks-cluster-1",
					"ecrRegistryAccountId": "ecr-account-1"
				}
			}`,
			ResponseCode: http.StatusOK,
			ResponseBody: `{
				"apiVersion": "v1",
				"status": "ok",
				"code": 200,
				"payload":{
					"id" : "new-cluster-id-1"
				}
			}`,
			T: t,
		},
	}

	input := CreateAWSCluster{
		CreateCluster: CreateCluster{
			Name:       "cluster-1",
			Version:    "2.0",
			SshKeyName: "ssh-key-1",
			ClusterConfiguration: ClusterConfiguration{
				Head: HeadConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
					HAEnabled: true,
				},
				Workers: []WorkerConfiguration{
					{
						NodeConfiguration: NodeConfiguration{
							InstanceType: "node-type-2",
							DiskSize:     256,
						},
						Count: 2,
					},
				},
			},
			IssueLetsEncrypt:      true,
			AttachPublicIP:        true,
			BackupRetentionPeriod: 10,
			ManagedUsers:          true,
			Tags: []ClusterTag{
				{
					Name:  "tag1",
					Value: "tag1-value1",
				},
			},
			RonDB: &RonDBConfiguration{
				Configuration: RonDBBaseConfiguration{
					NdbdDefault: RonDBNdbdDefaultConfiguration{
						ReplicationFactor: 2,
					},
					General: RonDBGeneralConfiguration{
						Benchmark: RonDBBenchmarkConfiguration{
							GrantUserPrivileges: false,
						},
					},
				},
				ManagementNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "mgm-node-1",
						DiskSize:     30,
					},
					Count: 1,
				},
				DataNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "data-node-1",
						DiskSize:     512,
					},
					Count: 2,
				},
				MYSQLNodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "mysqld-node-1",
						DiskSize:     100,
					},
					Count: 1,
				},
				APINodes: RonDBNodeConfiguration{
					NodeConfiguration: NodeConfiguration{
						InstanceType: "api-node-1",
						DiskSize:     50,
					},
					Count: 1,
				},
			},
		},
		AWSCluster: AWSCluster{
			Region:                 "region-1",
			BucketName:             "bucket-1",
			InstanceProfileArn:     "profile-1",
			HeadInstanceProfileArn: "profile-2",
			VpcId:                  "vpc-1",
			SubnetId:               "subnet-1",
			SecurityGroupId:        "security-group-1",
			EksClusterName:         "eks-cluster-1",
			EcrRegistryAccountId:   "ecr-account-1",
		},
	}

	clusterId, err := NewCluster(context.TODO(), apiClient, input)
	if err != nil {
		t.Fatalf("new cluster should not throw error, but got %s", err)
	}

	if clusterId != "new-cluster-id-1" {
		t.Fatalf("new cluster should return the new cluster id, expected: new-cluster-id-1, got %s", clusterId)
	}
}

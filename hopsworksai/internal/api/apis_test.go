package api

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type mockHttpClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (c *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return c.doFunc(req)
}

func TestInvalidAPIKey(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(strings.NewReader("Unauthorized")),
				}
				return resp, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "latest",
						}`)),
				}
				return resp, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("x-api-key") != "my-api-key" {
					t.Fatal("api key should be passed as a header")
				}
				if req.Header.Get("hopsai-api-version") != "testV1" {
					t.Fatal("api version should be passed as header")
				}
				if req.Header.Get("Content-Type") != "application/json" {
					t.Fatal("content-type should be application/json")
				}
				if req.URL.Path != "/api/clusters/cluster-id-1" {
					t.Fatalf("invalid path for get cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodGet {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				respString := `{
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
								"securityGroupId": "security-group-1",
								"eksClusterName": "eks-cluster-1",
								"ecrRegistryAccountId": "ecr-account-1"
							}
						}
					}
				}`
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(respString)),
				}
				return resp, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("x-api-key") != "my-api-key" {
					t.Fatal("api key should be passed as a header")
				}
				if req.Header.Get("hopsai-api-version") != "testV1" {
					t.Fatal("api version should be passed as header")
				}
				if req.Header.Get("Content-Type") != "application/json" {
					t.Fatal("content-type should be application/json")
				}
				if req.URL.Path != "/api/clusters/cluster-id-1" {
					t.Fatalf("invalid path for get cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodGet {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				respString := `{
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
								"securityGroupName": "security-group-name-1",
								"aksClusterName": "aks-cluster-name-1",
								"acrRegistryName": "acr-registry-name-1"
							}
						}
					}
				}`
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(respString)),
				}
				return resp, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "latest",
						"status": "error",
						"code": 404
					}`)),
				}
				return resp, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "latest",
						"status": "error",
						"code": 400,
						"message": "bad request you cannot retrieve cluster info"
					}`)),
				}
				return resp, nil
			},
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
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				var expectedQuery string = ""
				if provider != "" {
					expectedQuery = "cloud=" + provider
				}
				if req.URL.RawQuery != expectedQuery {
					t.Fatalf("should get correct url with filter, expected %s, got %s", expectedQuery, req.URL.RawQuery)
				}
				if req.URL.Path != "/api/clusters" {
					t.Fatalf("invalid path for get clusters, got %s", req.URL.Path)
				}
				if req.Method != http.MethodGet {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				return &http.Response{StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"statue": "ok",
						"code": 200,
						"payload":{
							"clusters": []
						}
					 }`))}, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
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
					}`)),
				}, nil
			},
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

func cleanJSONString(jsonString string) string {
	for _, v := range []string{"\n", "\t", " "} {
		jsonString = strings.ReplaceAll(jsonString, v, "")
	}
	return jsonString
}

func TestNewClusterAWS(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				reqBody, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				if req.URL.Path != "/api/clusters" {
					t.Fatalf("invalid path for new cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodPost {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				reqBodyString := string(reqBody)
				expectedReqBody := `{
					"cloudProvider": "AWS",
					"cluster": {
						"name": "cluster-1",
						"version": "2.0",
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
						"region": "region-1",
						"bucketName": "bucket-1",
						"instanceProfileArn": "profile-1",
						"vpcId": "vpc-1",
						"subnetId": "subnet-1",
						"securityGroupId": "security-group-1",
						"eksClusterName": "eks-cluster-1",
						"ecrRegistryAccountId": "ecr-account-1"
					}
				}`

				expectedReqBody = cleanJSONString(expectedReqBody)

				if reqBodyString != expectedReqBody {
					t.Fatalf("error matching req body, expected:\n%s, but got:\n%s", expectedReqBody, reqBodyString)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200,
						"payload":{
							"id" : "new-cluster-id-1"
						}
					}`)),
				}, nil
			},
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
		},
		AWSCluster: AWSCluster{
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				reqBody, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				if req.URL.Path != "/api/clusters" {
					t.Fatalf("invalid path for new cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodPost {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				reqBodyString := string(reqBody)
				expectedReqBody := `{
					"cloudProvider": "AZURE",
					"cluster": {
						"name": "cluster-1",
						"version": "2.0",
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
						"location": "location-1",
						"managedIdentity": "profile-1",
						"resourceGroup": "resource-group-1",
						"blobContainerName": "container-1",
						"storageAccount": "account-1",
						"virtualNetworkName": "network-1",
						"subnetName": "subnet-1",
						"securityGroupName": "security-group-1",
						"aksClusterName": "aks-cluster-1",
						"acrRegistryName": "acr-registry-1"
					}
				}`

				expectedReqBody = cleanJSONString(expectedReqBody)

				if reqBodyString != expectedReqBody {
					t.Fatalf("error matching req body, expected:\n%s, but got:\n%s", expectedReqBody, reqBodyString)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200,
						"payload":{
							"id" : "new-cluster-id-1"
						}
					}`)),
				}, nil
			},
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
		},
		AzureCluster: AzureCluster{
			Location:           "location-1",
			ManagedIdentity:    "profile-1",
			ResourceGroup:      "resource-group-1",
			BlobContainerName:  "container-1",
			StorageAccount:     "account-1",
			VirtualNetworkName: "network-1",
			SubnetName:         "subnet-1",
			SecurityGroupName:  "security-group-1",
			AksClusterName:     "aks-cluster-1",
			AcrRegistryName:    "acr-registry-1",
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
	if err.Error() != "unknown cloud provider #{}" {
		t.Fatalf("new cluster should throw an unknown cloud provider error, but got %s", err)
	}
	if clusterId != "" {
		t.Fatalf("new cluster should return empty id if encountered an error, but got %s", clusterId)
	}
}

func TestDeleteCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/clusters/cluster-id-1" {
					t.Fatalf("invalid path for delete cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodDelete {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200
					}`)),
				}, nil
			},
		},
	}
	err := DeleteCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("delete cluster should not throw an error, but got %s", err)
	}
}

func TestStopCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/clusters/cluster-id-1/stop" {
					t.Fatalf("invalid path for stop cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodPut {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200
					}`)),
				}, nil
			},
		},
	}
	err := StopCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("stop cluster should not throw an error, but got %s", err)
	}
}

func TestStartCluster(t *testing.T) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/clusters/cluster-id-1/start" {
					t.Fatalf("invalid path for start cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodPut {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200
					}`)),
				}, nil
			},
		},
	}
	err := StartCluster(context.TODO(), apiClient, "cluster-id-1")
	if err != nil {
		t.Fatalf("start cluster should not throw an error, but got %s", err)
	}
}

func testAddWorkers(t *testing.T, expectedReqBody string, toAdd []WorkerConfiguration) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/clusters/cluster-id-1/workers" {
					t.Fatalf("invalid path for update cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodPost {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				reqBody, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				reqBodyString := string(reqBody)
				expectedReqBody = cleanJSONString(expectedReqBody)
				if reqBodyString != expectedReqBody {
					t.Fatalf("error matching req body, expected:\n%s, but got:\n%s", expectedReqBody, reqBodyString)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200
					}`)),
				}, nil
			},
		},
	}

	err := AddWorkers(context.TODO(), apiClient, "cluster-id-1", toAdd)
	if err != nil {
		t.Fatalf("update cluster should not throw an error, but got %s", err)
	}
}

func testRemoveWorkers(t *testing.T, expectedReqBody string, toRemove []WorkerConfiguration) {
	apiClient := &HopsworksAIClient{
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/clusters/cluster-id-1/workers" {
					t.Fatalf("invalid path for update cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodDelete {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				reqBody, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				reqBodyString := string(reqBody)
				expectedReqBody = cleanJSONString(expectedReqBody)
				if reqBodyString != expectedReqBody {
					t.Fatalf("error matching req body, expected:\n%s, but got:\n%s", expectedReqBody, reqBodyString)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200
					}`)),
				}, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				t.Fatal("update cluster should not do http request if no updates")
				return nil, nil
			},
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
		Client: &mockHttpClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/clusters/cluster-id-1/ports" {
					t.Fatalf("invalid path for update cluster, got %s", req.URL.Path)
				}
				if req.Method != http.MethodPost {
					t.Fatalf("invalid http method, got %s", req.Method)
				}
				reqBody, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				reqBodyString := string(reqBody)
				expectedReqBody = cleanJSONString(expectedReqBody)
				if reqBodyString != expectedReqBody {
					t.Fatalf("error matching req body, expected:\n%s, but got:\n%s", expectedReqBody, reqBodyString)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"apiVersion": "v1",
						"status": "ok",
						"code": 200
					}`)),
				}, nil
			},
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

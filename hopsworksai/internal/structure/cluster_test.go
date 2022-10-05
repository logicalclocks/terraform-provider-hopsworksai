package structure

import (
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
)

func TestFlattenHeadConfiguration(t *testing.T) {
	input := &api.HeadConfigurationStatus{
		HeadConfiguration: api.HeadConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "head-type-1",
				DiskSize:     512,
			},
			HAEnabled: false,
		},
		NodeId:    "head-node-id-1",
		PrivateIp: "ip",
	}

	expected := []map[string]interface{}{
		{
			"instance_type": input.InstanceType,
			"disk_size":     input.DiskSize,
			"node_id":       input.NodeId,
			"ha_enabled":    false,
			"private_ip":    input.PrivateIp,
		},
	}

	output := flattenHead(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenWorkersConfiguration(t *testing.T) {
	cases := []struct {
		input    []api.WorkerConfiguration
		expected *schema.Set
	}{
		{
			input: []api.WorkerConfiguration{
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
					Count: 2,
				},
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-2",
						DiskSize:     256,
					},
					Count: 3,
				},
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-3",
						DiskSize:     1024,
					},
					Count: 1,
				},
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-4",
						DiskSize:     1024,
					},
					Count: 1,
					SpotInfo: &api.SpotConfiguration{
						MaxPrice:         10,
						FallBackOnDemand: false,
					},
				},
			},
			expected: schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
					"count":         2,
				},
				map[string]interface{}{
					"instance_type": "node-type-2",
					"disk_size":     256,
					"count":         3,
				},
				map[string]interface{}{
					"instance_type": "node-type-3",
					"disk_size":     1024,
					"count":         1,
				},
				map[string]interface{}{
					"instance_type": "node-type-4",
					"disk_size":     1024,
					"count":         1,
					"spot_config": []interface{}{
						map[string]interface{}{
							"max_price_percent":   10,
							"fall_back_on_demand": false,
						},
					},
				},
			}),
		},
		{
			input: []api.WorkerConfiguration{
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     512,
					},
					Count:      2,
					PrivateIps: []string{"ip1", "ip2"},
				},
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-1",
						DiskSize:     256,
					},
					Count:      3,
					PrivateIps: []string{"ip1", "ip2", "ip3"},
				},
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "node-type-3",
						DiskSize:     1024,
					},
					Count:      1,
					PrivateIps: []string{"ip1"},
				},
			},
			expected: schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
					"count":         2,
					"private_ips":   []string{"ip1", "ip2"},
				},
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     256,
					"count":         3,
					"private_ips":   []string{"ip1", "ip2", "ip3"},
				},
				map[string]interface{}{
					"instance_type": "node-type-3",
					"disk_size":     1024,
					"count":         1,
					"private_ips":   []string{"ip1"},
				},
			}),
		},
	}

	for i, c := range cases {
		output := flattenWorkers(nil, c.input)
		if c.expected.Difference(output).Len() != 0 && output.Difference(c.expected).Len() != 0 {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected, output)
		}
	}
}

func TestFlattenAWSAttributes(t *testing.T) {
	input := &api.Cluster{
		Provider: api.AWS,
		AWS: api.AWSCluster{
			Region:               "region-1",
			InstanceProfileArn:   "instance-profile-1",
			BucketName:           "bucket-name-1",
			VpcId:                "vpc-id-1",
			SubnetId:             "subnet-id-1",
			SecurityGroupId:      "security-group-1",
			EksClusterName:       "eks-cluster-name-1",
			EcrRegistryAccountId: "ecr-registry-account-1",
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"region":               input.AWS.Region,
			"instance_profile_arn": input.AWS.InstanceProfileArn,
			"network": []map[string]interface{}{
				{
					"vpc_id":            input.AWS.VpcId,
					"subnet_id":         input.AWS.SubnetId,
					"security_group_id": input.AWS.SecurityGroupId,
				},
			},
			"eks_cluster_name":        input.AWS.EksClusterName,
			"ecr_registry_account_id": input.AWS.EcrRegistryAccountId,
			"bucket": []map[string]interface{}{
				{
					"name": input.AWS.BucketName,
				},
			},
			"ebs_encryption": []map[string]interface{}{},
		},
	}

	output := flattenAWSAttributes(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}

	input.Provider = ""
	if flattenAWSAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AWS)
	}

	input.Provider = api.AZURE
	if flattenAWSAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AWS)
	}

	input.Provider = "aws"
	if flattenAWSAttributes(input) != nil {
		t.Fatal("cloud provider should be always capital")
	}
}

func TestFlattenAzureAttributes(t *testing.T) {
	input := &api.Cluster{
		Provider: api.AZURE,
		Azure: api.AzureCluster{
			Location:             "location-1",
			ResourceGroup:        "resource-group-1",
			ManagedIdentity:      "managed-identity-1",
			BlobContainerName:    "blob-container-name-1",
			StorageAccount:       "storage-account-1",
			VirtualNetworkName:   "virtual-network-name-1",
			SubnetName:           "subnet-name-1",
			SecurityGroupName:    "security-group-name-1",
			AksClusterName:       "aks-cluster-name-1",
			AcrRegistryName:      "acr-registry-name-1",
			NetworkResourceGroup: "network-resource-group-1",
			SearchDomain:         "internal.cloudapp.net",
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"location":                       input.Azure.Location,
			"resource_group":                 input.Azure.ResourceGroup,
			"user_assigned_managed_identity": input.Azure.ManagedIdentity,
			"network": []map[string]interface{}{
				{
					"resource_group":       input.Azure.NetworkResourceGroup,
					"virtual_network_name": input.Azure.VirtualNetworkName,
					"subnet_name":          input.Azure.SubnetName,
					"security_group_name":  input.Azure.SecurityGroupName,
					"search_domain":        input.Azure.SearchDomain,
				},
			},
			"aks_cluster_name":  input.Azure.AksClusterName,
			"acr_registry_name": input.Azure.AcrRegistryName,
			"container": []map[string]interface{}{
				{
					"name":            input.Azure.BlobContainerName,
					"storage_account": input.Azure.StorageAccount,
				},
			},
		},
	}

	output := flattenAzureAttributes(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}

	input.Provider = ""
	if flattenAzureAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AZURE)
	}

	input.Provider = api.AWS
	if flattenAzureAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AZURE)
	}

	input.Provider = "azure"
	if flattenAzureAttributes(input) != nil {
		t.Fatal("cloud provider should be always capital")
	}
}

func TestFlattenCluster(t *testing.T) {
	input := &api.Cluster{
		Id:                  "cluster-id-1",
		Name:                "cluster",
		State:               "state-1",
		ActivationState:     "activation-state-1",
		InitializationStage: "initializtion-stage-1",
		CreatedOn:           1605374387069,
		StartedOn:           1605374388069,
		Version:             "cluster-version",
		URL:                 "cluster-url",
		Provider:            api.AWS,
		Tags: []api.ClusterTag{
			{
				Name:  "tag1",
				Value: "tagvalue1",
			},
		},
		SshKeyName:            "ssh-key-1",
		PublicIPAttached:      true,
		LetsEncryptIssued:     true,
		ManagedUsers:          true,
		BackupRetentionPeriod: 0,
		ClusterConfiguration: api.ClusterConfigurationStatus{
			Head: api.HeadConfigurationStatus{
				HeadConfiguration: api.HeadConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "head-node-type-1",
						DiskSize:     512,
					},
					HAEnabled: false,
				},
				NodeId: "head-node-id-1",
			},
			Workers: []api.WorkerConfiguration{
				{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "worker-node-type-1",
						DiskSize:     256,
					},
					Count: 1,
				},
			},
		},
		Ports: api.ServiceOpenPorts{
			FeatureStore:       true,
			OnlineFeatureStore: false,
			Kafka:              true,
			SSH:                false,
		},
		RonDB: &api.RonDBConfiguration{
			Configuration: api.RonDBBaseConfiguration{
				NdbdDefault: api.RonDBNdbdDefaultConfiguration{
					ReplicationFactor: 2,
				},
				General: api.RonDBGeneralConfiguration{
					Benchmark: api.RonDBBenchmarkConfiguration{
						GrantUserPrivileges: false,
					},
				},
			},
			ManagementNodes: api.RonDBNodeConfiguration{
				NodeConfiguration: api.NodeConfiguration{
					InstanceType: "mgm-node-1",
					DiskSize:     30,
				},
				Count: 1,
			},
			DataNodes: api.RonDBNodeConfiguration{
				NodeConfiguration: api.NodeConfiguration{
					InstanceType: "data-node-1",
					DiskSize:     512,
				},
				Count: 2,
			},
			MYSQLNodes: api.RonDBNodeConfiguration{
				NodeConfiguration: api.NodeConfiguration{
					InstanceType: "mysqld-node-1",
					DiskSize:     100,
				},
				Count: 1,
			},
			APINodes: api.RonDBNodeConfiguration{
				NodeConfiguration: api.NodeConfiguration{
					InstanceType: "api-node-1",
					DiskSize:     50,
				},
				Count: 1,
			},
		},
		Autoscale: &api.AutoscaleConfiguration{
			NonGPU: &api.AutoscaleConfigurationBase{
				InstanceType:      "auto-node-1",
				DiskSize:          256,
				MinWorkers:        0,
				MaxWorkers:        10,
				StandbyWorkers:    0.5,
				DownscaleWaitTime: 300,
			},
			GPU: &api.AutoscaleConfigurationBase{
				InstanceType:      "auto-gpu-node-1",
				DiskSize:          512,
				MinWorkers:        1,
				MaxWorkers:        5,
				StandbyWorkers:    0.4,
				DownscaleWaitTime: 200,
			},
		},
		InitScript:         "#!/usr/bin/env bash\nset -e\necho 'Hello World'",
		RunInitScriptFirst: true,
		OS:                 "centos",
		UpgradeInProgress: &api.UpgradeInProgress{
			From: "v1",
			To:   "v2",
		},
		DeactivateLogReport: false,
		CollectLogs:         false,
	}

	var emptyAttributes []interface{} = nil
	expected := map[string]interface{}{
		"cluster_id":                            input.Id,
		"name":                                  input.Name,
		"url":                                   input.URL,
		"state":                                 input.State,
		"activation_state":                      input.ActivationState,
		"creation_date":                         time.Unix(input.CreatedOn, 0).Format(time.RFC3339),
		"start_date":                            time.Unix(input.StartedOn, 0).Format(time.RFC3339),
		"version":                               input.Version,
		"ssh_key":                               input.SshKeyName,
		"head":                                  flattenHead(&input.ClusterConfiguration.Head),
		"issue_lets_encrypt_certificate":        input.LetsEncryptIssued,
		"attach_public_ip":                      input.PublicIPAttached,
		"managed_users":                         input.ManagedUsers,
		"backup_retention_period":               input.BackupRetentionPeriod,
		"update_state":                          "none",
		"workers":                               flattenWorkers(input.Autoscale, input.ClusterConfiguration.Workers),
		"aws_attributes":                        emptyAttributes,
		"azure_attributes":                      emptyAttributes,
		"open_ports":                            flattenPorts(&input.Ports),
		"tags":                                  flattenTags(input.Tags),
		"rondb":                                 flattenRonDB(input.RonDB),
		"autoscale":                             flattenAutoscaleConfiguration(input.Autoscale),
		"init_script":                           input.InitScript,
		"run_init_script_first":                 input.RunInitScriptFirst,
		"os":                                    input.OS,
		"upgrade_in_progress":                   flattenUpgradeInProgress(input.UpgradeInProgress),
		"deactivate_hopsworksai_log_collection": input.DeactivateLogReport,
		"collect_logs":                          input.CollectLogs,
		"cluster_domain_prefix":                 input.ClusterDomainPrefix,
		"custom_hosted_zone":                    input.CustomHostedZone,
	}

	for _, cloud := range []api.CloudProvider{api.AWS, api.AZURE} {
		input.Provider = cloud
		if cloud == api.AWS {
			input.AWS = api.AWSCluster{
				Region:               "region-1",
				InstanceProfileArn:   "instance-profile-1",
				BucketName:           "bucket-name-1",
				VpcId:                "vpc-id-1",
				SubnetId:             "subnet-id-1",
				SecurityGroupId:      "security-group-1",
				EksClusterName:       "eks-cluster-name-1",
				EcrRegistryAccountId: "ecr-registry-account-1",
			}
			input.Azure = api.AzureCluster{}
			expected["aws_attributes"] = flattenAWSAttributes(input)
			expected["azure_attributes"] = emptyAttributes
		} else if cloud == api.AZURE {
			input.Azure = api.AzureCluster{
				Location:           "location-1",
				ResourceGroup:      "resource-group-1",
				ManagedIdentity:    "managed-identity-1",
				BlobContainerName:  "blob-container-name-1",
				StorageAccount:     "storage-account-1",
				VirtualNetworkName: "virtual-network-name-1",
				SubnetName:         "subnet-name-1",
				SecurityGroupName:  "security-group-name-1",
				AksClusterName:     "aks-cluster-name-1",
				AcrRegistryName:    "acr-registry-name-1",
				SearchDomain:       "internal.cloudapp.net",
			}
			input.AWS = api.AWSCluster{}
			expected["aws_attributes"] = emptyAttributes
			expected["azure_attributes"] = flattenAzureAttributes(input)
		}

		output := FlattenCluster(input)
		if len(expected) != len(output) {
			t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
		}

		for k, v := range expected {
			if k == "workers" {
				expectedWorker, outputWorker := v.(*schema.Set), output[k].(*schema.Set)
				if expectedWorker.Difference(outputWorker).Len() != 0 && outputWorker.Difference(expectedWorker).Len() != 0 {
					t.Fatalf("error while matching workers:\nexpected %#v \nbut got %#v", expectedWorker, outputWorker)
				}
			} else if !reflect.DeepEqual(v, output[k]) {
				t.Fatalf("error while matching %s:\nexpected %#v \nbut got %#v for %s", k, v, output[k], cloud.String())
			}
		}
	}
}

func TestExpandNode(t *testing.T) {
	input := map[string]interface{}{
		"instance_type": "instance-type-1",
		"disk_size":     512,
	}

	expected := api.NodeConfiguration{
		InstanceType: "instance-type-1",
		DiskSize:     512,
	}

	output := ExpandNode(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestExpandWorker(t *testing.T) {
	input := map[string]interface{}{
		"instance_type": "instance-type-1",
		"disk_size":     512,
		"count":         2,
		"spot_config": []interface{}{
			map[string]interface{}{
				"max_price_percent":   100,
				"fall_back_on_demand": true,
			},
		},
	}

	expected := api.WorkerConfiguration{
		NodeConfiguration: api.NodeConfiguration{
			InstanceType: "instance-type-1",
			DiskSize:     512,
		},
		Count: 2,
		SpotInfo: &api.SpotConfiguration{
			MaxPrice:         100,
			FallBackOnDemand: true,
		},
	}

	output := ExpandWorker(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestExpandWorkers(t *testing.T) {
	input := schema.NewSet(helpers.WorkerSetHash, []interface{}{
		map[string]interface{}{
			"instance_type": "node-type-1",
			"disk_size":     512,
			"count":         2,
			"spot_config": []interface{}{
				map[string]interface{}{
					"max_price_percent":   100,
					"fall_back_on_demand": true,
				},
			},
		},
		map[string]interface{}{
			"instance_type": "node-type-1",
			"disk_size":     256,
			"count":         3,
		},
		map[string]interface{}{
			"instance_type": "node-type-3",
			"disk_size":     1024,
			"count":         1,
		},
	})

	expected := map[string]api.WorkerConfiguration{
		"node-type-1-512-100-true-": {
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-1",
				DiskSize:     512,
			},
			Count: 2,
			SpotInfo: &api.SpotConfiguration{
				MaxPrice:         100,
				FallBackOnDemand: true,
			},
		},
		"node-type-1-256-": {
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-1",
				DiskSize:     256,
			},
			Count: 3,
		},
		"node-type-3-1024-": {
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-3",
				DiskSize:     1024,
			},
			Count: 1,
		},
	}

	output := ExpandWorkers(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenOpenPorts(t *testing.T) {
	cases := []struct {
		input    *api.ServiceOpenPorts
		expected []map[string]interface{}
	}{
		{
			input: &api.ServiceOpenPorts{
				FeatureStore:       true,
				OnlineFeatureStore: false,
				Kafka:              true,
				SSH:                false,
			},
			expected: []map[string]interface{}{
				{
					"feature_store":        true,
					"online_feature_store": false,
					"kafka":                true,
					"ssh":                  false,
				},
			},
		},
		{
			input: &api.ServiceOpenPorts{
				FeatureStore:       false,
				OnlineFeatureStore: true,
				Kafka:              false,
				SSH:                true,
			},
			expected: []map[string]interface{}{
				{
					"feature_store":        false,
					"online_feature_store": true,
					"kafka":                false,
					"ssh":                  true,
				},
			},
		},
		{
			input: &api.ServiceOpenPorts{
				FeatureStore:       false,
				OnlineFeatureStore: false,
				Kafka:              false,
				SSH:                false,
			},
			expected: []map[string]interface{}{
				{
					"feature_store":        false,
					"online_feature_store": false,
					"kafka":                false,
					"ssh":                  false,
				},
			},
		},
		{
			input: &api.ServiceOpenPorts{
				FeatureStore:       true,
				OnlineFeatureStore: true,
				Kafka:              true,
				SSH:                true,
			},
			expected: []map[string]interface{}{
				{
					"feature_store":        true,
					"online_feature_store": true,
					"kafka":                true,
					"ssh":                  true,
				},
			},
		},
	}

	for i, c := range cases {
		output := flattenPorts(c.input)
		if !reflect.DeepEqual(c.expected, output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected, output)
		}
	}
}

func TestExpandOpenPorts(t *testing.T) {
	cases := []struct {
		input    map[string]interface{}
		expected api.ServiceOpenPorts
	}{
		{
			input: map[string]interface{}{
				"feature_store":        true,
				"online_feature_store": false,
				"kafka":                true,
				"ssh":                  false,
			},
			expected: api.ServiceOpenPorts{
				FeatureStore:       true,
				OnlineFeatureStore: false,
				Kafka:              true,
				SSH:                false,
			},
		},
		{
			input: map[string]interface{}{
				"feature_store":        false,
				"online_feature_store": true,
				"kafka":                false,
				"ssh":                  true,
			},
			expected: api.ServiceOpenPorts{
				FeatureStore:       false,
				OnlineFeatureStore: true,
				Kafka:              false,
				SSH:                true,
			},
		},
		{
			input: map[string]interface{}{
				"feature_store":        false,
				"online_feature_store": false,
				"kafka":                false,
				"ssh":                  false,
			},
			expected: api.ServiceOpenPorts{
				FeatureStore:       false,
				OnlineFeatureStore: false,
				Kafka:              false,
				SSH:                false,
			},
		},
		{
			input: map[string]interface{}{
				"feature_store":        true,
				"online_feature_store": true,
				"kafka":                true,
				"ssh":                  true,
			},
			expected: api.ServiceOpenPorts{
				FeatureStore:       true,
				OnlineFeatureStore: true,
				Kafka:              true,
				SSH:                true,
			},
		},
	}

	for i, c := range cases {
		output := ExpandPorts(c.input)
		if !reflect.DeepEqual(c.expected, output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected, output)
		}
	}
}

func TestFlattenTags(t *testing.T) {
	input := []api.ClusterTag{
		{
			Name:  "tag1",
			Value: "tag1-value",
		},
		{
			Name:  "tag2",
			Value: "tag2-value",
		},
	}

	expected := map[string]interface{}{
		"tag1": "tag1-value",
		"tag2": "tag2-value",
	}

	output := flattenTags(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenRonDB(t *testing.T) {
	input := &api.RonDBConfiguration{
		Configuration: api.RonDBBaseConfiguration{
			NdbdDefault: api.RonDBNdbdDefaultConfiguration{
				ReplicationFactor: 2,
			},
			General: api.RonDBGeneralConfiguration{
				Benchmark: api.RonDBBenchmarkConfiguration{
					GrantUserPrivileges: false,
				},
			},
		},
		ManagementNodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "mgm-node-1",
				DiskSize:     30,
			},
			Count: 1,
		},
		DataNodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "data-node-1",
				DiskSize:     512,
			},
			Count: 2,
		},
		MYSQLNodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "mysqld-node-1",
				DiskSize:     100,
			},
			Count: 1,
		},
		APINodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "api-node-1",
				DiskSize:     50,
			},
			Count: 1,
		},
	}

	expected := []map[string]interface{}{
		{
			"configuration": []interface{}{
				map[string]interface{}{
					"ndbd_default": []interface{}{
						map[string]interface{}{
							"replication_factor": 2,
						},
					},
					"general": []interface{}{
						map[string]interface{}{
							"benchmark": []interface{}{
								map[string]interface{}{
									"grant_user_privileges": false,
								},
							},
						},
					},
				},
			},
			"management_nodes": []interface{}{
				map[string]interface{}{
					"instance_type": "mgm-node-1",
					"disk_size":     30,
					"count":         1,
				},
			},
			"data_nodes": []interface{}{
				map[string]interface{}{
					"instance_type": "data-node-1",
					"disk_size":     512,
					"count":         2,
				},
			},
			"mysql_nodes": []interface{}{
				map[string]interface{}{
					"instance_type": "mysqld-node-1",
					"disk_size":     100,
					"count":         1,
				},
			},
			"api_nodes": []interface{}{
				map[string]interface{}{
					"instance_type": "api-node-1",
					"disk_size":     50,
					"count":         1,
				},
			},
			"single_node": nil,
		},
	}

	output := flattenRonDB(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenRonDB_nil(t *testing.T) {
	output := flattenRonDB(nil)
	if output != nil {
		t.Fatalf("error while matching:\nexpected nil \nbut got %#v", output)
	}
}

func TestFlattenWorkersConfiguration_autoscaleEnabled(t *testing.T) {
	input := []api.WorkerConfiguration{
		{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-1",
				DiskSize:     512,
			},
			Count: 2,
		},
		{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-2",
				DiskSize:     256,
			},
			Count: 3,
		},
		{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-3",
				DiskSize:     1024,
			},
			Count: 1,
		},
	}

	expected := schema.NewSet(helpers.WorkerSetHash, []interface{}{})

	output := flattenWorkers(&api.AutoscaleConfiguration{}, input)
	if expected.Difference(output).Len() != 0 && output.Difference(expected).Len() != 0 {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenAutoscaleConfiguration(t *testing.T) {
	cases := []struct {
		input    *api.AutoscaleConfiguration
		expected []map[string]interface{}
	}{
		{
			input: &api.AutoscaleConfiguration{
				NonGPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "non-gpu-node",
					DiskSize:          256,
					MinWorkers:        0,
					MaxWorkers:        5,
					StandbyWorkers:    0.5,
					DownscaleWaitTime: 300,
					SpotInfo: &api.SpotConfiguration{
						MaxPrice:         1,
						FallBackOnDemand: true,
					},
				},
				GPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "gpu-node",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        10,
					StandbyWorkers:    0.4,
					DownscaleWaitTime: 200,
					SpotInfo: &api.SpotConfiguration{
						MaxPrice:         2,
						FallBackOnDemand: true,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"non_gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "non-gpu-node",
							"disk_size":           256,
							"min_workers":         0,
							"max_workers":         5,
							"standby_workers":     0.5,
							"downscale_wait_time": 300,
							"spot_config": []interface{}{
								map[string]interface{}{
									"max_price_percent":   1,
									"fall_back_on_demand": true,
								},
							},
						},
					},
					"gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "gpu-node",
							"disk_size":           512,
							"min_workers":         1,
							"max_workers":         10,
							"standby_workers":     0.4,
							"downscale_wait_time": 200,
							"spot_config": []interface{}{
								map[string]interface{}{
									"max_price_percent":   2,
									"fall_back_on_demand": true,
								},
							},
						},
					},
				},
			},
		},
		{
			input: &api.AutoscaleConfiguration{
				NonGPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "non-gpu-node",
					DiskSize:          256,
					MinWorkers:        0,
					MaxWorkers:        5,
					StandbyWorkers:    0.5,
					DownscaleWaitTime: 300,
				},
			},
			expected: []map[string]interface{}{
				{
					"non_gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "non-gpu-node",
							"disk_size":           256,
							"min_workers":         0,
							"max_workers":         5,
							"standby_workers":     0.5,
							"downscale_wait_time": 300,
						},
					},
					"gpu_workers": []interface{}{},
				},
			},
		},
		{
			input: &api.AutoscaleConfiguration{
				GPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "gpu-node",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        10,
					StandbyWorkers:    0.4,
					DownscaleWaitTime: 200,
				},
			},
			expected: []map[string]interface{}{
				{
					"non_gpu_workers": []interface{}{},
					"gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "gpu-node",
							"disk_size":           512,
							"min_workers":         1,
							"max_workers":         10,
							"standby_workers":     0.4,
							"downscale_wait_time": 200,
						},
					},
				},
			},
		},
		{
			input:    nil,
			expected: nil,
		},
	}

	for i, c := range cases {
		output := flattenAutoscaleConfiguration(c.input)
		if !reflect.DeepEqual(c.expected, output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected, output)
		}
	}
}

func TestExpandAutoscaleConfiguration(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *api.AutoscaleConfiguration
	}{
		{
			expected: &api.AutoscaleConfiguration{
				NonGPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "non-gpu-node",
					DiskSize:          256,
					MinWorkers:        0,
					MaxWorkers:        5,
					StandbyWorkers:    0.5,
					DownscaleWaitTime: 300,
					SpotInfo: &api.SpotConfiguration{
						MaxPrice:         1,
						FallBackOnDemand: true,
					},
				},
				GPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "gpu-node",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        10,
					StandbyWorkers:    0.4,
					DownscaleWaitTime: 200,
					SpotInfo: &api.SpotConfiguration{
						MaxPrice:         2,
						FallBackOnDemand: true,
					},
				},
			},
			input: []interface{}{
				map[string]interface{}{
					"non_gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "non-gpu-node",
							"disk_size":           256,
							"min_workers":         0,
							"max_workers":         5,
							"standby_workers":     0.5,
							"downscale_wait_time": 300,
							"spot_config": []interface{}{
								map[string]interface{}{
									"max_price_percent":   1,
									"fall_back_on_demand": true,
								},
							},
						},
					},
					"gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "gpu-node",
							"disk_size":           512,
							"min_workers":         1,
							"max_workers":         10,
							"standby_workers":     0.4,
							"downscale_wait_time": 200,
							"spot_config": []interface{}{
								map[string]interface{}{
									"max_price_percent":   2,
									"fall_back_on_demand": true,
								},
							},
						},
					},
				},
			},
		},
		{
			expected: &api.AutoscaleConfiguration{
				NonGPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "non-gpu-node",
					DiskSize:          256,
					MinWorkers:        0,
					MaxWorkers:        5,
					StandbyWorkers:    0.5,
					DownscaleWaitTime: 300,
				},
			},
			input: []interface{}{
				map[string]interface{}{
					"non_gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "non-gpu-node",
							"disk_size":           256,
							"min_workers":         0,
							"max_workers":         5,
							"standby_workers":     0.5,
							"downscale_wait_time": 300,
						},
					},
					"gpu_workers": []interface{}{},
				},
			},
		},
		{
			expected: &api.AutoscaleConfiguration{
				GPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "gpu-node",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        10,
					StandbyWorkers:    0.4,
					DownscaleWaitTime: 200,
				},
			},
			input: []interface{}{
				map[string]interface{}{
					"non_gpu_workers": []interface{}{},
					"gpu_workers": []interface{}{
						map[string]interface{}{
							"instance_type":       "gpu-node",
							"disk_size":           512,
							"min_workers":         1,
							"max_workers":         10,
							"standby_workers":     0.4,
							"downscale_wait_time": 200,
						},
					},
				},
			},
		},
		{
			input:    nil,
			expected: nil,
		},
	}

	for i, c := range cases {
		output := ExpandAutoscaleConfiguration(c.input)
		if !reflect.DeepEqual(c.expected, output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected, output)
		}
	}
}

func TestExpandTags(t *testing.T) {
	input := map[string]interface{}{
		"tag1": "tag1-value",
		"tag2": "tag2-value",
	}

	expected1 := []api.ClusterTag{
		{
			Name:  "tag1",
			Value: "tag1-value",
		},
		{
			Name:  "tag2",
			Value: "tag2-value",
		},
	}

	expected2 := []api.ClusterTag{
		{
			Name:  "tag2",
			Value: "tag2-value",
		},
		{
			Name:  "tag1",
			Value: "tag1-value",
		},
	}

	output := ExpandTags(input)
	if !(reflect.DeepEqual(expected1, output) || reflect.DeepEqual(expected2, output)) {
		t.Fatalf("error while matching:\nexpected %#v or %#v \nbut got %#v", expected1, expected2, output)
	}
}

func TestFlattenClusters(t *testing.T) {
	input := []api.Cluster{
		{
			Id:                  "cluster-id-1",
			Name:                "cluster",
			State:               "state-1",
			ActivationState:     "activation-state-1",
			InitializationStage: "initializtion-stage-1",
			CreatedOn:           1605374387069,
			StartedOn:           1605374388069,
			Version:             "cluster-version",
			URL:                 "cluster-url",
			Provider:            api.AWS,
			Tags: []api.ClusterTag{
				{
					Name:  "tag1",
					Value: "tagvalue1",
				},
			},
			SshKeyName:            "ssh-key-1",
			PublicIPAttached:      true,
			LetsEncryptIssued:     true,
			ManagedUsers:          true,
			BackupRetentionPeriod: 0,
			ClusterConfiguration: api.ClusterConfigurationStatus{
				Head: api.HeadConfigurationStatus{
					HeadConfiguration: api.HeadConfiguration{
						NodeConfiguration: api.NodeConfiguration{
							InstanceType: "head-node-type-1",
							DiskSize:     512,
						},
						HAEnabled: false,
					},
					NodeId: "head-node-id-1",
				},
				Workers: []api.WorkerConfiguration{
					{
						NodeConfiguration: api.NodeConfiguration{
							InstanceType: "worker-node-type-1",
							DiskSize:     256,
						},
						Count: 1,
					},
				},
			},
			Ports: api.ServiceOpenPorts{
				FeatureStore:       true,
				OnlineFeatureStore: false,
				Kafka:              true,
				SSH:                false,
			},
			RonDB: &api.RonDBConfiguration{
				Configuration: api.RonDBBaseConfiguration{
					NdbdDefault: api.RonDBNdbdDefaultConfiguration{
						ReplicationFactor: 2,
					},
					General: api.RonDBGeneralConfiguration{
						Benchmark: api.RonDBBenchmarkConfiguration{
							GrantUserPrivileges: false,
						},
					},
				},
				ManagementNodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "mgm-node-1",
						DiskSize:     30,
					},
					Count: 1,
				},
				DataNodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "data-node-1",
						DiskSize:     512,
					},
					Count: 2,
				},
				MYSQLNodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "mysqld-node-1",
						DiskSize:     100,
					},
					Count: 1,
				},
				APINodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "api-node-1",
						DiskSize:     50,
					},
					Count: 1,
				},
			},
			Autoscale: &api.AutoscaleConfiguration{
				NonGPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "auto-node-1",
					DiskSize:          256,
					MinWorkers:        0,
					MaxWorkers:        10,
					StandbyWorkers:    0.5,
					DownscaleWaitTime: 300,
				},
				GPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "auto-gpu-node-1",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        5,
					StandbyWorkers:    0.4,
					DownscaleWaitTime: 200,
				},
			},
			InitScript: "#!/usr/bin/env bash\nset -e\necho 'Hello World'",
			AWS: api.AWSCluster{
				Region:               "region-1",
				InstanceProfileArn:   "instance-profile-1",
				BucketName:           "bucket-name-1",
				VpcId:                "vpc-id-1",
				SubnetId:             "subnet-id-1",
				SecurityGroupId:      "security-group-1",
				EksClusterName:       "eks-cluster-name-1",
				EcrRegistryAccountId: "ecr-registry-account-1",
			},
		},
		{
			Id:                  "cluster-id-2",
			Name:                "cluster2",
			State:               "state-2",
			ActivationState:     "activation-state-1",
			InitializationStage: "initializtion-stage-1",
			CreatedOn:           1605374387010,
			StartedOn:           1605374388010,
			Version:             "cluster-version-2",
			URL:                 "cluster-url-2",
			Provider:            api.AZURE,
			Tags: []api.ClusterTag{
				{
					Name:  "tag1",
					Value: "tagvalue1",
				},
			},
			SshKeyName:            "ssh-key-1",
			PublicIPAttached:      true,
			LetsEncryptIssued:     true,
			ManagedUsers:          true,
			BackupRetentionPeriod: 0,
			ClusterConfiguration: api.ClusterConfigurationStatus{
				Head: api.HeadConfigurationStatus{
					HeadConfiguration: api.HeadConfiguration{
						NodeConfiguration: api.NodeConfiguration{
							InstanceType: "head-node-type-1",
							DiskSize:     512,
						},
					},
					NodeId: "head-node-id-1",
				},
				Workers: []api.WorkerConfiguration{
					{
						NodeConfiguration: api.NodeConfiguration{
							InstanceType: "worker-node-type-1",
							DiskSize:     256,
						},
						Count: 1,
					},
				},
			},
			Ports: api.ServiceOpenPorts{
				FeatureStore:       true,
				OnlineFeatureStore: false,
				Kafka:              true,
				SSH:                false,
			},
			RonDB: &api.RonDBConfiguration{
				Configuration: api.RonDBBaseConfiguration{
					NdbdDefault: api.RonDBNdbdDefaultConfiguration{
						ReplicationFactor: 2,
					},
					General: api.RonDBGeneralConfiguration{
						Benchmark: api.RonDBBenchmarkConfiguration{
							GrantUserPrivileges: false,
						},
					},
				},
				ManagementNodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "mgm-node-1",
						DiskSize:     30,
					},
					Count: 1,
				},
				DataNodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "data-node-1",
						DiskSize:     512,
					},
					Count: 2,
				},
				MYSQLNodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "mysqld-node-1",
						DiskSize:     100,
					},
					Count: 1,
				},
				APINodes: api.RonDBNodeConfiguration{
					NodeConfiguration: api.NodeConfiguration{
						InstanceType: "api-node-1",
						DiskSize:     50,
					},
					Count: 1,
				},
			},
			Autoscale: &api.AutoscaleConfiguration{
				NonGPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "auto-node-1",
					DiskSize:          256,
					MinWorkers:        0,
					MaxWorkers:        10,
					StandbyWorkers:    0.5,
					DownscaleWaitTime: 300,
				},
				GPU: &api.AutoscaleConfigurationBase{
					InstanceType:      "auto-gpu-node-1",
					DiskSize:          512,
					MinWorkers:        1,
					MaxWorkers:        5,
					StandbyWorkers:    0.4,
					DownscaleWaitTime: 200,
				},
			},
			InitScript: "#!/usr/bin/env bash\nset -e\necho 'Hello World 2'",
			Azure: api.AzureCluster{
				Location:           "location-1",
				ResourceGroup:      "resource-group-1",
				ManagedIdentity:    "managed-identity-1",
				BlobContainerName:  "blob-container-name-1",
				StorageAccount:     "storage-account-1",
				VirtualNetworkName: "virtual-network-name-1",
				SubnetName:         "subnet-name-1",
				SecurityGroupName:  "security-group-name-1",
				AksClusterName:     "aks-cluster-name-1",
				AcrRegistryName:    "acr-registry-name-1",
				SearchDomain:       "internal.cloudapp.net",
			},
		},
	}

	var emptyAttributes []interface{} = nil
	expected := []map[string]interface{}{
		{
			"cluster_id":                     input[0].Id,
			"name":                           input[0].Name,
			"url":                            input[0].URL,
			"state":                          input[0].State,
			"activation_state":               input[0].ActivationState,
			"creation_date":                  time.Unix(input[0].CreatedOn, 0).Format(time.RFC3339),
			"start_date":                     time.Unix(input[0].StartedOn, 0).Format(time.RFC3339),
			"version":                        input[0].Version,
			"ssh_key":                        input[0].SshKeyName,
			"head":                           flattenHead(&input[0].ClusterConfiguration.Head),
			"issue_lets_encrypt_certificate": input[0].LetsEncryptIssued,
			"attach_public_ip":               input[0].PublicIPAttached,
			"managed_users":                  input[0].ManagedUsers,
			"backup_retention_period":        input[0].BackupRetentionPeriod,
			"update_state":                   "none",
			"workers":                        flattenWorkers(input[0].Autoscale, input[0].ClusterConfiguration.Workers),
			"aws_attributes":                 flattenAWSAttributes(&input[0]),
			"azure_attributes":               emptyAttributes,
			"open_ports":                     flattenPorts(&input[0].Ports),
			"tags":                           flattenTags(input[0].Tags),
			"rondb":                          flattenRonDB(input[0].RonDB),
			"autoscale":                      flattenAutoscaleConfiguration(input[0].Autoscale),
			"init_script":                    input[0].InitScript,
		},
		{
			"cluster_id":                     input[1].Id,
			"name":                           input[1].Name,
			"url":                            input[1].URL,
			"state":                          input[1].State,
			"activation_state":               input[1].ActivationState,
			"creation_date":                  time.Unix(input[1].CreatedOn, 0).Format(time.RFC3339),
			"start_date":                     time.Unix(input[1].StartedOn, 0).Format(time.RFC3339),
			"version":                        input[1].Version,
			"ssh_key":                        input[1].SshKeyName,
			"head":                           flattenHead(&input[1].ClusterConfiguration.Head),
			"issue_lets_encrypt_certificate": input[1].LetsEncryptIssued,
			"attach_public_ip":               input[1].PublicIPAttached,
			"managed_users":                  input[1].ManagedUsers,
			"backup_retention_period":        input[1].BackupRetentionPeriod,
			"update_state":                   "none",
			"workers":                        flattenWorkers(input[1].Autoscale, input[1].ClusterConfiguration.Workers),
			"aws_attributes":                 emptyAttributes,
			"azure_attributes":               flattenAzureAttributes(&input[1]),
			"open_ports":                     flattenPorts(&input[1].Ports),
			"tags":                           flattenTags(input[1].Tags),
			"rondb":                          flattenRonDB(input[1].RonDB),
			"autoscale":                      flattenAutoscaleConfiguration(input[1].Autoscale),
			"init_script":                    input[1].InitScript,
		},
	}

	output := FlattenClusters(input)

	for i, expectedCluster := range expected {
		outputCluster := output[i]
		for k, v := range expectedCluster {
			if k == "workers" {
				expectedWorker, outputWorker := v.(*schema.Set), outputCluster[k].(*schema.Set)
				if expectedWorker.Difference(outputWorker).Len() != 0 && outputWorker.Difference(expectedWorker).Len() != 0 {
					t.Fatalf("error while matching workers:\nexpected %#v \nbut got %#v", expectedWorker, outputWorker)
				}
			} else if !reflect.DeepEqual(v, outputCluster[k]) {
				t.Fatalf("error while matching %s:\nexpected %#v \nbut got %#v ", k, v, outputCluster[k])
			}
		}
	}
}

func TestFlattenUpgradeInProgress(t *testing.T) {
	input := []*api.UpgradeInProgress{
		{
			From: "v1",
			To:   "v2",
		},
		nil,
	}

	expected := [][]interface{}{
		{
			map[string]interface{}{
				"from_version": "v1",
				"to_version":   "v2",
			},
		},
		nil,
	}

	for i := range input {
		output := flattenUpgradeInProgress(input[i])
		if !reflect.DeepEqual(expected[i], output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, expected[i], output)
		}
	}
}

func TestFlattenAWSAttributes_bucketConfiguration(t *testing.T) {
	input := &api.Cluster{
		Provider: api.AWS,
		AWS: api.AWSCluster{
			Region:               "region-1",
			InstanceProfileArn:   "instance-profile-1",
			BucketName:           "bucket-name-1",
			VpcId:                "vpc-id-1",
			SubnetId:             "subnet-id-1",
			SecurityGroupId:      "security-group-1",
			EksClusterName:       "eks-cluster-name-1",
			EcrRegistryAccountId: "ecr-registry-account-1",
			BucketConfiguration: &api.S3BucketConfiguration{
				Encryption: api.S3EncryptionConfiguration{
					Mode:       "SSE-KMS",
					KMSType:    "User",
					UserKeyArn: "arn-key",
					BucketKey:  true,
				},
				ACL: &api.S3ACLConfiguration{
					BucketOwnerFullControl: true,
				},
			},
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"region":               input.AWS.Region,
			"instance_profile_arn": input.AWS.InstanceProfileArn,
			"network": []map[string]interface{}{
				{
					"vpc_id":            input.AWS.VpcId,
					"subnet_id":         input.AWS.SubnetId,
					"security_group_id": input.AWS.SecurityGroupId,
				},
			},
			"eks_cluster_name":        input.AWS.EksClusterName,
			"ecr_registry_account_id": input.AWS.EcrRegistryAccountId,
			"bucket": []map[string]interface{}{
				{
					"name": input.AWS.BucketName,
					"encryption": []map[string]interface{}{
						{
							"mode":         "SSE-KMS",
							"kms_type":     "User",
							"user_key_arn": "arn-key",
							"bucket_key":   true,
						},
					},
					"acl": []map[string]interface{}{
						{
							"bucket_owner_full_control": true,
						},
					},
				},
			},
			"ebs_encryption": []map[string]interface{}{},
		},
	}

	output := flattenAWSAttributes(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}

	input.Provider = ""
	if flattenAWSAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AWS)
	}

	input.Provider = api.AZURE
	if flattenAWSAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AWS)
	}

	input.Provider = "aws"
	if flattenAWSAttributes(input) != nil {
		t.Fatal("cloud provider should be always capital")
	}
}

func TestFlattenAzureAttributes_containerConfiguration(t *testing.T) {
	input := &api.Cluster{
		Provider: api.AZURE,
		Azure: api.AzureCluster{
			Location:             "location-1",
			ResourceGroup:        "resource-group-1",
			ManagedIdentity:      "managed-identity-1",
			BlobContainerName:    "blob-container-name-1",
			StorageAccount:       "storage-account-1",
			VirtualNetworkName:   "virtual-network-name-1",
			SubnetName:           "subnet-name-1",
			SecurityGroupName:    "security-group-name-1",
			AksClusterName:       "aks-cluster-name-1",
			AcrRegistryName:      "acr-registry-name-1",
			NetworkResourceGroup: "network-resource-group-1",
			SearchDomain:         "internal.cloudapp.net",
			ContainerConfiguration: &api.AzureContainerConfiguration{
				Encryption: api.AzureEncryptionConfiguration{
					Mode: "None",
				},
			},
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"location":                       input.Azure.Location,
			"resource_group":                 input.Azure.ResourceGroup,
			"user_assigned_managed_identity": input.Azure.ManagedIdentity,
			"network": []map[string]interface{}{
				{
					"resource_group":       input.Azure.NetworkResourceGroup,
					"virtual_network_name": input.Azure.VirtualNetworkName,
					"subnet_name":          input.Azure.SubnetName,
					"security_group_name":  input.Azure.SecurityGroupName,
					"search_domain":        input.Azure.SearchDomain,
				},
			},
			"aks_cluster_name":  input.Azure.AksClusterName,
			"acr_registry_name": input.Azure.AcrRegistryName,
			"container": []map[string]interface{}{
				{
					"name":            input.Azure.BlobContainerName,
					"storage_account": input.Azure.StorageAccount,
					"encryption": []map[string]interface{}{
						{
							"mode": "None",
						},
					},
				},
			},
		},
	}

	output := flattenAzureAttributes(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}

	input.Provider = ""
	if flattenAzureAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AZURE)
	}

	input.Provider = api.AWS
	if flattenAzureAttributes(input) != nil {
		t.Fatalf("should return nil if the provider is not %s", api.AZURE)
	}

	input.Provider = "azure"
	if flattenAzureAttributes(input) != nil {
		t.Fatal("cloud provider should be always capital")
	}
}

func TestFlattenS3BucketConfiguration(t *testing.T) {
	bucketName := "bucket-name"
	input := []*api.S3BucketConfiguration{
		{
			Encryption: api.S3EncryptionConfiguration{
				Mode:       "SSE-KMS",
				KMSType:    "User",
				UserKeyArn: "arn-key",
				BucketKey:  true,
			},
			ACL: &api.S3ACLConfiguration{
				BucketOwnerFullControl: true,
			},
		},
		{
			Encryption: api.S3EncryptionConfiguration{
				Mode: "SSE-S3",
			},
		},
		nil,
	}

	expected := [][]map[string]interface{}{
		{
			map[string]interface{}{
				"name": bucketName,
				"encryption": []map[string]interface{}{
					{
						"mode":         "SSE-KMS",
						"kms_type":     "User",
						"user_key_arn": "arn-key",
						"bucket_key":   true,
					},
				},
				"acl": []map[string]interface{}{
					{
						"bucket_owner_full_control": true,
					},
				},
			},
		},
		{
			map[string]interface{}{
				"name": bucketName,
				"encryption": []map[string]interface{}{
					{
						"mode":         "SSE-S3",
						"kms_type":     "",
						"user_key_arn": "",
						"bucket_key":   false,
					},
				},
			},
		},
		{
			map[string]interface{}{
				"name": bucketName,
			},
		},
	}

	for i := range input {
		output := flattenS3BucketConfiguration(bucketName, input[i])
		if !reflect.DeepEqual(expected[i], output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, expected[i], output)
		}
	}
}

func TestFlattenAzureContainerConfiguration(t *testing.T) {
	containerName := "container-name"
	storageAccount := "storage-account"
	input := []*api.AzureContainerConfiguration{
		{
			Encryption: api.AzureEncryptionConfiguration{
				Mode: "None",
			},
		},
		nil,
	}

	expected := [][]map[string]interface{}{
		{
			map[string]interface{}{
				"name":            containerName,
				"storage_account": storageAccount,
				"encryption": []map[string]interface{}{
					{
						"mode": "None",
					},
				},
			},
		},
		{
			map[string]interface{}{
				"name":            containerName,
				"storage_account": storageAccount,
			},
		},
	}

	for i := range input {
		output := flattenAzureContainerConfiguration(storageAccount, containerName, input[i])
		if !reflect.DeepEqual(expected[i], output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, expected[i], output)
		}
	}
}

func TestFlattenAWSAttributes_ebsEncryption(t *testing.T) {
	input := &api.Cluster{
		Provider: api.AWS,
		AWS: api.AWSCluster{
			Region:               "region-1",
			InstanceProfileArn:   "instance-profile-1",
			BucketName:           "bucket-name-1",
			VpcId:                "vpc-id-1",
			SubnetId:             "subnet-id-1",
			SecurityGroupId:      "security-group-1",
			EksClusterName:       "eks-cluster-name-1",
			EcrRegistryAccountId: "ecr-registry-account-1",
			EBSEncryption: &api.EBSEncryption{
				KmsKey: "my-key-id",
			},
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"region":               input.AWS.Region,
			"instance_profile_arn": input.AWS.InstanceProfileArn,
			"network": []map[string]interface{}{
				{
					"vpc_id":            input.AWS.VpcId,
					"subnet_id":         input.AWS.SubnetId,
					"security_group_id": input.AWS.SecurityGroupId,
				},
			},
			"eks_cluster_name":        input.AWS.EksClusterName,
			"ecr_registry_account_id": input.AWS.EcrRegistryAccountId,
			"bucket": []map[string]interface{}{
				{
					"name": input.AWS.BucketName,
				},
			},
			"ebs_encryption": []map[string]interface{}{
				{
					"kms_key": "my-key-id",
				},
			},
		},
	}

	output := flattenAWSAttributes(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

func TestFlattenEBSEncryption(t *testing.T) {
	input := []*api.EBSEncryption{
		{},
		{
			KmsKey: "my-kms-key",
		},
		nil,
	}

	expected := [][]map[string]interface{}{
		{
			map[string]interface{}{
				"kms_key": "",
			},
		},
		{
			map[string]interface{}{
				"kms_key": "my-kms-key",
			},
		},
		{},
	}

	for i := range input {
		output := flattenEBSEncryption(input[i])
		if !reflect.DeepEqual(expected[i], output) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, expected[i], output)
		}
	}
}

func TestFlattenRonDBNodeConfiguration(t *testing.T) {
	input := api.RonDBNodeConfiguration{
		NodeConfiguration: api.NodeConfiguration{
			InstanceType: "node-1",
			DiskSize:     128,
		},
		Count:      2,
		PrivateIps: []string{"ip1", "ip2"},
	}

	expected := map[string]interface{}{
		"instance_type": "node-1",
		"disk_size":     128,
		"count":         2,
		"private_ips":   []interface{}{"ip1", "ip2"},
	}

	output := flattenRonDBNode(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}

	config := ExpandRonDBNodeConfiguration(output)

	if !reflect.DeepEqual(input, config) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", input, config)
	}
}

func TestFlattenRonDB_single_node(t *testing.T) {
	input := &api.RonDBConfiguration{
		Configuration: api.RonDBBaseConfiguration{
			NdbdDefault: api.RonDBNdbdDefaultConfiguration{
				ReplicationFactor: 1,
			},
		},
		ManagementNodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "mgm-node-1",
				DiskSize:     30,
			},
			Count: 1,
		},
		DataNodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "data-node-1",
				DiskSize:     512,
			},
			Count:      1,
			PrivateIps: []string{"ip1"},
		},
		MYSQLNodes: api.RonDBNodeConfiguration{
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "mysqld-node-1",
				DiskSize:     100,
			},
			Count: 1,
		},
	}

	expected := []map[string]interface{}{
		{
			"configuration":    nil,
			"management_nodes": nil,
			"data_nodes":       nil,
			"mysql_nodes":      nil,
			"api_nodes":        nil,
			"single_node": []interface{}{
				map[string]interface{}{
					"instance_type": "data-node-1",
					"disk_size":     512,
					"private_ips":   []interface{}{"ip1"},
				},
			},
		},
	}

	output := flattenRonDB(input)
	if !reflect.DeepEqual(expected, output) {
		t.Fatalf("error while matching:\nexpected %#v \nbut got %#v", expected, output)
	}
}

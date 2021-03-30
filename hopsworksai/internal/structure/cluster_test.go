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

	input := &api.HeadConfiguration{
		NodeConfiguration: api.NodeConfiguration{
			InstanceType: "head-type-1",
			DiskSize:     512,
		},
	}

	expected := []map[string]interface{}{
		{
			"instance_type": input.InstanceType,
			"disk_size":     input.DiskSize,
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
			}),
		},
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
						InstanceType: "node-type-1",
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
			},
			expected: schema.NewSet(helpers.WorkerSetHash, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
					"count":         2,
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
			}),
		},
	}

	for i, c := range cases {
		output := flattenWorkers(c.input)
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
			"bucket_name":          input.AWS.BucketName,
			"network": []map[string]interface{}{
				{
					"vpc_id":            input.AWS.VpcId,
					"subnet_id":         input.AWS.SubnetId,
					"security_group_id": input.AWS.SecurityGroupId,
				},
			},
			"eks_cluster_name":        input.AWS.EksClusterName,
			"ecr_registry_account_id": input.AWS.EcrRegistryAccountId,
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
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"location":                       input.Azure.Location,
			"resource_group":                 input.Azure.ResourceGroup,
			"storage_account":                input.Azure.StorageAccount,
			"storage_container_name":         input.Azure.BlobContainerName,
			"user_assigned_managed_identity": input.Azure.ManagedIdentity,
			"network": []map[string]interface{}{
				{
					"virtual_network_name": input.Azure.VirtualNetworkName,
					"subnet_name":          input.Azure.SubnetName,
					"security_group_name":  input.Azure.SecurityGroupName,
				},
			},
			"aks_cluster_name":  input.Azure.AksClusterName,
			"acr_registry_name": input.Azure.AcrRegistryName,
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
		ClusterConfiguration: api.ClusterConfiguration{
			Head: api.HeadConfiguration{
				NodeConfiguration: api.NodeConfiguration{
					InstanceType: "head-node-type-1",
					DiskSize:     512,
				},
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
	}

	var emptyAttributes []interface{} = nil
	expected := map[string]interface{}{
		"cluster_id":                     input.Id,
		"name":                           input.Name,
		"url":                            input.URL,
		"state":                          input.State,
		"activation_state":               input.ActivationState,
		"creation_date":                  time.Unix(input.CreatedOn, 0).Format(time.RFC3339),
		"start_date":                     time.Unix(input.StartedOn, 0).Format(time.RFC3339),
		"version":                        input.Version,
		"ssh_key":                        input.SshKeyName,
		"head":                           flattenHead(&input.ClusterConfiguration.Head),
		"issue_lets_encrypt_certificate": input.LetsEncryptIssued,
		"attach_public_ip":               input.PublicIPAttached,
		"managed_users":                  input.ManagedUsers,
		"backup_retention_period":        input.BackupRetentionPeriod,
		"update_state":                   "none",
		"workers":                        flattenWorkers(input.ClusterConfiguration.Workers),
		"aws_attributes":                 emptyAttributes,
		"azure_attributes":               emptyAttributes,
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
			} else {
				if !reflect.DeepEqual(v, output[k]) {
					t.Fatalf("error while matching %s:\nexpected %#v \nbut got %#v for %s", k, v, output[k], cloud.String())
				}
			}
		}
	}
}

func TestDiffWorkers(t *testing.T) {
	cases := []struct {
		input1    []api.WorkerConfiguration
		input2    *schema.Set
		expected1 bool
		expected2 string
	}{
		{
			input1: []api.WorkerConfiguration{
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
			},
			input2: schema.NewSet(helpers.WorkerSetHashIncludingCount, []interface{}{
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
			}),
			expected1: false,
			expected2: "",
		},
		{
			input1: []api.WorkerConfiguration{
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
			},
			input2: schema.NewSet(helpers.WorkerSetHashIncludingCount, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
					"count":         1,
				},
				map[string]interface{}{
					"instance_type": "node-type-3",
					"disk_size":     1024,
					"count":         1,
				},
			}),
			expected1: true,
			expected2: "Diff report:\n\tHopsworks.ai changes:\n\t\tinstance_type=node-type-1, disk_size=512, count=2\n\t\tinstance_type=node-type-2, disk_size=256, count=3\n\tLocal changes:\n\t\tinstance_type=node-type-1, disk_size=512, count=1\n",
		},
		{
			input1: []api.WorkerConfiguration{
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
			},
			input2: schema.NewSet(helpers.WorkerSetHashIncludingCount, []interface{}{
				map[string]interface{}{
					"instance_type": "node-type-1",
					"disk_size":     512,
					"count":         1,
				},
				map[string]interface{}{
					"instance_type": "node-type-3",
					"disk_size":     1024,
					"count":         1,
				},
			}),
			expected1: true,
			expected2: "Diff report:\n\tHopsworks.ai changes:\n\t\tinstance_type=node-type-1, disk_size=512, count=2\n\t\tinstance_type=node-type-2, disk_size=256, count=3\n\tLocal changes:\n\t\tinstance_type=node-type-1, disk_size=512, count=1\n\t\tinstance_type=node-type-3, disk_size=1024, count=1\n",
		},
	}

	for i, c := range cases {
		output1, output2 := DiffWorkers(c.input1, c.input2)
		if !reflect.DeepEqual(c.expected1, output1) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected1, output1)
		}
		if !reflect.DeepEqual(c.expected2, output2) {
			t.Fatalf("error while matching[%d]:\nexpected %#v \nbut got %#v", i, c.expected2, output2)
		}
	}

}

func TestExpectNode(t *testing.T) {
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

func TestExpectWorker(t *testing.T) {
	input := map[string]interface{}{
		"instance_type": "instance-type-1",
		"disk_size":     512,
		"count":         2,
	}

	expected := api.WorkerConfiguration{
		NodeConfiguration: api.NodeConfiguration{
			InstanceType: "instance-type-1",
			DiskSize:     512,
		},
		Count: 2,
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

	expected := map[api.NodeConfiguration]api.WorkerConfiguration{
		{
			InstanceType: "node-type-1",
			DiskSize:     512,
		}: {
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-1",
				DiskSize:     512,
			},
			Count: 2,
		},
		{
			InstanceType: "node-type-1",
			DiskSize:     256,
		}: {
			NodeConfiguration: api.NodeConfiguration{
				InstanceType: "node-type-1",
				DiskSize:     256,
			},
			Count: 3,
		},
		{
			InstanceType: "node-type-3",
			DiskSize:     1024,
		}: {
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
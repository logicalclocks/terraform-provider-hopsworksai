package structure

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/helpers"
)

func FlattenClusters(clustersArray []api.Cluster) []map[string]interface{} {
	clusters := make([]map[string]interface{}, 0)
	for _, v := range clustersArray {
		clusters = append(clusters, FlattenCluster(&v))
	}
	return clusters
}

func FlattenCluster(cluster *api.Cluster) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id":                     cluster.Id,
		"name":                           cluster.Name,
		"url":                            cluster.URL,
		"state":                          cluster.State,
		"activation_state":               cluster.ActivationState,
		"creation_date":                  time.Unix(cluster.CreatedOn, 0).Format(time.RFC3339),
		"start_date":                     time.Unix(cluster.StartedOn, 0).Format(time.RFC3339),
		"version":                        cluster.Version,
		"ssh_key":                        cluster.SshKeyName,
		"head":                           flattenHead(&cluster.ClusterConfiguration.Head),
		"issue_lets_encrypt_certificate": cluster.LetsEncryptIssued,
		"attach_public_ip":               cluster.PublicIPAttached,
		"managed_users":                  cluster.ManagedUsers,
		"backup_retention_period":        cluster.BackupRetentionPeriod,
		"update_state":                   "none",
		"workers":                        flattenWorkers(cluster.ClusterConfiguration.Workers),
		"aws_attributes":                 flattenAWSAttributes(cluster),
		"azure_attributes":               flattenAzureAttributes(cluster),
	}
}

func flattenHead(head *api.HeadConfiguration) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"instance_type": head.InstanceType,
			"disk_size":     head.DiskSize,
		},
	}
}

func flattenWorkers(workers []api.WorkerConfiguration) *schema.Set {
	workersArray := make([]interface{}, len(workers))
	for i, v := range workers {
		workersArray[i] = map[string]interface{}{
			"instance_type": v.InstanceType,
			"disk_size":     v.DiskSize,
			"count":         v.Count,
		}
	}
	return schema.NewSet(helpers.WorkerSetHash, workersArray)
}

func flattenAWSAttributes(cluster *api.Cluster) []interface{} {
	if !cluster.IsAWSCluster() {
		return nil
	}

	awsAttributes := make([]interface{}, 1)
	awsAttributes[0] = map[string]interface{}{
		"region":               cluster.AWS.Region,
		"bucket_name":          cluster.AWS.BucketName,
		"instance_profile_arn": cluster.AWS.InstanceProfileArn,
		"network": []map[string]interface{}{
			{
				"vpc_id":            cluster.AWS.VpcId,
				"subnet_id":         cluster.AWS.SubnetId,
				"security_group_id": cluster.AWS.SecurityGroupId,
			},
		},
		"eks_cluster_name":        cluster.AWS.EksClusterName,
		"ecr_registry_account_id": cluster.AWS.EcrRegistryAccountId,
	}
	return awsAttributes
}

func flattenAzureAttributes(cluster *api.Cluster) []interface{} {
	if !cluster.IsAzureCluster() {
		return nil
	}

	azureAttributes := make([]interface{}, 1)
	azureAttributes[0] = map[string]interface{}{
		"location":                       cluster.Azure.Location,
		"resource_group":                 cluster.Azure.ResourceGroup,
		"storage_account":                cluster.Azure.StorageAccount,
		"storage_container_name":         cluster.Azure.BlobContainerName,
		"user_assigned_managed_identity": cluster.Azure.ManagedIdentity,
		"network": []map[string]interface{}{
			{
				"virtual_network_name": cluster.Azure.VirtualNetworkName,
				"subnet_name":          cluster.Azure.SubnetName,
				"security_group_name":  cluster.Azure.SecurityGroupName,
			},
		},
		"aks_cluster_name":  cluster.Azure.AksClusterName,
		"acr_registry_name": cluster.Azure.AcrRegistryName,
	}
	return azureAttributes
}

func ExpandWorkers(workers *schema.Set) map[api.NodeConfiguration]api.WorkerConfiguration {
	workersMap := make(map[api.NodeConfiguration]api.WorkerConfiguration, workers.Len())
	for _, v := range workers.List() {
		worker := ExpandWorker(v.(map[string]interface{}))
		workersMap[worker.NodeConfiguration] = worker
	}
	return workersMap
}

func ExpandWorker(workerConfig map[string]interface{}) api.WorkerConfiguration {
	return api.WorkerConfiguration{
		NodeConfiguration: ExpandNode(workerConfig),
		Count:             workerConfig["count"].(int),
	}
}

func ExpandNode(config map[string]interface{}) api.NodeConfiguration {
	return api.NodeConfiguration{
		InstanceType: config["instance_type"].(string),
		DiskSize:     config["disk_size"].(int),
	}
}

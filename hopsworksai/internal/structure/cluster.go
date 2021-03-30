package structure

import (
	"bytes"
	"fmt"
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
	return flattenWorkersBase(workers, helpers.WorkerSetHash)
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

func DiffWorkers(workers []api.WorkerConfiguration, old *schema.Set) (bool, string) {
	hwWorkers := flattenWorkersBase(workers, helpers.WorkerSetHashIncludingCount)
	localWorkers := schema.NewSet(helpers.WorkerSetHashIncludingCount, old.List())
	diffToHW := hwWorkers.Difference(localWorkers)
	diffToLocal := localWorkers.Difference(hwWorkers)

	var message string = ""
	if diffToHW.Len() > 0 || diffToLocal.Len() > 0 {
		message = "Diff report:\n"
	}
	if diffToHW.Len() > 0 {
		message += fmt.Sprintf("\tHopsworks.ai changes:\n%s", workersString(diffToHW))
	}
	if diffToLocal.Len() > 0 {
		message += fmt.Sprintf("\tLocal changes:\n%s", workersString(diffToLocal))
	}
	return diffToHW.Len() > 0 && diffToLocal.Len() > 0, message
}

func workersString(workers *schema.Set) string {
	var buf bytes.Buffer
	for _, v := range workers.List() {
		worker := v.(map[string]interface{})
		buf.WriteString(fmt.Sprintf("\t\tinstance_type=%s", worker["instance_type"].(string)))
		buf.WriteString(fmt.Sprintf(", disk_size=%d", worker["disk_size"].(int)))
		buf.WriteString(fmt.Sprintf(", count=%d", worker["count"].(int)))
		buf.WriteString("\n")
	}
	return buf.String()
}

func flattenWorkersBase(workers []api.WorkerConfiguration, setHash schema.SchemaSetFunc) *schema.Set {
	workersArray := make([]interface{}, len(workers))
	for i, v := range workers {
		workersArray[i] = map[string]interface{}{
			"instance_type": v.InstanceType,
			"disk_size":     v.DiskSize,
			"count":         v.Count,
		}
	}
	return schema.NewSet(setHash, workersArray)
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

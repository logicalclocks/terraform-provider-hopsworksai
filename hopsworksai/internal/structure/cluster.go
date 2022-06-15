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
		"cluster_id":                            cluster.Id,
		"name":                                  cluster.Name,
		"url":                                   cluster.URL,
		"state":                                 cluster.State,
		"activation_state":                      cluster.ActivationState,
		"creation_date":                         time.Unix(cluster.CreatedOn, 0).Format(time.RFC3339),
		"start_date":                            time.Unix(cluster.StartedOn, 0).Format(time.RFC3339),
		"version":                               cluster.Version,
		"ssh_key":                               cluster.SshKeyName,
		"head":                                  flattenHead(&cluster.ClusterConfiguration.Head),
		"issue_lets_encrypt_certificate":        cluster.LetsEncryptIssued,
		"attach_public_ip":                      cluster.PublicIPAttached,
		"managed_users":                         cluster.ManagedUsers,
		"backup_retention_period":               cluster.BackupRetentionPeriod,
		"update_state":                          "none",
		"workers":                               flattenWorkers(cluster.Autoscale, cluster.ClusterConfiguration.Workers),
		"aws_attributes":                        flattenAWSAttributes(cluster),
		"azure_attributes":                      flattenAzureAttributes(cluster),
		"open_ports":                            flattenPorts(&cluster.Ports),
		"tags":                                  flattenTags(cluster.Tags),
		"rondb":                                 flattenRonDB(cluster.RonDB),
		"autoscale":                             flattenAutoscaleConfiguration(cluster.Autoscale),
		"init_script":                           cluster.InitScript,
		"run_init_script_first":                 cluster.RunInitScriptFirst,
		"os":                                    cluster.OS,
		"upgrade_in_progress":                   flattenUpgradeInProgress(cluster.UpgradeInProgress),
		"deactivate_hopsworksai_log_collection": cluster.DeactivateLogReport,
		"collect_logs":                          cluster.CollectLogs,
		"cluster_domain_prefix":                 cluster.ClusterDomainPrefix,
		"custom_hosted_zone":                    cluster.CustomHostedZone,
	}
}

func flattenHead(head *api.HeadConfigurationStatus) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"instance_type": head.InstanceType,
			"disk_size":     head.DiskSize,
			"node_id":       head.NodeId,
			"ha_enabled":    head.HAEnabled,
		},
	}
}

func flattenWorkers(autoscale *api.AutoscaleConfiguration, workers []api.WorkerConfiguration) *schema.Set {
	if autoscale != nil {
		return schema.NewSet(helpers.WorkerSetHash, []interface{}{})
	}
	workersArray := make([]interface{}, len(workers))
	for i, v := range workers {
		workersArray[i] = flattenWorker(v)
	}
	return schema.NewSet(helpers.WorkerSetHash, workersArray)
}

func flattenWorker(worker api.WorkerConfiguration) map[string]interface{} {
	workerConf := map[string]interface{}{
		"instance_type": worker.InstanceType,
		"disk_size":     worker.DiskSize,
		"count":         worker.Count,
	}
	if worker.SpotInfo != nil {
		workerConf["spot_config"] = flattenSpotInfo(worker.SpotInfo)
	}
	return workerConf
}

func flattenSpotInfo(spotInfo *api.SpotConfiguration) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"max_price_percent":   spotInfo.MaxPrice,
			"fall_back_on_demand": spotInfo.FallBackOnDemand,
		},
	}
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
		"bucket":                  flattenS3BucketConfiguration(cluster.AWS.BucketName, cluster.AWS.BucketConfiguration),
		"ebs_encryption":          flattenEBSEncryption(cluster.AWS.EBSEncryption),
	}
	return awsAttributes
}

func flattenS3BucketConfiguration(bucketName string, bucketConfiguration *api.S3BucketConfiguration) []map[string]interface{} {
	config := []map[string]interface{}{
		{
			"name": bucketName,
		},
	}

	if bucketConfiguration != nil {
		config[0]["encryption"] = []map[string]interface{}{
			{
				"mode":         bucketConfiguration.Encryption.Mode,
				"kms_type":     bucketConfiguration.Encryption.KMSType,
				"user_key_arn": bucketConfiguration.Encryption.UserKeyArn,
				"bucket_key":   bucketConfiguration.Encryption.BucketKey,
			},
		}

		if bucketConfiguration.ACL != nil {
			config[0]["acl"] = []map[string]interface{}{
				{
					"bucket_owner_full_control": bucketConfiguration.ACL.BucketOwnerFullControl,
				},
			}
		}
	}

	return config
}

func flattenEBSEncryption(ebsEncryption *api.EBSEncryption) []map[string]interface{} {
	if ebsEncryption == nil {
		return []map[string]interface{}{}
	}

	return []map[string]interface{}{
		{
			"kms_key": ebsEncryption.KmsKey,
		},
	}
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
				"resource_group":       cluster.Azure.NetworkResourceGroup,
				"virtual_network_name": cluster.Azure.VirtualNetworkName,
				"subnet_name":          cluster.Azure.SubnetName,
				"security_group_name":  cluster.Azure.SecurityGroupName,
				"search_domain":        cluster.Azure.SearchDomain,
			},
		},
		"aks_cluster_name":  cluster.Azure.AksClusterName,
		"acr_registry_name": cluster.Azure.AcrRegistryName,
		"search_domain":     cluster.Azure.SearchDomain,
		"container":         flattenAzureContainerConfiguration(cluster.Azure.StorageAccount, cluster.Azure.BlobContainerName, cluster.Azure.ContainerConfiguration),
	}
	return azureAttributes
}

func flattenAzureContainerConfiguration(storageAccount string, containerName string, containerConfiguration *api.AzureContainerConfiguration) []map[string]interface{} {
	config := []map[string]interface{}{
		{
			"name":            containerName,
			"storage_account": storageAccount,
		},
	}

	if containerConfiguration != nil {
		config[0]["encryption"] = []map[string]interface{}{
			{
				"mode": containerConfiguration.Encryption.Mode,
			},
		}
	}
	return config
}

func flattenPorts(ports *api.ServiceOpenPorts) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"feature_store":        ports.FeatureStore,
			"online_feature_store": ports.OnlineFeatureStore,
			"kafka":                ports.Kafka,
			"ssh":                  ports.SSH,
		},
	}
}

func flattenTags(tags []api.ClusterTag) map[string]interface{} {
	tagsMap := make(map[string]interface{}, len(tags))
	for _, tag := range tags {
		tagsMap[tag.Name] = tag.Value
	}
	return tagsMap
}

func flattenRonDB(ronDB *api.RonDBConfiguration) []map[string]interface{} {
	if ronDB == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"configuration": []interface{}{
				map[string]interface{}{
					"ndbd_default": []interface{}{
						map[string]interface{}{
							"replication_factor": ronDB.Configuration.NdbdDefault.ReplicationFactor,
						},
					},
					"general": []interface{}{
						map[string]interface{}{
							"benchmark": []interface{}{
								map[string]interface{}{
									"grant_user_privileges": ronDB.Configuration.General.Benchmark.GrantUserPrivileges,
								},
							},
						},
					},
				},
			},
			"management_nodes": []interface{}{
				flattenWorker(ronDB.ManagementNodes),
			},
			"data_nodes": []interface{}{
				flattenWorker(ronDB.DataNodes),
			},
			"mysql_nodes": []interface{}{
				flattenWorker(ronDB.MYSQLNodes),
			},
			"api_nodes": []interface{}{
				flattenWorker(ronDB.APINodes),
			},
		},
	}
}

func flattenAutoscaleConfiguration(autoscale *api.AutoscaleConfiguration) []map[string]interface{} {
	if autoscale == nil {
		return nil
	}

	var nonGPUNodes []interface{} = make([]interface{}, 0)
	var gpuNodes []interface{} = make([]interface{}, 0)
	if autoscale.NonGPU != nil {
		nonGPUNodes = append(nonGPUNodes, flattenAutoscaleConfigurationBase(autoscale.NonGPU))
	}
	if autoscale.GPU != nil {
		gpuNodes = append(gpuNodes, flattenAutoscaleConfigurationBase(autoscale.GPU))
	}

	return []map[string]interface{}{
		{
			"non_gpu_workers": nonGPUNodes,
			"gpu_workers":     gpuNodes,
		},
	}
}

func flattenAutoscaleConfigurationBase(autoscale *api.AutoscaleConfigurationBase) map[string]interface{} {
	autoscaleConf := map[string]interface{}{
		"instance_type":       autoscale.InstanceType,
		"disk_size":           autoscale.DiskSize,
		"min_workers":         autoscale.MinWorkers,
		"max_workers":         autoscale.MaxWorkers,
		"standby_workers":     autoscale.StandbyWorkers,
		"downscale_wait_time": autoscale.DownscaleWaitTime,
	}
	if autoscale.SpotInfo != nil {
		autoscaleConf["spot_config"] = flattenSpotInfo(autoscale.SpotInfo)
	}
	return autoscaleConf
}

func ExpandAutoscaleConfiguration(autoscaleConfig []interface{}) *api.AutoscaleConfiguration {
	if autoscaleConfig == nil {
		return nil
	}
	autoscale := &api.AutoscaleConfiguration{}
	if len(autoscaleConfig) > 0 {
		autoscaleConfigMap := autoscaleConfig[0].(map[string]interface{})
		if v, ok := autoscaleConfigMap["non_gpu_workers"]; ok && len(v.([]interface{})) > 0 {
			config := v.([]interface{})[0].(map[string]interface{})
			autoscale.NonGPU = ExpandAutoscaleConfigurationBase(config)
		}

		if v, ok := autoscaleConfigMap["gpu_workers"]; ok && len(v.([]interface{})) > 0 {
			config := v.([]interface{})[0].(map[string]interface{})
			autoscale.GPU = ExpandAutoscaleConfigurationBase(config)
		}
	}
	return autoscale
}

func ExpandAutoscaleConfigurationBase(config map[string]interface{}) *api.AutoscaleConfigurationBase {
	autoscaleConf := &api.AutoscaleConfigurationBase{
		InstanceType:      config["instance_type"].(string),
		DiskSize:          config["disk_size"].(int),
		MinWorkers:        config["min_workers"].(int),
		MaxWorkers:        config["max_workers"].(int),
		StandbyWorkers:    config["standby_workers"].(float64),
		DownscaleWaitTime: config["downscale_wait_time"].(int),
	}
	if _, ok := config["spot_config"]; ok {
		spot_configArr := config["spot_config"].([]interface{})
		if len(spot_configArr) > 0 && spot_configArr[0] != nil {
			spot_config := spot_configArr[0].(map[string]interface{})
			autoscaleConf.SpotInfo = &api.SpotConfiguration{
				MaxPrice:         spot_config["max_price_percent"].(int),
				FallBackOnDemand: spot_config["fall_back_on_demand"].(bool),
			}
		}
	}
	return autoscaleConf
}

func ExpandWorkers(workers *schema.Set) map[string]api.WorkerConfiguration {
	workersMap := make(map[string]api.WorkerConfiguration, workers.Len())
	for _, v := range workers.List() {
		worker := ExpandWorker(v.(map[string]interface{}))
		workersMap[helpers.WorkerKey(v)] = worker
	}
	return workersMap
}

func ExpandWorker(workerConfig map[string]interface{}) api.WorkerConfiguration {
	workerConf := api.WorkerConfiguration{
		NodeConfiguration: ExpandNode(workerConfig),
		Count:             workerConfig["count"].(int),
	}
	if _, ok := workerConfig["spot_config"]; ok {
		spot_configArr := workerConfig["spot_config"].([]interface{})
		if len(spot_configArr) > 0 && spot_configArr[0] != nil {
			spot_config := spot_configArr[0].(map[string]interface{})
			workerConf.SpotInfo = &api.SpotConfiguration{
				MaxPrice:         spot_config["max_price_percent"].(int),
				FallBackOnDemand: spot_config["fall_back_on_demand"].(bool),
			}
		}
	}
	return workerConf
}

func ExpandNode(config map[string]interface{}) api.NodeConfiguration {
	return api.NodeConfiguration{
		InstanceType: config["instance_type"].(string),
		DiskSize:     config["disk_size"].(int),
	}
}

func ExpandPorts(ports map[string]interface{}) api.ServiceOpenPorts {
	return api.ServiceOpenPorts{
		FeatureStore:       ports["feature_store"].(bool),
		OnlineFeatureStore: ports["online_feature_store"].(bool),
		Kafka:              ports["kafka"].(bool),
		SSH:                ports["ssh"].(bool),
	}
}

func ExpandTags(tags map[string]interface{}) []api.ClusterTag {
	tagsArr := make([]api.ClusterTag, len(tags))
	var index int = 0
	for k, v := range tags {
		tagsArr[index] = api.ClusterTag{
			Name:  k,
			Value: v.(string),
		}
		index++
	}
	return tagsArr
}

func flattenUpgradeInProgress(upgradeInProgress *api.UpgradeInProgress) []interface{} {
	if upgradeInProgress == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"from_version": upgradeInProgress.From,
			"to_version":   upgradeInProgress.To,
		},
	}
}

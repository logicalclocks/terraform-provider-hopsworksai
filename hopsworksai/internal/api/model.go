package api

import (
	"fmt"
	"net/http"
	"sort"
)

type CloudProvider string

const (
	AWS   CloudProvider = "AWS"
	AZURE CloudProvider = "AZURE"
)

func (c CloudProvider) String() string {
	return string(c)
}

// Cluster states
type ClusterState string

const worker = "worker"
const externally = "externally"
const secondary = "secondary"

const (
	Starting               ClusterState = "starting"
	Pending                ClusterState = "pending"
	Initializing           ClusterState = "initializing"
	Running                ClusterState = "running"
	Stopping               ClusterState = "stopping"
	Stopped                ClusterState = "stopped"
	Error                  ClusterState = "error"
	TerminationWarning     ClusterState = "termination-warning"
	ShuttingDown           ClusterState = "shutting-down"
	Updating               ClusterState = "updating"
	Decommissioning        ClusterState = "decommissioning"
	RonDBInitializing      ClusterState = "rondb-initializing"
	StartingHopsworks      ClusterState = "starting-hopsworks"
	CommandFailed          ClusterState = "command-failed"
	ExternallyStopped      ClusterState = externally + "-" + Stopped
	ExternallyShuttingDown ClusterState = externally + "-" + ShuttingDown
	ExternallyTerminated   ClusterState = externally + "-" + "terminated"
	// Worker states
	WorkerPending         ClusterState = worker + "-" + Pending
	WorkerInitializing    ClusterState = worker + "-" + Initializing
	WorkerStarting        ClusterState = worker + "-" + Starting
	WorkerError           ClusterState = worker + "-" + Error
	WorkerShuttingdown    ClusterState = worker + "-" + ShuttingDown
	WorkerDecommissioning ClusterState = worker + "-" + Decommissioning
	// Secondary states
	SecondaryInitializing ClusterState = secondary + "-" + Initializing
	SecondaryError        ClusterState = secondary + "-" + Error
	// local state not in Hopsworks.ai
	ClusterDeleted ClusterState = "tf-cluster-deleted"
)

func (s ClusterState) String() string {
	return string(s)
}

// Activation State
type ActivationState string

const (
	Startable  ActivationState = "startable"
	Stoppable  ActivationState = "stoppable"
	Terminable ActivationState = "terminable"
)

func (s ActivationState) String() string {
	return string(s)
}

// OS
type OS string

const (
	Ubuntu OS = "ubuntu"
	CentOS OS = "centos"
)

func (o OS) String() string {
	return string(o)
}

type BaseResponse struct {
	ApiVersion string `json:"apiVersion"`
	Status     string `json:"status"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
}

func (resp BaseResponse) validate() error {
	if resp.Code/100 != 2 && resp.Code != http.StatusNotFound {
		return fmt.Errorf(resp.Message)
	}
	return nil
}

type RonDBNdbdDefaultConfiguration struct {
	ReplicationFactor int `json:"replicationFactor"`
}

type RonDBBenchmarkConfiguration struct {
	GrantUserPrivileges bool `json:"grantUserPrivileges"`
}

type RonDBGeneralConfiguration struct {
	Benchmark RonDBBenchmarkConfiguration `json:"benchmark"`
}

type RonDBBaseConfiguration struct {
	NdbdDefault RonDBNdbdDefaultConfiguration `json:"ndbdDefault"`
	General     RonDBGeneralConfiguration     `json:"general"`
}

type RonDBNodeConfiguration struct {
	NodeConfiguration
	Count      int      `json:"count"`
	PrivateIps []string `json:"privateIps,omitempty"`
}

type RonDBConfiguration struct {
	Configuration   RonDBBaseConfiguration `json:"configuration"`
	ManagementNodes RonDBNodeConfiguration `json:"mgmd"`
	DataNodes       RonDBNodeConfiguration `json:"ndbd"`
	MYSQLNodes      RonDBNodeConfiguration `json:"mysqld"`
	APINodes        RonDBNodeConfiguration `json:"api"`
}

func (rondb *RonDBConfiguration) IsSingleNodeSetup() bool {
	return rondb.Configuration.NdbdDefault.ReplicationFactor == 1 && rondb.DataNodes.Count == 1 && rondb.MYSQLNodes.Count == 1
}

type SpotConfiguration struct {
	MaxPrice         int  `json:"maxPrice"`
	FallBackOnDemand bool `json:"fallBackOnDemand"`
}

type AutoscaleConfigurationBase struct {
	InstanceType      string             `json:"instanceType"`
	DiskSize          int                `json:"diskSize"`
	MinWorkers        int                `json:"minWorkers"`
	MaxWorkers        int                `json:"maxWorkers"`
	StandbyWorkers    float64            `json:"standbyWorkers"`
	DownscaleWaitTime int                `json:"downscaleWaitTime"`
	SpotInfo          *SpotConfiguration `json:"spotInfo,omitempty"`
}

type AutoscaleConfiguration struct {
	NonGPU *AutoscaleConfigurationBase `json:"nonGpu,omitempty"`
	GPU    *AutoscaleConfigurationBase `json:"gpu,omitempty"`
}

type UpgradeInProgress struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Cluster struct {
	Id                       string                     `json:"id"`
	Name                     string                     `json:"name"`
	State                    ClusterState               `json:"state"`
	ActivationState          ActivationState            `json:"activationState"`
	InitializationStage      string                     `json:"initializationStage"`
	CreatedOn                int64                      `json:"createdOn"`
	StartedOn                int64                      `json:"startedOn"`
	Version                  string                     `json:"version"`
	URL                      string                     `json:"url"`
	Provider                 CloudProvider              `json:"provider"`
	ErrorMessage             string                     `json:"errorMessage,omitempty"`
	Tags                     []ClusterTag               `json:"tags"`
	SshKeyName               string                     `json:"sshKeyName"`
	ClusterConfiguration     ClusterConfigurationStatus `json:"clusterConfiguration,omitempty"`
	PublicIPAttached         bool                       `json:"publicIPAttached"`
	LetsEncryptIssued        bool                       `json:"letsEncryptIssued"`
	ManagedUsers             bool                       `json:"managedUsers"`
	BackupRetentionPeriod    int                        `json:"backupRetentionPeriod"`
	Azure                    AzureCluster               `json:"azure,omitempty"`
	AWS                      AWSCluster                 `json:"aws,omitempty"`
	Ports                    ServiceOpenPorts           `json:"ports"`
	RonDB                    *RonDBConfiguration        `json:"ronDB,omitempty"`
	Autoscale                *AutoscaleConfiguration    `json:"autoscale,omitempty"`
	InitScript               string                     `json:"initScript"`
	RunInitScriptFirst       bool                       `json:"runInitScriptFirst"`
	OS                       OS                         `json:"os,omitempty"`
	UpgradeInProgress        *UpgradeInProgress         `json:"upgradeInProgress,omitempty"`
	DeactivateLogReport      bool                       `json:"deactivateLogReport"`
	CollectLogs              bool                       `json:"collectLogs"`
	BackupPipelineInProgress bool                       `json:"backupPipelineInProgress"`
	ClusterDomainPrefix      string                     `json:"clusterDomainPrefix,omitempty"`
	CustomHostedZone         string                     `json:"customHostedZone,omitempty"`
}

func (c *Cluster) IsAWSCluster() bool {
	return c.Provider == AWS
}

func (c *Cluster) IsAzureCluster() bool {
	return c.Provider == AZURE
}

type AzureEncryptionConfiguration struct {
	Mode string `json:"mode"`
}

type AzureContainerConfiguration struct {
	Encryption AzureEncryptionConfiguration `json:"encryption"`
}

type AzureCluster struct {
	Location               string                       `json:"location"`
	ManagedIdentity        string                       `json:"managedIdentity"`
	ResourceGroup          string                       `json:"resourceGroup"`
	BlobContainerName      string                       `json:"blobContainerName"`
	StorageAccount         string                       `json:"storageAccount"`
	NetworkResourceGroup   string                       `json:"networkResourceGroup"`
	VirtualNetworkName     string                       `json:"virtualNetworkName"`
	SubnetName             string                       `json:"subnetName"`
	SecurityGroupName      string                       `json:"securityGroupName"`
	AksClusterName         string                       `json:"aksClusterName"`
	AcrRegistryName        string                       `json:"acrRegistryName"`
	SearchDomain           string                       `json:"searchDomain"`
	ContainerConfiguration *AzureContainerConfiguration `json:"containerConfiguration,omitempty"`
}

type S3EncryptionConfiguration struct {
	Mode       string `json:"mode"`
	KMSType    string `json:"kmsType"`
	UserKeyArn string `json:"userKeyArn"`
	BucketKey  bool   `json:"bucketKey"`
}

type S3ACLConfiguration struct {
	BucketOwnerFullControl bool `json:"bucketOwnerFullControl"`
}

type S3BucketConfiguration struct {
	Encryption S3EncryptionConfiguration `json:"encryption"`
	ACL        *S3ACLConfiguration       `json:"acl,omitempty"`
}

type EBSEncryption struct {
	KmsKey string `json:"kmsKey"`
}

type AWSCluster struct {
	Region               string                 `json:"region"`
	BucketName           string                 `json:"bucketName"`
	InstanceProfileArn   string                 `json:"instanceProfileArn"`
	VpcId                string                 `json:"vpcId"`
	SubnetId             string                 `json:"subnetId"`
	SecurityGroupId      string                 `json:"securityGroupId"`
	EksClusterName       string                 `json:"eksClusterName"`
	EcrRegistryAccountId string                 `json:"ecrRegistryAccountId"`
	BucketConfiguration  *S3BucketConfiguration `json:"bucketConfiguration,omitempty"`
	EBSEncryption        *EBSEncryption         `json:"ebsEncryption,omitempty"`
}

type NodeConfiguration struct {
	InstanceType string `json:"instanceType"`
	DiskSize     int    `json:"diskSize"`
}

type HeadConfiguration struct {
	NodeConfiguration
	HAEnabled bool `json:"haEnabled"`
}

type WorkerConfiguration struct {
	NodeConfiguration
	Count      int                `json:"count"`
	SpotInfo   *SpotConfiguration `json:"spotInfo,omitempty"`
	PrivateIps []string           `json:"privateIps,omitempty"`
}

type ClusterConfiguration struct {
	Head    HeadConfiguration     `json:"head"`
	Workers []WorkerConfiguration `json:"workers"`
}

type HeadConfigurationStatus struct {
	HeadConfiguration
	NodeId    string `json:"nodeId"`
	PrivateIp string `json:"privateIp,omitempty"`
}

type ClusterConfigurationStatus struct {
	Head    HeadConfigurationStatus `json:"head"`
	Workers []WorkerConfiguration   `json:"workers"`
}

type ClusterTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CreateCluster struct {
	Name                  string                  `json:"name"`
	Version               string                  `json:"version"`
	SshKeyName            string                  `json:"sshKeyName"`
	ClusterConfiguration  ClusterConfiguration    `json:"clusterConfiguration"`
	IssueLetsEncrypt      bool                    `json:"issueLetsEncrypt"`
	AttachPublicIP        bool                    `json:"attachPublicIP"`
	BackupRetentionPeriod int                     `json:"backupRetentionPeriod"`
	ManagedUsers          bool                    `json:"managedUsers"`
	Tags                  []ClusterTag            `json:"tags,omitempty"`
	RonDB                 *RonDBConfiguration     `json:"ronDB,omitempty"`
	Autoscale             *AutoscaleConfiguration `json:"autoscale,omitempty"`
	InitScript            string                  `json:"initScript"`
	RunInitScriptFirst    bool                    `json:"runInitScriptFirst"`
	OS                    OS                      `json:"os,omitempty"`
	DeactivateLogReport   bool                    `json:"deactivateLogReport"`
	CollectLogs           bool                    `json:"collectLogs"`
	ClusterDomainPrefix   string                  `json:"clusterDomainPrefix,omitempty"`
	CustomHostedZone      string                  `json:"customHostedZone,omitempty"`
}

type CreateAzureCluster struct {
	CreateCluster
	AzureCluster
}

type CreateAWSCluster struct {
	CreateCluster
	AWSCluster
}

type NodeType string

const (
	HeadNode            NodeType = "head"
	WorkerNode          NodeType = "worker"
	RonDBManagementNode NodeType = "rondb_management"
	RonDBDataNode       NodeType = "rondb_data"
	RonDBMySQLNode      NodeType = "rondb_mysql"
	RonDBAPINode        NodeType = "rondb_api"
	RonDBAllInOneNode   NodeType = "rondb_aio"
)

func (n NodeType) String() string {
	return string(n)
}

func GetAllNodeTypes() []string {
	return []string{
		HeadNode.String(),
		WorkerNode.String(),
		RonDBManagementNode.String(),
		RonDBDataNode.String(),
		RonDBMySQLNode.String(),
		RonDBAPINode.String(),
	}
}

type SupportedInstanceType struct {
	Id     string  `json:"id"`
	CPUs   int     `json:"cpus"`
	Memory float64 `json:"memory"`
	GPUs   int     `json:"gpus"`
}

type SupportedInstanceTypeList []SupportedInstanceType

func (l SupportedInstanceTypeList) Sort() {
	sort.SliceStable(l, func(i, j int) bool {
		if l[i].GPUs != l[j].GPUs {
			return l[i].GPUs < l[j].GPUs
		}

		if l[i].CPUs != l[j].CPUs {
			return l[i].CPUs < l[j].CPUs
		}

		return l[i].Memory < l[j].Memory
	})
}

type SupportedRonDBInstanceTypes struct {
	ManagementNode SupportedInstanceTypeList `json:"mgmd"`
	DataNode       SupportedInstanceTypeList `json:"ndbd"`
	MySQLNode      SupportedInstanceTypeList `json:"mysqld"`
	APINode        SupportedInstanceTypeList `json:"api"`
}

type SupportedInstanceTypes struct {
	Head   SupportedInstanceTypeList   `json:"head"`
	Worker SupportedInstanceTypeList   `json:"worker"`
	RonDB  SupportedRonDBInstanceTypes `json:"ronDB"`
}

func (s *SupportedInstanceTypes) GetByNodeType(nodeType NodeType) SupportedInstanceTypeList {
	switch nodeType {
	case HeadNode:
		return s.Head
	case WorkerNode:
		return s.Worker
	case RonDBManagementNode:
		return s.RonDB.ManagementNode
	case RonDBDataNode:
		return s.RonDB.DataNode
	case RonDBMySQLNode:
		return s.RonDB.MySQLNode
	case RonDBAPINode:
		return s.RonDB.APINode
	}
	return nil
}

type GetClustersResponse struct {
	BaseResponse
	Payload struct {
		Clusters []Cluster `json:"clusters"`
	} `json:"payload"`
}

type GetClusterResponse struct {
	BaseResponse
	Payload struct {
		Cluster Cluster `json:"cluster"`
	} `json:"payload"`
}

type NewClusterRequest struct {
	CloudProvider CloudProvider `json:"cloudProvider"`
	CreateRequest interface{}   `json:"cluster"`
}

type NewAWSClusterRequest struct {
	CloudProvider CloudProvider    `json:"cloudProvider"`
	CreateRequest CreateAWSCluster `json:"cluster"`
}

type NewAzureClusterRequest struct {
	CloudProvider CloudProvider      `json:"cloudProvider"`
	CreateRequest CreateAzureCluster `json:"cluster"`
}

type NewClusterResponse struct {
	BaseResponse
	Payload struct {
		Id string `json:"id"`
	} `json:"payload"`
}

type UpdateWorkersRequest struct {
	Workers []WorkerConfiguration `json:"workers"`
}

type ServiceOpenPorts struct {
	FeatureStore       bool `json:"featureStore"`
	OnlineFeatureStore bool `json:"onlineFeatureStore"`
	Kafka              bool `json:"kafka"`
	SSH                bool `json:"ssh"`
}

type UpdateOpenPortsRequest struct {
	Ports ServiceOpenPorts `json:"ports"`
}

type GetSupportedInstanceTypesResponse struct {
	BaseResponse
	Payload struct {
		AWS   SupportedInstanceTypes `json:"aws"`
		AZURE SupportedInstanceTypes `json:"azure"`
	} `json:"payload"`
}

type ConfigureAutoscaleRequest struct {
	Autoscale *AutoscaleConfiguration `json:"autoscale,omitempty"`
}

type BackupState string

const (
	PendingBackup      BackupState = "pending"
	InitializingBackup BackupState = "initializing"
	ProcessingBackup   BackupState = "processing"
	DeletingBackup     BackupState = "deleting"
	BackupSucceed      BackupState = "succeed"
	BackupFailed       BackupState = "failed"
	// local state not in Hopsworks.ai
	BackupDeleted BackupState = "tf-backup-deleted"
)

func (s BackupState) String() string {
	return string(s)
}

type Backup struct {
	Id            string        `json:"backupId"`
	Name          string        `json:"backupName"`
	ClusterId     string        `json:"clusterId"`
	CreatedOn     int64         `json:"createdOn"`
	CloudProvider CloudProvider `json:"cloudProvider"`
	State         BackupState   `json:"state"`
	StateMessage  string        `json:"stateMessage"`
}

type NewBackupRequest struct {
	Backup struct {
		ClusterId  string `json:"clusterId"`
		BackupName string `json:"backupName"`
	} `json:"backup"`
}

type NewBackupResponse struct {
	BaseResponse
	Payload struct {
		Id string `json:"backupId"`
	} `json:"payload"`
}

type GetBackupResponse struct {
	BaseResponse
	Payload struct {
		Backup Backup `json:"backup"`
	} `json:"payload"`
}

type GetBackupsResponse struct {
	BaseResponse
	Payload struct {
		Backups []Backup `json:"backups"`
	} `json:"payload"`
}

type CreateClusterFromBackup struct {
	Name       string                  `json:"name,omitempty"`
	SshKeyName string                  `json:"sshKeyName,omitempty"`
	Tags       []ClusterTag            `json:"tags,omitempty"`
	Autoscale  *AutoscaleConfiguration `json:"autoscale,omitempty"`
}

type CreateAzureClusterFromBackup struct {
	CreateClusterFromBackup
	NetworkResourceGroup string `json:"networkResourceGroup,omitempty"`
	VirtualNetworkName   string `json:"virtualNetworkName,omitempty"`
	SubnetName           string `json:"subnetName,omitempty"`
	SecurityGroupName    string `json:"securityGroupName,omitempty"`
}

type CreateAWSClusterFromBackup struct {
	CreateClusterFromBackup
	InstanceProfileArn string `json:"instanceProfileArn,omitempty"`
	VpcId              string `json:"vpcId,omitempty"`
	SubnetId           string `json:"subnetId,omitempty"`
	SecurityGroupId    string `json:"securityGroupId,omitempty"`
}

type NewClusterFromBackupRequest struct {
	CreateRequest interface{} `json:"cluster"`
}

type SupportedVersionRegions struct {
	Ubuntu []string `json:"ubuntu,omitempty"`
	CentOS []string `json:"centos,omitempty"`
}

type SupportedVersion struct {
	Version               string                  `json:"version"`
	UpgradableFromVersion string                  `json:"upgradableFromVersion"`
	Default               bool                    `json:"default"`
	Experimental          bool                    `json:"experimental"`
	Regions               SupportedVersionRegions `json:"regions"`
	ReleaseNotesUrl       string                  `json:"releaseNotesUrl"`
}

type GetSupportedVersionsResponse struct {
	BaseResponse
	Payload struct {
		Versions []SupportedVersion `json:"versions"`
	} `json:"payload"`
}

type UpgradeClusterRequest struct {
	Version string `json:"version"`
}

type NodeInfo struct {
	NodeType     string `json:"nodeType"`
	InstanceType string `json:"instanceType"`
}

type ModifyInstanceTypeRequest struct {
	NodeInfo NodeInfo `json:"nodeInfo"`
}

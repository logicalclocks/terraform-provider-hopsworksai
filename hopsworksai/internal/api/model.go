package api

import (
	"fmt"
	"net/http"
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

const Worker = "worker"

const (
	Starting           ClusterState = "starting"
	Pending            ClusterState = "pending"
	Initializing       ClusterState = "initializing"
	Running            ClusterState = "running"
	Stopping           ClusterState = "stopping"
	Stopped            ClusterState = "stopped"
	Error              ClusterState = "error"
	TerminationWarning ClusterState = "termination-warning"
	ShuttingDown       ClusterState = "shutting-down"
	Updating           ClusterState = "updating"
	Decommissioning    ClusterState = "decommissioning"
	// Worker states
	WorkerPending      ClusterState = Worker + "-" + Pending
	WorkerInitializing ClusterState = Worker + "-" + Initializing
	WorkerStarting     ClusterState = Worker + "-" + Starting
	WorkerError        ClusterState = Worker + "-" + Error
	WorkerShuttingdown ClusterState = Worker + "-" + ShuttingDown
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

type Cluster struct {
	Id                    string               `json:"id"`
	Name                  string               `json:"name"`
	State                 ClusterState         `json:"state"`
	ActivationState       ActivationState      `json:"activationState"`
	InitializationStage   string               `json:"initializationStage"`
	CreatedOn             int64                `json:"createdOn"`
	StartedOn             int64                `json:"startedOn"`
	Version               string               `json:"version"`
	URL                   string               `json:"url"`
	Provider              CloudProvider        `json:"provider"`
	ErrorMessage          string               `json:"errorMessage,omitempty"`
	Tags                  []ClusterTag         `json:"tags"`
	SshKeyName            string               `json:"sshKeyName"`
	ClusterConfiguration  ClusterConfiguration `json:"clusterConfiguration,omitempty"`
	PublicIPAttached      bool                 `json:"publicIPAttached"`
	LetsEncryptIssued     bool                 `json:"letsEncryptIssued"`
	ManagedUsers          bool                 `json:"managedUsers"`
	BackupRetentionPeriod int                  `json:"backupRetentionPeriod"`
	Azure                 AzureCluster         `json:"azure,omitempty"`
	AWS                   AWSCluster           `json:"aws,omitempty"`
	Ports                 ServiceOpenPorts     `json:"ports"`
}

func (c *Cluster) IsAWSCluster() bool {
	return c.Provider == AWS
}

func (c *Cluster) IsAzureCluster() bool {
	return c.Provider == AZURE
}

type AzureCluster struct {
	Location           string `json:"location"`
	ManagedIdentity    string `json:"managedIdentity"`
	ResourceGroup      string `json:"resourceGroup"`
	BlobContainerName  string `json:"blobContainerName"`
	StorageAccount     string `json:"storageAccount"`
	VirtualNetworkName string `json:"virtualNetworkName"`
	SubnetName         string `json:"subnetName"`
	SecurityGroupName  string `json:"securityGroupName"`
	AksClusterName     string `json:"aksClusterName"`
	AcrRegistryName    string `json:"acrRegistryName"`
}

type AWSCluster struct {
	Region               string `json:"region"`
	BucketName           string `json:"bucketName"`
	InstanceProfileArn   string `json:"instanceProfileArn"`
	VpcId                string `json:"vpcId"`
	SubnetId             string `json:"subnetId"`
	SecurityGroupId      string `json:"securityGroupId"`
	EksClusterName       string `json:"eksClusterName"`
	EcrRegistryAccountId string `json:"ecrRegistryAccountId"`
}

type NodeConfiguration struct {
	InstanceType string `json:"instanceType"`
	DiskSize     int    `json:"diskSize"`
}

type HeadConfiguration struct {
	NodeConfiguration
}

type WorkerConfiguration struct {
	NodeConfiguration
	Count int `json:"count"`
}

type ClusterConfiguration struct {
	Head    HeadConfiguration     `json:"head"`
	Workers []WorkerConfiguration `json:"workers"`
}

type ClusterTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CreateCluster struct {
	Name                  string               `json:"name"`
	Version               string               `json:"version"`
	SshKeyName            string               `json:"sshKeyName"`
	ClusterConfiguration  ClusterConfiguration `json:"clusterConfiguration"`
	IssueLetsEncrypt      bool                 `json:"issueLetsEncrypt"`
	AttachPublicIP        bool                 `json:"attachPublicIP"`
	BackupRetentionPeriod int                  `json:"backupRetentionPeriod"`
	ManagedUsers          bool                 `json:"managedUsers"`
	Tags                  []ClusterTag         `json:"tags"`
}

type CreateAzureCluster struct {
	CreateCluster
	AzureCluster
}

type CreateAWSCluster struct {
	CreateCluster
	AWSCluster
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

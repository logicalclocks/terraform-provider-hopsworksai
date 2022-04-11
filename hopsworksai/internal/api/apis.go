package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func GetClusters(ctx context.Context, apiClient APIHandler, cloud CloudProvider) ([]Cluster, error) {
	var response GetClustersResponse
	var filter string = ""
	if cloud != "" {
		filter = "?cloud=" + cloud.String()
	}
	if err := apiClient.doRequest(ctx, http.MethodGet, "/api/clusters"+filter, nil, &response); err != nil {
		return nil, err
	}
	return response.Payload.Clusters, nil
}

func NewCluster(ctx context.Context, apiClient APIHandler, createRequest interface{}) (string, error) {
	var cloudProvider CloudProvider
	switch createRequest.(type) {
	case CreateAzureCluster, *CreateAzureCluster:
		tflog.Debug(ctx, fmt.Sprintf("new azure cluster: %#v", createRequest))
		cloudProvider = AZURE
	case CreateAWSCluster, *CreateAWSCluster:
		tflog.Debug(ctx, fmt.Sprintf("new aws cluster: %#v", createRequest))
		cloudProvider = AWS
	default:
		return "", fmt.Errorf("unknown cloud provider %#v", createRequest)
	}
	req := NewClusterRequest{
		CreateRequest: createRequest,
		CloudProvider: cloudProvider,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %s", err)
	}

	var response NewClusterResponse
	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters", bytes.NewBuffer(payload), &response); err != nil {
		return "", err
	}

	return response.Payload.Id, nil
}

func GetCluster(ctx context.Context, apiClient APIHandler, clusterId string) (*Cluster, error) {
	var response GetClusterResponse
	if err := apiClient.doRequest(ctx, http.MethodGet, "/api/clusters/"+clusterId, nil, &response); err != nil {
		return nil, err
	}
	if response.Code == http.StatusNotFound {
		tflog.Debug(ctx, fmt.Sprintf("cluster (id: %s) is not found", clusterId))
		return nil, nil
	}
	return &response.Payload.Cluster, nil
}

func DeleteCluster(ctx context.Context, apiClient APIHandler, clusterId string) error {
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodDelete, "/api/clusters/"+clusterId, nil, &response); err != nil {
		return err
	}
	return nil
}

func StopCluster(ctx context.Context, apiClient APIHandler, clusterId string) error {
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodPut, "/api/clusters/"+clusterId+"/stop", nil, &response); err != nil {
		return err
	}
	return nil
}

func StartCluster(ctx context.Context, apiClient APIHandler, clusterId string) error {
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodPut, "/api/clusters/"+clusterId+"/start", nil, &response); err != nil {
		return err
	}
	return nil
}

func AddWorkers(ctx context.Context, apiClient APIHandler, clusterId string, toAdd []WorkerConfiguration) error {
	if len(toAdd) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("skip update cluster %s due to no updates", clusterId))
		return nil
	}
	req := UpdateWorkersRequest{
		Workers: toAdd,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	var response BaseResponse

	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters/"+clusterId+"/workers", bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

func RemoveWorkers(ctx context.Context, apiClient APIHandler, clusterId string, toRemove []WorkerConfiguration) error {
	if len(toRemove) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("skip update cluster %s due to no updates", clusterId))
		return nil
	}
	req := UpdateWorkersRequest{
		Workers: toRemove,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	var response BaseResponse

	if err := apiClient.doRequest(ctx, http.MethodDelete, "/api/clusters/"+clusterId+"/workers", bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

func UpdateOpenPorts(ctx context.Context, apiClient APIHandler, clusterId string, ports *ServiceOpenPorts) error {
	req := UpdateOpenPortsRequest{
		Ports: *ports,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	var response BaseResponse

	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters/"+clusterId+"/ports", bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

func GetSupportedInstanceTypes(ctx context.Context, apiClient APIHandler, cloud CloudProvider) (*SupportedInstanceTypes, error) {
	var response GetSupportedInstanceTypesResponse
	if err := apiClient.doRequest(ctx, http.MethodGet, "/api/clusters/nodes/supported-types?cloud="+cloud.String(), nil, &response); err != nil {
		return nil, err
	}
	if cloud == AWS {
		return &response.Payload.AWS, nil
	} else if cloud == AZURE {
		return &response.Payload.AZURE, nil
	}
	return nil, fmt.Errorf("unknown cloud provider %s", cloud.String())
}

func ConfigureAutoscale(ctx context.Context, apiClient APIHandler, clusterId string, config *AutoscaleConfiguration) error {
	req := ConfigureAutoscaleRequest{
		Autoscale: config,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters/"+clusterId+"/autoscale", bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

func DisableAutoscale(ctx context.Context, apiClient APIHandler, clusterId string) error {
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodDelete, "/api/clusters/"+clusterId+"/autoscale", nil, &response); err != nil {
		return err
	}
	return nil
}

func NewBackup(ctx context.Context, apiClient APIHandler, clusterId string, backupName string) (string, error) {
	req := NewBackupRequest{}
	req.Backup.ClusterId = clusterId
	req.Backup.BackupName = backupName
	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %s", err)
	}

	var response NewBackupResponse
	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/backups", bytes.NewBuffer(payload), &response); err != nil {
		return "", err
	}

	return response.Payload.Id, nil
}

func GetBackup(ctx context.Context, apiClient APIHandler, backupId string) (*Backup, error) {
	var response GetBackupResponse
	if err := apiClient.doRequest(ctx, http.MethodGet, "/api/backups/"+backupId, nil, &response); err != nil {
		return nil, err
	}
	if response.Code == http.StatusNotFound {
		tflog.Debug(ctx, fmt.Sprintf("backup (id: %s) is not found", backupId))
		return nil, nil
	}
	return &response.Payload.Backup, nil
}

func DeleteBackup(ctx context.Context, apiClient APIHandler, backupId string) error {
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodDelete, "/api/backups/"+backupId, nil, &response); err != nil {
		return err
	}
	return nil
}

func GetBackups(ctx context.Context, apiClient APIHandler, clusterId string) ([]Backup, error) {
	var filter = ""
	if clusterId != "" {
		filter = "?clusterId=" + clusterId
	}
	var response GetBackupsResponse
	if err := apiClient.doRequest(ctx, http.MethodGet, "/api/backups"+filter, nil, &response); err != nil {
		return nil, err
	}
	return response.Payload.Backups, nil
}

func NewClusterFromBackup(ctx context.Context, apiClient APIHandler, backupId string, createRequest interface{}) (string, error) {
	switch createRequest.(type) {
	case CreateAWSClusterFromBackup, *CreateAWSClusterFromBackup:
		tflog.Debug(ctx, fmt.Sprintf("restore aws cluster: #%v", createRequest))
	case CreateAzureClusterFromBackup, *CreateAzureClusterFromBackup:
		tflog.Debug(ctx, fmt.Sprintf("restore azure cluster: #%v", createRequest))
	default:
		return "", fmt.Errorf("unknown create request #%v", createRequest)
	}
	req := NewClusterFromBackupRequest{
		CreateRequest: createRequest,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %s", err)
	}
	var response NewClusterResponse
	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters/restore/"+backupId, bytes.NewBuffer(payload), &response); err != nil {
		return "", err
	}
	return response.Payload.Id, nil
}

func GetSupportedVersions(ctx context.Context, apiClient APIHandler, cloud CloudProvider) ([]SupportedVersion, error) {
	var response GetSupportedVersionsResponse
	if err := apiClient.doRequest(ctx, http.MethodGet, "/api/clusters/hopsworks/versions/"+cloud.String(), nil, &response); err != nil {
		return nil, err
	}
	return response.Payload.Versions, nil
}

func UpgradeCluster(ctx context.Context, apiClient APIHandler, clusterId string, upgradeToVersion string) error {
	req := UpgradeClusterRequest{
		Version: upgradeToVersion,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters/"+clusterId+"/upgrade", bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

func RollbackUpgradeCluster(ctx context.Context, apiClient APIHandler, clusterId string) error {
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodPut, "/api/clusters/"+clusterId+"/upgrade/rollback", nil, &response); err != nil {
		return err
	}
	return nil
}

func ModifyInstanceType(ctx context.Context, apiClient APIHandler, clusterId string, nodeType NodeType, instanceType string) error {
	if nodeType == WorkerNode || nodeType == RonDBManagementNode {
		return fmt.Errorf("modifying instance type for %s is not supported", nodeType.String())
	}
	req := ModifyInstanceTypeRequest{
		NodeInfo: NodeInfo{
			NodeType:     nodeType.String(),
			InstanceType: instanceType,
		},
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}
	var response BaseResponse
	if err := apiClient.doRequest(ctx, http.MethodPut, "/api/clusters/"+clusterId+"/nodes/modify-instance-type", bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

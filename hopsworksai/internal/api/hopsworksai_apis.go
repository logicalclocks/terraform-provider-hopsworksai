package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		log.Printf("[DEBUG] new azure cluster: #%v", createRequest)
		cloudProvider = AZURE
	case CreateAWSCluster, *CreateAWSCluster:
		log.Printf("[DEBUG] new aws cluster: #%v", createRequest)
		cloudProvider = AWS
	default:
		return "", fmt.Errorf("unknown cloud provider #%v", createRequest)
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
		log.Printf("[DEBUG] cluster (id: %s) is not found", clusterId)
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

func UpdateCluster(ctx context.Context, apiClient APIHandler, clusterId string, toAdd []WorkerConfiguration, toRemove []WorkerConfiguration) error {
	req := UpdateClusterRequest{}
	req.UpdateRequest.Workers = UpdateWorkers{
		Add:    toAdd,
		Remove: toRemove,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	var response BaseResponse

	if err := apiClient.doRequest(ctx, http.MethodPost, "/api/clusters/"+clusterId, bytes.NewBuffer(payload), &response); err != nil {
		return err
	}
	return nil
}

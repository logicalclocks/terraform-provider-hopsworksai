package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("HOPSWORKSAI_API_KEY", ""),
				},
				"api_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Used for development",
					DefaultFunc: schema.EnvDefaultFunc("HOPSWORKSAI_API_HOST", "https://www.hopsworks.ai/"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"hopsworksai_data_source":         dataSourceHopsworksAI(),
				"hopsworksai_cluster_data_source": dataSourceClusterHopsworksAI(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"hopsworksai_azure_cluster": azureClusterResource(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	client    *http.Client
	host      string
	userAgent string
	apiKey    string
}

type InstancesReponse struct {
	Payload struct {
		Instances []Instance
	}
}

type InstanceReponse struct {
	Payload struct {
		InstanceData Instance
	}
}

type Instance struct {
	InstanceID          string
	InstanceName        string
	State               string
	InitializationStage string
	ActivationStage     string
}

type InstanceConfiguration struct {
	InstanceType string
	VolumeSize   int
}

type InstanceTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type InstancePorts struct {
	FeatureStore       string `json:"featureStore"`
	Kafka              string `json:"kafka"`
	OnlineFeatureStore string `json:"onlineFeatureStore"`
	SSH                string `json:"ssh"`
}

type InstanceUpdateRequest struct {
	WorkerConfiguration map[string]InstanceConfiguration `json:"instancesConfiguration"`
	NBNodes             int                              `json:"nbNodes"`
}

type InstanceRequest struct {
	BucketName                string                           `json:"bucketName"`
	DeleteBlocksRetentionDays int                              `json:"deletedBlocksRetentionDays"`
	InstanceName              string                           `json:"instanceName"`
	InstanceProfile           string                           `json:"instanceProfile"`
	InstanceConfiguration     map[string]InstanceConfiguration `json:"instancesConfiguration"`
	InstanceTags              []InstanceTag                    `json:"instanceTags"`
	IssueLetsEncrypt          bool                             `json:"issueLetsEncrypt"`
	KeyName                   string                           `json:"keyName"`
	ManagedUsers              bool                             `json:"managedUsers"`
	NBNodes                   int                              `json:"nbNodes"`
	Ports                     InstancePorts                    `json:"ports"`
	Region                    string                           `json:"region"`
	ResourceGroup             string                           `json:"resourceGroup"`
	StorageName               string                           `json:"storageName"`
	Version                   string                           `json:"version"`
}

func (a *apiClient) doRequest(method string, endpoint string, body io.Reader, result interface{}) error {
	req, err := http.NewRequest(method, a.host+endpoint, body)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	defer resp.Body.Close()

	log.Printf("resp: %#v", resp)

	if resp.StatusCode == http.StatusForbidden {
		bodyBytes, respErr := ioutil.ReadAll(resp.Body)
		if respErr != nil {
			return fmt.Errorf("the API token provided does not have access to hopsworks.ai, verify the token you specified matches the token hopsworks.ai created")
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("the API token provided does not have access to hopsworks.ai, verify the token you specified matches the token hopsworks.ai created, message from hops: %s", bodyString)
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("failed to decode json, resp: %s, path: %s err: %s", resp.Status, a.host+endpoint, err)
	}
	return nil
}

func (a *apiClient) GetInstances() (*InstancesReponse, error) {
	var payload InstancesReponse

	err := a.doRequest(http.MethodGet, "/api/instances", nil, &payload)
	if err != nil {
		return nil, fmt.Errorf("request failed with: %s", err)
	}

	return &payload, err
}

func (a *apiClient) GetInstance(id string) (*InstanceReponse, error) {
	var payload InstanceReponse

	err := a.doRequest(http.MethodGet, "/api/instances/"+id, nil, &payload)
	if err != nil {
		return nil, fmt.Errorf("request failed with: %s", err)
	}

	return &payload, err
}

func (a *apiClient) CreateInstance(instanceRequest *InstanceRequest, provider string) error {
	payload, err := json.Marshal(instanceRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	log.Printf("[INFO] req: %s", string(payload))

	m := make(map[string]interface{}, 0)

	return a.doRequest(http.MethodPost, "/api/instances/"+provider+"/create", bytes.NewBuffer(payload), &m)
}

func (a *apiClient) AddNodes(instanceRequest *InstanceUpdateRequest, instanceID string) error {
	payload, err := json.Marshal(instanceRequest)

	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	log.Printf("[INFO] req: %s", string(payload))

	m := make(map[string]interface{}, 0)

	return a.doRequest(http.MethodPut, "/api/instances/"+instanceID+"/addNodes", bytes.NewBuffer(payload), &m)
}

func (a *apiClient) RemoveNodes(instanceRequest *InstanceUpdateRequest, instanceID string) error {
	payload, err := json.Marshal(instanceRequest)

	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	log.Printf("[INFO] req: %s", string(payload))

	m := make(map[string]interface{}, 0)

	return a.doRequest(http.MethodPost, "/api/instances/"+instanceID+"/removeNodes", bytes.NewBuffer(payload), &m)
}

func (a *apiClient) DeleteInstance(instanceID string) error {
	log.Printf("[INFO] delete instance %s", instanceID)

	m := make(map[string]interface{}, 0)

	return a.doRequest(http.MethodDelete, "/api/instances/"+instanceID, nil, &m)
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &apiClient{
			userAgent: p.UserAgent("terraform-provider-hopsworksai", version),
			host:      d.Get("api_host").(string),
			apiKey:    d.Get("api_key").(string),
			client: &http.Client{
				Timeout: time.Second * 30,
			},
		}, nil
	}
}

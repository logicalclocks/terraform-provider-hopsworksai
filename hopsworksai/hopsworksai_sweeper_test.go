package hopsworksai

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedClient() (*api.HopsworksAIClient, error) {
	return &api.HopsworksAIClient{
		UserAgent:  "Terraform Acceptance Tests",
		ApiKey:     os.Getenv("HOPSWORKSAI_API_KEY"),
		ApiVersion: Default_API_VERSION,
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}, nil
}

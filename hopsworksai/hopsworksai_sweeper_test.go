package hopsworksai

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func hopsworksClient() *api.HopsworksAIClient {
	return &api.HopsworksAIClient{
		UserAgent:  "Terraform Acceptance Tests",
		ApiKey:     os.Getenv(env_API_KEY),
		ApiVersion: Default_API_VERSION,
		ApiGateway: api.DEFAULT_API_GATEWAY,
		Client: &http.Client{
			Timeout: time.Minute * 3,
		},
	}
}

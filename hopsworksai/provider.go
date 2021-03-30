package hopsworksai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

const Default_API_VERSION = "v1"

func init() {
	schema.DescriptionKind = schema.StringMarkdown
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil && s.Default != "" {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
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
				"hopsworksai_clusters": dataSourceClusters(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"hopsworksai_cluster": clusterResource(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &api.HopsworksAIClient{
			UserAgent:  p.UserAgent("terraform-provider-hopsworksai", version),
			Host:       d.Get("api_host").(string),
			ApiKey:     d.Get("api_key").(string),
			ApiVersion: Default_API_VERSION,
			Client: &http.Client{
				Timeout: time.Second * 30,
			},
		}, nil
	}
}
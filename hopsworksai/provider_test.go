package hopsworksai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

const (
	env_API_KEY = "HOPSWORKSAI_API_KEY"

	env_TEST_RUN_SUFFIX          = "TF_HOPSWORKSAI_TEST_SUFFIX"
	env_AWS_SKIP                 = "TF_HOPSWORKSAI_AWS_SKIP"
	env_AWS_REGION               = "TF_HOPSWORKSAI_AWS_REGION"
	env_AWS_SSH_KEY              = "TF_HOPSWORKSAI_AWS_SSH_KEY"
	env_AWS_INSTANCE_PROFILE_ARN = "TF_HOPSWORKSAI_AWS_INSTANCE_PROFILE_ARN"
	env_AWS_BUCKET_NAMES         = "TF_HOPSWORKSAI_AWS_BUCKET_NAMES"
	env_AWS_VPC_ID               = "TF_HOPSWORKSAI_AWS_VPC_ID"
	env_AWS_SUBNET_ID            = "TF_HOPSWORKSAI_AWS_SUBNET_ID"
	env_AWS_SECURITY_GROUP_ID    = "TF_HOPSWORKSAI_AWS_SECURITY_GROUP_ID"

	env_AZURE_SKIP                        = "TF_HOPSWORKSAI_AZURE_SKIP"
	env_AZURE_LOCATION                    = "TF_HOPSWORKSAI_AZURE_LOCATION"
	env_AZURE_RESOURCE_GROUP              = "TF_HOPSWORKSAI_AZURE_RESOURCE_GROUP"
	env_AZURE_STORAGE_ACCOUNT             = "TF_HOPSWORKSAI_AZURE_STORAGE_ACCOUNT_NAME"
	env_AZURE_USER_ASSIGNED_IDENTITY_NAME = "TF_HOPSWORKSAI_AZURE_USER_ASSIGNED_IDENTITY_NAME"
	env_AZURE_SSH_KEY                     = "TF_HOPSWORKSAI_AZURE_SSH_KEY"
	env_AZURE_VIRTUAL_NETWORK_NAME        = "TF_HOPSWORKSAI_AZURE_VIRTUAL_NETWORK_NAME"
	env_AZURE_SUBNET_NAME                 = "TF_HOPSWORKSAI_AZURE_SUBNET_NAME"
	env_AZURE_SECURITY_GROUP_NAME         = "TF_HOPSWORKSAI_AZURE_SECURITY_GROUP_NAME"
	env_AZURE_ACR_REGISTRY_NAME           = "TF_HOPSWORKSAI_AZURE_ACR_REGISTRY_NAME"

	num_AWS_BUCKETS_NEEDED = 11
)

const (
	default_CLUSTER_NAME_PREFIX = "tfacctest"
	default_CLUSTER_TAG_KEY     = "Purpose"
	default_CLUSTER_TAG_VALUE   = "acceptance-test"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"hopsworksai": func() (*schema.Provider, error) { return Provider("dev")(), nil }, //nolint:unparam
	}
}

func parallelTest(t *testing.T, cloud api.CloudProvider, test resource.TestCase) {
	if cloud == api.AWS {
		if isAWSAccSkipped() {
			t.Skipf("Skipping %s test as %s is set", t.Name(), env_AWS_SKIP)
		}
	} else if cloud == api.AZURE {
		if isAzureAccSkipped() {
			t.Skipf("Skipping %s test as %s is set", t.Name(), env_AZURE_SKIP)
		}
	}
	resource.ParallelTest(t, test)
}

func testAccPreCheck(t *testing.T) func() {
	return func() {
		testCheckEnv(t, "", env_API_KEY)
		testCheckEnv(t, "", env_TEST_RUN_SUFFIX)
		if !isAWSAccSkipped() {
			testCheckEnv(t, fmt.Sprintf("You can skip AWS tests by setting %s=true", env_AWS_SKIP),
				env_AWS_REGION,
				env_AWS_SSH_KEY,
				env_AWS_BUCKET_NAMES,
				env_AWS_INSTANCE_PROFILE_ARN,
				env_AWS_VPC_ID,
				env_AWS_SUBNET_ID,
				env_AWS_SECURITY_GROUP_ID)

			buckets := strings.Split(os.Getenv(env_AWS_BUCKET_NAMES), ",")
			if len(buckets) < num_AWS_BUCKETS_NEEDED {
				t.Fatalf("Incorrect number of buckets expected %d but got %d. Each AWS test case that create a cluster requires and empty bucket.", num_AWS_BUCKETS_NEEDED, len(buckets))
			}
		}

		if !isAzureAccSkipped() {
			testCheckEnv(t, fmt.Sprintf("You can skip AZURE tests by setting %s=true", env_AZURE_SKIP),
				env_AZURE_LOCATION,
				env_AZURE_RESOURCE_GROUP,
				env_AZURE_SSH_KEY,
				env_AZURE_STORAGE_ACCOUNT,
				env_AZURE_USER_ASSIGNED_IDENTITY_NAME,
				env_AZURE_VIRTUAL_NETWORK_NAME,
				env_AZURE_SUBNET_NAME,
				env_AZURE_SECURITY_GROUP_NAME,
				env_AZURE_ACR_REGISTRY_NAME)
		}
	}
}

func testCheckEnv(t *testing.T, msg string, envVars ...string) {
	for _, envVar := range envVars {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("Environment variable %s is not set. %s", envVar, msg)
		}
	}
}

func testAccClusterCloudSSHKeyAttribute(cloud api.CloudProvider) string {
	if cloud == api.AWS {
		return os.Getenv(env_AWS_SSH_KEY)
	} else if cloud == api.AZURE {
		return os.Getenv(env_AZURE_SSH_KEY)
	}
	return ""
}

func isAWSAccSkipped() bool {
	return os.Getenv(env_AWS_SKIP) == "true"
}

func isAzureAccSkipped() bool {
	return os.Getenv(env_AZURE_SKIP) == "true"
}

func testAccClusterCloudConfigAttributes(cloud api.CloudProvider, bucketIndex int, setNetwork bool) string {
	if cloud == api.AWS {
		bucketNames := strings.Split(os.Getenv(env_AWS_BUCKET_NAMES), ",")
		if bucketIndex >= len(bucketNames) {
			if os.Getenv(resource.EnvTfAcc) == "" || isAWSAccSkipped() {
				return ""
			} else {
				panic(fmt.Errorf("bucket index is out of range index: %d list: %#v", bucketIndex, bucketNames))
			}
		}
		bucketName := bucketNames[bucketIndex]
		baseConfig := fmt.Sprintf(`
			region               = "%s"
			instance_profile_arn = "%s"
			bucket {
				name = "%s"
			}
		`, os.Getenv(env_AWS_REGION), os.Getenv(env_AWS_INSTANCE_PROFILE_ARN), bucketName)

		var networkConfig = ""
		if setNetwork {
			networkConfig = fmt.Sprintf(`
				network {
					vpc_id = "%s"
					subnet_id = "%s"
					security_group_id = "%s"
				}
			`, os.Getenv(env_AWS_VPC_ID), os.Getenv(env_AWS_SUBNET_ID), os.Getenv(env_AWS_SECURITY_GROUP_ID))
		}
		return fmt.Sprintf(`
		aws_attributes {
			%s
			%s
		}
		`, baseConfig, networkConfig)
	} else if cloud == api.AZURE {
		baseConfig := fmt.Sprintf(`
			location                       = "%s"
			resource_group                 = "%s"
			container {
				storage_account            = "%s"
			}
			user_assigned_managed_identity = "%s"
			acr_registry_name              = "%s"
		`, os.Getenv(env_AZURE_LOCATION), os.Getenv(env_AZURE_RESOURCE_GROUP), os.Getenv(env_AZURE_STORAGE_ACCOUNT), os.Getenv(env_AZURE_USER_ASSIGNED_IDENTITY_NAME), os.Getenv(env_AZURE_ACR_REGISTRY_NAME))
		var networkConfig = ""
		if setNetwork {
			networkConfig = fmt.Sprintf(`
				network {
					virtual_network_name = "%s"
					subnet_name = "%s"
					security_group_name = "%s"
				}
			`, os.Getenv(env_AZURE_VIRTUAL_NETWORK_NAME), os.Getenv(env_AZURE_SUBNET_NAME), os.Getenv(env_AZURE_SECURITY_GROUP_NAME))
		}
		return fmt.Sprintf(`
		azure_attributes {
			%s
			%s
		}
		`, baseConfig, networkConfig)
	}
	return ""
}

func testAccDefaultTags() string {
	return fmt.Sprintf(`
		"Run" = "%s"
		"%s" = "%s"
	`, os.Getenv(env_TEST_RUN_SUFFIX), default_CLUSTER_TAG_KEY, default_CLUSTER_TAG_VALUE)
}

func testAccResourceDataSourceCheckAllAttributes(resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		for k := range rs.Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, k)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func TestProviderAPIGateway(t *testing.T) {
	testCases := []struct {
		config              *terraform.ResourceConfig
		expectedGateway     string
		expectedDiagnostics diag.Diagnostics
	}{
		{
			config: terraform.NewResourceConfigRaw(map[string]interface{}{
				"api_gateway": "https://api.gateway.ai",
			}),
			expectedGateway: "https://api.gateway.ai",
			expectedDiagnostics: diag.Diagnostics{
				ApiGatewayDevDiagnostic,
			},
		},

		{
			config:          terraform.NewResourceConfigRaw(map[string]interface{}{}),
			expectedGateway: api.DEFAULT_API_GATEWAY,
		},

		{
			config: terraform.NewResourceConfigRaw(map[string]interface{}{
				"api_gateway": "api.gateway.ai",
			}),
			expectedDiagnostics: diag.Diagnostics{
				ApiGatewayDevDiagnostic,
				{
					Severity: diag.Error,
					Summary:  "API Gateway URL is missing scheme http/https",
				},
			},
		},

		{
			config: terraform.NewResourceConfigRaw(map[string]interface{}{
				"api_gateway": ":foo",
			}),
			expectedDiagnostics: diag.Diagnostics{
				ApiGatewayDevDiagnostic,
				{
					Severity: diag.Error,
					Summary:  "Could not parse API Gateway URL",
				},
			},
		},
	}

	var provider *schema.Provider
	for _, test := range testCases {
		provider = Provider("dev")()
		diagnostics := provider.Validate(test.config)
		apiGatewayInTest, found := test.config.Get("api_gateway")
		if !found {
			apiGatewayInTest = api.DEFAULT_API_GATEWAY
		}
		t.Logf("Testing case api_gateway: %s", apiGatewayInTest)
		if len(test.expectedDiagnostics) > 0 {
			if test.expectedDiagnostics.HasError() && !diagnostics.HasError() {
				t.Fatalf("Expected Validate to return Diagnostic error %s but it did not",
					diagnosticsSummary(test.expectedDiagnostics))
			}
			if !validateDiagnosticErrors(test.expectedDiagnostics, diagnostics) {
				t.Fatalf("Expected to find Diagnostics summary \"%s\" in Validation but found \"%s\"",
					diagnosticsSummary(test.expectedDiagnostics), diagnosticsSummary(diagnostics))
			}
		} else {
			provider.Configure(context.Background(), test.config)
			c, ok := provider.Meta().(*api.HopsworksAIClient)
			if !ok {
				t.Error("client is not HopsworksAIClient")
			}
			if c.ApiGateway != test.expectedGateway {
				t.Errorf("Expected API Gateway to be %s but it is %s", test.expectedGateway, c.ApiGateway)
			}
		}
	}
}

func diagnosticsSummary(diagnostics diag.Diagnostics) string {
	var summary strings.Builder
	for i, d := range diagnostics {
		summary.WriteString(d.Summary)
		if i < len(diagnostics)-1 {
			summary.WriteString(" - ")
		}
	}
	return summary.String()
}

func validateDiagnosticErrors(expected, actual diag.Diagnostics) bool {
	if len(expected) != len(actual) {
		return false
	}
	for _, expectedDiag := range expected {
		var diagnosticFound bool
		for _, actualDiag := range actual {
			if expectedDiag.Summary == actualDiag.Summary {
				diagnosticFound = true
				break
			}
		}
		if !diagnosticFound {
			return false
		}
	}
	return true
}

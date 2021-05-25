package hopsworksai

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	env_API_KEY = "HOPSWORKSAI_API_KEY"

	env_AWS_SKIP                 = "TF_HOPSWORKSAI_AWS_SKIP"
	env_AWS_REGION               = "TF_HOPSWORKSAI_AWS_REGION"
	env_AWS_SSH_KEY              = "TF_HOPSWORKSAI_AWS_SSH_KEY"
	env_AWS_INSTANCE_PROFILE_ARN = "TF_HOPSWORKSAI_AWS_INSTANCE_PROFILE_ARN"
	env_AWS_BUCKET_NAME          = "TF_HOPSWORKSAI_AWS_BUCKET_NAME"

	env_AZURE_SKIP                        = "TF_HOPSWORKSAI_AZURE_SKIP"
	env_AZURE_LOCATION                    = "TF_HOPSWORKSAI_AZURE_LOCATION"
	env_AZURE_RESOURCE_GROUP              = "TF_HOPSWORKSAI_AZURE_RESOURCE_GROUP"
	env_AZURE_STORAGE_ACCOUNT             = "TF_HOPSWORKSAI_AZURE_STORAGE_ACCOUNT_NAME"
	env_AZURE_USER_ASSIGNED_IDENTITY_NAME = "TF_HOPSWORKSAI_AZURE_USER_ASSIGNED_IDENTITY_NAME"
	env_AZURE_SSH_KEY                     = "TF_HOPSWORKSAI_AZURE_SSH_KEY"
)

const clusterPrefixName = "tfacctest"

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider("0.1.0")()
	testAccProviders = map[string]*schema.Provider{
		"hopsworksai": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) func() {
	return func() {
		testCheckEnv(t, "", env_API_KEY)

		if v := os.Getenv(env_AWS_SKIP); v != "true" {
			testCheckEnv(t, fmt.Sprintf("You can skip AWS tests by setting %s=true", env_AWS_SKIP),
				env_AWS_REGION,
				env_AWS_SSH_KEY,
				env_AWS_BUCKET_NAME,
				env_AWS_INSTANCE_PROFILE_ARN)
		}

		if v := os.Getenv(env_AZURE_SKIP); v != "true" {
			testCheckEnv(t, fmt.Sprintf("You can skip AZURE tests by setting %s=true", env_AZURE_SKIP),
				env_AZURE_LOCATION,
				env_AZURE_RESOURCE_GROUP,
				env_AZURE_SSH_KEY,
				env_AZURE_STORAGE_ACCOUNT,
				env_AZURE_USER_ASSIGNED_IDENTITY_NAME)
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

func testSkipAWS(t *testing.T) {
	if v := os.Getenv(env_AWS_SKIP); v == "true" {
		t.Skip(fmt.Sprintf("Skipping %s test as %s is set", t.Name(), env_AWS_SKIP))
	}
}

func testSkipAZURE(t *testing.T) {
	if v := os.Getenv(env_AZURE_SKIP); v == "true" {
		t.Skip(fmt.Sprintf("Skipping %s test as %s is set", t.Name(), env_AZURE_SKIP))
	}
}

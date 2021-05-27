package hopsworksai

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

const (
	env_API_KEY = "HOPSWORKSAI_API_KEY"

	env_AWS_SKIP                 = "TF_HOPSWORKSAI_AWS_SKIP"
	env_AWS_REGION               = "TF_HOPSWORKSAI_AWS_REGION"
	env_AWS_SSH_KEY              = "TF_HOPSWORKSAI_AWS_SSH_KEY"
	env_AWS_INSTANCE_PROFILE_ARN = "TF_HOPSWORKSAI_AWS_INSTANCE_PROFILE_ARN"
	env_AWS_BUCKET_NAMES         = "TF_HOPSWORKSAI_AWS_BUCKET_NAMES"

	env_AZURE_SKIP                        = "TF_HOPSWORKSAI_AZURE_SKIP"
	env_AZURE_LOCATION                    = "TF_HOPSWORKSAI_AZURE_LOCATION"
	env_AZURE_RESOURCE_GROUP              = "TF_HOPSWORKSAI_AZURE_RESOURCE_GROUP"
	env_AZURE_STORAGE_ACCOUNT             = "TF_HOPSWORKSAI_AZURE_STORAGE_ACCOUNT_NAME"
	env_AZURE_USER_ASSIGNED_IDENTITY_NAME = "TF_HOPSWORKSAI_AZURE_USER_ASSIGNED_IDENTITY_NAME"
	env_AZURE_SSH_KEY                     = "TF_HOPSWORKSAI_AZURE_SSH_KEY"

	num_AWS_BUCKETS_NEEDED = 3
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

func parallelTest(t *testing.T, cloud api.CloudProvider, test resource.TestCase) {
	if cloud == api.AWS {
		if isAWSAccSkipped() {
			t.Skip(fmt.Sprintf("Skipping %s test as %s is set", t.Name(), env_AWS_SKIP))
		}
	} else if cloud == api.AZURE {
		if isAzureAccSkipped() {
			t.Skip(fmt.Sprintf("Skipping %s test as %s is set", t.Name(), env_AZURE_SKIP))
		}
	}
	resource.ParallelTest(t, test)
}

func testAccPreCheck(t *testing.T) func() {
	return func() {
		testCheckEnv(t, "", env_API_KEY)

		if !isAWSAccSkipped() {
			testCheckEnv(t, fmt.Sprintf("You can skip AWS tests by setting %s=true", env_AWS_SKIP),
				env_AWS_REGION,
				env_AWS_SSH_KEY,
				env_AWS_BUCKET_NAMES,
				env_AWS_INSTANCE_PROFILE_ARN)

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

func testAccClusterCloudConfigAttributes(cloud api.CloudProvider, bucketIndex int) string {
	if cloud == api.AWS {
		bucketNames := strings.Split(os.Getenv(env_AWS_BUCKET_NAMES), ",")
		if bucketIndex >= len(bucketNames) {
			if os.Getenv(resource.TestEnvVar) == "" || isAWSAccSkipped() {
				return ""
			} else {
				panic(fmt.Errorf("bucket index is out of range index: %d list: %#v", bucketIndex, bucketNames))
			}
		}
		bucketName := bucketNames[bucketIndex]
		return fmt.Sprintf(`
		aws_attributes {
			region               = "%s"
			instance_profile_arn = "%s"
			bucket_name          = "%s"
		  }
		`, os.Getenv(env_AWS_REGION), os.Getenv(env_AWS_INSTANCE_PROFILE_ARN), bucketName)
	} else if cloud == api.AZURE {
		return fmt.Sprintf(`
		azure_attributes {
			location                       = "%s"
			resource_group                 = "%s"
			storage_account                = "%s"
			user_assigned_managed_identity = "%s"
		  }
		`, os.Getenv(env_AZURE_LOCATION), os.Getenv(env_AZURE_RESOURCE_GROUP), os.Getenv(env_AZURE_STORAGE_ACCOUNT), os.Getenv(env_AZURE_USER_ASSIGNED_IDENTITY_NAME))
	}
	return ""
}

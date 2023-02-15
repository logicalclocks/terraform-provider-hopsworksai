package hopsworksai

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSInstanceProfilePolicy_basic(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/*"
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:*:repository/*/filebeat",
		"arn:aws:ecr:*:*:repository/*/base",
		"arn:aws:ecr:*:*:repository/*/onlinefs",
		"arn:aws:ecr:*:*:repository/*/airflow",
		"arn:aws:ecr:*:*:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:822623301872:repository/filebeat",
		"arn:aws:ecr:*:822623301872:repository/base",
		"arn:aws:ecr:*:822623301872:repository/onlinefs",
		"arn:aws:ecr:*:822623301872:repository/airflow",
		"arn:aws:ecr:*:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsEKSPermissions(allowDescribeEKSResource)...)
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_eks_restriction(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/cluster_name"
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:*:repository/*/filebeat",
		"arn:aws:ecr:*:*:repository/*/base",
		"arn:aws:ecr:*:*:repository/*/onlinefs",
		"arn:aws:ecr:*:*:repository/*/airflow",
		"arn:aws:ecr:*:*:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:822623301872:repository/filebeat",
		"arn:aws:ecr:*:822623301872:repository/base",
		"arn:aws:ecr:*:822623301872:repository/onlinefs",
		"arn:aws:ecr:*:822623301872:repository/airflow",
		"arn:aws:ecr:*:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsEKSPermissions(allowDescribeEKSResource)...)
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_eks_restriction(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_cluster_id(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/*"
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:*:repository/cluster_id/filebeat",
		"arn:aws:ecr:*:*:repository/cluster_id/base",
		"arn:aws:ecr:*:*:repository/cluster_id/onlinefs",
		"arn:aws:ecr:*:*:repository/cluster_id/airflow",
		"arn:aws:ecr:*:*:repository/cluster_id/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:822623301872:repository/filebeat",
		"arn:aws:ecr:*:822623301872:repository/base",
		"arn:aws:ecr:*:822623301872:repository/onlinefs",
		"arn:aws:ecr:*:822623301872:repository/airflow",
		"arn:aws:ecr:*:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsEKSPermissions(allowDescribeEKSResource)...)
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_cluster_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_singleBucket(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions([]string{"arn:aws:s3:::test/*", "arn:aws:s3:::test"}),
			awsBackupPermissions([]string{"arn:aws:s3:::test/*", "arn:aws:s3:::test"}),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/*"
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:*:repository/*/filebeat",
		"arn:aws:ecr:*:*:repository/*/base",
		"arn:aws:ecr:*:*:repository/*/onlinefs",
		"arn:aws:ecr:*:*:repository/*/airflow",
		"arn:aws:ecr:*:*:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:822623301872:repository/filebeat",
		"arn:aws:ecr:*:822623301872:repository/base",
		"arn:aws:ecr:*:822623301872:repository/onlinefs",
		"arn:aws:ecr:*:822623301872:repository/airflow",
		"arn:aws:ecr:*:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsEKSPermissions(allowDescribeEKSResource)...)
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_singleBucket(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_disableEKSAndECR(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_disableEKSAndECR(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_disableEKS(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:*:repository/*/filebeat",
		"arn:aws:ecr:*:*:repository/*/base",
		"arn:aws:ecr:*:*:repository/*/onlinefs",
		"arn:aws:ecr:*:*:repository/*/airflow",
		"arn:aws:ecr:*:*:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:822623301872:repository/filebeat",
		"arn:aws:ecr:*:822623301872:repository/base",
		"arn:aws:ecr:*:822623301872:repository/onlinefs",
		"arn:aws:ecr:*:822623301872:repository/airflow",
		"arn:aws:ecr:*:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_disableEKS(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_limitRegion(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:us-east-2:*:repository/*/filebeat",
		"arn:aws:ecr:us-east-2:*:repository/*/base",
		"arn:aws:ecr:us-east-2:*:repository/*/onlinefs",
		"arn:aws:ecr:us-east-2:*:repository/*/airflow",
		"arn:aws:ecr:us-east-2:*:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:us-east-2:822623301872:repository/filebeat",
		"arn:aws:ecr:us-east-2:822623301872:repository/base",
		"arn:aws:ecr:us-east-2:822623301872:repository/onlinefs",
		"arn:aws:ecr:us-east-2:822623301872:repository/airflow",
		"arn:aws:ecr:us-east-2:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_limitRegion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_limitUserAccount(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:user:repository/*/filebeat",
		"arn:aws:ecr:*:user:repository/*/base",
		"arn:aws:ecr:*:user:repository/*/onlinefs",
		"arn:aws:ecr:*:user:repository/*/airflow",
		"arn:aws:ecr:*:user:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:822623301872:repository/filebeat",
		"arn:aws:ecr:*:822623301872:repository/base",
		"arn:aws:ecr:*:822623301872:repository/onlinefs",
		"arn:aws:ecr:*:822623301872:repository/airflow",
		"arn:aws:ecr:*:822623301872:repository/git",
	}
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_limitUserAccount(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_limitHopsworksAccount(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
			awsBackupPermissions("*"),
		},
	}
	policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	var allowPushandPullImagesResource = []string{
		"arn:aws:ecr:*:*:repository/*/filebeat",
		"arn:aws:ecr:*:*:repository/*/base",
		"arn:aws:ecr:*:*:repository/*/onlinefs",
		"arn:aws:ecr:*:*:repository/*/airflow",
		"arn:aws:ecr:*:*:repository/*/git",
	}
	var allowPullImagesFromHopsworkAiResource = []string{
		"arn:aws:ecr:*:hopsworks:repository/filebeat",
		"arn:aws:ecr:*:hopsworks:repository/base",
		"arn:aws:ecr:*:hopsworks:repository/onlinefs",
		"arn:aws:ecr:*:hopsworks:repository/airflow",
		"arn:aws:ecr:*:hopsworks:repository/git",
	}
	policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_limitHopsworksAccount(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func TestAccAWSInstanceProfilePolicy_enableOnlyStorage(t *testing.T) {
	dataSourceName := "data.hopsworksai_aws_instance_profile_policy.test"
	policy := &awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			awsStoragePermissions("*"),
		},
	}

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_enableOnlyStorage(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSPolicyToJSONString(t, policy)),
				),
			},
		},
	})
}

func testAccAWSInstanceProfilePolicyConfig_basic() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_eks_restriction() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		eks_cluster_name = "cluster_name"
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_cluster_id() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		cluster_id = "cluster_id"
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_singleBucket() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		bucket_name = "test"
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_disableEKSAndECR() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks = false
		enable_ecr = false
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_enableOnlyStorage() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks = false
		enable_ecr = false
		enable_cloud_watch = false
		enable_backup = false
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_disableEKS() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks = false
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_limitRegion() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks = false
		region = "us-east-2"
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_limitUserAccount() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks = false
		user_ecr_account = "user"
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_limitHopsworksAccount() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks = false
		hopsworksai_ecr_account = "hopsworks"
	}
	`
}

func testAccAWSPolicyToJSONString(t *testing.T, policy *awsPolicy) string {
	policyJson, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(policyJson)
}

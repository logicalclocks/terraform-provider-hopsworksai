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
	policy.Statements = append(policy.Statements, awsUpgradePermissions())
	var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/*"
	policy.Statements = append(policy.Statements, awsEKSECRPermissions(allowDescribeEKSResource)...)

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
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
	policy.Statements = append(policy.Statements, awsUpgradePermissions())
	var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/*"
	policy.Statements = append(policy.Statements, awsEKSECRPermissions(allowDescribeEKSResource)...)

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
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
	policy.Statements = append(policy.Statements, awsUpgradePermissions())

	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
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

func TestAccAWSInstanceProfilePolicy_disableEKSAndUpgrade(t *testing.T) {
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
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_disableEKSAndUpgrade(),
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
		Providers: testAccProviders,
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

func testAccAWSInstanceProfilePolicyConfig_singleBucket() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		bucket_name = "test"
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_disableEKS() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks_and_ecr = false
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_disableEKSAndUpgrade() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks_and_ecr = false
		enable_upgrade = false
	}
	`
}

func testAccAWSInstanceProfilePolicyConfig_enableOnlyStorage() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_eks_and_ecr = false
		enable_upgrade = false
		enable_cloud_watch = false
		enable_backup = false
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

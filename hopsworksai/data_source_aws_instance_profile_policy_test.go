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
	}
	policy.Statements = append(policy.Statements, awsEKSECRPermissions(allowDescribeEKSResource, allowPushandPullImagesResource)...)

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
	}
	policy.Statements = append(policy.Statements, awsEKSECRPermissions(allowDescribeEKSResource, allowPushandPullImagesResource)...)

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
	}
	policy.Statements = append(policy.Statements, awsEKSECRPermissions(allowDescribeEKSResource, allowPushandPullImagesResource)...)

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
	}
	policy.Statements = append(policy.Statements, awsEKSECRPermissions(allowDescribeEKSResource, allowPushandPullImagesResource)...)

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
		ProviderFactories: testAccProviderFactories,
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

func TestAccAWSInstanceProfilePolicy_enableUpgrade(t *testing.T) {
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
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePolicyConfig_enableUpgrade(),
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

func testAccAWSInstanceProfilePolicyConfig_enableUpgrade() string {
	return `
	data "hopsworksai_aws_instance_profile_policy" "test" {
		enable_upgrade = true
		enable_eks_and_ecr = false
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

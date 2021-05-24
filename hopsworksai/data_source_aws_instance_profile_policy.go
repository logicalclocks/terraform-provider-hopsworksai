package hopsworksai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type awsPolicyStatement struct {
	Sid       string      `json:"Sid,omitempty"`
	Effect    string      `json:"Effect,omitempty"`
	Action    []string    `json:"Action,omitempty"`
	Resources interface{} `json:"Resource,omitempty"`
}

type awsPolicy struct {
	Version    string               `json:"Version,omitempty"`
	Statements []awsPolicyStatement `json:"Statement,omitempty"`
}

func dataSourceAWSInstanceProfilePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the aws instance profile policy needed by Hopsworks.ai",
		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Description: "Limit permissions to this S3 bucket.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"add_cloud_watch": {
				Description: "Add permissions required to allow collecting your cluster logs using cloud watch.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"add_upgrade": {
				Description: "Add permissions required to enable upgrade to newer versions of Hopsworks.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"add_eks_and_ecr": {
				Description: "Add permissions required to enable access to Amazon EKS and ECR from within your Hopsworks cluster.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"json": {
				Description: "The instance profile policy in JSON format.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		ReadContext: dataSourceAWSInstanceProfilePolicyRead,
	}
}

func dataSourceAWSInstanceProfilePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var s3Resources interface{} = "*"
	if v, ok := d.GetOk("bucket_name"); ok {
		bucketName := v.(string)
		s3Resources = []string{
			fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
			fmt.Sprintf("arn:aws:s3:::%s", bucketName),
		}
	}

	policy := awsPolicy{
		Version: "2012-10-17",
		Statements: []awsPolicyStatement{
			{
				Sid:    "S3Permissions",
				Effect: "Allow",
				Action: []string{
					"S3:PutObject",
					"S3:ListBucket",
					"S3:GetBucketLocation",
					"S3:GetObject",
					"S3:DeleteObject",
					"S3:AbortMultipartUpload",
					"S3:ListBucketMultipartUploads",
					"S3:PutLifecycleConfiguration",
					"S3:GetLifecycleConfiguration",
					"S3:PutBucketVersioning",
					"S3:GetBucketVersioning",
				},
				Resources: s3Resources,
			},
		},
	}

	if d.Get("add_cloud_watch").(bool) {
		policy.Statements = append(policy.Statements, awsPolicyStatement{
			Sid:    "CloudwatchPermissions",
			Effect: "Allow",
			Action: []string{
				"cloudwatch:PutMetricData",
				"ec2:DescribeVolumes",
				"ec2:DescribeTags",
				"logs:PutLogEvents",
				"logs:DescribeLogStreams",
				"logs:DescribeLogGroups",
				"logs:CreateLogStream",
				"logs:CreateLogGroup",
			},
			Resources: "*",
		}, awsPolicyStatement{
			Sid:    "HopsworksAICloudWatchParam",
			Effect: "Allow",
			Action: []string{
				"ssm:GetParameter",
			},
			Resources: "arn:aws:ssm:*:*:parameter/AmazonCloudWatch-*",
		})
	}

	if d.Get("add_upgrade").(bool) {
		policy.Statements = append(policy.Statements, awsPolicyStatement{
			Sid:    "UpgradePermissions",
			Effect: "Allow",
			Action: []string{
				"ec2:DescribeVolumes",
				"ec2:DetachVolume",
				"ec2:AttachVolume",
				"ec2:ModifyInstanceAttribute",
			},
			Resources: "*",
		})
	}

	if d.Get("add_eks_and_ecr").(bool) {
		policy.Statements = append(policy.Statements, awsPolicyStatement{
			Sid:    "AllowPullMainImages",
			Effect: "Allow",
			Action: []string{
				"ecr:GetDownloadUrlForLayer",
				"ecr:BatchGetImage",
			},
			Resources: []string{
				"arn:aws:ecr:*:*:repository/filebeat",
				"arn:aws:ecr:*:*:repository/base",
			},
		}, awsPolicyStatement{
			Sid:    "AllowPushandPullImages",
			Effect: "Allow",
			Action: []string{
				"ecr:CreateRepository",
				"ecr:GetDownloadUrlForLayer",
				"ecr:BatchGetImage",
				"ecr:CompleteLayerUpload",
				"ecr:UploadLayerPart",
				"ecr:InitiateLayerUpload",
				"ecr:DeleteRepository",
				"ecr:BatchCheckLayerAvailability",
				"ecr:PutImage",
				"ecr:ListImages",
				"ecr:BatchDeleteImage",
				"ecr:GetLifecyclePolicy",
				"ecr:PutLifecyclePolicy",
			},
			Resources: []string{
				"arn:aws:ecr:*:*:repository/*/filebeat",
				"arn:aws:ecr:*:*:repository/*/base",
			},
		}, awsPolicyStatement{
			Sid:    "AllowGetAuthToken",
			Effect: "Allow",
			Action: []string{
				"ecr:GetAuthorizationToken",
			},
			Resources: "*",
		}, awsPolicyStatement{
			Sid:    "AllowDescribeEKS",
			Effect: "Allow",
			Action: []string{
				"eks:DescribeCluster",
			},
			Resources: "arn:aws:eks:*:*:cluster/*",
		})
	}

	policyJson, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}

	policyString := string(policyJson)

	d.SetId(strconv.Itoa(schema.HashString(policyString)))
	if err := d.Set("json", policyString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

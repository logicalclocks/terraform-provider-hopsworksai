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
			"enable_storage": {
				Description: "Add permissions required to allow Hopsworks clusters to read and write from and to your aws S3 buckets.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_backup": {
				Description: "Add permissions required to allow creating backups of your clusters.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_cloud_watch": {
				Description: "Add permissions required to allow collecting your cluster logs using cloud watch.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_eks_and_ecr": {
				Description:   "Add permissions required to enable access to Amazon EKS and ECR from within your Hopsworks cluster.",
				Deprecated:    "Use enable_ecr and enable_eks instead",
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"enable_ecr", "enable_eks"},
			},
			"enable_ecr": {
				Description:   "Add permissions required to enable access to Amazon ECR from within your Hopsworks cluster.",
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       true,
				ConflictsWith: []string{"enable_eks_and_ecr"},
			},
			"enable_eks": {
				Description:   "Add permissions required to enable access to Amazon EKS from within your Hopsworks cluster.",
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       true,
				ConflictsWith: []string{"enable_eks_and_ecr"},
			},
			"eks_cluster_name": {
				Description: "Limit permissions to eks cluster.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cluster_id": {
				Description: "Limit docker repository permissions to the cluster id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"json": {
				Description: "The instance profile policy in JSON format.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "Limit docker repository permissions to a region",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user_ecr_account": {
				Description: "Limit docker repository permissions to the user aws account",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"hopsworksai_ecr_account": {
				Description: "Limit docker pull image from hopsworks.ai permissions to the hopsworks.ai aws account",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "822623301872",
			},
		},
		ReadContext: dataSourceAWSInstanceProfilePolicyRead,
	}
}

func awsStoragePermissions(s3Resources interface{}) awsPolicyStatement {
	return awsPolicyStatement{
		Sid:    "S3Permissions",
		Effect: "Allow",
		Action: []string{
			"S3:PutObject",
			"S3:ListBucket",
			"S3:GetObject",
			"S3:DeleteObject",
			"S3:AbortMultipartUpload",
			"S3:ListBucketMultipartUploads",
			"S3:GetBucketVersioning",
		},
		Resources: s3Resources,
	}
}

func awsBackupPermissions(s3Resources interface{}) awsPolicyStatement {
	return awsPolicyStatement{
		Sid:    "BackupsPermissions",
		Effect: "Allow",
		Action: []string{
			"S3:PutLifecycleConfiguration",
			"S3:GetLifecycleConfiguration",
			"S3:PutBucketVersioning",
			"S3:ListBucketVersions",
			"S3:DeleteObjectVersion",
		},
		Resources: s3Resources,
	}
}

func awsCloudWatchPermissions() []awsPolicyStatement {
	return []awsPolicyStatement{
		{
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
		}, {
			Sid:    "HopsworksAICloudWatchParam",
			Effect: "Allow",
			Action: []string{
				"ssm:GetParameter",
			},
			Resources: "arn:aws:ssm:*:*:parameter/AmazonCloudWatch-*",
		},
	}
}

func awsECRPermissions(allowPullImagesFromHopsworkAiResource interface{}, allowPushandPullImagesToUserRepoResource interface{}) []awsPolicyStatement {
	return []awsPolicyStatement{
		{
			Sid:    "AllowPullImagesFromHopsworkAi",
			Effect: "Allow",
			Action: []string{
				"ecr:GetDownloadUrlForLayer",
				"ecr:BatchGetImage",
			},
			Resources: allowPullImagesFromHopsworkAiResource,
		}, {
			Sid:    "AllowCreateRepository",
			Effect: "Allow",
			Action: []string{
				"ecr:CreateRepository",
			},
			Resources: "*",
		}, {
			Sid:    "AllowPushandPullImagesToUserRepo",
			Effect: "Allow",
			Action: []string{
				"ecr:GetDownloadUrlForLayer",
				"ecr:BatchGetImage",
				"ecr:CompleteLayerUpload",
				"ecr:UploadLayerPart",
				"ecr:InitiateLayerUpload",
				"ecr:BatchCheckLayerAvailability",
				"ecr:PutImage",
				"ecr:ListImages",
				"ecr:BatchDeleteImage",
				"ecr:GetLifecyclePolicy",
				"ecr:PutLifecyclePolicy",
				"ecr:TagResource",
			},
			Resources: allowPushandPullImagesToUserRepoResource,
		}, {
			Sid:    "AllowGetAuthToken",
			Effect: "Allow",
			Action: []string{
				"ecr:GetAuthorizationToken",
			},
			Resources: "*",
		},
	}
}

func awsEKSPermissions(allowDescribeEKSResource interface{}) []awsPolicyStatement {
	return []awsPolicyStatement{
		{
			Sid:    "AllowDescribeEKS",
			Effect: "Allow",
			Action: []string{
				"eks:DescribeCluster",
			},
			Resources: allowDescribeEKSResource,
		},
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
		Version:    "2012-10-17",
		Statements: []awsPolicyStatement{},
	}

	if d.Get("enable_storage").(bool) {
		policy.Statements = append(policy.Statements, awsStoragePermissions(s3Resources))
	}

	if d.Get("enable_backup").(bool) {
		policy.Statements = append(policy.Statements, awsBackupPermissions(s3Resources))
	}

	if d.Get("enable_cloud_watch").(bool) {
		policy.Statements = append(policy.Statements, awsCloudWatchPermissions()...)
	}

	if d.Get("enable_eks_and_ecr").(bool) || d.Get("enable_eks").(bool) {
		var allowDescribeEKSResource interface{} = "arn:aws:eks:*:*:cluster/*"
		if v, ok := d.GetOk("eks_cluster_name"); ok {
			eksClusterName := v.(string)
			allowDescribeEKSResource = fmt.Sprintf("arn:aws:eks:*:*:cluster/%s", eksClusterName)
		}

		policy.Statements = append(policy.Statements, awsEKSPermissions(allowDescribeEKSResource)...)
	}

	if d.Get("enable_eks_and_ecr").(bool) || d.Get("enable_ecr").(bool) {
		var region = "*"
		if v, ok := d.GetOk("region"); ok {
			region = v.(string)
		}
		var clusterId = "*"
		if v, ok := d.GetOk("cluster_id"); ok {
			clusterId = v.(string)
		}
		var userEcrAccount = "*"
		if v, ok := d.GetOk("user_ecr_account"); ok {
			userEcrAccount = v.(string)
		}
		var hopsworksaiEcrAccount = "822623301872"
		if v, ok := d.GetOk("hopsworksai_ecr_account"); ok {
			hopsworksaiEcrAccount = v.(string)
		}
		var allowPushandPullImagesResource = []string{
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/filebeat", region, userEcrAccount, clusterId),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/base", region, userEcrAccount, clusterId),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/onlinefs", region, userEcrAccount, clusterId),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/airflow", region, userEcrAccount, clusterId),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/git", region, userEcrAccount, clusterId),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/testconnector", region, userEcrAccount, clusterId),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s/flyingduck", region, userEcrAccount, clusterId),
		}
		var allowPullImagesFromHopsworkAiResource = []string{
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/filebeat", region, hopsworksaiEcrAccount),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/base", region, hopsworksaiEcrAccount),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/onlinefs", region, hopsworksaiEcrAccount),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/airflow", region, hopsworksaiEcrAccount),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/git", region, hopsworksaiEcrAccount),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/testconnector", region, hopsworksaiEcrAccount),
			fmt.Sprintf("arn:aws:ecr:%s:%s:repository/flyingduck", region, hopsworksaiEcrAccount),
		}
		policy.Statements = append(policy.Statements, awsECRPermissions(allowPullImagesFromHopsworkAiResource, allowPushandPullImagesResource)...)
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

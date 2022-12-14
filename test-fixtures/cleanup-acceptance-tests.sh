#!/usr/bin/env bash
set -e 

TF_VAR_aws_profile=${TF_VAR_aws_profile:-default}
TF_VAR_aws_region=${TF_VAR_aws_region:-us-east-2}

BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd ${BASE_DIR}


export TF_HOPSWORKSAI_TEST_SUFFIX=$(terraform output -raw test_random_suffix)
echo "Run acceptance test cleanup for test run - ${TF_HOPSWORKSAI_TEST_SUFFIX}"

ECR_REPOS=`aws resourcegroupstaggingapi get-resources --tag-filters Key=Purpose,Values=acceptance-test Key=Run,Values=${TF_HOPSWORKSAI_TEST_SUFFIX} --resource-type-filters ecr:repository  | jq -r '.ResourceTagMappingList | map(.ResourceARN | split(":repository/")[1]) | .[]'`
for repo in ${ECR_REPOS[@]}; do
    aws --profile ${TF_VAR_aws_profile} --region ${TF_VAR_aws_region} ecr delete-repository --repository-name $repo --force --output text || "$repo do not exists"
done

terraform destroy -auto-approve || terraform destroy -auto-approve
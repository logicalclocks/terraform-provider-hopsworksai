#!/usr/bin/env bash
set -e 
ACCTEST_TIMEOUT=${ACCTEST_TIMEOUT:-240m}
ACCTEST_PARALLELISM=${ACCTEST_PARALLELISM:-2}

TF_VAR_skip_aws=${TF_VAR_skip_aws:-false}
TF_VAR_skip_azure=${TF_VAR_skip_azure:-false}

TF_VAR_aws_profile=${TF_VAR_aws_profile:-default}
TF_VAR_aws_region=${TF_VAR_aws_region:-us-east-2}

TF_ACCTEST_LOG_DIR=${TF_ACCTEST_LOG_DIR:-/tmp/tf-acc-logs/$(date +%Y%m%d_%H%M%S)}
TF_ACCTEST_LOG_LEVEL=${TF_ACCTEST_LOG_LEVEL:-debug}

if [ -z ${HOPSWORKSAI_API_KEY} ] ; then 
    echo "Environment variable HOPSWORKSAI_API_KEY is not set, you need to export your Hopsworks API key to run the acceptance tests"
    exit 1
fi 

if [ -z ${TF_VAR_azure_resource_group} ] && [ ${TF_VAR_skip_azure} = false ] ; then 
    echo "You need to set the azure resource group for testing by setting the environment variable TF_VAR_azure_resource_group"
    exit 1
fi 

BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd ${BASE_DIR}

KEYS_DIR=".keys"
KEY_NAME="${KEYS_DIR}/tf"
mkdir -p ${KEYS_DIR}
if [ ! -f ${KEY_NAME} ]; then 
    echo "Generate SSH Key pair ${KEY_NAME}"
    ssh-keygen -q -t rsa -b 2048 -f ${KEY_NAME} -N ""
else
    echo "Use existing ssh key ${KEY_NAME}"
fi 

echo "Initialize test fixtures"
terraform init || ( rm -rf .terraform* && terraform init )
terraform destroy -auto-approve || terraform destroy -auto-approve
terraform apply -auto-approve || terraform apply -auto-approve

echo "Setting environment variables for testing"
export TF_HOPSWORKSAI_AWS_SKIP=${TF_VAR_skip_aws}
if [ ${TF_VAR_skip_aws} = false ]; then 
    export TF_HOPSWORKSAI_AWS_REGION=$(terraform output -raw aws_region)
    export TF_HOPSWORKSAI_AWS_BUCKET_NAMES=$(terraform output -raw aws_bucket_names)
    export TF_HOPSWORKSAI_AWS_INSTANCE_PROFILE_ARN=$(terraform output -raw aws_instance_profile_arn)
    export TF_HOPSWORKSAI_AWS_SSH_KEY=$(terraform output -raw aws_ssh_key_name)
    export TF_HOPSWORKSAI_AWS_VPC_ID=$(terraform output -raw aws_vpc_id)
    export TF_HOPSWORKSAI_AWS_SUBNET_ID=$(terraform output -raw aws_subnet_id)
    export TF_HOPSWORKSAI_AWS_SECURITY_GROUP_ID=$(terraform output -raw aws_security_group_id)
fi 

export TF_HOPSWORKSAI_AZURE_SKIP=${TF_VAR_skip_azure}
if [ ${TF_VAR_skip_azure} = false ]; then 
    export TF_HOPSWORKSAI_AZURE_LOCATION=$(terraform output -raw azure_location)
    export TF_HOPSWORKSAI_AZURE_RESOURCE_GROUP=$(terraform output -raw azure_resource_group)
    export TF_HOPSWORKSAI_AZURE_STORAGE_ACCOUNT_NAME=$(terraform output -raw azure_storage_account_name)
    export TF_HOPSWORKSAI_AZURE_USER_ASSIGNED_IDENTITY_NAME=$(terraform output -raw azure_user_assigned_identity_name)
    export TF_HOPSWORKSAI_AZURE_SSH_KEY=$(terraform output -raw azure_ssh_key_name)
    export TF_HOPSWORKSAI_AZURE_VIRTUAL_NETWORK_NAME=$(terraform output -raw azure_virtual_network_name)
    export TF_HOPSWORKSAI_AZURE_SUBNET_NAME=$(terraform output -raw azure_subnet_name)
    export TF_HOPSWORKSAI_AZURE_SECURITY_GROUP_NAME=$(terraform output -raw azure_security_group_name)
fi

popd

mkdir -p $TF_ACCTEST_LOG_DIR

echo "Run test cases with args ${TESTARGS} timeout ${ACCTEST_TIMEOUT} parallel ${ACCTEST_PARALLELISM}"
TF_LOG=${TF_ACCTEST_LOG_LEVEL} TF_LOG_PATH_MASK="${TF_ACCTEST_LOG_DIR}/%s" TF_ACC=1 go test ./... -v ${TESTARGS} --cover -timeout ${ACCTEST_TIMEOUT} --parallel ${ACCTEST_PARALLELISM} 2>&1 | go-junit-report -set-exit-code -iocopy -out report.xml

pushd ${BASE_DIR}
echo "Destroying test fixtures"
terraform destroy -auto-approve || terraform destroy -auto-approve
echo "Done"
popd
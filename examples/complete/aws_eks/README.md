# Integrate Hopsworks cluster with Amazon EKS and ECR

In this example, we create an EKS cluster and a Hopsworks cluster that is integrated with both EKS and ECR. We also create a VPC where both EKS and Hopsworks reside ensuring that they can communicate with each other.

## How to run the example 
First ensure that your aws credentials are setup correctly by running the following command 

```bash
aws configure 
```

Then, run the following commands. Replace the placeholder with your Hopsworks API Key. The EKS and Hopsworks clusters will be created in us-east-2 region by default, however, you can configure the region to use by setting the variable region when applying the changes `-var="region=YOUR_REGION"`

```bash
export HOPSWORKSAI_API_KEY=<YOUR_HOPSWORKSAI_API_KEY>
terraform init
terraform apply
```


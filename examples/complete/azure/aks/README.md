# Integrate Hopsworks cluster with Azure AKS and ACR

In this example, we create an AKS cluster, an ACR registry, and a Hopsworks cluster that is integrated with both AKS and ACR. We also create a virtual network where both AKS and Hopsworks reside ensuring that they can communicate with each other given that the AKS cluster is private with no public access allowed.

## How to run the example 
First ensure that your azure credentials are setup correctly by running the following command

```bash
az login 
```

Then, run the following commands. Replace the placeholders with your Hopsworks API Key and your Azure resource group

```bash
export HOPSWORKSAI_API_KEY=<YOUR_HOPSWORKSAI_API_KEY>
terraform init
terraform apply
```


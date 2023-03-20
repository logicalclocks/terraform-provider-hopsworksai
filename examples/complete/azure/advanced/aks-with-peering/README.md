# Integrate Hopsworks cluster with Azure AKS and ACR

In this example, we create an AKS cluster, an ACR registry, and a Hopsworks cluster that is integrated with both AKS and ACR. We create two different virtual networks for both AKS and Hopsworks and connect them together using virtual network peering. The AKS cluster and the ACR registry can reside in a different resource group than the Hopsworks cluster.

## How to run the example 
First ensure that your azure credentials are setup correctly by running the following command

```bash
az login 
```

Then, run the following commands. Replace the placeholders with your Hopsworks API Key and your Azure resource group(s). You can use the same resource group for both AKS and Hopsworks or use two differen resource groups.

```bash
export HOPSWORKSAI_API_KEY=<YOUR_HOPSWORKSAI_API_KEY>
terraform init
terraform apply  -var="aks_resource_group=<YOUR_AKS_RESOURCE_GROUP>" -var="hopsworks_resource_group=<YOUR_HOPSWORKS_RESOURCE_GROUP>"
```

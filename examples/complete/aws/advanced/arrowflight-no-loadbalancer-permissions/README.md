# Hopsworks cluster with ArrowFlight server with no load balancer permissions 

In this example, we create a Hopsworks cluster with arrow flight server enabled. This example assumes that the users have removed the permissions for Hopsworks.ai to manage load balancer on their behalf. If you have given Hopsworks.ai the manage load balancer permissions as shown in [the docs](https://docs.hopsworks.ai/latest/setup_installation/aws/restrictive_permissions/#load-balancers-permissions-for-external-access), then there is no need to run this example and instead you can directly just set the `rondb/mysql_nodes/arrow_flight_with_duckdb` attribute to true in your terraform cluster confiugraiton.


## How to run the example 
First ensure that your aws credentials are setup correctly by running the following command 

```bash
aws configure 
```

Then, run the following commands. Replace the placeholder with your Hopsworks API Key. The cluster will be created in us-east-2 region by default, however, you can configure which region to use by setting the variable region when applying the changes `-var="region=YOUR_REGION"`

```bash
export HOPSWORKSAI_API_KEY=<YOUR_HOPSWORKSAI_API_KEY>
terraform init
terraform apply
```

## Terminate the cluster

You can run `terraform destroy` to delete the cluster and all the other required cloud resources created in this example.
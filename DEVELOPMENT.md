# Development guide 

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `make install` command: 
```sh
$ make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](README.md#requirements) above).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the terraform plugin directory.

To generate or update documentation, run `make generate`.

## Testing the Provider

### Unit tests 
To run the unit tests, run `make test`

### Acceptance tests

**Note:** Acceptance tests create real resources, and cost money to run.

In order to run the full suite of Acceptance tests, you need do the following:
* Configure your AWS and Azure credentials locally 
* Export the following environment variables

```sh
   export HOPSWORKSAI_API_KEY=<YOUR HOPSWORKS API KEY>
   export TF_VAR_skip_aws=false # Setting it to true will not run any acceptance tests on AWS
   export TF_VAR_skip_azure=false # Setting it to true will not run any acceptance tests on Azure
   export TF_VAR_azure_resource_group=<YOUR AZURE RESOURCE GROUP> # no need to set if you skip tests on Azure
```
* Run all the acceptance tests using the following command 

```sh
$ make testacc 
```

You can also run only a single test or a some tests following some name pattern as follows
```sh
$ make testacc TESTARGS='-run=TestAcc*'
```

Acceptance tests provision real resources, and ideally these resources should be destroyed at the end of each test, however, it can happen that resources are leaked due to different reasons. For that, you can run the sweeper to clean up all resources created during acceptance testing.

```sh
$ make sweep 
```

## Using the Provider

With Terraform v0.14 and later, [development overrides for provider developers](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers) can be leveraged in order to use the provider built from source.

To do this, populate a Terraform CLI configuration file (`~/.terraformrc` for all platforms other than Windows; `terraform.rc` in the `%APPDATA%` directory when using Windows) with at least the following options:

```hcl
provider_installation {
  dev_overrides {
    "logicalclocks/hopsworksai" = "[REPLACE WITH THE REPOSITORY LOCAL DIRECTORY]/bin"
  }
  direct {}
}
```

## Releasing the Provider

We use a GitHub Action that is configured to automatically build and publish assets for release when a tag is pushed that matches the pattern v* (ie. v0.1.0).

[The Goreleaser configuration](.goreleaser.yml) produces build artifacts matching the [layout required](https://www.terraform.io/docs/registry/providers/publishing.html#manually-preparing-a-release) to publish the provider in the Terraform Registry.

Releases will appear as drafts. Once marked as published on the GitHub Releases page, they will become available via the Terraform Registry.

## Recommended Docs

- [Terraform Plugin Best Practices](https://www.terraform.io/docs/extend/best-practices/index.html)
- [Acceptance Tests](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html)
- [Debugging Providers](https://www.terraform.io/docs/extend/debugging.html)

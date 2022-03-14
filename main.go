package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai"
)

// Generate docs
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// Will be set by the goreleaser configuration
	version string = "dev"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: hopsworksai.Provider(version)}

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/logicalclocks/hopsworksai", opts)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}

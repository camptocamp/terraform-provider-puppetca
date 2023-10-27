package main

import (
	"context"
	"flag"
	"log"

	"github.com/camptocamp/terraform-provider-puppetca/internal/datasources"
	"github.com/camptocamp/terraform-provider-puppetca/internal/provider"
	"github.com/camptocamp/terraform-provider-puppetca/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	name    = "puppetca"
	version = "dev"
	address = "registry.terraform.io/camptocamp/" + name
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like Delve")
	flag.Parse()

	err := providerserver.Serve(
		context.Background(),
		provider.NewFactory(name, version, datasources.DataSources(), resources.Resources()),
		providerserver.ServeOpts{
			Address: address,
			Debug:   debug,
		})

	if err != nil {
		log.Fatal(err.Error())
	}
}

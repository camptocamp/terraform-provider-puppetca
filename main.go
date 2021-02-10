package main

import (
	"github.com/greennosedmule/terraform-provider-puppetca/puppetca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	opts := plugin.ServeOpts{
		ProviderFunc: puppetca.Provider,
	}
	plugin.Serve(&opts)
}

package main

import (
	"github.com/camptocamp/terraform-provider-puppetca/puppetca"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: puppetca.Provider})
}

package main

import (
	"github.com/greennosedmule/terraform-provider-puppetca/puppetca"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: puppetca.Provider})
}

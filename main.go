package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		ProviderAddr: "registry.terraform.io/zenlayercloud/test",
		Debug: true,
		ProviderFunc: func() *schema.Provider {
			return zenlayercloud.Provider()
		},
	})
}

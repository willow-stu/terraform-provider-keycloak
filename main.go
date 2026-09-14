package main

// Modified by willow-stu for its community distribution. See NOTICE.

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/keycloak/terraform-provider-keycloak/provider"
)

func main() {

	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return provider.KeycloakProvider(nil)
		},
		Debug:        debugMode,
		ProviderAddr: "registry.terraform.io/willow-stu/keycloak",
	}
	plugin.Serve(opts)
}

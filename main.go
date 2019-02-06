package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/syoya/terraform-provider-sendgrid/sendgrid"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sendgrid.Provider,
	})
}

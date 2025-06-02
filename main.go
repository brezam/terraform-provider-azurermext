package main

import (
	"context"
	"terraform-provider-azurermext/internal"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), internal.NewProvider, providerserver.ServeOpts{
		Address:         "registry.terraform.io/brezam/azurermext",
		ProtocolVersion: 6,
	})
}

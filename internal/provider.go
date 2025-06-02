package internal

import (
	"context"
	"terraform-provider-azurermext/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewProvider() provider.Provider {
	return &azureRMExtProvider{}
}

type azureRMExtProvider struct{}

type azureRMExtProviderModel struct {
}

// Metadata returns the provider type name.
func (p *azureRMExtProvider) Metadata(ctx context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azurermext"
}

// Schema defines the provider-level schema for configuration data.
func (p *azureRMExtProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: providerDescription,
		Attributes:  map[string]schema.Attribute{},
	}
}

// Configure defines the provider configuration and what is passed onto resource and datasources.
func (p *azureRMExtProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Azure RM Ext client")

	var config azureRMExtProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	client_ := client.New()
	resp.DataSourceData = client_
	resp.ResourceData = client_
}

// DataSources defines the data sources implemented in the provider.
func (p *azureRMExtProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *azureRMExtProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

package internal

import (
	"context"
	"os"
	"terraform-provider-azurermext/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewProvider() provider.Provider {
	return &azureRMExtProvider{}
}

type azureRMExtProvider struct{}

type azureRMExtProviderModel struct {
	TenantId     types.String `tfsdk:"tenant_id"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Metadata returns the provider type name.
func (p *azureRMExtProvider) Metadata(ctx context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azurermext"
}

// Schema defines the provider-level schema for configuration data.
func (p *azureRMExtProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: providerDescription,
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "Service Principal Client ID.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "Service Principal Client ID.",
			},
			"client_secret": schema.StringAttribute{
				Sensitive:   true,
				Optional:    true,
				Description: "Service Principal Client Secret.",
			},
		},
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

	var tenantId string
	if config.TenantId.IsNull() {
		tenantId = os.Getenv("ARM_TENANT_ID")
	}

	var clientId string
	if config.ClientId.IsNull() {
		clientId = os.Getenv("ARM_CLIENT_ID")
	}

	var clientSecret string
	if config.ClientSecret.IsNull() {
		clientSecret = os.Getenv("ARM_CLIENT_SECRET")
	}

	tflog.Info(ctx, "client configuration finished", map[string]any{"client_id": clientId, "tenant_id": tenantId})

	client_ := client.New(clientId, clientSecret, tenantId)
	resp.DataSourceData = client_
	resp.ResourceData = client_
}

// DataSources defines the data sources implemented in the provider.
func (p *azureRMExtProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *azureRMExtProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCosmosDBMongoDBIpFilterResource,
	}
}

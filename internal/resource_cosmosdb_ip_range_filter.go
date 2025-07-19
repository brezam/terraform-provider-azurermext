package internal

import (
	"context"
	"fmt"
	"terraform-provider-azurermext/internal/client"

	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure = (*CosmosDBIpFilterResource)(nil)
)

type CosmosDBIpFilterResource struct {
	client *client.Client
}

type CosmosDBMongoDBIpFilterResourceModel struct {
	ID                types.String `tfsdk:"id"`
	CosmosDBAccountId types.String `tfsdk:"cosmosdb_account_id"`
	IpRules           types.List   `tfsdk:"ip_rules"`
}

func NewCosmosDBMongoDBIpFilterResource() resource.Resource {
	return &CosmosDBIpFilterResource{}
}

func (r *CosmosDBIpFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cosmosdb_ip_range_filter"
}

func (r *CosmosDBIpFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *CosmosDBIpFilterResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: cosmosDbIpRangeFilterDescription,
		Version:     1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Computed:      true,
			},
			"cosmosdb_account_id": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:      true,
				Description:   "Resource ID of the Azure CosmosDB Account.",
			},
			"ip_rules": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "List of IP addresses or CIDR ranges to allow access to the Azure CosmosDB Account.",
			},
		},
	}
}

func (r *CosmosDBIpFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CosmosDBMongoDBIpFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cosmo, err := r.client.ReadCosmosDB(ctx, state.CosmosDBAccountId.ValueString())
	if err != nil {
		// TODO: Add not found removing state as opposed to being an error of read
		//       resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError(
			"Could not read CosmosDB",
			"Failed to read CosmosDB account with ID "+state.CosmosDBAccountId.ValueString()+": "+err.Error(),
		)
		return
	}
	currentIpRules := parseCurrentIpRulesFromResponse(cosmo)

	if !cosmo.Properties.PublicNetworkAccess.IsEnabled() {
		resp.Diagnostics.AddError(
			"CosmosDB account is not publicly accessible",
			"CosmosDB account "+state.CosmosDBAccountId.ValueString()+" is not publicly accessible. Please enable public network access to add IP rules.",
		)
		return
	}

	if len(currentIpRules) == 0 {
		// In this case the CosmosDB account is public
		return
	}

	newStateIpRules := []attr.Value{}
	for _, stateIPT := range state.IpRules.Elements() {
		stateIP := stateIPT.(types.String).ValueString()
		if slices.Contains(currentIpRules, stateIP) {
			newStateIpRules = append(newStateIpRules, stateIPT)
		}
	}
	newIpRulesState, diags := types.ListValue(types.StringType, newStateIpRules)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	state.IpRules = newIpRulesState
	state.ID = types.StringValue(cosmo.ID)
	resp.State.Set(ctx, &state)
}

func (r *CosmosDBIpFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CosmosDBMongoDBIpFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsertCosmosDB(ctx, nil, &plan, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, &plan)
}

func (r *CosmosDBIpFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CosmosDBMongoDBIpFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state CosmosDBMongoDBIpFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsertCosmosDB(ctx, &state, &plan, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, &plan)
}

func (r *CosmosDBIpFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CosmosDBMongoDBIpFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// Instead of no-op, we might consider removing the rules that are currently in our state, essentially having the
	// same behavior as if the person modified the tfvars to pass an empty list.
	// However, if the user uses this resource alongside the official `azurerm_cosmosdb_account`, and they want to perform
	// a full `terraform destroy`, we'd have the situation where there's 10-15 minutes of select IPs removal only to then
	// later have 10-15 minutes of CosmosDB account destruction, wasting everyone's time.
	// I believe it's better to just give people the option of the workflow <state with IPs> -> <tfvars modification to an empty list>
	// causing the removal of all managed IPs, and have destruction be no-op with the assumption of destruction of CosmosDB account
	// being done by the official resource right after.
}

// This method modifies state and diags inplace
func (r *CosmosDBIpFilterResource) upsertCosmosDB(ctx context.Context, state, plan *CosmosDBMongoDBIpFilterResourceModel, diags diag.Diagnostics) {
	cosmosID := plan.CosmosDBAccountId.ValueString()
	cosmo, err := r.client.ReadCosmosDB(ctx, cosmosID)
	if err != nil {
		diags.AddError(
			"Could not read CosmosDB",
			"Failed to read CosmosDB account with ID "+cosmosID+": "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(cosmo.ID)

	if !cosmo.Properties.PublicNetworkAccess.IsEnabled() {
		diags.AddError(
			"CosmosDB account is not publicly accessible",
			"CosmosDB account "+cosmosID+" is not publicly accessible. Please enable public network access to add IP rules.",
		)
		return
	}

	currentIpRules := parseCurrentIpRulesFromResponse(cosmo)
	if len(currentIpRules) == 0 {
		// In this case the CosmosDB account is public, so we avoid adding any IP rules otherwise we would block access.
		// Technically speaking we should check that there are no approved private endpoints as well, but I'd rather err on the side of caution here.
		// Attempting to add ip rules when public removes the 'publicness' of the account.
		plan.ID = types.StringValue(cosmo.ID)
		return
	}

	// figuring out which rules to remove
	rulesToRemove := []string{}
	if state != nil {
		for _, stateIPT := range state.IpRules.Elements() {
			stateIP := stateIPT.(types.String).ValueString()
			foundInPlan := false
			for _, planIPT := range plan.IpRules.Elements() {
				planIP := planIPT.(types.String).ValueString()
				if stateIP == planIP {
					foundInPlan = true
					break
				}
			}
			if !foundInPlan {
				rulesToRemove = append(rulesToRemove, stateIP)
			}
		}
	}
	// figuring out which rules to add
	newRules := []string{}
	for _, planIPT := range plan.IpRules.Elements() {
		planIP := planIPT.(types.String).ValueString()
		if !slices.Contains(currentIpRules, planIP) {
			newRules = append(newRules, planIP)
		}
	}

	// finalizing the IP rules to be set
	finalIPRules := []client.CosmosDBIpRule{}
	for _, rule := range cosmo.Properties.IpRules {
		if rule.IpAddressOrRange == "" {
			continue
		}
		if slices.Contains(currentIpRules, rule.IpAddressOrRange) && !slices.Contains(rulesToRemove, rule.IpAddressOrRange) {
			finalIPRules = append(finalIPRules, rule)
		}
	}
	for _, newRule := range newRules {
		finalIPRules = append(finalIPRules, client.CosmosDBIpRule{IpAddressOrRange: newRule})
	}

	if len(newRules) != 0 || len(rulesToRemove) != 0 {
		tflog.Info(ctx, fmt.Sprintf("New Rules: %v", newRules))
		tflog.Info(ctx, fmt.Sprintf("Rules to remove: %v", rulesToRemove))
		err = r.client.UpdateCosmosDBIpRulesAndPoll(ctx, plan.CosmosDBAccountId.ValueString(), finalIPRules)
		if err != nil {
			diags.AddError(
				"Could not update CosmosDB IP rules",
				err.Error(),
			)
			return
		}
	}
}

func parseCurrentIpRulesFromResponse(cosmo *client.CosmosDBResponse) []string {
	ipRules := make([]string, 0, len(cosmo.Properties.IpRules))
	for _, rule := range cosmo.Properties.IpRules {
		if rule.IpAddressOrRange != "" {
			ipRules = append(ipRules, rule.IpAddressOrRange)
		}
	}
	return ipRules
}

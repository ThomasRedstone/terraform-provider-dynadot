package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ThomasRedstone/terraform-provider-dynadot/internal/client"
)

var _ resource.Resource = &NameserversResource{}

type NameserversResource struct{ client *client.Client }

type NameserversModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.List   `tfsdk:"nameservers"`
}

func NewNameserversResource() resource.Resource { return &NameserversResource{} }

func (r *NameserversResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_nameservers"
}

func (r *NameserversResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets the nameservers for a Dynadot-registered domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to configure.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nameservers": schema.ListAttribute{
				Description: "Ordered list of nameservers (up to 13).",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *NameserversResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *NameserversResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NameserversModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := toStringSlice(ctx, plan.Nameservers)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read nameservers", err.Error())
		return
	}

	if err := r.client.SetNameservers(plan.Domain.ValueString(), ns); err != nil {
		resp.Diagnostics.AddError("Failed to set nameservers", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NameserversResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NameserversModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := r.client.GetNameservers(state.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read nameservers", err.Error())
		return
	}

	elems := make([]attr.Value, len(ns))
	for i, n := range ns {
		elems[i] = types.StringValue(n)
	}
	list, diags := types.ListValue(types.StringType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Nameservers = list
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *NameserversResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NameserversModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := toStringSlice(ctx, plan.Nameservers)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read nameservers", err.Error())
		return
	}

	if err := r.client.SetNameservers(plan.Domain.ValueString(), ns); err != nil {
		resp.Diagnostics.AddError("Failed to update nameservers", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NameserversResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nameservers cannot be deleted, only changed — no-op on destroy.
}

func toStringSlice(ctx context.Context, list types.List) ([]string, error) {
	var elems []types.String
	if diags := list.ElementsAs(ctx, &elems, false); diags.HasError() {
		return nil, fmt.Errorf("reading nameservers list")
	}
	result := make([]string, len(elems))
	for i, e := range elems {
		result[i] = e.ValueString()
	}
	return result, nil
}

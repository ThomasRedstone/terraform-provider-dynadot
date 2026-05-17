package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ThomasRedstone/terraform-provider-dynadot/internal/client"
)

var _ provider.Provider = &DynadotProvider{}

type DynadotProvider struct{ version string }

type DynadotProviderModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &DynadotProvider{version: version} }
}

func (p *DynadotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dynadot"
	resp.Version = p.version
}

func (p *DynadotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Dynadot domain nameservers.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Dynadot API key. Can also be set via DYNADOT_API_KEY.",
				Optional:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "Dynadot secret key. Can also be set via DYNADOT_SECRET_KEY.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DynadotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DynadotProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := envFallback(config.APIKey, "DYNADOT_API_KEY")
	secretKey := envFallback(config.SecretKey, "DYNADOT_SECRET_KEY")

	if apiKey == "" {
		resp.Diagnostics.AddError("Missing API Key", "Set api_key in the provider or DYNADOT_API_KEY env var.")
	}
	if secretKey == "" {
		resp.Diagnostics.AddError("Missing Secret Key", "Set secret_key in the provider or DYNADOT_SECRET_KEY env var.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(apiKey, secretKey)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *DynadotProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewNameserversResource}
}

func (p *DynadotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func envFallback(val types.String, env string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(env)
}

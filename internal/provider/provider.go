package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &quismonProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &quismonProvider{
			version: version,
		}
	}
}

// quismonProvider is the provider implementation.
type quismonProvider struct {
	version string
}

// quismonProviderModel maps provider schema data to a Go type.
type quismonProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

// Metadata returns the provider type name.
func (p *quismonProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "quismon"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *quismonProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Quismon monitoring resources as Infrastructure as Code.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Quismon API key. Can also be set via QUISMON_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Quismon API base URL. Defaults to https://api.quismon.com. Can also be set via QUISMON_BASE_URL.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares an API client for data sources and resources.
func (p *quismonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config quismonProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Quismon API Key",
			"The provider cannot create the Quismon API client as there is an unknown configuration value for the API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the QUISMON_API_KEY environment variable.",
		)
	}

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown Quismon Base URL",
			"The provider cannot create the Quismon API client as there is an unknown configuration value for the base URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the QUISMON_BASE_URL environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	apiKey := os.Getenv("QUISMON_API_KEY")
	baseURL := os.Getenv("QUISMON_BASE_URL")

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	// Default base URL if not set
	if baseURL == "" {
		baseURL = "https://api.quismon.com"
	}

	// If no API key provided, try to read from terraform state
	// This enables seamless usage after initial signup creation
	if apiKey == "" {
		stateKey, warning := readAPIKeyFromState()
		if stateKey != "" {
			apiKey = stateKey
			// Add warning to diagnostics (visible in terraform output)
			resp.Diagnostics.AddWarning(
				"API Key Read from Terraform State",
				warning,
			)
		}
	}

	// Note: API key is optional - it's only required for authenticated resources.
	// The signup resource can work without an API key.
	// We create the client even with empty API key - individual resources will
	// fail if they need auth and no key is provided.

	// Create a new Quismon client using the configuration values
	c, err := client.New(baseURL, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Quismon API Client",
			"An unexpected error occurred when creating the Quismon API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Quismon Client Error: "+err.Error(),
		)
		return
	}

	// Make the Quismon client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = c
	resp.ResourceData = c
}

// DataSources defines the data sources implemented in the provider.
func (p *quismonProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCheckDataSource,
		NewChecksDataSource,
		NewNotificationChannelDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *quismonProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCheckResource,
		NewAlertRuleResource,
		NewNotificationChannelResource,
		NewSignupResource,
	}
}

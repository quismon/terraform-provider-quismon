package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RegionInfo represents a monitoring region
type RegionInfo struct {
	Code        types.String `tfsdk:"code"`
	DisplayName types.String `tfsdk:"display_name"`
	City        types.String `tfsdk:"city"`
	Country     types.String `tfsdk:"country"`
	Continent   types.String `tfsdk:"continent"`
}

// RegionsDataSourceModel maps the data source schema data
type RegionsDataSourceModel struct {
	Regions []RegionInfo `tfsdk:"regions"`
}

// regionsDataSource is the data source implementation
type regionsDataSource struct{}

// NewRegionsDataSource returns a new regions data source
func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

// Metadata returns the data source type name
func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

// Schema defines the schema for the data source
func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all available monitoring regions.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Description: "List of available regions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Description: "Region code (e.g., 'na-east-ewr').",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "Human-readable region name.",
							Computed:    true,
						},
						"city": schema.StringAttribute{
							Description: "City name.",
							Computed:    true,
						},
						"country": schema.StringAttribute{
							Description: "Country name.",
							Computed:    true,
						},
						"continent": schema.StringAttribute{
							Description: "Continent code (na, sa, eu, ap, au, af).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read fetches the regions list
func (d *regionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state RegionsDataSourceModel

	// The catalog is generated from quismon-api/pkg/constants/platform.go
	// (VultrRegions + RegionDisplayNames); regenerate with
	// `make gen-provider-regions` and verify with `make audit-provider-regions`.
	for _, r := range regionCatalog {
		state.Regions = append(state.Regions, RegionInfo{
			Code:        types.StringValue(r.code),
			DisplayName: types.StringValue(r.displayName),
			City:        types.StringValue(r.city),
			Country:     types.StringValue(r.country),
			Continent:   types.StringValue(r.continent),
		})
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ValidateRegion checks if a region code is valid. The valid set is the
// generated regionCatalog (source: platform.go VultrRegions), so provider-side
// validation cannot accept codes the API rejects.
func ValidateRegion(code string) error {
	for _, r := range regionCatalog {
		if r.code == code {
			return nil
		}
	}
	return fmt.Errorf("invalid region code: %s. Use quismon_regions data source to list available regions", code)
}

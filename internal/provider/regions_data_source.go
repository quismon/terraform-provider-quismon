package provider

import (
	"context"
	"fmt"
	"sort"

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

	// Define all available regions (from quismon-region-manager/config)
	regions := []struct {
		code        string
		displayName string
		city        string
		country     string
		continent   string
	}{
		// North America
		{"na-east-ewr", "Newark, USA (NYC Metro)", "Newark", "United States", "na"},
		{"na-west-sjc", "Silicon Valley, USA", "Silicon Valley", "United States", "na"},
		{"na-west-lax", "Los Angeles, USA", "Los Angeles", "United States", "na"},
		{"na-west-sea", "Seattle, USA", "Seattle", "United States", "na"},
		{"na-central-dfw", "Dallas, USA", "Dallas", "United States", "na"},
		{"na-central-ord", "Chicago, USA", "Chicago", "United States", "na"},
		{"na-east-mia", "Miami, USA", "Miami", "United States", "na"},
		{"na-east-atl", "Atlanta, USA", "Atlanta", "United States", "na"},
		{"na-east-yto", "Toronto, Canada", "Toronto", "Canada", "na"},
		{"na-central-mex", "Mexico City, Mexico", "Mexico City", "Mexico", "na"},
		// South America
		{"sa-east-sao", "São Paulo, Brazil", "São Paulo", "Brazil", "sa"},
		{"sa-west-scl", "Santiago, Chile", "Santiago", "Chile", "sa"},
		// Europe
		{"eu-west-ams", "Amsterdam, Netherlands", "Amsterdam", "Netherlands", "eu"},
		{"eu-west-lhr", "London, UK", "London", "United Kingdom", "eu"},
		{"eu-west-man", "Manchester, UK", "Manchester", "United Kingdom", "eu"},
		{"eu-central-fra", "Frankfurt, Germany", "Frankfurt", "Germany", "eu"},
		{"eu-west-cdg", "Paris, France", "Paris", "France", "eu"},
		{"eu-south-mad", "Madrid, Spain", "Madrid", "Spain", "eu"},
		{"eu-north-waw", "Warsaw, Poland", "Warsaw", "Poland", "eu"},
		{"eu-north-sto", "Stockholm, Sweden", "Stockholm", "Sweden", "eu"},
		// Asia Pacific
		{"ap-northeast-nrt", "Tokyo, Japan", "Tokyo", "Japan", "ap"},
		{"ap-northeast-itm", "Osaka, Japan", "Osaka", "Japan", "ap"},
		{"ap-northeast-icn", "Seoul, South Korea", "Seoul", "South Korea", "ap"},
		{"ap-southeast-sin", "Singapore", "Singapore", "Singapore", "ap"},
		{"ap-south-bom", "Mumbai, India", "Mumbai", "India", "ap"},
		{"ap-south-del", "Delhi NCR, India", "Delhi NCR", "India", "ap"},
		{"ap-south-blr", "Bangalore, India", "Bangalore", "India", "ap"},
		{"ap-west-tlv", "Tel Aviv, Israel", "Tel Aviv", "Israel", "ap"},
		// Australia
		{"au-southeast-syd", "Sydney, Australia", "Sydney", "Australia", "au"},
		{"au-south-mel", "Melbourne, Australia", "Melbourne", "Australia", "au"},
		// Africa
		{"af-south-jnb", "Johannesburg, South Africa", "Johannesburg", "South Africa", "af"},
		// Legacy region codes (for backward compatibility)
		{"fr-par-1", "Paris, France (Legacy)", "Paris", "France", "eu"},
	}

	// Sort by code for consistent output
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].code < regions[j].code
	})

	// Convert to state
	for _, r := range regions {
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

// ValidateRegion checks if a region code is valid
func ValidateRegion(code string) error {
	validRegions := map[string]bool{
		// North America
		"na-east-ewr": true, "na-west-sjc": true, "na-west-lax": true,
		"na-west-sea": true, "na-central-dfw": true, "na-central-ord": true,
		"na-east-mia": true, "na-east-atl": true, "na-east-yto": true,
		"na-central-mex": true,
		// South America
		"sa-east-sao": true, "sa-west-scl": true,
		// Europe
		"eu-west-ams": true, "eu-west-lhr": true, "eu-west-man": true,
		"eu-central-fra": true, "eu-west-cdg": true, "eu-south-mad": true,
		"eu-north-waw": true, "eu-north-sto": true,
		// Asia Pacific
		"ap-northeast-nrt": true, "ap-northeast-itm": true,
		"ap-northeast-icn": true, "ap-southeast-sin": true,
		"ap-south-bom": true, "ap-south-del": true, "ap-south-blr": true,
		"ap-west-tlv": true,
		// Australia
		"au-southeast-syd": true, "au-south-mel": true,
		// Africa
		"af-south-jnb": true,
		// Legacy (for backward compatibility)
		"fr-par-1": true,
		// Old-style codes (for backward compatibility - will be deprecated)
		"us-east-1": true, "eu-west-1": true, "ap-southeast-1": true,
	}

	if !validRegions[code] {
		return fmt.Errorf("invalid region code: %s. Use quismon_regions data source to list available regions", code)
	}
	return nil
}

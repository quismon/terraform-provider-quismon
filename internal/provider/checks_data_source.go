package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

var _ datasource.DataSource = &checksDataSource{}

func NewChecksDataSource() datasource.DataSource {
	return &checksDataSource{}
}

type checksDataSource struct {
	client *client.Client
}

func (d *checksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_checks"
}

func (d *checksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Quismon checks.",
		Attributes: map[string]schema.Attribute{
			"checks": schema.ListNestedAttribute{
				Description: "List of checks.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"health_status": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *checksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

func (d *checksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	checks, err := d.client.ListChecks()
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Checks", err.Error())
		return
	}

	var data struct {
		Checks []struct {
			ID           types.String `tfsdk:"id"`
			Name         types.String `tfsdk:"name"`
			Type         types.String `tfsdk:"type"`
			HealthStatus types.String `tfsdk:"health_status"`
		} `tfsdk:"checks"`
	}

	for _, check := range checks {
		data.Checks = append(data.Checks, struct {
			ID           types.String `tfsdk:"id"`
			Name         types.String `tfsdk:"name"`
			Type         types.String `tfsdk:"type"`
			HealthStatus types.String `tfsdk:"health_status"`
		}{
			ID:           types.StringValue(check.ID),
			Name:         types.StringValue(check.Name),
			Type:         types.StringValue(check.Type),
			HealthStatus: types.StringValue(check.HealthStatus),
		})
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

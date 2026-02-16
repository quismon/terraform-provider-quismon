package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

var _ datasource.DataSource = &checkDataSource{}

func NewCheckDataSource() datasource.DataSource {
	return &checkDataSource{}
}

type checkDataSource struct {
	client *client.Client
}

func (d *checkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

func (d *checkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Quismon check by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Check name.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Check ID.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Check type.",
				Computed:    true,
			},
			"health_status": schema.StringAttribute{
				Description: "Health status.",
				Computed:    true,
			},
		},
	}
}

func (d *checkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*client.Client)
}

func (d *checkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data struct {
		Name         types.String `tfsdk:"name"`
		ID           types.String `tfsdk:"id"`
		Type         types.String `tfsdk:"type"`
		HealthStatus types.String `tfsdk:"health_status"`
	}

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := d.client.GetCheckByName(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Check", err.Error())
		return
	}

	data.ID = types.StringValue(check.ID)
	data.Type = types.StringValue(check.Type)
	data.HealthStatus = types.StringValue(check.HealthStatus)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

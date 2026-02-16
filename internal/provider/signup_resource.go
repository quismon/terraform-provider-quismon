package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

var (
	_ resource.Resource                = &signupResource{}
	_ resource.ResourceWithConfigure   = &signupResource{}
	_ resource.ResourceWithImportState = &signupResource{}
)

func NewSignupResource() resource.Resource {
	return &signupResource{}
}

type signupResource struct {
	baseURL string
}

type signupResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Email                types.String `tfsdk:"email"`
	OrgName              types.String `tfsdk:"org_name"`
	OrgID                types.String `tfsdk:"org_id"`
	APIKey               types.String `tfsdk:"api_key"`
	VerificationRequired types.Bool   `tfsdk:"verification_required"`
}

func (r *signupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_signup"
}

func (r *signupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a new Quismon organization with an API key for quick start. No existing credentials required.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this signup resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address for the organization.",
				Required:    true,
			},
			"org_name": schema.StringAttribute{
				Description: "Name for the organization. Defaults to 'My Organization'.",
				Optional:    true,
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "The ID of the created organization.",
				Computed:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for the organization. Use this for subsequent resource creation.",
				Computed:    true,
				Sensitive:   true,
			},
			"verification_required": schema.BoolAttribute{
				Description: "Whether email verification is required before checks can run.",
				Computed:    true,
			},
		},
	}
}

func (r *signupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.baseURL = client.BaseURL
}

func (r *signupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan signupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create signup request
	signupReq := client.QuickSignupRequest{
		Email:   plan.Email.ValueString(),
		OrgName: plan.OrgName.ValueString(),
	}

	// Perform signup
	signupResp, err := client.QuickSignup(ctx, r.baseURL, signupReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating signup", err.Error())
		return
	}

	// Set the computed values
	plan.ID = types.StringValue(signupResp.OrgID)
	plan.OrgID = types.StringValue(signupResp.OrgID)
	plan.APIKey = types.StringValue(signupResp.APIKey)
	plan.VerificationRequired = types.BoolValue(signupResp.VerificationRequired)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *signupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state signupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Signup resources are not readable after creation
	// Just return the current state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *signupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan signupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Signup resources cannot be updated
	// Just return the current state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *signupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state signupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: Deleting a signup resource does not delete the organization
	// The organization must be deleted via the API if needed
}

func (r *signupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

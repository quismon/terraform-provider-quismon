package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/quismon/terraform-provider-quismon/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &statusPageResource{}
	_ resource.ResourceWithConfigure   = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
)

// NewStatusPageResource is a helper function to simplify the provider implementation.
func NewStatusPageResource() resource.Resource {
	return &statusPageResource{}
}

// statusPageResource manages the organization's status page (singleton: one per org).
type statusPageResource struct {
	client *client.Client
}

// statusPageResourceModel maps the resource schema data.
type statusPageResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Slug                 types.String `tfsdk:"slug"`
	Title                types.String `tfsdk:"title"`
	Description          types.String `tfsdk:"description"`
	LogoURL              types.String `tfsdk:"logo_url"`
	Theme                types.String `tfsdk:"theme"`
	PrimaryColor         types.String `tfsdk:"primary_color"`
	AccessMode           types.String `tfsdk:"access_mode"`
	AccessToken          types.String `tfsdk:"access_token"`
	AccessPassword       types.String `tfsdk:"access_password"`
	AllowedDomains       types.List   `tfsdk:"allowed_domains"`
	ShowResponseTimes    types.Bool   `tfsdk:"show_response_times"`
	ShowUptimePercentage types.Bool   `tfsdk:"show_uptime_percentage"`
	ShowIncidentHistory  types.Bool   `tfsdk:"show_incident_history"`
	IncidentHistoryDays  types.Int64  `tfsdk:"incident_history_days"`
	ShowAllChecks        types.Bool   `tfsdk:"show_all_checks"`
	VisibleCheckIDs      types.List   `tfsdk:"visible_check_ids"`
	URL                  types.String `tfsdk:"url"`
}

// statusPageAPIResponse mirrors the API's status_page object.
type statusPageAPIResponse struct {
	ID                   string   `json:"id"`
	Slug                 string   `json:"slug"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	LogoURL              *string  `json:"logo_url"`
	Theme                *string  `json:"theme"`
	PrimaryColor         *string  `json:"primary_color"`
	AccessMode           *string  `json:"access_mode"`
	AllowedDomains       []string `json:"allowed_domains"`
	ShowResponseTimes    bool     `json:"show_response_times"`
	ShowUptimePercentage bool     `json:"show_uptime_percentage"`
	ShowIncidentHistory  bool     `json:"show_incident_history"`
	IncidentHistoryDays  int      `json:"incident_history_days"`
	ShowAllChecks        bool     `json:"show_all_checks"`
	VisibleCheckIDs      []string `json:"visible_check_ids"`
}

// Metadata returns the resource type name.
func (r *statusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

// Schema defines the schema for the resource.
func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The organization's public status page (one per organization). " +
			"Checks appear on the page when their show_on_status_page flag is set, or all of them when show_all_checks is true.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Status page ID (UUID)",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "URL slug (immutable after creation; page lives at /status/<slug>)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"title": schema.StringAttribute{
				Description: "Page title",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional page description",
				Optional:    true,
			},
			"logo_url": schema.StringAttribute{
				Description: "Optional logo image URL",
				Optional:    true,
			},
			"theme": schema.StringAttribute{
				Description: "Color theme: light or dark (default dark)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("light", "dark"),
				},
			},
			"primary_color": schema.StringAttribute{
				Description: "Primary accent color as hex (e.g. #8b5cf6)",
				Optional:    true,
			},
			"access_mode": schema.StringAttribute{
				Description: "Access protection: public, token, domain, or password. Free tier is always public.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "token", "domain", "password"),
				},
			},
			"access_token": schema.StringAttribute{
				Description: "Access token for token mode (write-only, not returned by the API)",
				Optional:    true,
				Sensitive:   true,
			},
			"access_password": schema.StringAttribute{
				Description: "Access password for password mode (write-only, not returned by the API)",
				Optional:    true,
				Sensitive:   true,
			},
			"allowed_domains": schema.ListAttribute{
				Description: "Allowed email domains for domain mode",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"show_response_times": schema.BoolAttribute{
				Description: "Show response time charts on the page",
				Required:    true,
			},
			"show_uptime_percentage": schema.BoolAttribute{
				Description: "Show uptime percentages on the page",
				Required:    true,
			},
			"show_incident_history": schema.BoolAttribute{
				Description: "Show incident history on the page",
				Required:    true,
			},
			"incident_history_days": schema.Int64Attribute{
				Description: "How many days of incident history to show (default 30)",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 365),
				},
			},
			"show_all_checks": schema.BoolAttribute{
				Description: "Show every enabled check instead of only visible_check_ids",
				Required:    true,
			},
			"visible_check_ids": schema.ListAttribute{
				Description: "Check IDs shown when show_all_checks is false",
				Optional:    true,
				ElementType: types.StringType,
			},
			"url": schema.StringAttribute{
				Description: "Public URL of the status page",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *statusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// requestBody builds the API request payload from the model.
func (r *statusPageResource) requestBody(plan *statusPageResourceModel) map[string]interface{} {
	body := map[string]interface{}{
		"slug":                   plan.Slug.ValueString(),
		"title":                  plan.Title.ValueString(),
		"show_response_times":    plan.ShowResponseTimes.ValueBool(),
		"show_uptime_percentage": plan.ShowUptimePercentage.ValueBool(),
		"show_incident_history":  plan.ShowIncidentHistory.ValueBool(),
		"incident_history_days":  30,
		"show_all_checks":        plan.ShowAllChecks.ValueBool(),
	}

	if !plan.Description.IsNull() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.LogoURL.IsNull() {
		body["logo_url"] = plan.LogoURL.ValueString()
	}
	if !plan.Theme.IsNull() {
		body["theme"] = plan.Theme.ValueString()
	}
	if !plan.PrimaryColor.IsNull() {
		body["primary_color"] = plan.PrimaryColor.ValueString()
	}
	if !plan.AccessMode.IsNull() {
		body["access_mode"] = plan.AccessMode.ValueString()
	}
	if !plan.AccessToken.IsNull() && plan.AccessToken.ValueString() != "" {
		body["access_token"] = plan.AccessToken.ValueString()
	}
	if !plan.AccessPassword.IsNull() && plan.AccessPassword.ValueString() != "" {
		body["access_password"] = plan.AccessPassword.ValueString()
	}
	if !plan.AllowedDomains.IsNull() {
		var domains []string
		plan.AllowedDomains.ElementsAs(context.Background(), &domains, false)
		body["allowed_domains"] = domains
	}
	if !plan.IncidentHistoryDays.IsNull() {
		body["incident_history_days"] = int(plan.IncidentHistoryDays.ValueInt64())
	}
	if !plan.VisibleCheckIDs.IsNull() {
		var ids []string
		plan.VisibleCheckIDs.ElementsAs(context.Background(), &ids, false)
		body["visible_check_ids"] = ids
	}

	return body
}

// Create creates the status page and sets initial Terraform state.
func (r *statusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan statusPageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("POST", "/v1/status-pages", r.requestBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating status page",
			"Could not create status page: "+err.Error(),
		)
		return
	}

	// Read back the created page to populate state
	page, err := r.getStatusPage(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading status page after create",
			"Could not read status page: "+err.Error(),
		)
		return
	}
	if page == nil {
		resp.Diagnostics.AddError(
			"Error reading status page after create",
			"Status page not found after creation",
		)
		return
	}

	r.stateFromAPI(&plan, page)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *statusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state statusPageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.getStatusPage(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading status page",
			"Could not read status page: "+err.Error(),
		)
		return
	}
	if page == nil {
		// Status page was deleted outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	r.stateFromAPI(&state, page)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the status page and sets the updated Terraform state on success.
func (r *statusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan statusPageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("PUT", "/v1/status-pages?slug="+plan.Slug.ValueString(), r.requestBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating status page",
			"Could not update status page: "+err.Error(),
		)
		return
	}

	page, err := r.getStatusPage(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading status page after update",
			"Could not read status page: "+err.Error(),
		)
		return
	}

	r.stateFromAPI(&plan, page)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the status page and removes the Terraform state on success.
func (r *statusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state statusPageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", "/v1/status-pages?slug="+state.Slug.ValueString(), nil)
	if err != nil {
		// Treat 404 as already deleted
		if !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error deleting status page",
				"Could not delete status page: "+err.Error(),
			)
			return
		}
	}
}

// ImportState imports the existing status page into Terraform (singleton, no ID needed).
func (r *statusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	page, err := r.getStatusPage(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing status page",
			"Could not read status page: "+err.Error(),
		)
		return
	}
	if page == nil {
		resp.Diagnostics.AddError(
			"Error importing status page",
			"This organization has no status page to import",
		)
		return
	}

	state := statusPageResourceModel{
		ID:             types.StringValue(page.ID),
		Slug:           types.StringValue(page.Slug),
		Title:          types.StringValue(page.Title),
		AccessToken:    types.StringNull(),
		AccessPassword: types.StringNull(),
	}
	r.stateFromAPI(&state, page)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// getStatusPage fetches the org's status page; nil means none exists.
func (r *statusPageResource) getStatusPage(_ context.Context) (*statusPageAPIResponse, error) {
	data, err := r.client.DoRequest("GET", "/v1/status-pages", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		StatusPage *statusPageAPIResponse `json:"status_page"`
	}
	// the status-pages endpoint returns {"status_page": ...} directly (no
	// {"data": ...} envelope), so unmarshal the body without UnmarshalAPIResponse
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.StatusPage, nil
}

// stateFromAPI maps the API response onto the model. Empty strings from the API
// are mapped back to null so omitted optional attributes don't cause diffs.
func (r *statusPageResource) stateFromAPI(m *statusPageResourceModel, page *statusPageAPIResponse) {
	m.ID = types.StringValue(page.ID)
	m.Slug = types.StringValue(page.Slug)
	m.Title = types.StringValue(page.Title)
	m.URL = types.StringValue(fmt.Sprintf("https://status.quismon.com/%s", page.Slug))

	m.Description = strOrNull(page.Description)
	if page.LogoURL != nil {
		m.LogoURL = strOrNull(*page.LogoURL)
	} else {
		m.LogoURL = types.StringNull()
	}
	if page.Theme != nil {
		m.Theme = strOrNull(*page.Theme)
	} else {
		m.Theme = types.StringNull()
	}
	if page.PrimaryColor != nil {
		m.PrimaryColor = strOrNull(*page.PrimaryColor)
	} else {
		m.PrimaryColor = types.StringNull()
	}
	if page.AccessMode != nil {
		m.AccessMode = strOrNull(*page.AccessMode)
	} else {
		m.AccessMode = types.StringNull()
	}

	if len(page.AllowedDomains) > 0 {
		domains, _ := types.ListValueFrom(context.Background(), types.StringType, page.AllowedDomains)
		m.AllowedDomains = domains
	} else {
		m.AllowedDomains = types.ListNull(types.StringType)
	}

	m.ShowResponseTimes = types.BoolValue(page.ShowResponseTimes)
	m.ShowUptimePercentage = types.BoolValue(page.ShowUptimePercentage)
	m.ShowIncidentHistory = types.BoolValue(page.ShowIncidentHistory)
	if page.IncidentHistoryDays > 0 {
		m.IncidentHistoryDays = types.Int64Value(int64(page.IncidentHistoryDays))
	} else {
		m.IncidentHistoryDays = types.Int64Value(30)
	}
	m.ShowAllChecks = types.BoolValue(page.ShowAllChecks)

	if len(page.VisibleCheckIDs) > 0 {
		ids, _ := types.ListValueFrom(context.Background(), types.StringType, page.VisibleCheckIDs)
		m.VisibleCheckIDs = ids
	} else {
		m.VisibleCheckIDs = types.ListNull(types.StringType)
	}
}

func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

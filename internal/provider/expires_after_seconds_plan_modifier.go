package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// expiresAfterSecondsModifier is a plan modifier that preserves the API value
// for expires_after_seconds unless the user explicitly sets it.
// This prevents Terraform from trying to "fix" expiring checks.
type expiresAfterSecondsModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m expiresAfterSecondsModifier) Description(_ context.Context) string {
	return "Preserves expires_after_seconds from API unless explicitly set in configuration."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m expiresAfterSecondsModifier) MarkdownDescription(_ context.Context) string {
	return "Preserves `expires_after_seconds` from API unless explicitly set in configuration. This prevents Terraform from tampering with temporary/expiring checks."
}

// PlanModifyInt64 implements the plan modifier interface.
func (m expiresAfterSecondsModifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// If the plan already has a value set (user explicitly configured it), use that
	if !req.PlanValue.IsNull() && req.PlanValue.ValueInt64() > 0 {
		resp.PlanValue = req.PlanValue
		return
	}

	// If the state has a value (from API), preserve it
	if !req.StateValue.IsNull() {
		resp.PlanValue = req.StateValue
		return
	}

	// Otherwise, keep the plan value as-is (null)
	resp.PlanValue = req.PlanValue
}

// ExpiresAfterSecondsModifier returns a plan modifier that preserves
// the expires_after_seconds value from the API unless explicitly configured.
func ExpiresAfterSecondsModifier() planmodifier.Int64 {
	return expiresAfterSecondsModifier{}
}

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mergeStartsBrokenAlias resolves the starts_broken family: starts_broken is
// the preferred attribute; alert_after_first_success is the deprecated alias.
// starts_broken wins when both are provided.
func mergeStartsBrokenAlias(startsBroken, alias types.Bool) *bool {
	if !startsBroken.IsNull() && !startsBroken.IsUnknown() {
		v := startsBroken.ValueBool()
		return &v
	}
	if !alias.IsNull() && !alias.IsUnknown() {
		v := alias.ValueBool()
		return &v
	}
	return nil
}

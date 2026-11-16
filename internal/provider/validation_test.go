package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Table-driven tests for validateDomConfig: the plan-time gate for dom
// checks. Mirrors the API's write-time contract (url-or-script, 16KB cap) so
// configuration errors surface at plan, not apply.

func TestValidateDomConfig(t *testing.T) {
	sixteenKBPlusOne := make([]byte, 16*1024+1)
	for i := range sixteenKBPlusOne {
		sixteenKBPlusOne[i] = 'a'
	}
	exactly16KB := make([]byte, 16*1024)
	for i := range exactly16KB {
		exactly16KB[i] = 'a'
	}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errHint string
	}{
		{
			name:   "declarative url only",
			config: map[string]interface{}{"url": "https://example.com/"},
		},
		{
			name:   "declarative url with assertions",
			config: map[string]interface{}{"url": "https://example.com/", "expected_text": "Welcome", "expected_status_code": "200"},
		},
		{
			name:   "script only",
			config: map[string]interface{}{"script": "return { ok: true };"},
		},
		{
			name:   "url and script together",
			config: map[string]interface{}{"url": "https://example.com/", "script": "return { ok: true };"},
		},
		{
			name:   "script with secrets map",
			config: map[string]interface{}{"script": "return { ok: true };", "secrets": map[string]interface{}{"APP_PASSWORD": "hunter2"}},
		},
		{
			name:   "script at exactly 16KB is allowed",
			config: map[string]interface{}{"script": string(exactly16KB)},
		},
		{
			name:    "script over 16KB",
			config:  map[string]interface{}{"script": string(sixteenKBPlusOne)},
			wantErr: true,
			errHint: "16KB",
		},
		{
			name:    "empty script",
			config:  map[string]interface{}{"script": ""},
			wantErr: true,
			errHint: "must not be empty",
		},
		{
			name:    "script wrong type",
			config:  map[string]interface{}{"script": 42},
			wantErr: true,
			errHint: "must be a string",
		},
		{
			name:    "neither url nor script",
			config:  map[string]interface{}{"expected_text": "Welcome"},
			wantErr: true,
			errHint: "requires either",
		},
		{
			name:    "empty url without script",
			config:  map[string]interface{}{"url": "", "expected_text": "Welcome"},
			wantErr: true,
			errHint: "must be a non-empty string",
		},
		{
			name:   "empty url with script is fine (script navigates)",
			config: map[string]interface{}{"url": "", "script": "return { ok: true };"},
		},
		{
			name:    "empty config",
			config:  map[string]interface{}{},
			wantErr: true,
			errHint: "requires either",
		},
		{
			name:   "expected_text array",
			config: map[string]interface{}{"url": "https://example.com/", "expected_text": []interface{}{"Welcome", "Sign out"}},
		},
		{
			name:   "expected_text string list",
			config: map[string]interface{}{"url": "https://example.com/", "expected_text": []string{"Welcome"}},
		},
		{
			name:    "expected_text empty array",
			config:  map[string]interface{}{"url": "https://example.com/", "expected_text": []interface{}{}},
			wantErr: true,
			errHint: "at least one",
		},
		{
			name:    "expected_text blank entry",
			config:  map[string]interface{}{"url": "https://example.com/", "expected_text": []interface{}{"ok", "  "}},
			wantErr: true,
			errHint: "must not be empty",
		},
		{
			name:    "expected_text non-string entry",
			config:  map[string]interface{}{"url": "https://example.com/", "expected_text": []interface{}{"ok", 2}},
			wantErr: true,
			errHint: "must be a string",
		},
		{
			name:    "expected_text wrong type",
			config:  map[string]interface{}{"url": "https://example.com/", "expected_text": 2},
			wantErr: true,
			errHint: "string or an array",
		},
		{
			name:    "script and expected_text conflict",
			config:  map[string]interface{}{"script": "return { ok: true };", "expected_text": "Welcome"},
			wantErr: true,
			errHint: "cannot be combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDomConfig(%v) expected error, got nil", tt.config)
					return
				}
				if tt.errHint != "" && !contains(err.Error(), tt.errHint) {
					t.Errorf("validateDomConfig(%v) error %q does not contain hint %q", tt.config, err.Error(), tt.errHint)
				}
				return
			}
			if err != nil {
				t.Errorf("validateDomConfig(%v) unexpected error: %v", tt.config, err)
			}
		})
	}
}

// TestValidateDomConfigDoesNotRejectNestedSecrets: a regression class from the
// flat-string config era - the nested secrets object must pass through
// untouched (masking is the API's job, not the provider's).
func TestValidateDomConfigDoesNotRejectNestedSecrets(t *testing.T) {
	cfg := map[string]interface{}{
		"url":     "https://example.com/",
		"secrets": map[string]interface{}{"TOKEN": "abc", "EMPTY": ""},
	}
	if err := validateDomConfig(cfg); err != nil {
		t.Errorf("unexpected error for config with nested secrets: %v", err)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// --- Q-059: result_webhook scheme preflight (plan-time) ---

func TestResultWebhookValidator(t *testing.T) {
	type tc struct {
		name    string
		value   string
		null    bool
		wantErr bool
	}
	cases := []tc{
		{"null passes (clears on update)", "", true, false},
		{"empty passes (clears on update)", "", false, false},
		{"https passes", "https://ops.example.com/hook", false, false},
		{"http passes", "http://intranet.local/hook", false, false},
		{"token-in-query passes (scheme is the only provider rule)", "https://ops.example.com/h?token=abc", false, false},
		{"ftp rejected", "ftp://ops.example.com/hook", false, true},
		{"relative rejected", "/hook", false, true},
		{"no host rejected", "https://", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := resultWebhookValidator{}
			var cv types.String
			if c.null {
				cv = types.StringNull()
			} else {
				cv = types.StringValue(c.value)
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{ConfigValue: cv}, resp)
			if c.wantErr && !resp.Diagnostics.HasError() {
				t.Errorf("expected error for %q", c.value)
			}
			if !c.wantErr && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error for %q", c.value)
			}
		})
	}
}

// --- Q-059: regionThresholdPtr / resultWebhookPtr contract helpers ---

func TestRegionThresholdPtrAlwaysConcrete(t *testing.T) {
	if got := *regionThresholdPtr(types.Int64Null()); got != 1 {
		t.Errorf("null -> %d, want 1 (create default)", got)
	}
	if got := *regionThresholdPtr(types.Int64Value(3)); got != 3 {
		t.Errorf("3 -> %d, want 3", got)
	}
	if got := *regionThresholdPtr(types.Int64Value(0)); got != 1 {
		t.Errorf("0 -> %d, want clamp-to-1", got)
	}
}

func TestResultWebhookPtrClearsOnNull(t *testing.T) {
	if got := *resultWebhookPtr(types.StringNull()); got != "" {
		t.Errorf("null -> %q, want empty (clears)", got)
	}
	if got := *resultWebhookPtr(types.StringValue("https://x.example.com")); got != "https://x.example.com" {
		t.Errorf("value -> %q", got)
	}
}

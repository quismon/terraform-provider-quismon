package provider

import "testing"

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

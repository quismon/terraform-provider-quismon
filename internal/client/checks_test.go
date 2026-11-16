package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// Q-059 payload-construction contract tests. The design's two hard
// constraints live HERE, at the marshaling layer, so a future refactor that
// flips them fails these tests before it can re-open the bricking window or
// make webhooks unremovable:
//   - region_threshold is ALWAYS present in update payloads (a concrete
//     pointer; null config resolves to the create default 1)
//   - result_webhook null CLEARS on update (pointer-to-empty serializes as
//     "") and is omitted on create when empty (omitempty)

func TestUpdateCheckRequestRegionThresholdAlwaysPresent(t *testing.T) {
	one := 1
	three := 3
	cases := []struct {
		name string
		req  UpdateCheckRequest
		want string
	}{
		{"explicit 3", UpdateCheckRequest{RegionThreshold: &three}, `"region_threshold":3`},
		{"null-config fallback 1", UpdateCheckRequest{RegionThreshold: &one}, `"region_threshold":1`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(b), tc.want) {
				t.Errorf("update payload missing %s: %s", tc.want, b)
			}
		})
	}
}

func TestUpdateCheckRequestResultWebhookClearsOnNull(t *testing.T) {
	// A NIL pointer preserves (raw API semantics, non-resource callers);
	// the RESOURCE always builds a pointer — these tests pin both halves.
	empty := ""
	set := "https://ops.example.com/hook"
	b, _ := json.Marshal(UpdateCheckRequest{ResultWebhook: &empty})
	if !strings.Contains(string(b), `"result_webhook":""`) {
		t.Errorf("pointer-to-empty must serialize as an explicit empty string (clears): %s", b)
	}
	b, _ = json.Marshal(UpdateCheckRequest{ResultWebhook: &set})
	if !strings.Contains(string(b), `"result_webhook":"https://ops.example.com/hook"`) {
		t.Errorf("pointer-to-value must serialize the value: %s", b)
	}
	b, _ = json.Marshal(UpdateCheckRequest{})
	if strings.Contains(string(b), "result_webhook") {
		t.Errorf("nil pointer must omit the field entirely (preserves): %s", b)
	}
}

func TestCreateCheckRequestWebhookOmittedWhenEmpty(t *testing.T) {
	b, _ := json.Marshal(CreateCheckRequest{})
	if strings.Contains(string(b), "result_webhook") {
		t.Errorf("empty create webhook must be omitted (omitempty): %s", b)
	}
	b, _ = json.Marshal(CreateCheckRequest{ResultWebhook: "https://ops.example.com/hook"})
	if !strings.Contains(string(b), `"result_webhook":"https://ops.example.com/hook"`) {
		t.Errorf("create webhook must serialize: %s", b)
	}
}

func TestCheckResponseParsesNewFields(t *testing.T) {
	// Refresh parity: Read maps these from the API response.
	raw := `{"id":"11111111-1111-4111-8111-111111111111","org_id":"22222222-2222-4222-8222-222222222222",
		"name":"n","type":"https","config":{},"interval_seconds":60,"regions":["na-east-ewr"],
		"region_threshold":2,"enabled":true,"result_webhook":"https://ops.example.com/hook",
		"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`
	var c Check
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.RegionThreshold != 2 {
		t.Errorf("region_threshold = %d, want 2", c.RegionThreshold)
	}
	if c.ResultWebhook != "https://ops.example.com/hook" {
		t.Errorf("result_webhook = %q", c.ResultWebhook)
	}
}

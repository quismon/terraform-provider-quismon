package provider

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// validatePlaywrightConfig validates the Playwright check configuration
// Returns an error if the configuration is invalid
func validatePlaywrightConfig(configMap map[string]interface{}) error {
	// Validate that at least one script source is provided
	scriptID, _ := configMap["script_id"].(string)
	scriptInline, _ := configMap["script_inline"].(string)
	scriptURL, _ := configMap["script_url"].(string)

	if scriptID == "" && scriptInline == "" && scriptURL == "" {
		return fmt.Errorf("playwright check requires at least one of: script_id, script_inline, or script_url")
	}

	// If using low-code actions, validate script_inline is not also set
	if actions, ok := configMap["actions"].([]interface{}); ok && len(actions) > 0 {
		if scriptInline != "" {
			return fmt.Errorf("cannot use both 'script_inline' and 'actions' in the same check")
		}
	}

	// Validate browser type if provided
	if browser, ok := configMap["browser"].(string); ok && browser != "" {
		validBrowsers := map[string]bool{
			"chromium": true,
			"firefox":  true,
			"webkit":   true,
		}
		if !validBrowsers[browser] {
			return fmt.Errorf("browser must be one of: chromium, firefox, webkit (got: %s)", browser)
		}
	}

	// Validate viewport dimensions if provided
	if viewportWidth, ok := configMap["viewport_width"]; ok && viewportWidth != nil {
		width := parseIntFromInterface(viewportWidth, 0)
		if width < 100 || width > 4096 {
			return fmt.Errorf("viewport_width must be between 100 and 4096 pixels (got: %d)", width)
		}
	}

	if viewportHeight, ok := configMap["viewport_height"]; ok && viewportHeight != nil {
		height := parseIntFromInterface(viewportHeight, 0)
		if height < 100 || height > 4096 {
			return fmt.Errorf("viewport_height must be between 100 and 4096 pixels (got: %d)", height)
		}
	}

	// Validate device_scale_factor if provided
	if deviceScale, ok := configMap["device_scale_factor"]; ok && deviceScale != nil {
		scale := parseFloatFromInterface(deviceScale, 1.0)
		if scale < 0.1 || scale > 10.0 {
			return fmt.Errorf("device_scale_factor must be between 0.1 and 10.0 (got: %.1f)", scale)
		}
	}

	// Validate timeout_seconds (1-300 seconds)
	if timeout, ok := configMap["timeout_seconds"]; ok && timeout != nil {
		timeoutSeconds := parseIntFromInterface(timeout, 30)
		if timeoutSeconds < 1 || timeoutSeconds > 300 {
			return fmt.Errorf("timeout_seconds must be between 1 and 300 seconds (got: %d)", timeoutSeconds)
		}
	}

	// Validate low-code actions if provided
	if actionsRaw, ok := configMap["actions"]; ok {
		actions, ok := actionsRaw.([]interface{})
		if !ok {
			return fmt.Errorf("'actions' must be an array")
		}

		if len(actions) > 0 {
			for i, actionRaw := range actions {
				action, ok := actionRaw.(map[string]interface{})
				if !ok {
					return fmt.Errorf("action %d must be an object", i+1)
				}

				// Validate action name
				name, _ := action["name"].(string)
				if name == "" {
					return fmt.Errorf("action %d must have a 'name' field", i+1)
				}

				// Validate action type
				actionType, _ := action["action"].(string)
				if actionType == "" {
					return fmt.Errorf("action %d must have an 'action' field", i+1)
				}

				validActionTypes := map[string]bool{
					"goto": true, "click": true, "fill": true, "type": true,
					"wait": true, "screenshot": true, "evaluate": true,
					"select": true, "check": true, "uncheck": true,
					"press": true, "hover": true, "scroll": true,
				}

				if !validActionTypes[actionType] {
					return fmt.Errorf("action %d has invalid action type '%s'", i+1, actionType)
				}

				// Validate required fields based on action type
				switch actionType {
				case "goto":
					url, _ := action["url"].(string)
					if url == "" {
						return fmt.Errorf("action %d (goto) requires 'url' field", i+1)
					}
				case "click", "check", "uncheck", "hover", "scroll":
					selector, _ := action["selector"].(string)
					if selector == "" {
						return fmt.Errorf("action %d (%s) requires 'selector' field", i+1, actionType)
					}
				case "fill", "type":
					selector, _ := action["selector"].(string)
					if selector == "" {
						return fmt.Errorf("action %d (%s) requires 'selector' field", i+1, actionType)
					}
					value, _ := action["value"].(string)
					if value == "" {
						return fmt.Errorf("action %d (%s) requires 'value' field", i+1, actionType)
					}
				case "select":
					selector, _ := action["selector"].(string)
					if selector == "" {
						return fmt.Errorf("action %d (select) requires 'selector' field", i+1)
					}
					value, _ := action["value"].(string)
					if value == "" {
						return fmt.Errorf("action %d (select) requires 'value' field", i+1)
					}
				case "press":
					value, _ := action["value"].(string)
					if value == "" {
						return fmt.Errorf("action %d (press) requires 'value' field", i+1)
					}
				case "evaluate":
					script, _ := action["script"].(string)
					if script == "" {
						return fmt.Errorf("action %d (evaluate) requires 'script' field", i+1)
					}
				}
			}
		}
	}

	// Validate extracts if provided
	if extractsRaw, ok := configMap["extracts"]; ok {
		extracts, ok := extractsRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'extracts' must be an object")
		}

		for varName, extractRaw := range extracts {
			extract, ok := extractRaw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("extract '%s' must be an object", varName)
			}
			// At least one extraction method must be specified
			hasMethod := false
			if _, ok := extract["selector"].(string); ok {
				hasMethod = true
			}
			if _, ok := extract["jsonpath"].(string); ok {
				hasMethod = true
			}
			if _, ok := extract["regex"].(string); ok {
				hasMethod = true
			}
			if _, ok := extract["attribute"].(string); ok {
				hasMethod = true
			}
			if _, ok := extract["property"].(string); ok {
				hasMethod = true
			}
			if !hasMethod {
				return fmt.Errorf("extract '%s' must have at least one of: selector, jsonpath, regex, attribute, or property", varName)
			}
		}
	}

	// Validate assertions if provided
	if assertionsRaw, ok := configMap["assertions"]; ok {
		assertions, ok := assertionsRaw.([]interface{})
		if !ok {
			return fmt.Errorf("'assertions' must be an array")
		}

		for i, assertionRaw := range assertions {
			assertion, ok := assertionRaw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("assertion %d must be an object", i+1)
			}

			// Validate assertion name
			name, _ := assertion["name"].(string)
			if name == "" {
				return fmt.Errorf("assertion %d must have a 'name' field", i+1)
			}

			// Validate assertion type
			assertionType, _ := assertion["type"].(string)
			if assertionType == "" {
				return fmt.Errorf("assertion %d must have a 'type' field", i+1)
			}

			validAssertionTypes := map[string]bool{
				"text": true, "value": true, "visible": true, "enabled": true,
				"url": true, "title": true, "count": true, "contains": true,
			}

			if !validAssertionTypes[assertionType] {
				return fmt.Errorf("assertion %d has invalid type '%s'", i+1, assertionType)
			}

			// Validate expected value is provided
			expected, _ := assertion["expected"].(string)
			if expected == "" {
				return fmt.Errorf("assertion %d must have an 'expected' value", i+1)
			}
		}
	}

	return nil
}

// parseIntFromInterface extracts an int from an interface{}
// Handles int, float64, and string types
func parseIntFromInterface(v interface{}, defaultValue int) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// parseFloatFromInterface extracts a float64 from an interface{}
// Handles int, float64, and string types
func parseFloatFromInterface(v interface{}, defaultValue float64) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	case string:
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// validatePlaywrightConfigJSON validates Playwright config from a JSON string
// This is used for Terraform plan-time validation
func validatePlaywrightConfigJSON(configJSON string, checkType string) error {
	if checkType != "playwright" {
		return nil // Not a Playwright check, no validation needed
	}

	if configJSON == "" {
		return fmt.Errorf("playwright check requires 'config_json' to be specified")
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &configMap); err != nil {
		return fmt.Errorf("could not parse config_json as JSON: %w", err)
	}

	return validatePlaywrightConfig(configMap)
}

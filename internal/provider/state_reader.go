package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TerraformState represents the structure of a Terraform state file
type TerraformState struct {
	Version   int `json:"version"`
	Resources []struct {
		Type      string `json:"type"`
		Name      string `json:"name"`
		Instances []struct {
			Attributes map[string]interface{} `json:"attributes"`
		} `json:"instances"`
	} `json:"resources"`
}

// findStateFile searches for terraform state files in common locations
func findStateFile() string {
	// Common state file locations in order of priority
	candidates := []string{
		"terraform.tfstate",
		".terraform/terraform.tfstate",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absPath, _ := filepath.Abs(candidate)
			return absPath
		}
	}

	return ""
}

// readAPIKeyFromState attempts to read the API key from a terraform.tfstate file
// Returns the API key if found, empty string otherwise, and any warning message
func readAPIKeyFromState() (apiKey string, warning string) {
	statePath := findStateFile()
	if statePath == "" {
		return "", ""
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", ""
	}

	var state TerraformState
	if err := json.Unmarshal(data, &state); err != nil {
		return "", ""
	}

	// Look for quismon_signup resources
	for _, resource := range state.Resources {
		if resource.Type == "quismon_signup" && len(resource.Instances) > 0 {
			attrs := resource.Instances[0].Attributes
			if key, ok := attrs["api_key"].(string); ok && key != "" {
				warning = fmt.Sprintf(
					"Using API key from terraform state (%s).\n\n"+
						"💡 Tip: For future runs, export the key:\n"+
						"   export QUISMON_API_KEY=$(terraform output -raw api_key)\n\n"+
						"📧 Verify your email to increase your free tier check frequency from 20/hr to 60/hr!",
					statePath,
				)
				return key, warning
			}
		}
	}

	return "", ""
}

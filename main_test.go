package main

import (
	"log"
	"strings"
	"testing"
)

func TestCheckAllFields(t *testing.T) {
	// TODO: Write more test cases for version validation

	// Valid input
	validInput := "* RC: No\r\n * dapr/components-contrib: VERSION\r\n * dapr/dapr: VERSION\r\n * dapr/cli: VERSION\r\n * dapr/dashboard: VERSION\r\n * SDKs:\r\n   * go: VERSION\r\n   * rust: VERSION\r\n   * python: VERSION\r\n   * dotnet: VERSION\r\n   * java: VERSION\r\n   * js: VERSION"

	// Parse the valid input
	core, sdks, err := ParseMarkdown(validInput)
	if err != nil {
		t.Fatalf("Failed to parse valid input: %v", err)
	}

	log.Println(core)
	log.Println(sdks)

	// Test case 1: Valid input should not return an error
	t.Run("ValidInput", func(t *testing.T) {
		err := CheckAllFields(core, sdks)
		if err != nil {
			t.Errorf("CheckAllFields returned an error for valid input: %v", err)
		}
	})

	// Test case 2: Missing core field
	t.Run("MissingCoreField", func(t *testing.T) {
		incompleteCoreFields := []DaprCore{
			{Name: "RC", Value: "No"},
			{Name: "dapr/components-contrib", Value: "VERSION"},
			{Name: "dapr/dapr", Value: "VERSION"},
			{Name: "dapr/cli", Value: "VERSION"},
			// Missing dapr/dashboard
		}
		err := CheckAllFields(incompleteCoreFields, sdks)
		if err == nil {
			t.Error("CheckAllFields should return an error for missing core field")
		}
		if err != nil && !strings.Contains(err.Error(), "missing core definition: dapr/dashboard") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	// Test case 3: Missing SDK field
	t.Run("MissingSDKField", func(t *testing.T) {
		incompleteSDKFields := []DaprSDK{
			{Name: "go", Value: "VERSION"},
			{Name: "rust", Value: "VERSION"},
			{Name: "python", Value: "VERSION"},
			{Name: "dotnet", Value: "VERSION"},
			{Name: "java", Value: "VERSION"},
			// Missing js
		}
		err := CheckAllFields(core, incompleteSDKFields)
		if err == nil {
			t.Error("CheckAllFields should return an error for missing SDK field")
		}
		if err != nil && !strings.Contains(err.Error(), "missing SDK definition: js") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	// Test case 4: Empty input
	t.Run("EmptyInput", func(t *testing.T) {
		err := CheckAllFields([]DaprCore{}, []DaprSDK{})
		if err == nil {
			t.Error("CheckAllFields should return an error for empty input")
		}
	})
}

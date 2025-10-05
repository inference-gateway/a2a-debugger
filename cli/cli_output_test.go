package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func TestGetOutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		configValue    string
		expectedFormat OutputFormat
	}{
		{
			name:           "Default YAML format",
			configValue:    "",
			expectedFormat: OutputFormatYAML,
		},
		{
			name:           "Explicit YAML format",
			configValue:    "yaml",
			expectedFormat: OutputFormatYAML,
		},
		{
			name:           "JSON format",
			configValue:    "json",
			expectedFormat: OutputFormatJSON,
		},
		{
			name:           "Case insensitive JSON",
			configValue:    "JSON",
			expectedFormat: OutputFormatJSON,
		},
		{
			name:           "Invalid format defaults to YAML",
			configValue:    "invalid",
			expectedFormat: OutputFormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up viper configuration
			if tt.configValue != "" {
				viper.Set("output", tt.configValue)
			} else {
				viper.Set("output", "yaml") // Set default
			}

			format := getOutputFormat()
			if format != tt.expectedFormat {
				t.Errorf("Expected format %v, got %v", tt.expectedFormat, format)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	testData := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"active":  true,
	}

	tests := []struct {
		name         string
		outputFormat string
		data         interface{}
		validate     func([]byte) error
	}{
		{
			name:         "YAML output",
			outputFormat: "yaml",
			data:         testData,
			validate: func(output []byte) error {
				var result map[string]interface{}
				if err := yaml.Unmarshal(output, &result); err != nil {
					return fmt.Errorf("failed to parse YAML: %w", err)
				}
				if result["name"] != "test" {
					return fmt.Errorf("expected name 'test', got %v", result["name"])
				}
				return nil
			},
		},
		{
			name:         "JSON output",
			outputFormat: "json",
			data:         testData,
			validate: func(output []byte) error {
				var result map[string]interface{}
				if err := json.Unmarshal(output, &result); err != nil {
					return fmt.Errorf("failed to parse JSON: %w", err)
				}
				if result["name"] != "test" {
					return fmt.Errorf("expected name 'test', got %v", result["name"])
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("output", tt.outputFormat)

			output, err := formatOutput(tt.data)
			if err != nil {
				t.Fatalf("formatOutput failed: %v", err)
			}

			if err := tt.validate(output); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

func TestPrintFormatted(t *testing.T) {
	testData := map[string]interface{}{
		"message": "hello world",
		"status":  "success",
	}

	tests := []struct {
		name         string
		outputFormat string
		expectError  bool
	}{
		{
			name:         "YAML print success",
			outputFormat: "yaml",
			expectError:  false,
		},
		{
			name:         "JSON print success",
			outputFormat: "json",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("output", tt.outputFormat)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := printFormatted(testData)

			_ = w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(output) == 0 {
					t.Error("Expected output but got none")
				}

				// Verify the output can be parsed back
				if tt.outputFormat == "json" {
					var result map[string]interface{}
					if err := json.Unmarshal([]byte(output), &result); err != nil {
						t.Errorf("Failed to parse JSON output: %v", err)
					}
				} else {
					var result map[string]interface{}
					if err := yaml.Unmarshal([]byte(output), &result); err != nil {
						t.Errorf("Failed to parse YAML output: %v", err)
					}
				}
			}
		})
	}
}

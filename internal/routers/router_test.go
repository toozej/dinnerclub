package routers

import (
	"html/template"
	"strings"
	"testing"
)

func TestCreateTemplateFuncMap(t *testing.T) {
	funcMap := createTemplateFuncMap()

	// Test that all expected helper functions are registered
	expectedFunctions := []string{
		"FormatBoolean",
		"FormatRating",
		"GetRatingDescription",
		"GetRatingLabel",
		"UserProfileLink",
		"TooltipText",
		"GetRatingOptions",
		"SafeHTML",
		"Add",
		"add", // lowercase alias for compatibility
		"StringToInt",
		"IsEmpty",
		"FormatDate",
		"FormatDateTime",
		"CSRFToken",
		"CSRFField",
		"CSRFInput",
		"sub",
		"Sub",
		"mul",
		"Mul",
		"default",
	}

	for _, funcName := range expectedFunctions {
		if _, exists := funcMap[funcName]; !exists {
			t.Errorf("Expected function %s not found in template function map", funcName)
		}
	}

	// Test that the function map has the expected number of functions
	if len(funcMap) != len(expectedFunctions) {
		t.Errorf("Expected %d functions in template function map, got %d", len(expectedFunctions), len(funcMap))
	}
}

func TestSetupTemplatesWithHelpers(t *testing.T) {
	// This test verifies that templates can be parsed with the helper functions
	// without actually starting the full router

	funcMap := createTemplateFuncMap()

	// Test that we can create a template with the function map
	tmpl := template.New("test").Funcs(funcMap)

	// Test a simple template that uses one of our helper functions
	testTemplate := `{{FormatBoolean true}}`

	_, err := tmpl.Parse(testTemplate)
	if err != nil {
		t.Errorf("Failed to parse template with helper functions: %v", err)
	}
}
func TestTemplateHelperIntegration(t *testing.T) {
	// Test that helper functions work correctly when used in templates
	funcMap := createTemplateFuncMap()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "FormatBoolean true",
			template: `{{FormatBoolean true}}`,
			expected: "Yes",
		},
		{
			name:     "FormatBoolean false",
			template: `{{FormatBoolean false}}`,
			expected: "No",
		},
		{
			name:     "FormatRating",
			template: `{{FormatRating 3}}`,
			expected: "3 - Good",
		},
		{
			name:     "UserProfileLink",
			template: `{{UserProfileLink "testuser"}}`,
			expected: `<a href="/entries/submittedby/testuser">testuser</a>`,
		},
		{
			name:     "Add function",
			template: `{{Add 2 3}}`,
			expected: "5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(funcMap).Parse(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var buf strings.Builder
			err = tmpl.Execute(&buf, nil)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			result := buf.String()
			if result != test.expected {
				t.Errorf("Template execution result = %q, expected %q", result, test.expected)
			}
		})
	}
}

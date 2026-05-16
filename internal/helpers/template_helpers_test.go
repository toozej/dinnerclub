package helpers

import (
	"html/template"
	"testing"
	"time"
)

func TestFormatBoolean(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    bool
		expected string
	}{
		{true, "Yes"},
		{false, "No"},
	}

	for _, test := range tests {
		result := h.FormatBoolean(test.input)
		if result != test.expected {
			t.Errorf("FormatBoolean(%v) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestFormatRating(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    int
		expected string
	}{
		{0, "Not Rated"},
		{1, "1 - Poor"},
		{3, "3 - Good"},
		{5, "5 - Outstanding"},
	}

	for _, test := range tests {
		result := h.FormatRating(test.input)
		if result != test.expected {
			t.Errorf("FormatRating(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGetRatingLabel(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    int
		expected string
	}{
		{0, "Not Rated"},
		{1, "Poor"},
		{3, "Good"},
		{5, "Outstanding"},
	}

	for _, test := range tests {
		result := h.GetRatingLabel(test.input)
		if result != test.expected {
			t.Errorf("GetRatingLabel(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestUserProfileLink(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    string
		expected template.HTML
	}{
		{"", template.HTML("Anonymous")},
		{"testuser", template.HTML(`<a href="/entries/submittedby/testuser">testuser</a>`)},
		{"user with spaces", template.HTML(`<a href="/entries/submittedby/user with spaces">user with spaces</a>`)},
	}

	for _, test := range tests {
		result := h.UserProfileLink(test.input)
		if result != test.expected {
			t.Errorf("UserProfileLink(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestTooltipText(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    string
		expected string
	}{
		{"name", "Enter the restaurant name"},
		{"foodrating", "Rate the food quality (1=Poor, 2=Fair, 3=Good, 4=Very Good, 5=Outstanding)"},
		{"ordered", "List items you ordered (separate multiple items with commas)"},
		{"nonexistent", ""},
	}

	for _, test := range tests {
		result := h.TooltipText(test.input)
		if result != test.expected {
			t.Errorf("TooltipText(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGetRatingOptions(t *testing.T) {
	h := &TemplateHelpers{}

	options := h.GetRatingOptions()

	// Should return 5 options (1-5, excluding 0)
	if len(options) != 5 {
		t.Errorf("GetRatingOptions() returned %d options, expected 5", len(options))
	}

	// First option should be rating 1
	if options[0].Value != 1 {
		t.Errorf("First rating option value = %d, expected 1", options[0].Value)
	}

	// Last option should be rating 5
	if options[4].Value != 5 {
		t.Errorf("Last rating option value = %d, expected 5", options[4].Value)
	}
}

func TestAdd(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 3},
		{0, 0, 0},
		{-1, 1, 0},
		{10, 5, 15},
	}

	for _, test := range tests {
		result := h.Add(test.a, test.b)
		if result != test.expected {
			t.Errorf("Add(%d, %d) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestStringToInt(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"-5", -5},
		{"invalid", 0}, // Should return 0 for invalid strings
		{"", 0},
	}

	for _, test := range tests {
		result := h.StringToInt(test.input)
		if result != test.expected {
			t.Errorf("StringToInt(%s) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	h := &TemplateHelpers{}

	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"hello", false},
		{" hello ", false},
		{"0", false},
	}

	for _, test := range tests {
		result := h.IsEmpty(test.input)
		if result != test.expected {
			t.Errorf("IsEmpty(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestFormatDate(t *testing.T) {
	h := &TemplateHelpers{}

	// Create a specific test time
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	expected := "Dec 25 2023"

	result := h.FormatDate(testTime)
	if result != expected {
		t.Errorf("FormatDate(%v) = %s, expected %s", testTime, result, expected)
	}
}

func TestFormatDateTime(t *testing.T) {
	h := &TemplateHelpers{}

	// Create a specific test time
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	expected := "Dec 25 2023 3:30 PM"

	result := h.FormatDateTime(testTime)
	if result != expected {
		t.Errorf("FormatDateTime(%v) = %s, expected %s", testTime, result, expected)
	}
}

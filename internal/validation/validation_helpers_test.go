package validation

import (
	"strings"
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
)

func TestValidationHelper_ValidateEntry(t *testing.T) {
	vh := NewValidationHelper()

	tests := []struct {
		name           string
		entry          *models.Entry
		expectedValid  bool
		expectedErrors int
	}{
		{
			name: "Valid visited entry",
			entry: &models.Entry{
				Name:           "Test Restaurant",
				Location:       "N",
				Cuisine:        "Italian",
				Visited:        true,
				FoodRating:     4,
				AmbienceRating: 3,
				ValueRating:    5,
			},
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name: "Valid unvisited entry",
			entry: &models.Entry{
				Name:             "Future Restaurant",
				Location:         "SE",
				Cuisine:          "Mexican",
				Visited:          false,
				LookingForwardTo: "Tacos and margaritas",
			},
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name: "Missing required fields",
			entry: &models.Entry{
				Name:     "",
				Location: "",
				Cuisine:  "",
				Visited:  false,
			},
			expectedValid:  false,
			expectedErrors: 4, // name, location, cuisine, lookingforwardto
		},
		{
			name: "Invalid ratings for visited entry",
			entry: &models.Entry{
				Name:           "Test Restaurant",
				Location:       "N",
				Cuisine:        "Italian",
				Visited:        true,
				FoodRating:     6,  // Invalid
				AmbienceRating: -1, // Invalid
				ValueRating:    10, // Invalid
			},
			expectedValid:  false,
			expectedErrors: 3, // All three ratings invalid
		},
		{
			name: "Invalid location",
			entry: &models.Entry{
				Name:     "Test Restaurant",
				Location: "INVALID",
				Cuisine:  "Italian",
				Visited:  true,
			},
			expectedValid:  false,
			expectedErrors: 1, // Invalid location
		},
		{
			name: "Name too long",
			entry: &models.Entry{
				Name:     "This is a very long restaurant name that exceeds the maximum allowed length of 100 characters and should trigger a validation error",
				Location: "N",
				Cuisine:  "Italian",
				Visited:  true,
			},
			expectedValid:  false,
			expectedErrors: 1, // Name too long
		},
		{
			name: "Cuisine too long",
			entry: &models.Entry{
				Name:     "Test Restaurant",
				Location: "N",
				Cuisine:  "This is a very long cuisine type that exceeds the maximum allowed length",
				Visited:  true,
			},
			expectedValid:  false,
			expectedErrors: 1, // Cuisine too long
		},
		{
			name: "Unvisited entry missing looking forward to",
			entry: &models.Entry{
				Name:             "Future Restaurant",
				Location:         "SE",
				Cuisine:          "Mexican",
				Visited:          false,
				LookingForwardTo: "",
			},
			expectedValid:  false,
			expectedErrors: 1, // Missing lookingforwardto
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vh.ValidateEntry(tt.entry)

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateEntry() IsValid = %v, expected %v", result.IsValid, tt.expectedValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateEntry() errors count = %d, expected %d", len(result.Errors), tt.expectedErrors)
				for _, err := range result.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidationHelper_ValidateLoginCredentials(t *testing.T) {
	vh := NewValidationHelper()

	tests := []struct {
		name           string
		username       string
		password       string
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "Valid credentials",
			username:       "testuser",
			password:       "validpassword123",
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "Empty username",
			username:       "",
			password:       "validpassword123",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Empty password",
			username:       "testuser",
			password:       "",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Password too short",
			username:       "testuser",
			password:       "short",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Invalid username format",
			username:       "test-user!",
			password:       "validpassword123",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Both invalid",
			username:       "",
			password:       "short",
			expectedValid:  false,
			expectedErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vh.ValidateLoginCredentials(tt.username, tt.password)

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateLoginCredentials() IsValid = %v, expected %v", result.IsValid, tt.expectedValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateLoginCredentials() errors count = %d, expected %d", len(result.Errors), tt.expectedErrors)
				for _, err := range result.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidationHelper_ValidateRegistrationData(t *testing.T) {
	vh := NewValidationHelper()

	tests := []struct {
		name           string
		user           *models.User
		expectedValid  bool
		expectedErrors int
	}{
		{
			name: "Valid registration data",
			user: &models.User{
				Username:     "testuser",
				Password:     "validpassword123",
				ReferralCode: "VALID123",
			},
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name: "Missing username",
			user: &models.User{
				Username:     "",
				Password:     "validpassword123",
				ReferralCode: "VALID123",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "Missing password",
			user: &models.User{
				Username:     "testuser",
				Password:     "",
				ReferralCode: "VALID123",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "Missing referral code",
			user: &models.User{
				Username:     "testuser",
				Password:     "validpassword123",
				ReferralCode: "",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "Invalid username format",
			user: &models.User{
				Username:     "test-user!",
				Password:     "validpassword123",
				ReferralCode: "VALID123",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "Password too short",
			user: &models.User{
				Username:     "testuser",
				Password:     "short",
				ReferralCode: "VALID123",
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "All fields invalid",
			user: &models.User{
				Username:     "",
				Password:     "short",
				ReferralCode: "",
			},
			expectedValid:  false,
			expectedErrors: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vh.ValidateRegistrationData(tt.user)

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateRegistrationData() IsValid = %v, expected %v", result.IsValid, tt.expectedValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateRegistrationData() errors count = %d, expected %d", len(result.Errors), tt.expectedErrors)
				for _, err := range result.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidationHelper_isValidLocation(t *testing.T) {
	vh := NewValidationHelper()

	validLocations := []string{"N", "NE", "NW", "SW", "SE", "S"}
	invalidLocations := []string{"", "INVALID", "north", "ne", "123", "X"}

	for _, loc := range validLocations {
		if !vh.isValidLocation(loc) {
			t.Errorf("isValidLocation(%s) = false, expected true", loc)
		}
	}

	for _, loc := range invalidLocations {
		if vh.isValidLocation(loc) {
			t.Errorf("isValidLocation(%s) = true, expected false", loc)
		}
	}
}

func TestValidationHelper_isValidRating(t *testing.T) {
	vh := NewValidationHelper()

	validRatings := []int{1, 2, 3, 4, 5}
	invalidRatings := []int{0, -1, 6, 10, -5}

	for _, rating := range validRatings {
		if !vh.isValidRating(rating) {
			t.Errorf("isValidRating(%d) = false, expected true", rating)
		}
	}

	for _, rating := range invalidRatings {
		if vh.isValidRating(rating) {
			t.Errorf("isValidRating(%d) = true, expected false", rating)
		}
	}
}

func TestValidationHelper_isValidUsername(t *testing.T) {
	vh := NewValidationHelper()

	validUsernames := []string{"testuser", "user123", "abc", "test1test2"}
	invalidUsernames := []string{"", "Test", "test-user", "test_user", "test@user", "test user", "123Test"}

	for _, username := range validUsernames {
		if !vh.isValidUsername(username) {
			t.Errorf("isValidUsername(%s) = false, expected true", username)
		}
	}

	for _, username := range invalidUsernames {
		if vh.isValidUsername(username) {
			t.Errorf("isValidUsername(%s) = true, expected false", username)
		}
	}
}

func TestValidationHelper_isValidEmail(t *testing.T) {
	vh := NewValidationHelper()

	validEmails := []string{"test@example.com", "user.name@domain.co.uk", "test123@test.org"}
	invalidEmails := []string{"", "invalid", "test@", "@domain.com", "test@domain", "test.domain.com"}

	for _, email := range validEmails {
		if !vh.isValidEmail(email) {
			t.Errorf("isValidEmail(%s) = false, expected true", email)
		}
	}

	for _, email := range invalidEmails {
		if vh.isValidEmail(email) {
			t.Errorf("isValidEmail(%s) = true, expected false", email)
		}
	}
}

func TestValidationHelper_ValidateUserUpdate(t *testing.T) {
	vh := NewValidationHelper()

	tests := []struct {
		name           string
		username       string
		password       string
		email          string
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "Valid update data",
			username:       "newuser",
			password:       "newpassword123",
			email:          "new@example.com",
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "Empty fields (valid for update)",
			username:       "",
			password:       "",
			email:          "",
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "Invalid username format",
			username:       "New-User",
			password:       "validpassword123",
			email:          "valid@example.com",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Password too short",
			username:       "validuser",
			password:       "short",
			email:          "valid@example.com",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Invalid email",
			username:       "validuser",
			password:       "validpassword123",
			email:          "invalid-email",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Multiple invalid fields",
			username:       "Invalid-User",
			password:       "short",
			email:          "invalid-email",
			expectedValid:  false,
			expectedErrors: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vh.ValidateUserUpdate(tt.username, tt.password, tt.email)

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateUserUpdate() IsValid = %v, expected %v", result.IsValid, tt.expectedValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateUserUpdate() errors count = %d, expected %d", len(result.Errors), tt.expectedErrors)
				for _, err := range result.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidationHelper_getFieldErrorMessage(t *testing.T) {
	// This test would require creating mock validator.FieldError instances
	// For now, we'll test the basic functionality through the other validation methods
	// which internally use getFieldErrorMessage
}

func TestNewValidationHelper(t *testing.T) {
	vh := NewValidationHelper()

	if vh == nil {
		t.Error("NewValidationHelper() returned nil")
		return
	}

	if vh.validator == nil {
		t.Error("NewValidationHelper() validator is nil")
	}
}
func TestValidationHelper_ValidateSearchParameters(t *testing.T) {
	vh := NewValidationHelper()

	tests := []struct {
		name           string
		query          string
		pageStr        string
		limitStr       string
		sortBy         string
		sortOrder      string
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "Valid search parameters",
			query:          "test restaurant",
			pageStr:        "1",
			limitStr:       "20",
			sortBy:         "name",
			sortOrder:      "asc",
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "Empty query",
			query:          "",
			pageStr:        "1",
			limitStr:       "20",
			sortBy:         "name",
			sortOrder:      "asc",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Query too long",
			query:          strings.Repeat("a", 201),
			pageStr:        "1",
			limitStr:       "20",
			sortBy:         "name",
			sortOrder:      "asc",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Invalid page",
			query:          "test",
			pageStr:        "0",
			limitStr:       "20",
			sortBy:         "name",
			sortOrder:      "asc",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Invalid limit",
			query:          "test",
			pageStr:        "1",
			limitStr:       "101",
			sortBy:         "name",
			sortOrder:      "asc",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Invalid sort field",
			query:          "test",
			pageStr:        "1",
			limitStr:       "20",
			sortBy:         "invalid_field",
			sortOrder:      "asc",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Invalid sort order",
			query:          "test",
			pageStr:        "1",
			limitStr:       "20",
			sortBy:         "name",
			sortOrder:      "invalid",
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "Multiple invalid parameters",
			query:          "",
			pageStr:        "0",
			limitStr:       "101",
			sortBy:         "invalid",
			sortOrder:      "invalid",
			expectedValid:  false,
			expectedErrors: 5,
		},
		{
			name:           "Empty optional parameters (valid)",
			query:          "test",
			pageStr:        "",
			limitStr:       "",
			sortBy:         "",
			sortOrder:      "",
			expectedValid:  true,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vh.ValidateSearchParameters(tt.query, tt.pageStr, tt.limitStr, tt.sortBy, tt.sortOrder)

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateSearchParameters() IsValid = %v, expected %v", result.IsValid, tt.expectedValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateSearchParameters() errors count = %d, expected %d", len(result.Errors), tt.expectedErrors)
				for _, err := range result.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

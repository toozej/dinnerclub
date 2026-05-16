package services

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

func setupProfileTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.ProfileChange{},
		&models.EmailForwardingRule{},
		&models.PasswordReset{},
		&models.SocialNotification{},
		&models.BackupCode{},
		&models.UserFollow{},
	)
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

func TestProfileService_RequestEmailChange(t *testing.T) {
	db := setupProfileTestDB()
	emailService := createTestEmailService(db)
	profileService := NewProfileService(db, emailService)

	// Create test user
	user := &models.User{
		Username: "testuser",
		Email:    "old@example.com",
		Password: "hashedpassword",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("successful email change request", func(t *testing.T) {
		token, err := profileService.RequestEmailChange(user.ID, "new@example.com")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify profile change record was created
		var profileChange models.ProfileChange
		err = db.Where("user_id = ? AND change_type = ?", user.ID, "email").First(&profileChange).Error
		assert.NoError(t, err)
		assert.Equal(t, "old@example.com", profileChange.OldValue)
		assert.Equal(t, "new@example.com", profileChange.NewValue)
		assert.Equal(t, token, profileChange.Token)
		assert.False(t, profileChange.Confirmed)
	})

	t.Run("invalid email format", func(t *testing.T) {
		_, err := profileService.RequestEmailChange(user.ID, "invalid-email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("email already in use", func(t *testing.T) {
		// Create another user with the target email
		existingUser := &models.User{
			Username: "existing",
			Email:    "existing@example.com",
			Password: "hashedpassword",
		}
		require.NoError(t, db.Create(existingUser).Error)

		_, err := profileService.RequestEmailChange(user.ID, "existing@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already in use")
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := profileService.RequestEmailChange(99999, "new@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestProfileService_ConfirmEmailChange(t *testing.T) {
	db := setupProfileTestDB()
	emailService := createTestEmailService(db)
	profileService := NewProfileService(db, emailService)

	// Create test user
	user := &models.User{
		Username: "testuser",
		Email:    "old@example.com",
		Password: "hashedpassword",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("successful email change confirmation", func(t *testing.T) {
		// Create profile change record
		token, err := profileService.GenerateConfirmationToken()
		require.NoError(t, err)

		profileChange := &models.ProfileChange{
			UserID:     user.ID,
			ChangeType: "email",
			OldValue:   "old@example.com",
			NewValue:   "new@example.com",
			Token:      token,
			ExpiresAt:  time.Now().Add(24 * time.Hour),
			Confirmed:  false,
		}
		require.NoError(t, db.Create(profileChange).Error)

		// Confirm the change
		err = profileService.ConfirmEmailChange(token)
		assert.NoError(t, err)

		// Verify user email was updated
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "new@example.com", updatedUser.Email)

		// Verify profile change was marked as confirmed
		var confirmedChange models.ProfileChange
		err = db.First(&confirmedChange, profileChange.ID).Error
		assert.NoError(t, err)
		assert.True(t, confirmedChange.Confirmed)
	})

	t.Run("invalid token", func(t *testing.T) {
		err := profileService.ConfirmEmailChange("invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := profileService.GenerateConfirmationToken()
		require.NoError(t, err)

		profileChange := &models.ProfileChange{
			UserID:     user.ID,
			ChangeType: "email",
			OldValue:   "old@example.com",
			NewValue:   "expired@example.com",
			Token:      token,
			ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired
			Confirmed:  false,
		}
		require.NoError(t, db.Create(profileChange).Error)

		err = profileService.ConfirmEmailChange(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has expired")
	})

	t.Run("already used token", func(t *testing.T) {
		token, err := profileService.GenerateConfirmationToken()
		require.NoError(t, err)

		profileChange := &models.ProfileChange{
			UserID:     user.ID,
			ChangeType: "email",
			OldValue:   "old@example.com",
			NewValue:   "used@example.com",
			Token:      token,
			ExpiresAt:  time.Now().Add(24 * time.Hour),
			Confirmed:  true, // Already confirmed
		}
		require.NoError(t, db.Create(profileChange).Error)

		err = profileService.ConfirmEmailChange(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token has already been used")
	})
}

func TestProfileService_UpdateProfile(t *testing.T) {
	db := setupProfileTestDB()
	emailService := createTestEmailService(db)
	profileService := NewProfileService(db, emailService)

	// Create test user
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("successful non-sensitive update", func(t *testing.T) {
		updates := map[string]interface{}{
			"bio":              "New bio",
			"location_enabled": true,
			"privacy_level":    1, // 1 = followers only
		}

		err := profileService.UpdateProfile(user.ID, updates)
		assert.NoError(t, err)

		// Verify updates were applied
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "New bio", updatedUser.Bio)
		assert.True(t, updatedUser.LocationEnabled)
		assert.Equal(t, 1, updatedUser.PrivacyLevel)
	})

	t.Run("sensitive field requires confirmation", func(t *testing.T) {
		updates := map[string]interface{}{
			"email": "new@example.com",
		}

		err := profileService.UpdateProfile(user.ID, updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sensitive profile changes require confirmation")
	})

	t.Run("no valid fields to update", func(t *testing.T) {
		updates := map[string]interface{}{
			"invalid_field": "value",
		}

		err := profileService.UpdateProfile(user.ID, updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid fields to update")
	})
}

func TestProfileService_GenerateConfirmationToken(t *testing.T) {
	db := setupProfileTestDB()
	emailService := createTestEmailService(db)
	profileService := NewProfileService(db, emailService)

	t.Run("generates unique tokens", func(t *testing.T) {
		token1, err1 := profileService.GenerateConfirmationToken()
		token2, err2 := profileService.GenerateConfirmationToken()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEmpty(t, token1)
		assert.NotEmpty(t, token2)
		assert.NotEqual(t, token1, token2)
		assert.Equal(t, 64, len(token1)) // 32 bytes = 64 hex characters
	})
}

// Property-based tests for profile change confirmations
func TestProfileService_PropertyTests(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	// Property 7: Sensitive Change Confirmation Requirement
	properties.Property("Property 7: Sensitive fields require confirmation", prop.ForAll(
		func(userID uint, email string) bool {
			if userID == 0 || len(email) == 0 {
				return true // Skip invalid inputs
			}

			db := setupProfileTestDB()
			emailService := createTestEmailService(db)
			profileService := NewProfileService(db, emailService)

			// Create test user
			user := &models.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
			}
			if err := db.Create(user).Error; err != nil {
				return false
			}

			// Try to update sensitive field directly
			updates := map[string]interface{}{
				"email": email,
			}

			err := profileService.UpdateProfile(user.ID, updates)

			// Should always require confirmation for sensitive fields
			return err != nil &&
				(err.Error() == "sensitive profile changes require confirmation - use specific methods")
		},
		gen.UIntRange(1, 1000),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 8: Confirmation Token Security
	properties.Property("Property 8: Confirmation tokens are secure and unique", prop.ForAll(
		func() bool {
			db := setupProfileTestDB()
			emailService := createTestEmailService(db)
			profileService := NewProfileService(db, emailService)

			// Generate multiple tokens
			tokens := make(map[string]bool)
			for i := 0; i < 10; i++ {
				token, err := profileService.GenerateConfirmationToken()
				if err != nil {
					return false
				}

				// Check uniqueness
				if tokens[token] {
					return false
				}
				tokens[token] = true

				// Check length (should be 64 hex characters for 32 bytes)
				if len(token) != 64 {
					return false
				}
			}

			return true
		},
	))

	// Property 9: Valid Token Processing
	properties.Property("Property 9: Valid tokens are processed correctly", prop.ForAll(
		func(userID uint, oldEmail, newEmail string) bool {
			if userID == 0 || len(oldEmail) == 0 || len(newEmail) == 0 || oldEmail == newEmail {
				return true // Skip invalid inputs
			}

			db := setupProfileTestDB()
			emailService := createTestEmailService(db)
			profileService := NewProfileService(db, emailService)

			// Create test user
			user := &models.User{
				Username: "testuser",
				Email:    oldEmail,
				Password: "hashedpassword",
			}
			if err := db.Create(user).Error; err != nil {
				return false
			}

			// Generate token and create profile change
			token, err := profileService.GenerateConfirmationToken()
			if err != nil {
				return false
			}

			profileChange := &models.ProfileChange{
				UserID:     user.ID,
				ChangeType: "email",
				OldValue:   oldEmail,
				NewValue:   newEmail,
				Token:      token,
				ExpiresAt:  time.Now().Add(24 * time.Hour),
				Confirmed:  false,
			}
			if err := db.Create(profileChange).Error; err != nil {
				return false
			}

			// Validate token should work
			validatedChange, err := profileService.ValidateConfirmationToken(token)
			if err != nil {
				return false
			}

			return validatedChange.Token == token &&
				validatedChange.UserID == userID &&
				validatedChange.OldValue == oldEmail &&
				validatedChange.NewValue == newEmail &&
				!validatedChange.Confirmed
		},
		gen.UIntRange(1, 1000),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 10: Expired Token Rejection
	properties.Property("Property 10: Expired tokens are rejected", prop.ForAll(
		func(userID uint, expirationMinutes int) bool {
			if userID == 0 || expirationMinutes >= 0 {
				return true // Skip invalid inputs - we want negative expiration
			}

			db := setupProfileTestDB()
			emailService := createTestEmailService(db)
			profileService := NewProfileService(db, emailService)

			// Create test user
			user := &models.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
			}
			if err := db.Create(user).Error; err != nil {
				return false
			}

			// Generate token and create expired profile change
			token, err := profileService.GenerateConfirmationToken()
			if err != nil {
				return false
			}

			// Create profile change with past expiration time
			profileChange := &models.ProfileChange{
				UserID:     user.ID,
				ChangeType: "email",
				OldValue:   "old@example.com",
				NewValue:   "new@example.com",
				Token:      token,
				ExpiresAt:  time.Now().Add(time.Duration(expirationMinutes) * time.Minute), // Negative = expired
				Confirmed:  false,
			}
			if err := db.Create(profileChange).Error; err != nil {
				return false
			}

			// Validate token should fail for expired tokens
			_, err = profileService.ValidateConfirmationToken(token)

			// Should return error containing "expired"
			return err != nil && (err.Error() == "token has expired")
		},
		gen.UIntRange(1, 1000),
		gen.IntRange(-1440, -1), // Negative minutes (1 minute to 24 hours in the past)
	))

	properties.TestingRun(t)
}

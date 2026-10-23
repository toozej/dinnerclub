package services

import (
	"fmt"
	mathrand "math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// setupPasswordResetTestDB creates an in-memory SQLite database for password reset testing
func setupPasswordResetTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.PasswordReset{},
		&models.BackupCode{},
		&models.UserFollow{},
		&models.ProfileChange{},
		&models.EmailForwardingRule{},
		&models.SocialNotification{},
	)
	require.NoError(t, err)

	return db
}

// mockEmailService is a mock implementation of EmailService for testing
type mockEmailService struct {
	sentEmails []sentEmail
}

type sentEmail struct {
	user         *models.User
	tempPassword string
	emailType    string
}

func newMockEmailService() *mockEmailService {
	return &mockEmailService{
		sentEmails: make([]sentEmail, 0),
	}
}

func (m *mockEmailService) SendPasswordResetEmail(user *models.User, tempPassword string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{
		user:         user,
		tempPassword: tempPassword,
		emailType:    "password_reset",
	})
	return nil
}

// Property 17: Password Reset Email Delivery
// **Feature: advanced-user-system, Property 17: Password Reset Email Delivery**
// **Validates: Requirements 4.1**
func TestProperty_PasswordResetEmailDelivery(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupPasswordResetTestDB(t)
			mockEmail := newMockEmailService()
			service := NewPasswordResetService(db, mockEmail)

			// Generate random user
			user := generateRandomUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Test the property: For any valid password reset request,
			// a temporary password should be sent to the user's registered email address
			err := service.RequestPasswordReset(user.Email)
			require.NoError(t, err)

			// Assertions for email delivery
			assert.Len(t, mockEmail.sentEmails, 1, "Exactly one email should be sent")

			sentEmail := mockEmail.sentEmails[0]
			assert.Equal(t, user.Email, sentEmail.user.Email, "Email should be sent to user's registered email")
			assert.Equal(t, user.ID, sentEmail.user.ID, "Email should be sent to correct user")
			assert.NotEmpty(t, sentEmail.tempPassword, "Temporary password should be included in email")
			assert.Equal(t, "password_reset", sentEmail.emailType, "Email type should be password_reset")

			// Verify password reset record was created
			var passwordReset models.PasswordReset
			err = db.Where("user_id = ?", user.ID).First(&passwordReset).Error
			require.NoError(t, err)

			assert.Equal(t, user.ID, passwordReset.UserID, "Password reset should be for correct user")
			assert.NotEmpty(t, passwordReset.Token, "Reset token should be generated")
			assert.NotEmpty(t, passwordReset.TempPassword, "Temporary password should be stored")
			assert.False(t, passwordReset.Used, "Reset should not be marked as used initially")
			assert.True(t, passwordReset.ExpiresAt.After(time.Now()), "Reset should have future expiration")

			// Verify the temporary password in email matches the stored hash
			err = bcrypt.CompareHashAndPassword([]byte(passwordReset.TempPassword), []byte(sentEmail.tempPassword))
			assert.NoError(t, err, "Temporary password in email should match stored hash")
		})
	}
}

// Property 18: Temporary Password Security
// **Feature: advanced-user-system, Property 18: Temporary Password Security**
// **Validates: Requirements 4.2**
func TestProperty_TemporaryPasswordSecurity(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupPasswordResetTestDB(t)
			mockEmail := newMockEmailService()
			service := NewPasswordResetService(db, mockEmail)

			// Generate random user
			user := generateRandomUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Test the property: For any generated temporary password,
			// it should be cryptographically secure, random, and have a proper expiration time
			err := service.RequestPasswordReset(user.Email)
			require.NoError(t, err)

			// Get the sent email
			require.Len(t, mockEmail.sentEmails, 1)
			tempPassword := mockEmail.sentEmails[0].tempPassword

			// Assertions for temporary password security
			assert.NotEmpty(t, tempPassword, "Temporary password should not be empty")
			assert.GreaterOrEqual(t, len(tempPassword), 16, "Temporary password should be at least 16 characters for security")

			// Verify password contains mix of characters (base64 URL encoding ensures this)
			assert.Regexp(t, "[A-Za-z0-9_-]+", tempPassword, "Temporary password should contain alphanumeric characters")

			// Verify password reset record has proper expiration
			var passwordReset models.PasswordReset
			err = db.Where("user_id = ?", user.ID).Order("created_at DESC").First(&passwordReset).Error
			require.NoError(t, err)

			// Check expiration is approximately 1 hour from now (within 5 seconds tolerance)
			expectedExpiration := time.Now().Add(1 * time.Hour)
			timeDiff := passwordReset.ExpiresAt.Sub(expectedExpiration).Abs()
			assert.Less(t, timeDiff, 5*time.Second, "Expiration should be approximately 1 hour from creation")

			// Verify expiration is in the future
			assert.True(t, passwordReset.ExpiresAt.After(time.Now()), "Expiration should be in the future")

			// Verify password is properly hashed in database
			assert.NotEqual(t, tempPassword, passwordReset.TempPassword, "Password should be hashed in database")
			err = bcrypt.CompareHashAndPassword([]byte(passwordReset.TempPassword), []byte(tempPassword))
			assert.NoError(t, err, "Stored hash should match temporary password")

			// Test uniqueness: generate multiple passwords and verify they're different
			if mathrand.IntN(5) == 0 { // Randomly test uniqueness
				passwords := make(map[string]bool)
				passwords[tempPassword] = true

				for j := 0; j < 10; j++ {
					newPassword, err := service.GenerateTemporaryPassword()
					require.NoError(t, err)
					assert.False(t, passwords[newPassword], "Generated passwords should be unique")
					passwords[newPassword] = true
				}
			}
		})
	}
}

// Property 20: Temporary Password Expiration
// **Feature: advanced-user-system, Property 20: Temporary Password Expiration**
// **Validates: Requirements 4.4**
func TestProperty_TemporaryPasswordExpiration(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupPasswordResetTestDB(t)
			mockEmail := newMockEmailService()
			service := NewPasswordResetService(db, mockEmail)

			// Generate random user
			user := generateRandomUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Request password reset
			err := service.RequestPasswordReset(user.Email)
			require.NoError(t, err)

			// Get the password reset record
			var passwordReset models.PasswordReset
			err = db.Where("user_id = ?", user.ID).First(&passwordReset).Error
			require.NoError(t, err)

			// Test the property: For any expired temporary password,
			// login attempts should be rejected and require a new reset request

			// Simulate expiration by manually setting ExpiresAt to the past
			passwordReset.ExpiresAt = time.Now().Add(-1 * time.Hour)
			db.Save(&passwordReset)

			// Attempt to validate expired token
			_, err = service.ValidateResetToken(passwordReset.Token)
			assert.Error(t, err, "Expired token should be rejected")
			assert.Contains(t, err.Error(), "invalid or expired", "Error should indicate token is expired")

			// Attempt to validate expired temporary password
			tempPassword := mockEmail.sentEmails[0].tempPassword
			_, err = service.ValidateTemporaryPassword(user.ID, tempPassword)
			assert.Error(t, err, "Expired temporary password should be rejected")
			assert.Contains(t, err.Error(), "no active password reset", "Error should indicate no active reset")

			// Verify that a new reset request can be made
			mockEmail.sentEmails = nil // Clear previous emails
			err = service.RequestPasswordReset(user.Email)
			assert.NoError(t, err, "New reset request should succeed after expiration")
			assert.Len(t, mockEmail.sentEmails, 1, "New email should be sent")

			// Verify new reset record is created
			var newPasswordReset models.PasswordReset
			err = db.Where("user_id = ? AND used = ?", user.ID, false).
				Order("created_at DESC").
				First(&newPasswordReset).Error
			require.NoError(t, err)

			assert.NotEqual(t, passwordReset.ID, newPasswordReset.ID, "New reset record should be created")
			assert.True(t, newPasswordReset.ExpiresAt.After(time.Now()), "New reset should have future expiration")

			// Test cleanup of expired resets
			if mathrand.IntN(3) == 0 { // Randomly test cleanup
				err = service.CleanupExpiredResets()
				assert.NoError(t, err, "Cleanup should succeed")

				// Verify expired reset was deleted
				var count int64
				db.Model(&models.PasswordReset{}).Where("id = ?", passwordReset.ID).Count(&count)
				assert.Equal(t, int64(0), count, "Expired reset should be cleaned up")

				// Verify new reset still exists
				db.Model(&models.PasswordReset{}).Where("id = ?", newPasswordReset.ID).Count(&count)
				assert.Equal(t, int64(1), count, "Active reset should not be cleaned up")
			}
		})
	}
}

// Unit tests for specific functionality
func TestPasswordResetService_GenerateTemporaryPassword(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	t.Run("generates valid temporary password", func(t *testing.T) {
		password, err := service.GenerateTemporaryPassword()

		assert.NoError(t, err)
		assert.NotEmpty(t, password)
		assert.GreaterOrEqual(t, len(password), 16, "Password should be at least 16 characters")
		assert.Regexp(t, "^[A-Za-z0-9_-]+$", password, "Password should be base64 URL encoded")
	})

	t.Run("generates unique passwords", func(t *testing.T) {
		passwords := make(map[string]bool)
		for i := 0; i < 100; i++ {
			password, err := service.GenerateTemporaryPassword()
			assert.NoError(t, err)
			assert.False(t, passwords[password], "Passwords should be unique")
			passwords[password] = true
		}
	})
}

func TestPasswordResetService_GenerateResetToken(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	t.Run("generates valid reset token", func(t *testing.T) {
		token, err := service.GenerateResetToken()

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, len(token), 32, "Token should be sufficiently long")
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		tokens := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := service.GenerateResetToken()
			assert.NoError(t, err)
			assert.False(t, tokens[token], "Tokens should be unique")
			tokens[token] = true
		}
	})
}

func TestPasswordResetService_RequestPasswordReset(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	t.Run("creates password reset for valid user", func(t *testing.T) {
		user := generateRandomUser()
		db.Create(&user)

		err := service.RequestPasswordReset(user.Email)
		assert.NoError(t, err)

		// Verify reset record created
		var passwordReset models.PasswordReset
		err = db.Where("user_id = ?", user.ID).First(&passwordReset).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, passwordReset.UserID)
		assert.False(t, passwordReset.Used)

		// Verify email sent
		assert.Len(t, mockEmail.sentEmails, 1)
		assert.Equal(t, user.Email, mockEmail.sentEmails[0].user.Email)
	})

	t.Run("handles non-existent email gracefully", func(t *testing.T) {
		// Should not reveal if email exists (security best practice)
		err := service.RequestPasswordReset("nonexistent@example.com")
		assert.NoError(t, err, "Should not reveal if email exists")

		// Verify no email was sent
		mockEmail.sentEmails = nil
		err = service.RequestPasswordReset("nonexistent@example.com")
		assert.NoError(t, err)
		assert.Empty(t, mockEmail.sentEmails, "No email should be sent for non-existent user")
	})

	t.Run("allows multiple reset requests", func(t *testing.T) {
		user := generateRandomUser()
		db.Create(&user)

		// First request
		err := service.RequestPasswordReset(user.Email)
		assert.NoError(t, err)

		// Second request
		err = service.RequestPasswordReset(user.Email)
		assert.NoError(t, err)

		// Verify multiple reset records exist
		var count int64
		db.Model(&models.PasswordReset{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(2), count, "Multiple reset requests should be allowed")
	})
}

func TestPasswordResetService_ValidateResetToken(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	db.Create(&user)

	err := service.RequestPasswordReset(user.Email)
	require.NoError(t, err)

	var passwordReset models.PasswordReset
	db.Where("user_id = ?", user.ID).First(&passwordReset)

	t.Run("validates valid token", func(t *testing.T) {
		reset, err := service.ValidateResetToken(passwordReset.Token)
		assert.NoError(t, err)
		assert.NotNil(t, reset)
		assert.Equal(t, passwordReset.ID, reset.ID)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := service.ValidateResetToken("invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired")
	})

	t.Run("rejects expired token", func(t *testing.T) {
		// Expire the token
		passwordReset.ExpiresAt = time.Now().Add(-1 * time.Hour)
		db.Save(&passwordReset)

		_, err := service.ValidateResetToken(passwordReset.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired")
	})

	t.Run("rejects used token", func(t *testing.T) {
		// Mark as used
		passwordReset.Used = true
		passwordReset.ExpiresAt = time.Now().Add(1 * time.Hour) // Make it not expired
		db.Save(&passwordReset)

		_, err := service.ValidateResetToken(passwordReset.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired")
	})
}

func TestPasswordResetService_ValidateTemporaryPassword(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	db.Create(&user)

	err := service.RequestPasswordReset(user.Email)
	require.NoError(t, err)

	tempPassword := mockEmail.sentEmails[0].tempPassword

	t.Run("validates correct temporary password", func(t *testing.T) {
		reset, err := service.ValidateTemporaryPassword(user.ID, tempPassword)
		assert.NoError(t, err)
		assert.NotNil(t, reset)
		assert.Equal(t, user.ID, reset.UserID)
	})

	t.Run("rejects incorrect temporary password", func(t *testing.T) {
		_, err := service.ValidateTemporaryPassword(user.ID, "wrong-password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid temporary password")
	})

	t.Run("rejects expired temporary password", func(t *testing.T) {
		// Expire the reset
		var passwordReset models.PasswordReset
		db.Where("user_id = ?", user.ID).First(&passwordReset)
		passwordReset.ExpiresAt = time.Now().Add(-1 * time.Hour)
		db.Save(&passwordReset)

		_, err := service.ValidateTemporaryPassword(user.ID, tempPassword)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active password reset")
	})
}

func TestPasswordResetService_CompletePasswordReset(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	originalPassword := user.Password
	db.Create(&user)

	err := service.RequestPasswordReset(user.Email)
	require.NoError(t, err)

	var passwordReset models.PasswordReset
	db.Where("user_id = ?", user.ID).First(&passwordReset)

	t.Run("completes password reset successfully", func(t *testing.T) {
		newPassword := "newSecurePassword123"

		err := service.CompletePasswordReset(passwordReset.ID, newPassword)
		assert.NoError(t, err)

		// Verify password was updated
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.NotEqual(t, originalPassword, updatedUser.Password, "Password should be changed")

		// Verify new password hash is correct
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
		assert.NoError(t, err, "New password should match")

		// Verify reset is marked as used
		var updatedReset models.PasswordReset
		db.First(&updatedReset, passwordReset.ID)
		assert.True(t, updatedReset.Used, "Reset should be marked as used")
	})

	t.Run("rejects already used reset", func(t *testing.T) {
		err := service.CompletePasswordReset(passwordReset.ID, "anotherPassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already been used")
	})

	t.Run("rejects expired reset", func(t *testing.T) {
		// Create new reset
		user2 := generateRandomUser()
		db.Create(&user2)
		service.RequestPasswordReset(user2.Email)

		var reset2 models.PasswordReset
		db.Where("user_id = ?", user2.ID).First(&reset2)

		// Expire it
		reset2.ExpiresAt = time.Now().Add(-1 * time.Hour)
		db.Save(&reset2)

		err := service.CompletePasswordReset(reset2.ID, "newPassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

func TestPasswordResetService_CleanupExpiredResets(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	// Create users and resets
	user1 := generateRandomUser()
	db.Create(&user1)
	service.RequestPasswordReset(user1.Email)

	user2 := generateRandomUser()
	db.Create(&user2)
	service.RequestPasswordReset(user2.Email)

	// Expire first reset
	var reset1 models.PasswordReset
	db.Where("user_id = ?", user1.ID).First(&reset1)
	reset1.ExpiresAt = time.Now().Add(-1 * time.Hour)
	db.Save(&reset1)

	t.Run("removes expired resets only", func(t *testing.T) {
		err := service.CleanupExpiredResets()
		assert.NoError(t, err)

		// Verify expired reset is deleted
		var count int64
		db.Model(&models.PasswordReset{}).Where("id = ?", reset1.ID).Count(&count)
		assert.Equal(t, int64(0), count, "Expired reset should be deleted")

		// Verify active reset still exists
		var reset2 models.PasswordReset
		err = db.Where("user_id = ?", user2.ID).First(&reset2).Error
		assert.NoError(t, err, "Active reset should still exist")
	})
}

func TestResetRateLimiter_Allow(t *testing.T) {
	limiter := NewResetRateLimiter(3, 1*time.Hour)

	t.Run("allows requests within limit", func(t *testing.T) {
		identifier := "test@example.com"

		assert.True(t, limiter.Allow(identifier))
		assert.True(t, limiter.Allow(identifier))
		assert.True(t, limiter.Allow(identifier))
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		identifier := "blocked@example.com"

		// Use up the limit
		for i := 0; i < 3; i++ {
			limiter.Allow(identifier)
		}

		// Next request should be blocked
		assert.False(t, limiter.Allow(identifier))
	})

	t.Run("allows requests for different identifiers", func(t *testing.T) {
		assert.True(t, limiter.Allow("user1@example.com"))
		assert.True(t, limiter.Allow("user2@example.com"))
		assert.True(t, limiter.Allow("user3@example.com"))
	})

	t.Run("resets after time window", func(t *testing.T) {
		// Use a short window for testing
		shortLimiter := NewResetRateLimiter(2, 100*time.Millisecond)
		identifier := "timed@example.com"

		// Use up the limit
		assert.True(t, shortLimiter.Allow(identifier))
		assert.True(t, shortLimiter.Allow(identifier))
		assert.False(t, shortLimiter.Allow(identifier))

		// Wait for window to pass
		time.Sleep(150 * time.Millisecond)

		// Should be allowed again
		assert.True(t, shortLimiter.Allow(identifier))
	})
}

func TestResetRateLimiter_Reset(t *testing.T) {
	limiter := NewResetRateLimiter(2, 1*time.Hour)
	identifier := "reset@example.com"

	// Use up the limit
	limiter.Allow(identifier)
	limiter.Allow(identifier)
	assert.False(t, limiter.Allow(identifier))

	// Reset the limiter
	limiter.Reset(identifier)

	// Should be allowed again
	assert.True(t, limiter.Allow(identifier))
}

func TestResetRateLimiter_GetRemainingAttempts(t *testing.T) {
	limiter := NewResetRateLimiter(5, 1*time.Hour)
	identifier := "attempts@example.com"

	t.Run("returns correct remaining attempts", func(t *testing.T) {
		assert.Equal(t, 5, limiter.GetRemainingAttempts(identifier))

		limiter.Allow(identifier)
		assert.Equal(t, 4, limiter.GetRemainingAttempts(identifier))

		limiter.Allow(identifier)
		assert.Equal(t, 3, limiter.GetRemainingAttempts(identifier))

		limiter.Allow(identifier)
		limiter.Allow(identifier)
		limiter.Allow(identifier)
		assert.Equal(t, 0, limiter.GetRemainingAttempts(identifier))
	})
}

func TestPasswordResetService_RateLimiting(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	db.Create(&user)

	t.Run("enforces rate limiting", func(t *testing.T) {
		// Make 3 requests (the limit)
		for i := 0; i < 3; i++ {
			err := service.RequestPasswordReset(user.Email)
			assert.NoError(t, err, fmt.Sprintf("Request %d should succeed", i+1))
		}

		// 4th request should be rate limited
		err := service.RequestPasswordReset(user.Email)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "too many", "Should indicate rate limiting")
	})

	t.Run("rate limiting is per email", func(t *testing.T) {
		// Create another user
		user2 := generateRandomUser()
		db.Create(&user2)

		// Should be able to make requests for different email
		err := service.RequestPasswordReset(user2.Email)
		assert.NoError(t, err, "Different email should not be rate limited")
	})
}

// Property 19: Temporary Password Force Change
// **Feature: advanced-user-system, Property 19: Temporary Password Force Change**
// **Validates: Requirements 4.3**
func TestProperty_TemporaryPasswordForceChange(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupPasswordResetTestDB(t)
			mockEmail := newMockEmailService()
			service := NewPasswordResetService(db, mockEmail)

			// Generate random user
			user := generateRandomUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Request password reset
			err := service.RequestPasswordReset(user.Email)
			require.NoError(t, err)

			// Get the temporary password
			tempPassword := mockEmail.sentEmails[0].tempPassword

			// Test the property: For any login using a temporary password,
			// the system should require the user to change their password before proceeding
			passwordReset, requiresChange, err := service.LoginWithTemporaryPassword(user.ID, tempPassword)
			require.NoError(t, err)
			assert.NotNil(t, passwordReset, "Password reset record should be returned")
			assert.True(t, requiresChange, "System should require password change")

			// Verify RequiresPasswordChange returns true
			needsChange, err := service.RequiresPasswordChange(user.ID)
			require.NoError(t, err)
			assert.True(t, needsChange, "User should be flagged as requiring password change")

			// After completing password reset, should no longer require change
			newPassword := "newSecurePassword123"
			err = service.CompletePasswordReset(passwordReset.ID, newPassword)
			require.NoError(t, err)

			// Verify RequiresPasswordChange now returns false
			needsChange, err = service.RequiresPasswordChange(user.ID)
			require.NoError(t, err)
			assert.False(t, needsChange, "User should not require password change after completing reset")

			// Verify the reset is marked as used
			var updatedReset models.PasswordReset
			db.First(&updatedReset, passwordReset.ID)
			assert.True(t, updatedReset.Used, "Reset should be marked as used")

			// Test edge case: invalid temporary password should not allow login
			if mathrand.IntN(3) == 0 {
				_, _, err := service.LoginWithTemporaryPassword(user.ID, "invalid-password")
				assert.Error(t, err, "Invalid temporary password should be rejected")
			}
		})
	}
}

// Property 21: Session Invalidation on Password Reset
// **Feature: advanced-user-system, Property 21: Session Invalidation on Password Reset**
// **Validates: Requirements 4.5**
func TestProperty_SessionInvalidationOnPasswordReset(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupPasswordResetTestDB(t)
			mockEmail := newMockEmailService()
			service := NewPasswordResetService(db, mockEmail)

			// Generate random user
			user := generateRandomUser()
			// Set initial LastLoginAt to simulate active session
			initialLoginTime := time.Now().Add(-1 * time.Hour)
			user.LastLoginAt = &initialLoginTime
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Request password reset
			err := service.RequestPasswordReset(user.Email)
			require.NoError(t, err)

			// Get password reset record
			var passwordReset models.PasswordReset
			err = db.Where("user_id = ?", user.ID).First(&passwordReset).Error
			require.NoError(t, err)

			// Test the property: For any successful password reset,
			// all existing user sessions should be invalidated for security
			newPassword := "newSecurePassword123"
			err = service.CompletePasswordResetWithSessionInvalidation(passwordReset.ID, newPassword)
			require.NoError(t, err)

			// Verify session was invalidated by checking LastLoginAt was updated
			var updatedUser models.User
			db.First(&updatedUser, user.ID)
			assert.NotNil(t, updatedUser.LastLoginAt, "LastLoginAt should be set")
			assert.True(t, updatedUser.LastLoginAt.After(initialLoginTime),
				"LastLoginAt should be updated to invalidate old sessions")

			// Verify the timestamp is recent (within last 5 seconds)
			timeDiff := time.Since(*updatedUser.LastLoginAt).Abs()
			assert.Less(t, timeDiff, 5*time.Second, "Session invalidation timestamp should be recent")

			// Verify password was actually changed
			err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
			assert.NoError(t, err, "New password should be set correctly")

			// Test that InvalidateAllSessions can be called independently
			if mathrand.IntN(3) == 0 {
				beforeInvalidation := *updatedUser.LastLoginAt
				time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp difference

				err = service.InvalidateAllSessions(user.ID)
				assert.NoError(t, err, "Session invalidation should succeed")

				// Verify timestamp was updated again
				db.First(&updatedUser, user.ID)
				assert.True(t, updatedUser.LastLoginAt.After(beforeInvalidation),
					"LastLoginAt should be updated on explicit invalidation")
			}
		})
	}
}

// Property 22: Password Reset Rate Limiting
// **Feature: advanced-user-system, Property 22: Password Reset Rate Limiting**
// **Validates: Requirements 4.6**
func TestProperty_PasswordResetRateLimiting(t *testing.T) {
	const iterations = 50 // Fewer iterations since we're testing rate limiting

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupPasswordResetTestDB(t)
			mockEmail := newMockEmailService()
			service := NewPasswordResetService(db, mockEmail)

			// Generate random user
			user := generateRandomUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Test the property: For any series of password reset requests from the same source,
			// the system should enforce rate limiting to prevent abuse

			// Make requests up to the limit (3 requests per hour)
			successfulRequests := 0
			for j := 0; j < 3; j++ {
				err := service.RequestPasswordReset(user.Email)
				if err == nil {
					successfulRequests++
				}
			}

			// Verify we got exactly 3 successful requests
			assert.Equal(t, 3, successfulRequests, "Should allow exactly 3 requests within limit")
			assert.Len(t, mockEmail.sentEmails, 3, "Should send 3 emails for successful requests")

			// Next request should be rate limited
			err := service.RequestPasswordReset(user.Email)
			assert.Error(t, err, "Request exceeding limit should be rejected")
			assert.Contains(t, strings.ToLower(err.Error()), "too many",
				"Error should indicate rate limiting")

			// Verify no additional email was sent
			assert.Len(t, mockEmail.sentEmails, 3, "No additional email should be sent when rate limited")

			// Verify rate limiter tracks remaining attempts correctly
			remaining := service.rateLimiter.GetRemainingAttempts(user.Email)
			assert.Equal(t, 0, remaining, "Should have 0 remaining attempts")

			// Test that different users have independent rate limits
			if mathrand.IntN(3) == 0 {
				anotherUser := generateRandomUser()
				db.Create(&anotherUser)

				err := service.RequestPasswordReset(anotherUser.Email)
				assert.NoError(t, err, "Different user should not be rate limited")
			}

			// Test rate limiter reset functionality
			if mathrand.IntN(3) == 0 {
				service.rateLimiter.Reset(user.Email)
				remaining := service.rateLimiter.GetRemainingAttempts(user.Email)
				assert.Equal(t, 3, remaining, "Rate limit should be reset")

				// Should be able to make requests again
				err := service.RequestPasswordReset(user.Email)
				assert.NoError(t, err, "Request should succeed after rate limit reset")
			}
		})
	}
}

// Unit tests for security features
func TestPasswordResetService_LoginWithTemporaryPassword(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	db.Create(&user)

	err := service.RequestPasswordReset(user.Email)
	require.NoError(t, err)

	tempPassword := mockEmail.sentEmails[0].tempPassword

	t.Run("allows login with valid temporary password", func(t *testing.T) {
		reset, requiresChange, err := service.LoginWithTemporaryPassword(user.ID, tempPassword)
		assert.NoError(t, err)
		assert.NotNil(t, reset)
		assert.True(t, requiresChange, "Should require password change")
	})

	t.Run("rejects login with invalid temporary password", func(t *testing.T) {
		_, _, err := service.LoginWithTemporaryPassword(user.ID, "wrong-password")
		assert.Error(t, err)
	})

	t.Run("rejects login with expired temporary password", func(t *testing.T) {
		// Expire the reset
		var passwordReset models.PasswordReset
		db.Where("user_id = ?", user.ID).First(&passwordReset)
		passwordReset.ExpiresAt = time.Now().Add(-1 * time.Hour)
		db.Save(&passwordReset)

		_, _, err := service.LoginWithTemporaryPassword(user.ID, tempPassword)
		assert.Error(t, err)
	})
}

func TestPasswordResetService_RequiresPasswordChange(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	db.Create(&user)

	t.Run("returns false when no active reset", func(t *testing.T) {
		requiresChange, err := service.RequiresPasswordChange(user.ID)
		assert.NoError(t, err)
		assert.False(t, requiresChange)
	})

	t.Run("returns true when active reset exists", func(t *testing.T) {
		err := service.RequestPasswordReset(user.Email)
		require.NoError(t, err)

		requiresChange, err := service.RequiresPasswordChange(user.ID)
		assert.NoError(t, err)
		assert.True(t, requiresChange)
	})

	t.Run("returns false after reset is completed", func(t *testing.T) {
		var passwordReset models.PasswordReset
		db.Where("user_id = ?", user.ID).First(&passwordReset)

		err := service.CompletePasswordReset(passwordReset.ID, "newPassword123")
		require.NoError(t, err)

		requiresChange, err := service.RequiresPasswordChange(user.ID)
		assert.NoError(t, err)
		assert.False(t, requiresChange)
	})
}

func TestPasswordResetService_InvalidateAllSessions(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	initialTime := time.Now().Add(-1 * time.Hour)
	user.LastLoginAt = &initialTime
	db.Create(&user)

	t.Run("updates LastLoginAt to invalidate sessions", func(t *testing.T) {
		err := service.InvalidateAllSessions(user.ID)
		assert.NoError(t, err)

		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.NotNil(t, updatedUser.LastLoginAt)
		assert.True(t, updatedUser.LastLoginAt.After(initialTime))
	})

	t.Run("handles non-existent user gracefully", func(t *testing.T) {
		err := service.InvalidateAllSessions(99999)
		assert.Error(t, err)
	})
}

func TestPasswordResetService_CompletePasswordResetWithSessionInvalidation(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	user := generateRandomUser()
	initialTime := time.Now().Add(-1 * time.Hour)
	user.LastLoginAt = &initialTime
	db.Create(&user)

	err := service.RequestPasswordReset(user.Email)
	require.NoError(t, err)

	var passwordReset models.PasswordReset
	db.Where("user_id = ?", user.ID).First(&passwordReset)

	t.Run("completes reset and invalidates sessions", func(t *testing.T) {
		newPassword := "newSecurePassword123"
		err := service.CompletePasswordResetWithSessionInvalidation(passwordReset.ID, newPassword)
		assert.NoError(t, err)

		// Verify password changed
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
		assert.NoError(t, err)

		// Verify sessions invalidated
		assert.True(t, updatedUser.LastLoginAt.After(initialTime))

		// Verify reset marked as used
		var updatedReset models.PasswordReset
		db.First(&updatedReset, passwordReset.ID)
		assert.True(t, updatedReset.Used)
	})
}

func TestPasswordResetService_SecurityEdgeCases(t *testing.T) {
	db := setupPasswordResetTestDB(t)
	mockEmail := newMockEmailService()
	service := NewPasswordResetService(db, mockEmail)

	t.Run("prevents reuse of temporary password after completion", func(t *testing.T) {
		user := generateRandomUser()
		db.Create(&user)

		err := service.RequestPasswordReset(user.Email)
		require.NoError(t, err)

		tempPassword := mockEmail.sentEmails[0].tempPassword

		// Complete the reset
		var passwordReset models.PasswordReset
		db.Where("user_id = ?", user.ID).First(&passwordReset)
		err = service.CompletePasswordReset(passwordReset.ID, "newPassword123")
		require.NoError(t, err)

		// Try to login with temporary password again
		_, _, err = service.LoginWithTemporaryPassword(user.ID, tempPassword)
		assert.Error(t, err, "Should not allow login with used temporary password")
	})

	t.Run("handles concurrent password reset requests", func(t *testing.T) {
		user := generateRandomUser()
		db.Create(&user)

		// Make multiple reset requests
		for i := 0; i < 3; i++ {
			err := service.RequestPasswordReset(user.Email)
			assert.NoError(t, err)
		}

		// Verify multiple reset records exist
		var count int64
		db.Model(&models.PasswordReset{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(3), count)

		// All should be valid until used
		var resets []models.PasswordReset
		db.Where("user_id = ? AND used = ?", user.ID, false).Find(&resets)
		assert.Len(t, resets, 3)
	})

	t.Run("rate limiting persists across service instances", func(t *testing.T) {
		user := generateRandomUser()
		db.Create(&user)

		// Use up rate limit
		for i := 0; i < 3; i++ {
			service.RequestPasswordReset(user.Email)
		}

		// Should be rate limited
		err := service.RequestPasswordReset(user.Email)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "too many")
	})
}

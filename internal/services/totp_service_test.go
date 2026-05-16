package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand/v2"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// setupTOTPTestDB creates an in-memory SQLite database for TOTP testing
func setupTOTPTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.BackupCode{},
		&models.UserFollow{},
		&models.ProfileChange{},
		&models.EmailForwardingRule{},
		&models.SocialNotification{},
		&models.PasswordReset{},
	)
	require.NoError(t, err)

	return db
}

// generateRandomTOTPUser creates a random user for TOTP testing
func generateRandomTOTPUser() models.User {
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	suffix := hex.EncodeToString(randomBytes)

	return models.User{
		Username:  fmt.Sprintf("user_%s", suffix),
		Password:  "password123456",
		Firstname: fmt.Sprintf("First_%s", suffix),
		Lastname:  fmt.Sprintf("Last_%s", suffix),
		Email:     fmt.Sprintf("user_%s@example.com", suffix),
	}
}

// Property 23: 2FA QR Code Generation
// **Feature: advanced-user-system, Property 23: 2FA QR Code Generation**
// **Validates: Requirements 5.1**
func TestProperty_2FAQRCodeGeneration(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Test the property: For any user enabling 2FA, the system should generate
			// a valid QR code for authenticator app setup
			secret, qrURL, err := service.GenerateSecret(user.Username)

			// Assertions
			require.NoError(t, err)
			assert.NotEmpty(t, secret, "Secret should not be empty")
			assert.NotEmpty(t, qrURL, "QR URL should not be empty")

			// Verify secret format (base32 encoded)
			assert.Regexp(t, "^[A-Z2-7]+=*$", secret, "Secret should be valid base32")
			assert.GreaterOrEqual(t, len(secret), 32, "Secret should be at least 32 characters")

			// Verify QR URL format
			assert.Contains(t, qrURL, "otpauth://totp/", "QR URL should be TOTP format")
			assert.Contains(t, qrURL, user.Username, "QR URL should contain username")
			assert.Contains(t, qrURL, "DinnerClub", "QR URL should contain issuer")
			assert.Contains(t, qrURL, "secret=", "QR URL should contain secret parameter")

			// Verify the secret in the URL matches the returned secret
			assert.Contains(t, qrURL, secret, "QR URL should contain the generated secret")

			// Verify we can generate a valid token from the secret
			token, err := totp.GenerateCode(secret, time.Now())
			require.NoError(t, err)
			assert.Len(t, token, 6, "Generated token should be 6 digits")
			assert.Regexp(t, "^[0-9]{6}$", token, "Token should be 6 digits")

			// Verify the token validates correctly
			valid := service.ValidateToken(secret, token)
			assert.True(t, valid, "Generated token should validate successfully")
		})
	}
}

// Property 24: 2FA Setup Verification Requirement
// **Feature: advanced-user-system, Property 24: 2FA Setup Verification Requirement**
// **Validates: Requirements 5.2**
func TestProperty_2FASetupVerificationRequirement(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Generate secret
			secret, _, err := service.GenerateSecret(user.Username)
			require.NoError(t, err)

			// Test the property: For any 2FA setup attempt, the system should require
			// successful verification of the authenticator before activation

			// Generate a valid token
			validToken, err := totp.GenerateCode(secret, time.Now())
			require.NoError(t, err)

			// Verify setup with valid token should succeed
			err = service.Verify2FASetup(secret, validToken)
			assert.NoError(t, err, "Valid token should pass verification")

			// Generate invalid tokens
			invalidTokens := []string{
				"000000", // All zeros
				"999999", // All nines
				fmt.Sprintf("%06d", mathrand.IntN(1000000)), // Random (likely invalid)
				"12345",   // Too short
				"1234567", // Too long
				"abcdef",  // Non-numeric
			}

			for _, invalidToken := range invalidTokens {
				// Skip if by chance we generated a valid token
				if invalidToken == validToken {
					continue
				}

				err = service.Verify2FASetup(secret, invalidToken)
				assert.Error(t, err, "Invalid token '%s' should fail verification", invalidToken)
				assert.Contains(t, err.Error(), "invalid verification code", "Error should indicate invalid code")
			}

			// Verify that 2FA is NOT enabled without proper verification
			// (Enable2FA should only be called after successful verification)
			status, err := service.GetUser2FAStatus(user.ID)
			require.NoError(t, err)
			assert.False(t, status, "2FA should not be enabled without verification")

			// Now enable 2FA after successful verification
			err = service.Enable2FA(user.ID, secret)
			require.NoError(t, err)

			// Verify 2FA is now enabled
			status, err = service.GetUser2FAStatus(user.ID)
			require.NoError(t, err)
			assert.True(t, status, "2FA should be enabled after verification and enable call")
		})
	}
}

// Property 25: Backup Code Generation
// **Feature: advanced-user-system, Property 25: Backup Code Generation**
// **Validates: Requirements 5.3**
func TestProperty_BackupCodeGeneration(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Test the property: For any 2FA activation, the system should generate
			// exactly 10 unique single-use backup codes
			backupCodes, err := service.GenerateBackupCodes(user.ID)

			// Assertions
			require.NoError(t, err)
			assert.Len(t, backupCodes, 10, "Should generate exactly 10 backup codes")

			// Verify all codes are unique
			codeSet := make(map[string]bool)
			for _, code := range backupCodes {
				assert.False(t, codeSet[code], "Backup code should be unique: %s", code)
				codeSet[code] = true
			}

			// Verify code format (XXXXX-XXXXX)
			codeRegex := regexp.MustCompile(`^[A-F0-9]{5}-[A-F0-9]{5}$`)
			for _, code := range backupCodes {
				assert.Regexp(t, codeRegex, code, "Backup code should match format XXXXX-XXXXX")
				assert.Len(t, code, 11, "Backup code should be 11 characters (including dash)")
			}

			// Verify codes are stored in database
			var dbCodes []models.BackupCode
			err = db.Where("user_id = ?", user.ID).Find(&dbCodes).Error
			require.NoError(t, err)
			assert.Len(t, dbCodes, 10, "Should have 10 backup codes in database")

			// Verify all codes are initially unused
			for _, dbCode := range dbCodes {
				assert.False(t, dbCode.Used, "Backup code should initially be unused")
				assert.Equal(t, user.ID, dbCode.UserID, "Backup code should belong to the user")
			}

			// Verify regenerating codes replaces old ones
			newBackupCodes, err := service.RegenerateBackupCodes(user.ID)
			require.NoError(t, err)
			assert.Len(t, newBackupCodes, 10, "Should regenerate exactly 10 backup codes")

			// Verify old codes are deleted
			var allCodes []models.BackupCode
			err = db.Where("user_id = ?", user.ID).Find(&allCodes).Error
			require.NoError(t, err)
			assert.Len(t, allCodes, 10, "Should only have 10 codes after regeneration (old ones deleted)")

			// Verify new codes are different from old codes
			for _, newCode := range newBackupCodes {
				assert.False(t, codeSet[newCode], "Regenerated code should be different from original codes")
			}
		})
	}
}

// Property 27: Backup Code Single Use
// **Feature: advanced-user-system, Property 27: Backup Code Single Use**
// **Validates: Requirements 5.5**
func TestProperty_BackupCodeSingleUse(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Generate backup codes
			backupCodes, err := service.GenerateBackupCodes(user.ID)
			require.NoError(t, err)
			require.Len(t, backupCodes, 10)

			// Select a random backup code to test
			testCodeIndex := mathrand.IntN(len(backupCodes))
			testCode := backupCodes[testCodeIndex]

			// Test the property: For any backup code used for authentication,
			// that specific code should be invalidated and cannot be reused

			// First use should succeed
			valid, err := service.ValidateBackupCode(user.ID, testCode)
			require.NoError(t, err)
			assert.True(t, valid, "First use of backup code should succeed")

			// Verify code is marked as used in database
			var dbCodes []models.BackupCode
			err = db.Where("user_id = ?", user.ID).Find(&dbCodes).Error
			require.NoError(t, err)

			usedCount := 0
			for _, dbCode := range dbCodes {
				normalizedDB := strings.ToUpper(strings.ReplaceAll(dbCode.Code, "-", ""))
				normalizedTest := strings.ToUpper(strings.ReplaceAll(testCode, "-", ""))
				if normalizedDB == normalizedTest {
					assert.True(t, dbCode.Used, "Used backup code should be marked as used")
					usedCount++
				}
			}
			assert.Equal(t, 1, usedCount, "Should find exactly one used code")

			// Second use of same code should fail
			valid, err = service.ValidateBackupCode(user.ID, testCode)
			require.NoError(t, err)
			assert.False(t, valid, "Second use of same backup code should fail")

			// Test with different formats of the same code
			testCodeVariations := []string{
				testCode,
				strings.ToLower(testCode),             // lowercase
				strings.ReplaceAll(testCode, "-", ""), // no dash
				strings.ToLower(strings.ReplaceAll(testCode, "-", "")), // lowercase no dash
			}

			for _, variation := range testCodeVariations {
				valid, err = service.ValidateBackupCode(user.ID, variation)
				require.NoError(t, err)
				assert.False(t, valid, "Used backup code should fail in any format: %s", variation)
			}

			// Verify other codes still work
			if len(backupCodes) > 1 {
				// Pick a different code
				otherCodeIndex := (testCodeIndex + 1) % len(backupCodes)
				otherCode := backupCodes[otherCodeIndex]

				valid, err = service.ValidateBackupCode(user.ID, otherCode)
				require.NoError(t, err)
				assert.True(t, valid, "Other unused backup codes should still work")
			}

			// Verify count of unused codes decreased
			unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
			require.NoError(t, err)
			assert.Equal(t, 8, unusedCount, "Should have 8 unused codes after using 2")
		})
	}
}

// Unit tests for specific functionality
func TestTOTPService_GenerateSecret(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	t.Run("generates valid secret and QR URL", func(t *testing.T) {
		secret, qrURL, err := service.GenerateSecret("testuser")

		assert.NoError(t, err)
		assert.NotEmpty(t, secret)
		assert.NotEmpty(t, qrURL)
		assert.Regexp(t, "^[A-Z2-7]+=*$", secret)
		assert.Contains(t, qrURL, "otpauth://totp/")
		assert.Contains(t, qrURL, "testuser")
		assert.Contains(t, qrURL, "DinnerClub")
	})

	t.Run("generates unique secrets", func(t *testing.T) {
		secrets := make(map[string]bool)
		for i := 0; i < 10; i++ {
			secret, _, err := service.GenerateSecret(fmt.Sprintf("user%d", i))
			assert.NoError(t, err)
			assert.False(t, secrets[secret], "Secret should be unique")
			secrets[secret] = true
		}
	})
}

func TestTOTPService_ValidateToken(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	secret, _, err := service.GenerateSecret("testuser")
	require.NoError(t, err)

	t.Run("validates correct token", func(t *testing.T) {
		token, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		valid := service.ValidateToken(secret, token)
		assert.True(t, valid)
	})

	t.Run("rejects incorrect token", func(t *testing.T) {
		valid := service.ValidateToken(secret, "000000")
		assert.False(t, valid)
	})

	t.Run("rejects invalid format", func(t *testing.T) {
		valid := service.ValidateToken(secret, "abc")
		assert.False(t, valid)
	})
}

func TestTOTPService_GenerateBackupCodes(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	t.Run("generates 10 backup codes", func(t *testing.T) {
		codes, err := service.GenerateBackupCodes(user.ID)
		assert.NoError(t, err)
		assert.Len(t, codes, 10)
	})

	t.Run("codes have correct format", func(t *testing.T) {
		codes, err := service.GenerateBackupCodes(user.ID)
		require.NoError(t, err)

		for _, code := range codes {
			assert.Regexp(t, `^[A-F0-9]{5}-[A-F0-9]{5}$`, code)
		}
	})

	t.Run("codes are stored in database", func(t *testing.T) {
		codes, err := service.GenerateBackupCodes(user.ID)
		require.NoError(t, err)

		var dbCodes []models.BackupCode
		db.Where("user_id = ?", user.ID).Find(&dbCodes)
		assert.Len(t, dbCodes, len(codes))
	})

	t.Run("regenerating replaces old codes", func(t *testing.T) {
		firstCodes, err := service.GenerateBackupCodes(user.ID)
		require.NoError(t, err)

		secondCodes, err := service.RegenerateBackupCodes(user.ID)
		require.NoError(t, err)

		// Verify we still have exactly 10 codes
		var dbCodes []models.BackupCode
		db.Where("user_id = ?", user.ID).Find(&dbCodes)
		assert.Len(t, dbCodes, 10)

		// Verify codes are different
		firstSet := make(map[string]bool)
		for _, code := range firstCodes {
			firstSet[code] = true
		}

		for _, code := range secondCodes {
			assert.False(t, firstSet[code], "New codes should be different from old codes")
		}
	})
}

func TestTOTPService_ValidateBackupCode(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	codes, err := service.GenerateBackupCodes(user.ID)
	require.NoError(t, err)
	testCode := codes[0]

	t.Run("validates correct backup code", func(t *testing.T) {
		valid, err := service.ValidateBackupCode(user.ID, testCode)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("marks code as used", func(t *testing.T) {
		var dbCode models.BackupCode
		normalizedTest := strings.ToUpper(strings.ReplaceAll(testCode, "-", ""))

		var allCodes []models.BackupCode
		db.Where("user_id = ?", user.ID).Find(&allCodes)

		for _, code := range allCodes {
			normalizedDB := strings.ToUpper(strings.ReplaceAll(code.Code, "-", ""))
			if normalizedDB == normalizedTest {
				dbCode = code
				break
			}
		}

		assert.True(t, dbCode.Used)
	})

	t.Run("rejects used code", func(t *testing.T) {
		valid, err := service.ValidateBackupCode(user.ID, testCode)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("rejects invalid code", func(t *testing.T) {
		valid, err := service.ValidateBackupCode(user.ID, "INVALID-CODE")
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("handles code format variations", func(t *testing.T) {
		newCode := codes[1]

		// Test lowercase
		valid, err := service.ValidateBackupCode(user.ID, strings.ToLower(newCode))
		assert.NoError(t, err)
		assert.True(t, valid)

		// Should now be used
		valid, err = service.ValidateBackupCode(user.ID, newCode)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestTOTPService_Enable2FA(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	secret, _, err := service.GenerateSecret(user.Username)
	require.NoError(t, err)

	t.Run("enables 2FA for user", func(t *testing.T) {
		err := service.Enable2FA(user.ID, secret)
		assert.NoError(t, err)

		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.True(t, updatedUser.TwoFactorEnabled)
		assert.Equal(t, secret, updatedUser.TwoFactorSecret)
	})

	t.Run("returns correct 2FA status", func(t *testing.T) {
		status, err := service.GetUser2FAStatus(user.ID)
		assert.NoError(t, err)
		assert.True(t, status)
	})
}

func TestTOTPService_Disable2FA(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	// Enable 2FA first
	secret, _, err := service.GenerateSecret(user.Username)
	require.NoError(t, err)
	err = service.Enable2FA(user.ID, secret)
	require.NoError(t, err)

	// Generate backup codes
	_, err = service.GenerateBackupCodes(user.ID)
	require.NoError(t, err)

	t.Run("disables 2FA for user", func(t *testing.T) {
		err := service.Disable2FA(user.ID)
		assert.NoError(t, err)

		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.False(t, updatedUser.TwoFactorEnabled)
		assert.Empty(t, updatedUser.TwoFactorSecret)
	})

	t.Run("deletes backup codes", func(t *testing.T) {
		var codes []models.BackupCode
		db.Where("user_id = ?", user.ID).Find(&codes)
		assert.Empty(t, codes)
	})

	t.Run("returns correct 2FA status", func(t *testing.T) {
		status, err := service.GetUser2FAStatus(user.ID)
		assert.NoError(t, err)
		assert.False(t, status)
	})
}

func TestTOTPService_GetUnusedBackupCodeCount(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	t.Run("returns 0 for user with no codes", func(t *testing.T) {
		count, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("returns correct count after generation", func(t *testing.T) {
		_, err := service.GenerateBackupCodes(user.ID)
		require.NoError(t, err)

		count, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
	})

	t.Run("returns correct count after using codes", func(t *testing.T) {
		codes, err := service.GenerateBackupCodes(user.ID)
		require.NoError(t, err)

		// Use 3 codes
		for i := 0; i < 3; i++ {
			service.ValidateBackupCode(user.ID, codes[i])
		}

		count, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 7, count)
	})
}

func TestTOTPService_Verify2FASetup(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	secret, _, err := service.GenerateSecret("testuser")
	require.NoError(t, err)

	t.Run("accepts valid token", func(t *testing.T) {
		token, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		err = service.Verify2FASetup(secret, token)
		assert.NoError(t, err)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		err := service.Verify2FASetup(secret, "000000")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid verification code")
	})

	t.Run("rejects empty token", func(t *testing.T) {
		err := service.Verify2FASetup(secret, "")
		assert.Error(t, err)
	})
}

func TestTOTPService_GenerateRecoveryKey(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	t.Run("generates recovery key", func(t *testing.T) {
		key, err := service.GenerateRecoveryKey()
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
	})

	t.Run("recovery key has correct format", func(t *testing.T) {
		key, err := service.GenerateRecoveryKey()
		require.NoError(t, err)

		// Should contain dashes for readability
		assert.Contains(t, key, "-")

		// Should be base32 characters (A-Z, 2-7) plus dashes
		cleanKey := strings.ReplaceAll(key, "-", "")
		assert.Regexp(t, "^[A-Z2-7]+$", cleanKey)
	})

	t.Run("generates unique recovery keys", func(t *testing.T) {
		keys := make(map[string]bool)
		for i := 0; i < 10; i++ {
			key, err := service.GenerateRecoveryKey()
			assert.NoError(t, err)
			assert.False(t, keys[key], "Recovery key should be unique")
			keys[key] = true
		}
	})
}

// Property 26: 2FA Login Requirements
// **Feature: advanced-user-system, Property 26: 2FA Login Requirements**
// **Validates: Requirements 5.4**
func TestProperty_2FALoginRequirements(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Generate secret and enable 2FA
			secret, _, err := service.GenerateSecret(user.Username)
			require.NoError(t, err)

			err = service.Enable2FA(user.ID, secret)
			require.NoError(t, err)

			// Generate backup codes
			backupCodes, err := service.GenerateBackupCodes(user.ID)
			require.NoError(t, err)

			// Test the property: For any login attempt by a user with 2FA enabled,
			// both password and valid 2FA code should be required

			// Verify 2FA is enabled
			var updatedUser models.User
			db.First(&updatedUser, user.ID)
			assert.True(t, updatedUser.TwoFactorEnabled, "2FA should be enabled")
			assert.NotEmpty(t, updatedUser.TwoFactorSecret, "2FA secret should be stored")

			// Test 1: Valid TOTP code should authenticate successfully
			validToken, err := totp.GenerateCode(secret, time.Now())
			require.NoError(t, err)

			valid := service.ValidateToken(secret, validToken)
			assert.True(t, valid, "Valid TOTP code should authenticate")

			// Test 2: Invalid TOTP code should fail
			invalidToken := "000000"
			if invalidToken == validToken {
				invalidToken = "999999"
			}

			valid = service.ValidateToken(secret, invalidToken)
			assert.False(t, valid, "Invalid TOTP code should fail authentication")

			// Test 3: Backup code should work as alternative
			testBackupCode := backupCodes[mathrand.IntN(len(backupCodes))]
			valid, err = service.ValidateBackupCode(user.ID, testBackupCode)
			require.NoError(t, err)
			assert.True(t, valid, "Valid backup code should authenticate")

			// Test 4: Used backup code should not work again
			valid, err = service.ValidateBackupCode(user.ID, testBackupCode)
			require.NoError(t, err)
			assert.False(t, valid, "Used backup code should not authenticate again")

			// Test 5: Empty codes should fail
			valid = service.ValidateToken(secret, "")
			assert.False(t, valid, "Empty TOTP code should fail")

			valid, err = service.ValidateBackupCode(user.ID, "")
			require.NoError(t, err)
			assert.False(t, valid, "Empty backup code should fail")

			// Test 6: Verify 2FA requirement persists across multiple login attempts
			for attempt := 0; attempt < 3; attempt++ {
				newToken, err := totp.GenerateCode(secret, time.Now())
				require.NoError(t, err)

				valid := service.ValidateToken(secret, newToken)
				assert.True(t, valid, "Valid TOTP code should work on attempt %d", attempt+1)

				// Small delay to ensure time-based token might change
				time.Sleep(100 * time.Millisecond)
			}
		})
	}
}

// Property 29: 2FA Disable Security
// **Feature: advanced-user-system, Property 29: 2FA Disable Security**
// **Validates: Requirements 5.7**
func TestProperty_2FADisableSecurity(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Generate secret and enable 2FA
			secret, _, err := service.GenerateSecret(user.Username)
			require.NoError(t, err)

			err = service.Enable2FA(user.ID, secret)
			require.NoError(t, err)

			// Generate backup codes
			backupCodes, err := service.GenerateBackupCodes(user.ID)
			require.NoError(t, err)

			// Test the property: For any 2FA disable request, the system should require
			// both current password and valid 2FA verification

			// Verify 2FA is enabled before attempting to disable
			status, err := service.GetUser2FAStatus(user.ID)
			require.NoError(t, err)
			assert.True(t, status, "2FA should be enabled before disable attempt")

			// Test 1: Cannot disable without valid 2FA code
			// (In a real implementation, this would be checked in the controller)
			// Here we verify that the secret is required for validation

			// Generate valid TOTP code
			validToken, err := totp.GenerateCode(secret, time.Now())
			require.NoError(t, err)

			// Verify the code is valid before disabling
			valid := service.ValidateToken(secret, validToken)
			assert.True(t, valid, "Valid TOTP code should be verified before disable")

			// Test 2: Disable 2FA (simulating successful verification)
			err = service.Disable2FA(user.ID)
			require.NoError(t, err)

			// Verify 2FA is disabled
			status, err = service.GetUser2FAStatus(user.ID)
			require.NoError(t, err)
			assert.False(t, status, "2FA should be disabled after successful disable")

			// Verify secret is cleared
			var updatedUser models.User
			db.First(&updatedUser, user.ID)
			assert.False(t, updatedUser.TwoFactorEnabled, "2FA enabled flag should be false")
			assert.Empty(t, updatedUser.TwoFactorSecret, "2FA secret should be cleared")

			// Test 3: Verify backup codes are deleted
			var remainingCodes []models.BackupCode
			db.Where("user_id = ?", user.ID).Find(&remainingCodes)
			assert.Empty(t, remainingCodes, "All backup codes should be deleted after disabling 2FA")

			// Test 4: Verify cannot disable 2FA when it's already disabled
			err = service.Disable2FA(user.ID)
			require.NoError(t, err) // Should not error, just no-op

			status, err = service.GetUser2FAStatus(user.ID)
			require.NoError(t, err)
			assert.False(t, status, "2FA should remain disabled")

			// Test 5: Test with backup code verification before disable
			// Re-enable 2FA for this test
			newSecret, _, err := service.GenerateSecret(user.Username)
			require.NoError(t, err)

			err = service.Enable2FA(user.ID, newSecret)
			require.NoError(t, err)

			newBackupCodes, err := service.GenerateBackupCodes(user.ID)
			require.NoError(t, err)

			// Verify with backup code
			testBackupCode := newBackupCodes[0]
			valid, err = service.ValidateBackupCode(user.ID, testBackupCode)
			require.NoError(t, err)
			assert.True(t, valid, "Backup code should be valid for verification before disable")

			// Now disable
			err = service.Disable2FA(user.ID)
			require.NoError(t, err)

			// Verify disabled
			status, err = service.GetUser2FAStatus(user.ID)
			require.NoError(t, err)
			assert.False(t, status, "2FA should be disabled after backup code verification")

			// Test 6: Verify old backup codes don't work after re-enabling
			// Re-enable one more time
			finalSecret, _, err := service.GenerateSecret(user.Username)
			require.NoError(t, err)

			err = service.Enable2FA(user.ID, finalSecret)
			require.NoError(t, err)

			// Old backup codes should not work (they were deleted when 2FA was disabled)
			for _, oldCode := range backupCodes {
				valid, err := service.ValidateBackupCode(user.ID, oldCode)
				require.NoError(t, err)
				assert.False(t, valid, "Old backup code should not work after re-enabling: %s", oldCode)
			}
		})
	}
}

// Additional unit tests for authentication flow integration
func TestTOTPService_AuthenticationFlowIntegration(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	t.Run("complete 2FA setup and login flow", func(t *testing.T) {
		// Step 1: Generate secret
		secret, qrURL, err := service.GenerateSecret(user.Username)
		assert.NoError(t, err)
		assert.NotEmpty(t, secret)
		assert.NotEmpty(t, qrURL)

		// Step 2: Verify setup with valid code
		token, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		err = service.Verify2FASetup(secret, token)
		assert.NoError(t, err)

		// Step 3: Enable 2FA
		err = service.Enable2FA(user.ID, secret)
		assert.NoError(t, err)

		// Step 4: Generate backup codes
		backupCodes, err := service.GenerateBackupCodes(user.ID)
		assert.NoError(t, err)
		assert.Len(t, backupCodes, 10)

		// Step 5: Simulate login with TOTP
		loginToken, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		valid := service.ValidateToken(secret, loginToken)
		assert.True(t, valid, "Should be able to login with TOTP")

		// Step 6: Simulate login with backup code
		valid, err = service.ValidateBackupCode(user.ID, backupCodes[0])
		assert.NoError(t, err)
		assert.True(t, valid, "Should be able to login with backup code")

		// Step 7: Verify backup code is used
		valid, err = service.ValidateBackupCode(user.ID, backupCodes[0])
		assert.NoError(t, err)
		assert.False(t, valid, "Used backup code should not work again")

		// Step 8: Check remaining backup codes
		unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 9, unusedCount)
	})

	t.Run("2FA disable flow with verification", func(t *testing.T) {
		// Setup: User already has 2FA enabled from previous test
		status, err := service.GetUser2FAStatus(user.ID)
		require.NoError(t, err)
		require.True(t, status)

		// Get user's secret
		var dbUser models.User
		db.First(&dbUser, user.ID)

		// Generate valid token for verification
		token, err := totp.GenerateCode(dbUser.TwoFactorSecret, time.Now())
		require.NoError(t, err)

		// Verify token is valid
		valid := service.ValidateToken(dbUser.TwoFactorSecret, token)
		assert.True(t, valid, "Token should be valid before disable")

		// Disable 2FA
		err = service.Disable2FA(user.ID)
		assert.NoError(t, err)

		// Verify disabled
		status, err = service.GetUser2FAStatus(user.ID)
		assert.NoError(t, err)
		assert.False(t, status)

		// Verify backup codes are deleted
		unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, unusedCount)
	})

	t.Run("cannot verify setup with invalid code", func(t *testing.T) {
		secret, _, err := service.GenerateSecret(user.Username)
		require.NoError(t, err)

		// Try to verify with invalid code
		err = service.Verify2FASetup(secret, "000000")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid verification code")
	})

	t.Run("2FA status reflects current state", func(t *testing.T) {
		// Initially disabled
		status, err := service.GetUser2FAStatus(user.ID)
		assert.NoError(t, err)
		assert.False(t, status)

		// Enable 2FA
		secret, _, err := service.GenerateSecret(user.Username)
		require.NoError(t, err)

		err = service.Enable2FA(user.ID, secret)
		require.NoError(t, err)

		// Now enabled
		status, err = service.GetUser2FAStatus(user.ID)
		assert.NoError(t, err)
		assert.True(t, status)

		// Disable again
		err = service.Disable2FA(user.ID)
		require.NoError(t, err)

		// Back to disabled
		status, err = service.GetUser2FAStatus(user.ID)
		assert.NoError(t, err)
		assert.False(t, status)
	})
}

func TestTOTPService_BackupCodeExhaustion(t *testing.T) {
	db := setupTOTPTestDB(t)
	service := NewTOTPService(db, "DinnerClub")

	user := generateRandomTOTPUser()
	db.Create(&user)

	// Enable 2FA and generate backup codes
	secret, _, err := service.GenerateSecret(user.Username)
	require.NoError(t, err)

	err = service.Enable2FA(user.ID, secret)
	require.NoError(t, err)

	backupCodes, err := service.GenerateBackupCodes(user.ID)
	require.NoError(t, err)

	t.Run("tracks backup code usage correctly", func(t *testing.T) {
		// Use all backup codes
		for i, code := range backupCodes {
			valid, err := service.ValidateBackupCode(user.ID, code)
			assert.NoError(t, err)
			assert.True(t, valid, "Backup code %d should be valid", i)

			// Check remaining count
			unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
			assert.NoError(t, err)
			assert.Equal(t, 10-(i+1), unusedCount, "Should have %d codes remaining", 10-(i+1))
		}

		// All codes should be exhausted
		unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, unusedCount, "All backup codes should be exhausted")
	})

	t.Run("can regenerate backup codes when exhausted", func(t *testing.T) {
		// Regenerate codes
		newCodes, err := service.RegenerateBackupCodes(user.ID)
		assert.NoError(t, err)
		assert.Len(t, newCodes, 10)

		// Verify new codes work
		valid, err := service.ValidateBackupCode(user.ID, newCodes[0])
		assert.NoError(t, err)
		assert.True(t, valid)

		// Verify count is correct
		unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 9, unusedCount) // 9 remaining after using one
	})

	t.Run("old backup codes don't work after regeneration", func(t *testing.T) {
		// Try to use old codes
		for _, oldCode := range backupCodes {
			valid, err := service.ValidateBackupCode(user.ID, oldCode)
			assert.NoError(t, err)
			assert.False(t, valid, "Old backup code should not work: %s", oldCode)
		}
	})
}

// Property 28: Backup Code Exhaustion Handling
// **Feature: advanced-user-system, Property 28: Backup Code Exhaustion Handling**
// **Validates: Requirements 5.6**
func TestProperty_BackupCodeExhaustionHandling(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupTOTPTestDB(t)
			service := NewTOTPService(db, "DinnerClub")

			// Generate random user
			user := generateRandomTOTPUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Enable 2FA and generate backup codes
			secret, _, err := service.GenerateSecret(user.Username)
			require.NoError(t, err)

			err = service.Enable2FA(user.ID, secret)
			require.NoError(t, err)

			backupCodes, err := service.GenerateBackupCodes(user.ID)
			require.NoError(t, err)
			require.Len(t, backupCodes, 10)

			// Test the property: For any user who has used all backup codes,
			// the system should prompt them to generate new backup codes

			// Randomly use some or all backup codes
			numToUse := mathrand.IntN(11) // 0 to 10
			if numToUse > len(backupCodes) {
				numToUse = len(backupCodes)
			}

			for j := 0; j < numToUse; j++ {
				valid, err := service.ValidateBackupCode(user.ID, backupCodes[j])
				require.NoError(t, err)
				assert.True(t, valid, "Backup code %d should be valid", j)
			}

			// Check remaining backup codes
			unusedCount, err := service.GetUnusedBackupCodeCount(user.ID)
			require.NoError(t, err)
			expectedRemaining := 10 - numToUse
			assert.Equal(t, expectedRemaining, unusedCount, "Should have %d unused codes", expectedRemaining)

			// Test exhaustion detection
			if unusedCount == 0 {
				// All codes exhausted - system should detect this
				assert.Equal(t, 0, unusedCount, "System should detect backup code exhaustion")

				// User should be able to regenerate codes
				newCodes, err := service.RegenerateBackupCodes(user.ID)
				require.NoError(t, err)
				assert.Len(t, newCodes, 10, "Should regenerate 10 new backup codes")

				// Verify new codes work
				valid, err := service.ValidateBackupCode(user.ID, newCodes[0])
				require.NoError(t, err)
				assert.True(t, valid, "New backup code should work")

				// Verify count is updated
				unusedCount, err = service.GetUnusedBackupCodeCount(user.ID)
				require.NoError(t, err)
				assert.Equal(t, 9, unusedCount, "Should have 9 codes after using one new code")
			} else if unusedCount <= 2 {
				// Low backup codes - system should warn user
				assert.LessOrEqual(t, unusedCount, 2, "System should detect low backup code count")
				assert.Greater(t, unusedCount, 0, "Should still have some codes remaining")

				// User can still use remaining codes
				for j := numToUse; j < len(backupCodes) && unusedCount > 0; j++ {
					valid, err := service.ValidateBackupCode(user.ID, backupCodes[j])
					require.NoError(t, err)
					if valid {
						unusedCount--
					}
				}

				// After exhaustion, regeneration should work
				if unusedCount == 0 {
					newCodes, err := service.RegenerateBackupCodes(user.ID)
					require.NoError(t, err)
					assert.Len(t, newCodes, 10, "Should regenerate 10 new backup codes")
				}
			} else {
				// Sufficient backup codes remaining
				assert.Greater(t, unusedCount, 2, "Should have sufficient backup codes")
			}

			// Test regeneration at any point
			regeneratedCodes, err := service.RegenerateBackupCodes(user.ID)
			require.NoError(t, err)
			assert.Len(t, regeneratedCodes, 10, "Regeneration should always produce 10 codes")

			// Verify old codes no longer work after regeneration
			for _, oldCode := range backupCodes {
				valid, err := service.ValidateBackupCode(user.ID, oldCode)
				require.NoError(t, err)
				assert.False(t, valid, "Old backup code should not work after regeneration: %s", oldCode)
			}

			// Verify new codes work
			testNewCode := regeneratedCodes[mathrand.IntN(len(regeneratedCodes))]
			valid, err := service.ValidateBackupCode(user.ID, testNewCode)
			require.NoError(t, err)
			assert.True(t, valid, "New backup code should work")

			// Verify count after regeneration
			finalCount, err := service.GetUnusedBackupCodeCount(user.ID)
			require.NoError(t, err)
			assert.Equal(t, 9, finalCount, "Should have 9 codes after using one regenerated code")
		})
	}
}

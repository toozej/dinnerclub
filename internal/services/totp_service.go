// Package services provides business logic services for the dinnerclub application
package services

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// TOTPService handles two-factor authentication using TOTP (Time-based One-Time Password)
type TOTPService struct {
	db     *gorm.DB
	issuer string
}

// NewTOTPService creates a new TOTP service instance
func NewTOTPService(db *gorm.DB, issuer string) *TOTPService {
	return &TOTPService{
		db:     db,
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret and QR code URL for a user
// Returns: secret (base32 encoded), QR code URL, error
func (t *TOTPService) GenerateSecret(username string) (string, string, error) {
	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: username,
		SecretSize:  32, // 256 bits
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Get the secret in base32 format
	secret := key.Secret()

	// Generate QR code URL
	qrURL := key.URL()

	return secret, qrURL, nil
}

// ValidateToken validates a TOTP token against a secret
func (t *TOTPService) ValidateToken(secret, token string) bool {
	// Validate the token
	valid := totp.Validate(token, secret)
	return valid
}

// GenerateBackupCodes generates 10 single-use backup codes for a user
func (t *TOTPService) GenerateBackupCodes(userID uint) ([]string, error) {
	// Delete any existing backup codes for this user
	if err := t.db.Where("user_id = ?", userID).Delete(&models.BackupCode{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete existing backup codes: %w", err)
	}

	// Generate 10 backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := t.generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}

		// Store in database
		backupCodeModel := models.BackupCode{
			UserID: userID,
			Code:   code,
			Used:   false,
		}

		if err := t.db.Create(&backupCodeModel).Error; err != nil {
			return nil, fmt.Errorf("failed to save backup code: %w", err)
		}

		backupCodes[i] = code
	}

	return backupCodes, nil
}

// generateBackupCode generates a single secure backup code
func (t *TOTPService) generateBackupCode() (string, error) {
	// Generate 5 bytes of random data (will be 10 hex characters)
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex and uppercase
	code := strings.ToUpper(hex.EncodeToString(bytes))

	// Format as XXXXX-XXXXX for readability
	if len(code) >= 10 {
		code = code[:5] + "-" + code[5:10]
	}

	return code, nil
}

// ValidateBackupCode validates and marks a backup code as used
func (t *TOTPService) ValidateBackupCode(userID uint, code string) (bool, error) {
	// Normalize the code (remove dashes, uppercase)
	normalizedCode := strings.ToUpper(strings.ReplaceAll(code, "-", ""))

	// Check all unused codes
	var codes []models.BackupCode
	if err := t.db.Where("user_id = ? AND used = ?", userID, false).Find(&codes).Error; err != nil {
		return false, fmt.Errorf("failed to find backup codes: %w", err)
	}

	for _, bc := range codes {
		// Normalize stored code for comparison
		storedNormalized := strings.ToUpper(strings.ReplaceAll(bc.Code, "-", ""))

		if storedNormalized == normalizedCode {
			// Mark as used
			if err := t.db.Model(&bc).Update("used", true).Error; err != nil {
				return false, fmt.Errorf("failed to mark backup code as used: %w", err)
			}
			return true, nil
		}
	}

	return false, nil
}

// RegenerateBackupCodes regenerates all backup codes for a user
func (t *TOTPService) RegenerateBackupCodes(userID uint) ([]string, error) {
	// This is the same as generating new codes (which deletes old ones first)
	return t.GenerateBackupCodes(userID)
}

// GetUnusedBackupCodeCount returns the number of unused backup codes for a user
func (t *TOTPService) GetUnusedBackupCodeCount(userID uint) (int, error) {
	var count int64
	if err := t.db.Model(&models.BackupCode{}).
		Where("user_id = ? AND used = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unused backup codes: %w", err)
	}
	return int(count), nil
}

// Enable2FA enables two-factor authentication for a user
func (t *TOTPService) Enable2FA(userID uint, secret string) error {
	// Update user's 2FA settings
	if err := t.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"two_factor_enabled": true,
			"two_factor_secret":  secret,
		}).Error; err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// Disable2FA disables two-factor authentication for a user
func (t *TOTPService) Disable2FA(userID uint) error {
	// Update user's 2FA settings
	if err := t.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"two_factor_enabled": false,
			"two_factor_secret":  "",
		}).Error; err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Delete all backup codes
	if err := t.db.Where("user_id = ?", userID).Delete(&models.BackupCode{}).Error; err != nil {
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}

	return nil
}

// Verify2FASetup verifies that a user can successfully generate valid TOTP codes
// This should be called before enabling 2FA to ensure the user has properly set up their authenticator
func (t *TOTPService) Verify2FASetup(secret, token string) error {
	if !t.ValidateToken(secret, token) {
		return fmt.Errorf("invalid verification code")
	}
	return nil
}

// GetUser2FAStatus returns the 2FA status for a user
func (t *TOTPService) GetUser2FAStatus(userID uint) (bool, error) {
	var user models.User
	if err := t.db.First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}
	return user.TwoFactorEnabled, nil
}

// GenerateRecoveryKey generates a recovery key that can be used to disable 2FA if locked out
func (t *TOTPService) GenerateRecoveryKey() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base32 for easy typing
	key := base32.StdEncoding.EncodeToString(bytes)

	// Remove padding
	key = strings.TrimRight(key, "=")

	// Format as groups of 4 characters for readability
	var formatted strings.Builder
	for i, char := range key {
		if i > 0 && i%4 == 0 {
			formatted.WriteRune('-')
		}
		formatted.WriteRune(char)
	}

	return formatted.String(), nil
}

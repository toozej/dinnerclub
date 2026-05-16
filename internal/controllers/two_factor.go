// Package controllers provides HTTP request handlers for two-factor authentication
package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
)

// Setup2FARequest represents the request to setup 2FA
type Setup2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// Verify2FARequest represents the request to verify 2FA setup
type Verify2FARequest struct {
	Secret string `json:"secret" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

// Disable2FARequest represents the request to disable 2FA
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"` // TOTP code or backup code
}

// PostSetup2FA handles POST /auth/2fa/setup requests to initiate 2FA setup
func PostSetup2FA(c *gin.Context) {
	Auth := authentication.Resolve()

	// Get authenticated user ID
	userID, err := Auth.UserID(c)
	if err != nil {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Bind request data
	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Response.BadRequest(c, "Invalid request data", err)
		return
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Get user from database
	var user models.User
	result := database.WithContext(ctx).First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Response.NotFound(c, "User not found")
			return
		}
		Response.DatabaseError(c, "find", result.Error)
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		Response.AuthenticationError(c, "invalid_password", err)
		return
	}

	// Check if 2FA is already enabled
	if user.TwoFactorEnabled {
		Response.Conflict(c, "Two-factor authentication is already enabled")
		return
	}

	// Initialize TOTP service
	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	// Generate secret and QR code
	secret, qrURL, err := totpService.GenerateSecret(user.Username)
	if err != nil {
		Response.InternalError(c, "Failed to generate 2FA secret", err)
		return
	}

	log.Debugf("Generated 2FA secret for user %s", user.Username)

	// Return secret and QR code URL
	Response.Success(c, gin.H{
		"secret":  secret,
		"qr_url":  qrURL,
		"message": "Scan the QR code with your authenticator app and verify with a code to complete setup",
	}, "2FA setup initiated")
}

// PostVerify2FA handles POST /auth/2fa/verify requests to complete 2FA setup
func PostVerify2FA(c *gin.Context) {
	Auth := authentication.Resolve()

	// Get authenticated user ID
	userID, err := Auth.UserID(c)
	if err != nil {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Bind request data
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Response.BadRequest(c, "Invalid request data", err)
		return
	}

	// Initialize TOTP service
	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	// Verify the setup with the provided code
	err = totpService.Verify2FASetup(req.Secret, req.Code)
	if err != nil {
		Response.AuthenticationError(c, "invalid_verification_code", err)
		return
	}

	// Enable 2FA for the user
	err = totpService.Enable2FA(userID, req.Secret)
	if err != nil {
		Response.InternalError(c, "Failed to enable 2FA", err)
		return
	}

	// Generate backup codes
	backupCodes, err := totpService.GenerateBackupCodes(userID)
	if err != nil {
		Response.InternalError(c, "Failed to generate backup codes", err)
		return
	}

	log.Infof("2FA enabled for user ID %d", userID)

	// Return backup codes
	Response.Success(c, gin.H{
		"backup_codes": backupCodes,
		"message":      "Two-factor authentication enabled successfully. Save these backup codes in a safe place.",
	}, "2FA enabled")
}

// PostDisable2FA handles POST /auth/2fa/disable requests to disable 2FA
func PostDisable2FA(c *gin.Context) {
	Auth := authentication.Resolve()

	// Get authenticated user ID
	userID, err := Auth.UserID(c)
	if err != nil {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Bind request data
	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Response.BadRequest(c, "Invalid request data", err)
		return
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Get user from database
	var user models.User
	result := database.WithContext(ctx).First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Response.NotFound(c, "User not found")
			return
		}
		Response.DatabaseError(c, "find", result.Error)
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		Response.AuthenticationError(c, "invalid_password", err)
		return
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		Response.BadRequest(c, "Two-factor authentication is not enabled", nil)
		return
	}

	// Initialize TOTP service
	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	// Verify 2FA code (TOTP or backup code)
	var validAuth bool

	// Try TOTP first
	if totpService.ValidateToken(user.TwoFactorSecret, req.Code) {
		validAuth = true
	} else {
		// Try backup code
		valid, err := totpService.ValidateBackupCode(userID, req.Code)
		if err != nil {
			Response.InternalError(c, "Code validation failed", err)
			return
		}
		validAuth = valid
	}

	if !validAuth {
		Response.AuthenticationError(c, "invalid_2fa_code", errors.New("invalid 2FA code"))
		return
	}

	// Disable 2FA
	err = totpService.Disable2FA(userID)
	if err != nil {
		Response.InternalError(c, "Failed to disable 2FA", err)
		return
	}

	log.Infof("2FA disabled for user ID %d", userID)

	Response.Success(c, nil, "Two-factor authentication disabled successfully")
}

// GetBackupCodes handles GET /auth/2fa/backup-codes requests to get backup code status
func GetBackupCodes(c *gin.Context) {
	Auth := authentication.Resolve()

	// Get authenticated user ID
	userID, err := Auth.UserID(c)
	if err != nil {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Get user from database
	var user models.User
	result := database.WithContext(ctx).First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Response.NotFound(c, "User not found")
			return
		}
		Response.DatabaseError(c, "find", result.Error)
		return
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		Response.BadRequest(c, "Two-factor authentication is not enabled", nil)
		return
	}

	// Initialize TOTP service
	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	// Get unused backup code count
	unusedCount, err := totpService.GetUnusedBackupCodeCount(userID)
	if err != nil {
		Response.InternalError(c, "Failed to get backup code count", err)
		return
	}

	Response.Success(c, gin.H{
		"unused_count": unusedCount,
		"total_count":  10,
	}, "Backup code status retrieved")
}

// PostRegenerateBackupCodes handles POST /auth/2fa/backup-codes/regenerate requests
func PostRegenerateBackupCodes(c *gin.Context) {
	Auth := authentication.Resolve()

	// Get authenticated user ID
	userID, err := Auth.UserID(c)
	if err != nil {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Bind request data (requires password for security)
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Response.BadRequest(c, "Invalid request data", err)
		return
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Get user from database
	var user models.User
	result := database.WithContext(ctx).First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			Response.NotFound(c, "User not found")
			return
		}
		Response.DatabaseError(c, "find", result.Error)
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		Response.AuthenticationError(c, "invalid_password", err)
		return
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		Response.BadRequest(c, "Two-factor authentication is not enabled", nil)
		return
	}

	// Initialize TOTP service
	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	// Regenerate backup codes
	backupCodes, err := totpService.RegenerateBackupCodes(userID)
	if err != nil {
		Response.InternalError(c, "Failed to regenerate backup codes", err)
		return
	}

	log.Infof("Backup codes regenerated for user ID %d", userID)

	Response.Success(c, gin.H{
		"backup_codes": backupCodes,
		"message":      "Backup codes regenerated successfully. Save these new codes in a safe place.",
	}, "Backup codes regenerated")
}

// Get2FAStatus handles GET /auth/2fa/status requests to check if 2FA is enabled
func Get2FAStatus(c *gin.Context) {
	Auth := authentication.Resolve()

	// Get authenticated user ID
	userID, err := Auth.UserID(c)
	if err != nil {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Initialize TOTP service
	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	// Get 2FA status
	enabled, err := totpService.GetUser2FAStatus(userID)
	if err != nil {
		Response.InternalError(c, "Failed to get 2FA status", err)
		return
	}

	var unusedBackupCodes int
	if enabled {
		unusedBackupCodes, _ = totpService.GetUnusedBackupCodeCount(userID)
	}

	Response.Success(c, gin.H{
		"enabled":             enabled,
		"unused_backup_codes": unusedBackupCodes,
	}, "2FA status retrieved")
}

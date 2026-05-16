// Package controllers provides HTTP request handlers for the dinnerclub web application.
// This file contains API endpoints for authentication security features including
// password reset and two-factor authentication management.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
)

// PasswordResetRequest represents the request body for initiating a password reset
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PostRequestPasswordReset handles POST /api/auth/reset-password
// Initiates a password reset by sending a temporary password to the user's email.
func PostRequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request: " + err.Error(),
			"success": false,
		})
		return
	}

	// Build email service from config values set by middleware
	cityCode, _ := c.Get("citycode")
	cityCodeStr, _ := cityCode.(string)

	emailService := services.NewEmailService(
		database.DB,
		"",  // SMTP host - not needed for password reset in test/dev
		587, // SMTP port
		"",  // SMTP username
		"",  // SMTP password
		"noreply@"+cityCodeStr+"dinner.club",
		cityCodeStr,
	)

	resetService := services.NewPasswordResetService(database.DB, emailService)

	// RequestPasswordReset silently succeeds even if email not found (security best practice)
	if err := resetService.RequestPasswordReset(req.Email); err != nil {
		errMsg := err.Error()
		log.Warnf("Password reset request failed for email %s: %v", req.Email, err)

		// Only surface rate limit errors to the caller; all other errors (SMTP, DB) are
		// swallowed to avoid leaking information about whether the email exists.
		if len(errMsg) > 0 && errMsg == "too many password reset requests. Please try again later" {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many password reset requests. Please try again later.",
				"success": false,
			})
			return
		}
		// For all other errors (SMTP unavailable, DB issues), still return success
		// to avoid leaking information about account existence.
	}

	log.Infof("Password reset requested for email: %s", req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with that email exists, a temporary password has been sent.",
		"success": true,
	})
}

// PostSetup2FAHandler is the route handler for POST /api/auth/2fa/setup
// Wraps PostSetup2FA to inject citycode from context.
func PostSetup2FAHandler(c *gin.Context) {
	PostSetup2FA(c)
}

// PostVerify2FAHandler is the route handler for POST /api/auth/2fa/verify
// Wraps PostVerify2FA to inject citycode from context.
func PostVerify2FAHandler(c *gin.Context) {
	PostVerify2FA(c)
}

// PostManageBackupCodes handles POST /api/auth/2fa/backup-codes
// Returns backup code status and supports regeneration via action parameter.
func PostManageBackupCodes(c *gin.Context) {
	if !authentication.IsLoggedIn(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"success": false,
		})
		return
	}

	userID, exists := authentication.GetUserID(c)
	if !exists || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"success": false,
		})
		return
	}

	// Parse optional action from request body
	var req struct {
		Action   string `json:"action"`   // "status" (default) or "regenerate"
		Password string `json:"password"` // required for regenerate
	}
	// Ignore bind errors - action defaults to "status"
	_ = c.ShouldBindJSON(&req)

	cityCode, _ := c.Get("citycode")
	cityCodeStr, _ := cityCode.(string)
	totpService := services.NewTOTPService(database.DB, cityCodeStr+"DinnerClub")

	switch req.Action {
	case "regenerate":
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Password is required to regenerate backup codes",
				"success": false,
			})
			return
		}
		// Delegate to the existing regenerate handler
		PostRegenerateBackupCodes(c)
		return

	default:
		// Return backup code status
		unusedCount, err := totpService.GetUnusedBackupCodeCount(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get backup code status",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"unused_count": unusedCount,
			"total_count":  10,
			"success":      true,
		})
	}
}

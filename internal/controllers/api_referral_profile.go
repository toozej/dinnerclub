// Package controllers provides HTTP request handlers for the dinnerclub web application.
// This file contains API endpoints for referral and profile management functionality.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/pkg/database"
)

// ReferralProfileAPIController handles API endpoints for referral and profile management
type ReferralProfileAPIController struct {
	referralService *services.ReferralService
	profileService  *services.ProfileService
	emailService    *services.EmailService
}

// NewReferralProfileAPIController creates a new controller instance
func NewReferralProfileAPIController() *ReferralProfileAPIController {
	db := database.DB

	// Initialize services with proper configuration
	// These would typically come from environment variables in a real application
	emailService := services.NewEmailService(
		db,
		"smtp.example.com",    // SMTP host
		587,                   // SMTP port
		"username",            // SMTP username
		"password",            // SMTP password
		"noreply@dinner.club", // From address
		"pdx",                 // City code - this should come from config
	)

	profileService := services.NewProfileService(db, emailService)
	referralService := services.NewReferralService(db, "pdx", "https://pdx.dinner.club")

	return &ReferralProfileAPIController{
		referralService: referralService,
		profileService:  profileService,
		emailService:    emailService,
	}
}

// GenerateReferralRequest represents the request body for referral message generation
type GenerateReferralRequest struct {
	CustomMessage string `json:"custom_message,omitempty"`
}

// GenerateReferralResponse represents the response for referral message generation
type GenerateReferralResponse struct {
	Message      string `json:"message"`
	ReferralCode string `json:"referral_code"`
	Success      bool   `json:"success"`
}

// PostGenerateReferral generates a referral message for the authenticated user
// POST /api/referrals/generate
func (ctrl *ReferralProfileAPIController) PostGenerateReferral(c *gin.Context) {
	// Get user ID from context (set by authentication middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"success": false,
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid user ID format",
			"success": false,
		})
		return
	}

	// Parse request body
	var req GenerateReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"success": false,
		})
		return
	}

	// Generate referral message
	message, err := ctrl.referralService.GenerateReferralMessage(userIDUint, req.CustomMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate referral message",
			"success": false,
		})
		return
	}

	// Get the actual referral code from the database using the proper User model
	var dbUser models.User
	if err := database.DB.Select("referral_code").Where("id = ?", userIDUint).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get referral code",
			"success": false,
		})
		return
	}

	response := GenerateReferralResponse{
		Message:      message,
		ReferralCode: dbUser.ReferralCode,
		Success:      true,
	}

	c.JSON(http.StatusOK, response)
}

// ReferralStatsResponse represents the response for referral statistics
type ReferralStatsResponse struct {
	UserID          uint `json:"user_id"`
	TotalReferrals  int  `json:"total_referrals"`
	ActiveReferrals int  `json:"active_referrals"`
	Success         bool `json:"success"`
}

// GetReferralStats returns referral statistics for the authenticated user
// GET /api/referrals/stats
func (ctrl *ReferralProfileAPIController) GetReferralStats(c *gin.Context) {
	// Get user ID from context (set by authentication middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"success": false,
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid user ID format",
			"success": false,
		})
		return
	}

	// Get referral statistics
	stats, err := ctrl.referralService.GetReferralStats(userIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get referral statistics",
			"success": false,
		})
		return
	}

	response := ReferralStatsResponse{
		UserID:          stats.UserID,
		TotalReferrals:  stats.TotalReferrals,
		ActiveReferrals: stats.ActiveReferrals,
		Success:         true,
	}

	c.JSON(http.StatusOK, response)
}

// ChangeEmailRequest represents the request body for email change
type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

// ChangeEmailResponse represents the response for email change request
type ChangeEmailResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// PostChangeEmail initiates an email change request
// POST /api/profile/change-email
func (ctrl *ReferralProfileAPIController) PostChangeEmail(c *gin.Context) {
	// Get user ID from context (set by authentication middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"success": false,
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid user ID format",
			"success": false,
		})
		return
	}

	// Parse request body
	var req ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format: " + err.Error(),
			"success": false,
		})
		return
	}

	// Request email change
	token, err := ctrl.profileService.RequestEmailChange(userIDUint, req.NewEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	response := ChangeEmailResponse{
		Token:   token,
		Message: "Email change confirmation sent to new email address",
		Success: true,
	}

	c.JSON(http.StatusOK, response)
}

// ConfirmRequest represents the request body for confirmation token processing
type ConfirmRequest struct {
	Token string `json:"token" binding:"required"`
}

// ConfirmResponse represents the response for confirmation processing
type ConfirmResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// PostConfirm processes a confirmation token for profile changes
// POST /api/profile/confirm
func (ctrl *ReferralProfileAPIController) PostConfirm(c *gin.Context) {
	// Parse request body
	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format: " + err.Error(),
			"success": false,
		})
		return
	}

	// Process confirmation
	err := ctrl.profileService.ConfirmProfileChange(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	response := ConfirmResponse{
		Message: "Profile change confirmed successfully",
		Success: true,
	}

	c.JSON(http.StatusOK, response)
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

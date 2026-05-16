// Package controllers provides HTTP request handlers for the dinnerclub web application.
// This file contains HTML page handlers for security settings, social feed, and notifications.
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
)

// Get2FASetupPage renders the 2FA management page for the authenticated user.
// GET /auth/2fa
func Get2FASetupPage(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

	enabled, err := totpService.GetUser2FAStatus(userID)
	if err != nil {
		Response.InternalError(c, "Failed to get 2FA status", err)
		return
	}

	unusedBackupCodes := 0
	if enabled {
		unusedBackupCodes, _ = totpService.GetUnusedBackupCodeCount(userID)
	}

	c.HTML(http.StatusOK, "auth/2fa_setup.html", gin.H{
		"twoFactorEnabled":  enabled,
		"unusedBackupCodes": unusedBackupCodes,
		"is_logged_in":      c.MustGet("is_logged_in").(bool),
		"citycode":          c.MustGet("citycode").(string),
		"messages":          Flashes(c),
		"context":           c,
	})
}

// GetPasswordResetPage renders the password reset request page.
// GET /auth/reset-password
func GetPasswordResetPage(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/password_reset.html", gin.H{
		"citycode": c.MustGet("citycode").(string),
		"messages": Flashes(c),
		"context":  c,
	})
}

// GetEmailChangePage renders the email change request page for the authenticated user.
// GET /profile/change-email
func GetEmailChangePage(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	// Get current user email
	var currentEmail string
	type emailResult struct {
		Email string
	}
	var result emailResult
	if err := database.DB.Table("users").Select("email").Where("id = ?", userID).Scan(&result).Error; err == nil {
		currentEmail = result.Email
	}

	c.HTML(http.StatusOK, "users/email_change.html", gin.H{
		"currentEmail": currentEmail,
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     Flashes(c),
		"context":      c,
	})
}

// GetEmailConfirmPage handles the email confirmation link click.
// GET /profile/confirm-email
func GetEmailConfirmPage(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.HTML(http.StatusOK, "users/email_change.html", gin.H{
			"tokenError":   true,
			"is_logged_in": c.MustGet("is_logged_in").(bool),
			"citycode":     c.MustGet("citycode").(string),
			"messages":     Flashes(c),
			"context":      c,
		})
		return
	}

	c.HTML(http.StatusOK, "users/email_change.html", gin.H{
		"token":        token,
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     Flashes(c),
		"context":      c,
	})
}

// GetSocialFeedPage renders the social feed page showing entries from followed users.
// GET /social/feed
func GetSocialFeedPage(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	socialService := services.NewSocialService(database.DB)

	feedEntries, err := socialService.GetFollowingFeedWithDetails(userID, limit+1, offset)
	if err != nil {
		feedEntries = nil
	}

	hasNext := len(feedEntries) > limit
	if hasNext {
		feedEntries = feedEntries[:limit]
	}
	hasPrev := page > 1

	c.HTML(http.StatusOK, "social/feed.html", gin.H{
		"entries":      feedEntries,
		"currentPage":  page,
		"hasNext":      hasNext,
		"hasPrev":      hasPrev,
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     Flashes(c),
		"context":      c,
	})
}

// GetNotificationsPage renders the notifications page for the authenticated user.
// GET /social/notifications
func GetNotificationsPage(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	socialService := services.NewSocialService(database.DB)

	// Fetch enough notifications to support pagination
	allNotifications, err := socialService.GetSocialNotifications(userID, offset+limit+1)
	var notifications []models.SocialNotification
	if err == nil && len(allNotifications) > offset {
		notifications = allNotifications[offset:]
	}

	hasNext := len(notifications) > limit
	if hasNext {
		notifications = notifications[:limit]
	}
	hasPrev := page > 1

	// Count unread notifications
	unreadCount := 0
	for _, n := range notifications {
		if !n.Read {
			unreadCount++
		}
	}

	c.HTML(http.StatusOK, "social/notifications.html", gin.H{
		"notifications": notifications,
		"unreadCount":   unreadCount,
		"currentPage":   page,
		"hasNext":       hasNext,
		"hasPrev":       hasPrev,
		"is_logged_in":  c.MustGet("is_logged_in").(bool),
		"citycode":      c.MustGet("citycode").(string),
		"messages":      Flashes(c),
		"context":       c,
	})
}

// PostMarkAllNotificationsRead marks all notifications as read for the authenticated user.
// POST /api/social/notifications/mark-all-read
func PostMarkAllNotificationsRead(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "Authentication required")
		return
	}

	socialService := services.NewSocialService(database.DB)
	if err := socialService.MarkAllNotificationsAsRead(userID); err != nil {
		Response.InternalError(c, "Failed to mark notifications as read", err)
		return
	}

	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/social/notifications", "All notifications marked as read.")
	} else {
		Response.Success(c, nil, "All notifications marked as read")
	}
}

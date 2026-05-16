// Package controllers provides HTTP request handlers for the dinnerclub web application
// including enhanced user profile display and management functionality.
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

// ProfileDisplayController handles enhanced user profile display operations
type ProfileDisplayController struct {
	profileService *services.ProfileDisplayService
}

// NewProfileDisplayController creates a new ProfileDisplayController instance
func NewProfileDisplayController() *ProfileDisplayController {
	// Initialize services
	socialService := services.NewSocialService(database.DB)
	avatarService := services.NewAvatarService(database.DB, "uploads/avatars", 5*1024*1024, []string{"jpg", "jpeg", "png", "gif"})
	profileService := services.NewProfileDisplayService(database.DB, socialService, avatarService)

	return &ProfileDisplayController{
		profileService: profileService,
	}
}

// GetUserProfile displays a user's public profile page
// GET /users/:username/profile
func (pc *ProfileDisplayController) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		Response.BadRequest(c, "Username is required", nil)
		return
	}

	// Get viewer ID if logged in
	viewerID, _ := authentication.GetUserID(c)

	// Get enhanced user profile
	profile, err := pc.profileService.GetUserProfileByUsername(username, viewerID)
	if err != nil {
		Response.NotFound(c, "User")
		return
	}

	// Check if profile is visible to viewer
	if !profile.IsVisible {
		Response.Forbidden(c, "This profile is private")
		return
	}

	// Handle different response types
	if Response.IsHTMLRequest(c) {
		c.HTML(http.StatusOK, "users/public_profile.html", gin.H{
			"profile":      profile,
			"is_logged_in": c.MustGet("is_logged_in").(bool),
			"citycode":     c.MustGet("citycode").(string),
			"messages":     Flashes(c),
			"context":      c,
		})
	} else {
		Response.Success(c, profile, "")
	}
}

// GetUserEntries displays all entries for a user with pagination
// GET /users/:username/entries
func (pc *ProfileDisplayController) GetUserEntries(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		Response.BadRequest(c, "Username is required", nil)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// Get filter parameters
	category := c.Query("category")
	value := c.Query("value")

	// Get viewer ID if logged in
	viewerID, _ := authentication.GetUserID(c)

	// Get user profile to check visibility
	profile, err := pc.profileService.GetUserProfileByUsername(username, viewerID)
	if err != nil {
		Response.NotFound(c, "User")
		return
	}

	if !profile.IsVisible {
		Response.Forbidden(c, "This profile is private")
		return
	}

	var entries []models.Entry
	var total int64

	// Get entries with optional filtering
	if category != "" && value != "" {
		entries, total, err = pc.profileService.GetUserEntriesByCategory(profile.User.ID, category, value, limit, offset)
	} else {
		entries, total, err = pc.profileService.GetAllUserEntries(profile.User.ID, limit, offset)
	}

	if err != nil {
		Response.InternalError(c, "Failed to get user entries", err)
		return
	}

	// Get entry categories for navigation
	categories, err := pc.profileService.GetUserEntryCategories(profile.User.ID)
	if err != nil {
		Response.InternalError(c, "Failed to get entry categories", err)
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	if Response.IsHTMLRequest(c) {
		c.HTML(http.StatusOK, "users/user_entries.html", gin.H{
			"profile":         profile,
			"entries":         entries,
			"categories":      categories,
			"current_page":    page,
			"total_pages":     totalPages,
			"has_next":        hasNext,
			"has_prev":        hasPrev,
			"total_count":     total,
			"filter_category": category,
			"filter_value":    value,
			"is_logged_in":    c.MustGet("is_logged_in").(bool),
			"citycode":        c.MustGet("citycode").(string),
			"messages":        Flashes(c),
			"context":         c,
		})
	} else {
		Response.Success(c, gin.H{
			"entries":      entries,
			"categories":   categories,
			"current_page": page,
			"total_pages":  totalPages,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
			"total_count":  total,
		}, "")
	}
}

// GetEnhancedProfile displays the enhanced profile page for the current user
// GET /profile/enhanced
func (pc *ProfileDisplayController) GetEnhancedProfile(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get enhanced profile
	profile, err := pc.profileService.GetEnhancedUserProfile(userID, userID)
	if err != nil {
		Response.InternalError(c, "Failed to get profile", err)
		return
	}

	// Get profile statistics
	stats, err := pc.profileService.GetProfileStatistics(userID)
	if err != nil {
		Response.InternalError(c, "Failed to get profile statistics", err)
		return
	}

	if Response.IsHTMLRequest(c) {
		c.HTML(http.StatusOK, "users/enhanced_profile.html", gin.H{
			"profile":      profile,
			"stats":        stats,
			"is_logged_in": c.MustGet("is_logged_in").(bool),
			"citycode":     c.MustGet("citycode").(string),
			"messages":     Flashes(c),
			"context":      c,
		})
	} else {
		Response.Success(c, gin.H{
			"profile": profile,
			"stats":   stats,
		}, "")
	}
}

// PostUpdatePrivacy updates user privacy settings
// POST /profile/privacy
func (pc *ProfileDisplayController) PostUpdatePrivacy(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	privacyLevelStr := c.PostForm("privacy_level")
	privacyLevel, err := strconv.Atoi(privacyLevelStr)
	if err != nil {
		Response.BadRequest(c, "Invalid privacy level", err)
		return
	}

	if err := pc.profileService.UpdatePrivacySettings(userID, privacyLevel); err != nil {
		Response.BadRequest(c, "Failed to update privacy settings", err)
		return
	}

	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/profile/enhanced", "Privacy settings updated successfully.")
	} else {
		Response.Success(c, nil, "Privacy settings updated successfully")
	}
}

// PostUpdateBio updates user bio
// POST /profile/update-bio
func (pc *ProfileDisplayController) PostUpdateBio(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	bio := c.PostForm("bio")
	if len(bio) > 500 {
		Response.BadRequest(c, "Bio must be 500 characters or less", nil)
		return
	}

	// Update bio using database directly
	result := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("bio", bio)
	if result.Error != nil {
		Response.InternalError(c, "Failed to update bio", result.Error)
		return
	}

	if result.RowsAffected == 0 {
		Response.NotFound(c, "User")
		return
	}

	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/profile/enhanced", "Bio updated successfully.")
	} else {
		Response.Success(c, nil, "Bio updated successfully")
	}
}

// PostFollowUser handles following/unfollowing a user
// POST /users/:username/follow
func (pc *ProfileDisplayController) PostFollowUser(c *gin.Context) {
	followerID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	username := c.Param("username")
	if username == "" {
		Response.BadRequest(c, "Username is required", nil)
		return
	}

	// Get the user to follow
	profile, err := pc.profileService.GetUserProfileByUsername(username, followerID)
	if err != nil {
		Response.NotFound(c, "User")
		return
	}

	// Prevent self-following
	if profile.User.ID == followerID {
		Response.BadRequest(c, "Cannot follow yourself", nil)
		return
	}

	// Get social service to handle follow/unfollow
	socialService := services.NewSocialService(database.DB)

	var message string
	if profile.IsFollowing {
		// Unfollow
		if err := socialService.UnfollowUser(followerID, profile.User.ID); err != nil {
			Response.InternalError(c, "Failed to unfollow user", err)
			return
		}
		message = "Successfully unfollowed " + profile.User.Username
	} else {
		// Follow
		if err := socialService.FollowUser(followerID, profile.User.ID); err != nil {
			Response.InternalError(c, "Failed to follow user", err)
			return
		}
		message = "Successfully followed " + profile.User.Username
	}

	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/users/"+username+"/profile", message)
	} else {
		Response.Success(c, gin.H{
			"following": !profile.IsFollowing,
		}, message)
	}
}

// GetProfileStatistics returns profile statistics as JSON
// GET /api/users/:username/stats
func (pc *ProfileDisplayController) GetProfileStatistics(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		Response.BadRequest(c, "Username is required", nil)
		return
	}

	// Get viewer ID if logged in
	viewerID, _ := authentication.GetUserID(c)

	// Get user profile to check visibility
	profile, err := pc.profileService.GetUserProfileByUsername(username, viewerID)
	if err != nil {
		Response.NotFound(c, "User")
		return
	}

	if !profile.IsVisible {
		Response.Forbidden(c, "This profile is private")
		return
	}

	// Get statistics
	stats, err := pc.profileService.GetProfileStatistics(profile.User.ID)
	if err != nil {
		Response.InternalError(c, "Failed to get profile statistics", err)
		return
	}

	Response.Success(c, stats, "")
}

// GetPrivacySettings returns current privacy settings for the authenticated user
// GET /api/profile/privacy
func (pc *ProfileDisplayController) GetPrivacySettings(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	settings, err := pc.profileService.GetPrivacySettings(userID)
	if err != nil {
		Response.InternalError(c, "Failed to get privacy settings", err)
		return
	}

	Response.Success(c, settings, "")
}

// CheckProfileAccess checks if the current user can access a profile
// GET /api/users/:username/access
func (pc *ProfileDisplayController) CheckProfileAccess(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		Response.BadRequest(c, "Username is required", nil)
		return
	}

	// Get viewer ID if logged in
	viewerID, _ := authentication.GetUserID(c)

	// Get user profile
	profile, err := pc.profileService.GetUserProfileByUsername(username, viewerID)
	if err != nil {
		Response.NotFound(c, "User")
		return
	}

	// Check access
	canAccess, reason, err := pc.profileService.CheckProfileAccess(profile.User.ID, viewerID)
	if err != nil {
		Response.InternalError(c, "Failed to check profile access", err)
		return
	}

	Response.Success(c, gin.H{
		"can_access":    canAccess,
		"reason":        reason,
		"privacy_level": profile.User.PrivacyLevel,
	}, "")
}

// GetVisibleUsers returns a list of users visible to the current viewer
// GET /api/users/visible
func (pc *ProfileDisplayController) GetVisibleUsers(c *gin.Context) {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// Get viewer ID if logged in
	viewerID, _ := authentication.GetUserID(c)

	// Get visible users
	users, total, err := pc.profileService.GetVisibleUsers(viewerID, limit, offset)
	if err != nil {
		Response.InternalError(c, "Failed to get visible users", err)
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	Response.Success(c, gin.H{
		"users":        users,
		"current_page": page,
		"total_pages":  totalPages,
		"has_next":     hasNext,
		"has_prev":     hasPrev,
		"total_count":  total,
	}, "")
}

package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/internal/validation"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/gorm"
)

type UserUpdate struct {
	Username  string `gorm:"size:255;not null;unique" form:"username" json:"username" binding:"alphanum,lowercase"`
	Password  string `gorm:"size:255;not null;" form:"password" json:"-" binding:"min=10"`
	Firstname string `gorm:"size:255" form:"firstname" json:"firstname"`
	Lastname  string `gorm:"size:255" form:"lastname" json:"lastname"`
	Email     string `gorm:"size:255;unique" form:"email" json:"email" binding:"email"`
}

// GET /users/:id
// Get a user by ID with timeout support
func GetUserByID(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var user models.User

	if err := database.WithContext(ctx).Where("id = ?", c.Param("id")).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Response.NotFound(c, "User")
		} else {
			Response.DatabaseError(c, "find", err)
		}
		return
	}

	Response.Success(c, user, "")
}

// GET /users/:username
// Get a user by username
func GetUserByUsername(c *gin.Context) {
	var user models.User

	if err := database.DB.Where("username = ?", c.Param("username")).First(&user).Error; err != nil {
		Response.NotFound(c, "User")
		return
	}

	Response.Success(c, user, "")
}

// POST /profile/delete
// Delete a user with timeout support
func PostProfileDelete(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := database.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		Response.NotFound(c, "User")
		return
	}

	if err := database.WithContext(ctx).Delete(&user).Error; err != nil {
		Response.DatabaseError(c, "delete", err)
		return
	}

	// Invalidate user cache
	if services.UserCache != nil {
		services.UserCache.InvalidateUser(user.Username)
	}

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		flashMessage(c, fmt.Sprintf("User '%s' deleted successfully.", user.Username))
		// Force the user to logout such that they can no longer post new entries, etc.
		PostLogout(c)
	} else {
		Response.Success(c, nil, fmt.Sprintf("User '%s' deleted successfully", user.Username))
	}
}

// GET /profile
// Get currently logged in user's profile
func GetProfile(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		Response.NotFound(c, "User")
		return
	}

	c.HTML(http.StatusOK, "users/profile.html",
		gin.H{"user": user, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// POST /profile/update
// Update user profile
func PostProfileUpdate(c *gin.Context) {
	userID, exists := authentication.GetUserID(c)
	if !exists {
		Response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get model if exist
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		Response.NotFound(c, "User")
		return
	}

	// Get form data
	newUsername := c.PostForm("newusername")
	newPassword := c.PostForm("newpassword")

	// Bind other form data
	var input UserUpdate
	if err := c.ShouldBind(&input); err != nil {
		Response.BadRequest(c, "Invalid form data", err)
		return
	}

	// Validate the update data using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateUserUpdate(newUsername, newPassword, input.Email)
	if !validationResult.IsValid {
		validation.ValidationHelperInstance.HandleValidationErrors(c, validationResult, "Please check all required fields")
		return
	}

	// Process validated data
	if newUsername != "" {
		input.Username = newUsername
	}
	if newPassword != "" {
		hashedPassword, err := hashPassword(newPassword)
		if err != nil {
			Response.InternalError(c, "Password processing failed", err)
			return
		}
		input.Password = hashedPassword
	}

	// map UserUpdate input struct fields to models.User struct
	updateUser := models.User{
		Username:  input.Username,
		Password:  input.Password,
		Firstname: input.Firstname,
		Lastname:  input.Lastname,
		Email:     input.Email,
	}

	if err := database.DB.Model(&user).Updates(&updateUser).Error; err != nil {
		Response.DatabaseError(c, "update", err)
		return
	}

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/profile", "Profile updated successfully.")
	} else {
		Response.Success(c, gin.H{"username": user.Username}, "Profile updated successfully")
	}
}

func GetCurrentUsername(c *gin.Context) string {
	// First try to get username directly from JWT context
	if username, exists := authentication.GetUsername(c); exists && username != "" {
		return username
	}

	// Fallback to database lookup if username not in context
	userID, exists := authentication.GetUserID(c)
	if !exists {
		log.Info("No authenticated user found")
		return ""
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Infof("No username found matching currently logged in UserID %d", userID)
		return ""
	}

	return user.Username
}

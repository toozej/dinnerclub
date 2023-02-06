package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
)

type UserUpdate struct {
	Username  string `gorm:"size:255;not null;unique" form:"username" json:"username"`
	Password  string `gorm:"size:255;not null;" form:"password" json:"-"`
	Firstname string `gorm:"size:255" form:"firstname" json:"firstname"`
	Lastname  string `gorm:"size:255" form:"lastname" json:"lastname"`
	Email     string `gorm:"size:255;unique" form:"email" json:"email"`
}

// TODO determine if I need FindUserByID function
// GET /users/:id
// Find a user
func FindUserByID(c *gin.Context) {
	var user models.User

	if err := database.DB.Where("id = ?", c.Param("id")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// TODO handle HTML and JSON returns here
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// TODO determine if I need FindUserByUsername function
// GET /users/:username
// Find a user by username
func FindUserByUsername(c *gin.Context) {
	var user models.User

	if err := database.DB.Where("username = ?", c.Param("username")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// TODO handle HTML and JSON returns here
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DELETE /profile/delete
// Delete a user
func DeleteUser(c *gin.Context) {
	Auth := authentication.Resolve()
	userID, _ := Auth.UserID(c)
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	database.DB.Delete(&user)
	flashMessage(c, fmt.Sprintf("User '%s' deleted successfully.", user.Username))

	// force the user to logout such that they can no longer post new entries, etc.
	Logout(c)
}

// GET /profile
// Get currently logged in user's profile
func GetProfile(c *gin.Context) {
	Auth := authentication.Resolve()
	userID, _ := Auth.UserID(c)
	var user models.User

	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}
	c.HTML(http.StatusOK, "users/profile.html",
		gin.H{"user": user, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// PATCH /profile/update
// Update user profile
func UpdateProfilePatch(c *gin.Context) {
	Auth := authentication.Resolve()
	userID, _ := Auth.UserID(c)
	// Get model if exist
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Validate input
	var input UserUpdate
	if c.PostForm("newusername") != "" {
		input.Username = c.PostForm("newusername")
	}
	if c.PostForm("newpassword") != "" {
		input.Password, _ = hashPassword(c.PostForm("newpassword"))
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// flash an update profile message and redirect to the profile page
	flashMessage(c, "Profile updated successfully.")
	c.Redirect(http.StatusFound, "/profile")
}

func GetCurrentUsername(c *gin.Context) string {
	Auth := authentication.Resolve()
	userID, _ := Auth.UserID(c)

	var user models.User

	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Infof("No username found matching currently logged in UserID %d", userID)
		return ""
	}

	return user.Username
}

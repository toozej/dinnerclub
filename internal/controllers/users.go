package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
)

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

// TODO make this handle redirect to /entries/
// TODO make this callable from the profile page with some button to delete account
// DELETE /users/:id
// Delete a user
func DeleteUser(c *gin.Context) {
	var user models.User
	if err := database.DB.Where("id = ?", c.Param("id")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	database.DB.Delete(&user)
	flashMessage(c, fmt.Sprintf("User '%s' deleted successfully.", user.Username))

	// TODO handle HTML and JSON returns here
	c.Redirect(http.StatusAccepted, "/entries")
	// c.JSON(http.StatusAccepted, gin.H{"data": true, "messages": flashes(c)})
}

func GetProfile(c *gin.Context) {
	c.HTML(http.StatusOK, "users/profile.html", gin.H{"is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

func GetCurrentUsername(c *gin.Context) string {
	Auth := authentication.Resolve()
	userID, _ := Auth.UserID(c)

	var user models.User

	// TODO better return hygene here, likely return nil and add an error if username not found
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return "No username found matching currently logged in UserID"
	}

	return user.Username
}

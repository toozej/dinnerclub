package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
)

// GET /restaurants
// Find all restaurants
func FindRestaurants(c *gin.Context) {
	var restaurants []models.Restaurant
	// TODO sort restaurants from newest to oldest
	database.DB.Find(&restaurants)

	c.HTML(http.StatusOK, "restaurants/index.html",
		gin.H{"restaurants": restaurants, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// GET /restaurants/:id
// Find an restaurant
func FindRestaurant(c *gin.Context) {
	var restaurant models.Restaurant

	if err := database.DB.Where("id = ?", c.Param("id")).First(&restaurant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.HTML(http.StatusOK, "restaurants/restaurant.html",
		gin.H{"restaurant": restaurant, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

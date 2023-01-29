package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
)

// GET /restaurants
// Find all restaurants
func FindRestaurants(c *gin.Context) {
	var restaurants []models.Restaurant
	// TODO sort restaurants from newest to oldest
	database.DB.Order("id desc").Find(&restaurants)

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

// Create new restaurant
func CreateRestaurantPost(c *gin.Context) {
	// Validate input
	restaurant := &models.Restaurant{}
	if err := c.ShouldBind(restaurant); err != nil {
		verrs := err.(validator.ValidationErrors)
		messages := make([]string, len(verrs))
		for i, verr := range verrs {
			messages[i] = fmt.Sprintf(
				"%s is required, but was empty.",
				verr.Field())
		}
		c.HTML(http.StatusBadRequest, "entries/new.html",
			gin.H{"errors": messages, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
		return
	}

	// if no record with restaurant name already exists, create one. Otherwise error
	if err := database.DB.Where("name = ?", c.Param("name")).First(&restaurant).Error; err != nil {
		// create the new restaurant
		if err := database.DB.Create(&restaurant).Error; err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// flash a message that the new restaurant entry was saved successfully
		flashMessage(c, fmt.Sprintf("New restaurant '%s' saved successfully.", restaurant.Name))
	} else {
		flashMessage(c, fmt.Sprintf("Restaurant '%s' already exists.", restaurant.Name))
	}
}

func DeleteRestaurantPost(c *gin.Context) {
	// TODO use same validation and ShouldBind() from controllers.CreateEntry()
	var restaurant models.Restaurant
	if err := database.DB.Where("id = ?", c.Param("id")).First(&restaurant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if err := database.DB.Delete(&restaurant).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	flashMessage(c, fmt.Sprintf("Restaurant '%s' deleted successfully.", restaurant.Name))
	c.Redirect(http.StatusOK, "/restaurants")
}

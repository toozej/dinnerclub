package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/helpers"
	"github.com/toozej/dinnerclub/pkg/pagination"
)

// GET /restaurants
// Find all restaurants
func FindRestaurants(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	var restaurantsCount int64
	if err := database.DB.Table("restaurants").Count(&restaurantsCount).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	const restaurantsPerPage = 10
	p, err := pagination.Paginate(pageStr, int(restaurantsCount), restaurantsPerPage)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var restaurants []models.Restaurant
	if err := database.DB.Order("id desc").Limit(restaurantsPerPage).Offset(p.Offset).Find(&restaurants).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "restaurants/index.html",
		gin.H{"restaurants": restaurants, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c), "p": p})
}

func getPopularItems(restaurant models.Restaurant) ([]string, error) {
	var entry models.Entry
	var rawpopularitems []string
	if err := database.DB.Model(&entry).Select("Ordered").Where("Name = ? AND Ordered <> '' AND Ordered IS NOT NULL", restaurant.Name).Find(&rawpopularitems).Error; err != nil {
		return nil, err
	}

	var popularitems []string

	// remove leading and trailing whitespace on ordered items
	// then split comma-separated items into the popularitems slice
	for _, i := range rawpopularitems {
		trimmeditem := strings.TrimSpace(i)
		splititem := strings.Split(trimmeditem, ",")
		popularitems = append(popularitems, splititem...)
	}

	// sort and then deduplicate popularitems, most to least popular
	popularitems = helpers.RemoveDuplicates(helpers.SortByFrequency(popularitems))

	return popularitems, nil
}

// GET /restaurants/:id
// Find an restaurant
func FindRestaurant(c *gin.Context) {
	var restaurant models.Restaurant

	if err := database.DB.Where("id = ?", c.Param("id")).First(&restaurant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	popularitems, err := getPopularItems(restaurant)
	if err != nil {
		flashMessage(c, fmt.Sprintf("Restaurant '%s' doesn't have any popular items yet.", restaurant.Name))
		return
	}

	var entries []models.Entry
	if err := database.DB.Model(&entries).Select("Submitter", "MealService", "FoodRating", "AmbienceRating", "ValueRating", "Comments").Where("Name = ?", restaurant.Name).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.HTML(http.StatusOK, "restaurants/restaurant.html",
		gin.H{"restaurant": restaurant, "popularitems": popularitems, "entries": entries, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// Create new restaurant
func CreateRestaurantPost(c *gin.Context) {
	// Validate input
	closed, _ := strconv.ParseBool(c.PostForm("closed"))
	restaurant := models.Restaurant{
		Name:     c.PostForm("name"),
		Location: c.PostForm("location"),
		Cuisine:  c.PostForm("cuisine"),
		Closed:   closed,
	}

	// if no record with restaurant name already exists, create one
	if err := database.DB.Model(&restaurant).Where("name = ?", c.Param("name")).Error; err != nil {
		// create the new restaurant
		if err := database.DB.Create(&restaurant).Error; err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// flash a message that the new restaurant entry was saved successfully
		flashMessage(c, fmt.Sprintf("New restaurant '%s' saved successfully.", restaurant.Name))
	}
}

// Delete restaurant when no entries reference it
func DeleteRestaurantPost(c *gin.Context, restaurantName string) {
	var entry models.Entry
	var restaurant models.Restaurant
	// if no entries with restaurant exist then delete the restaurant record
	// (a.k.a. if this call to DeleteRestaurantPost was from deleting the only entry referencing
	// this restaurant, then delete the restaurant record)
	// TODO likely need to check deleted_at field is null or something, not working currently
	if err := database.DB.Model(&entry).Where("name = ?", restaurantName).Error; err != nil {
		if err := database.DB.Delete(&restaurant).Error; err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		flashMessage(c, fmt.Sprintf("Restaurant '%s' deleted successfully.", restaurant.Name))
		c.Redirect(http.StatusOK, "/restaurants")
	}
}

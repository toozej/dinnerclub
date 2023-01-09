package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
)

// GET /entries
// Find all entries
func FindEntries(c *gin.Context) {
	var entries []models.Entry
	database.DB.Find(&entries)

	c.JSON(http.StatusOK, gin.H{"data": entries})
}

// GET /entries/:id
// Find an entry
func FindEntry(c *gin.Context) {
	var entry models.Entry

	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// POST /entries
// Create new entry
func CreateEntry(c *gin.Context) {
	// Validate input
	var input models.Entry
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create entry
	entry := models.Entry{
		Submitter:      input.Submitter,
		Name:           input.Name,
		Location:       input.Location,
		Cuisine:        input.Cuisine,
		Visited:        input.Visited,
		Closed:         input.Closed,
		MealService:    input.MealService,
		Ordered:        input.Ordered,
		FoodRating:     input.FoodRating,
		AmbienceRating: input.AmbienceRating,
		ValueRating:    input.ValueRating,
		Comments:       input.Comments,
	}
	database.DB.Create(&entry)

	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// PATCH /entries/:id
// Update an entry
func UpdateEntry(c *gin.Context) {
	// Get model if exist
	var entry models.Entry
	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Validate input
	var input models.Entry
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&entry).Updates(input)

	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// DELETE /entries/:id
// Delete an entry
func DeleteEntry(c *gin.Context) {
	var entry models.Entry
	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	database.DB.Delete(&entry)

	c.JSON(http.StatusOK, gin.H{"data": true})
}

package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
)

// GET /entries
// Find all entries
func FindEntries(c *gin.Context) {
	var entries []models.Entry
	database.DB.Find(&entries)

	c.HTML(http.StatusOK, "entries/index.html", gin.H{"entries": entries})
}

// GET /entries/:id
// Find an entry
func FindEntry(c *gin.Context) {
	var entry models.Entry

	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.HTML(http.StatusOK, "entries/entry.html", gin.H{"entry": entry})
}

// POST /entries
// Create new entry
func CreateEntry(c *gin.Context) {
	// Validate input
	entry := &models.Entry{}
	// var input models.Entry
	// if err := c.ShouldBind(&input); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// // Create entry
	// entry := models.Entry{
	// 	Submitter:      input.Submitter,
	// 	Name:           input.Name,
	// 	Location:       input.Location,
	// 	Cuisine:        input.Cuisine,
	// 	Visited:        input.Visited,
	// 	Closed:         input.Closed,
	// 	MealService:    input.MealService,
	// 	Ordered:        input.Ordered,
	// 	FoodRating:     input.FoodRating,
	// 	AmbienceRating: input.AmbienceRating,
	// 	ValueRating:    input.ValueRating,
	// 	Comments:       input.Comments,
	// }

	if err := c.ShouldBind(entry); err != nil {
		verrs := err.(validator.ValidationErrors)
		messages := make([]string, len(verrs))
		for i, verr := range verrs {
			messages[i] = fmt.Sprintf(
				"%s is required, but was empty.",
				verr.Field())
		}
		c.HTML(http.StatusBadRequest, "entries/new.html",
			gin.H{"errors": messages})
		return
	}

	if err := database.DB.Create(&entry).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO add flash for successful creation

	redirectPath := fmt.Sprintf("/entries/%d", entry.ID)
	c.Redirect(http.StatusFound, redirectPath)
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
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&entry).Updates(input).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO add flash for successful update
	redirectPath := fmt.Sprintf("/entry/%d", entry.ID)
	c.Redirect(http.StatusFound, redirectPath)
}

// DELETE /entries/:id
// Delete an entry
func DeleteEntry(c *gin.Context) {
	var entry models.Entry
	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if err := database.DB.Delete(&entry).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO add flash for successful deletion
	c.Redirect(http.StatusOK, "/entries")
}

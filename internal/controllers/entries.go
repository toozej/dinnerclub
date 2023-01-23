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
	// TODO sort entries from newest to oldest
	database.DB.Find(&entries)

	c.HTML(http.StatusOK, "entries/index.html",
		gin.H{"entries": entries, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// GET /entries/:id
// Find an entry
func FindEntry(c *gin.Context) {
	var entry models.Entry

	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.HTML(http.StatusOK, "entries/entry.html",
		gin.H{"entry": entry, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// GET /entries
// Create new entry HTML form
func CreateEntryGet(c *gin.Context) {
	c.HTML(http.StatusOK, "entries/new.html", gin.H{"citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// POST /entries
// Create new entry
func CreateEntryPost(c *gin.Context) {
	// Validate input
	entry := &models.Entry{}
	if err := c.ShouldBind(entry); err != nil {
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

	// TODO set entry.Submitter field to the username of the person submitting the form
	entry.Submitter = GetCurrentUsername(c)

	if err := database.DB.Create(&entry).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	flashMessage(c, fmt.Sprintf("New entry '%s' saved successfully.", entry.Name))
	redirectPath := fmt.Sprintf("/entries/%d", entry.ID)
	c.Redirect(http.StatusFound, redirectPath)
}

// GET /entries/:id/update
// Update an entry HTML form
func UpdateEntryGet(c *gin.Context) {
	c.HTML(http.StatusOK, "entries/update.html", gin.H{"citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// PATCH /entries/:id
// Update an entry
func UpdateEntryPatch(c *gin.Context) {
	// TODO use same validation and ShouldBind() from CreateEntry()
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

	flashMessage(c, fmt.Sprintf("Entry '%s' updated successfully.", entry.Name))
	redirectPath := fmt.Sprintf("/entry/%d", entry.ID)
	c.Redirect(http.StatusFound, redirectPath)
}

// DELETE /entries/:id
// Delete an entry
func DeleteEntry(c *gin.Context) {
	// TODO use same validation and ShouldBind() from CreateEntry()
	var entry models.Entry
	if err := database.DB.Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if err := database.DB.Delete(&entry).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	flashMessage(c, fmt.Sprintf("Entry '%s' deleted successfully.", entry.Name))
	c.Redirect(http.StatusOK, "/entries")
}

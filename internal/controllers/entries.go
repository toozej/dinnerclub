package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/internal/validation"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/pagination"
	"gorm.io/gorm"
)

// EntryUpdate represents the data structure for entry update operations
// with form binding tags for all entry fields.
type EntryUpdate struct {
	Name             string `form:"name" json:"name" gorm:"notNull"`
	Location         string `form:"location" json:"location" gorm:"notNull"`
	Cuisine          string `form:"cuisine" json:"cuisine" gorm:"notNull"`
	Visited          bool   `form:"visited" json:"visited"`
	LookingForwardTo string `form:"lookingforwardto" json:"lookingforwardto"`
	Closed           bool   `form:"closed" json:"closed" gorm:"default:false"`
	MealService      string `form:"mealservice" json:"mealservice"`
	Ordered          string `form:"ordered" json:"ordered"`
	OrderNext        string `form:"ordernext" json:"ordernext"`
	DoNotOrder       string `form:"donotorder" json:"donotorder"`
	FoodRating       int    `form:"foodrating" json:"foodrating"`
	AmbienceRating   int    `form:"ambiencerating" json:"ambiencerating"`
	ValueRating      int    `form:"valuerating" json:"valuerating"`
	Comments         string `form:"comments" json:"comments"`
}

// GetEntries handles GET /entries requests and returns a paginated list of all
// entries ordered by ID in descending order with timeout support.
func GetEntries(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	pageStr := c.DefaultQuery("page", "1")
	var entriesCount int64

	// Use timeout-aware database connection
	if err := database.WithContext(ctx).Table("entries").Count(&entriesCount).Error; err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	const entriesPerPage = 10
	p, err := pagination.Paginate(pageStr, int(entriesCount), entriesPerPage)
	if err != nil {
		Response.BadRequest(c, "Invalid pagination parameters", err)
		return
	}

	var entries []models.Entry
	if err := database.WithContext(ctx).Order("id desc").Limit(entriesPerPage).Offset(p.Offset).Find(&entries).Error; err != nil {
		Response.DatabaseError(c, "find", err)
		return
	}

	c.HTML(http.StatusOK, "entries/index.html",
		gin.H{"entries": entries, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "p": p, "context": c})
}

// GetEntry handles GET /entries/:id requests and returns a single entry by ID
// with ownership information for edit/delete permissions using timeout-aware queries.
func GetEntry(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var entry models.Entry

	if err := database.WithContext(ctx).Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Response.NotFound(c, "Entry")
		} else {
			Response.DatabaseError(c, "find", err)
		}
		return
	}

	// determine if the entry's submitter is the currently logged in user
	// if so, additional entry options like update and delete will be available
	mine := entry.Submitter == GetCurrentUsername(c)

	c.HTML(http.StatusOK, "entries/entry.html",
		gin.H{"entry": entry, "is_logged_in": c.MustGet("is_logged_in").(bool), "is_mine": mine, "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// GetEntriesByUsername handles GET /entries/submittedby/:username requests and returns
// a paginated list of entries submitted by a specific user with caching and timeout support.
func GetEntriesByUsername(c *gin.Context) {
	username := c.Param("username")
	pageStr := c.DefaultQuery("page", "1")

	// Initialize cache if not already done
	if services.UserCache == nil {
		services.InitUserCache()
	}

	// Try to get entries from cache first (for first page only)
	if pageStr == "1" {
		if cachedEntries, found := services.UserCache.GetUserEntries(username); found {
			const entriesPerPage = 10
			// Calculate pagination for cached entries
			totalEntries := len(cachedEntries)
			p, err := pagination.Paginate(pageStr, totalEntries, entriesPerPage)
			if err != nil {
				Response.BadRequest(c, "Invalid pagination parameters", err)
				return
			}

			// Get the first page from cached entries
			endIdx := entriesPerPage
			if endIdx > len(cachedEntries) {
				endIdx = len(cachedEntries)
			}
			entries := cachedEntries[:endIdx]

			c.HTML(http.StatusOK, "entries/index.html",
				gin.H{"entries": entries, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "p": p, "context": c})
			return
		}
	}

	// Fallback to database with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	userEntriesService := services.NewUserEntriesService()
	const entriesPerPage = 10

	p, err := pagination.Paginate(pageStr, 0, entriesPerPage) // We'll get count from service
	if err != nil {
		Response.BadRequest(c, "Invalid pagination parameters", err)
		return
	}

	entries, totalCount, err := userEntriesService.GetUserEntriesPaginated(ctx, username, entriesPerPage, p.Offset)
	if err != nil {
		Response.InternalError(c, "Failed to retrieve entries", err)
		return
	}

	// Update pagination with actual count
	p, err = pagination.Paginate(pageStr, int(totalCount), entriesPerPage)
	if err != nil {
		Response.BadRequest(c, "Invalid pagination parameters", err)
		return
	}

	c.HTML(http.StatusOK, "entries/index.html",
		gin.H{"entries": entries, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "p": p, "context": c})
}

// GetEntryForm handles GET /entries/new requests and renders the new entry
// creation form template.
func GetEntryForm(c *gin.Context) {
	c.HTML(http.StatusOK, "entries/new.html", gin.H{"is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// PostEntry handles POST /entries requests, validates entry data, creates new entries,
// and automatically creates associated restaurant records with timeout support.
func PostEntry(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Bind input data
	entry := &models.Entry{}
	if err := c.ShouldBind(entry); err != nil {
		Response.BadRequest(c, "Invalid form data", err)
		return
	}

	// Validate entry using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateEntry(entry)
	if !validationResult.IsValid {
		validation.ValidationHelperInstance.HandleValidationErrors(c, validationResult, "Please check all required fields")
		return
	}

	// set entry.Submitter field to the username of the person submitting the form
	entry.Submitter = GetCurrentUsername(c)

	// Use service for timeout-aware entry creation
	userEntriesService := services.NewUserEntriesService()
	if err := userEntriesService.CreateEntryWithTimeout(ctx, entry); err != nil {
		Response.DatabaseError(c, "create", err)
		return
	}

	// Invalidate user cache since we added a new entry
	if services.UserCache != nil {
		services.UserCache.InvalidateUser(entry.Submitter)
	}

	// create restaurant using new entry form information
	PostRestaurant(c)

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, fmt.Sprintf("/entries/%d", entry.ID), fmt.Sprintf("New entry '%s' saved successfully.", entry.Name))
	} else {
		Response.Created(c, gin.H{"id": entry.ID, "name": entry.Name}, "Entry created successfully")
	}
}

// GetEntryUpdateForm handles GET /entries/:id/update requests and renders the entry
// update form with pre-populated data using timeout-aware queries.
func GetEntryUpdateForm(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var entry models.Entry

	if err := database.WithContext(ctx).Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		Response.NotFound(c, "Entry")
		return
	}
	c.HTML(http.StatusOK, "entries/update.html", gin.H{"entry": entry, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// PutEntry handles PUT /entries/:id requests, validates updated entry data,
// and updates the entry record in the database with timeout support.
func PutEntry(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get model if exist
	var entry models.Entry
	if err := database.WithContext(ctx).Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		Response.NotFound(c, "Entry")
		return
	}

	// Bind input data
	var input EntryUpdate
	if err := c.ShouldBind(&input); err != nil {
		Response.BadRequest(c, "Invalid form data", err)
		return
	}

	// map EntryUpdate input struct fields to models.Entry struct for validation
	updateEntry := models.Entry{
		Name:             input.Name,
		Location:         input.Location,
		Cuisine:          input.Cuisine,
		Visited:          input.Visited,
		LookingForwardTo: input.LookingForwardTo,
		Closed:           input.Closed,
		MealService:      input.MealService,
		Ordered:          input.Ordered,
		OrderNext:        input.OrderNext,
		DoNotOrder:       input.DoNotOrder,
		FoodRating:       input.FoodRating,
		AmbienceRating:   input.AmbienceRating,
		ValueRating:      input.ValueRating,
		Comments:         input.Comments,
	}

	// Validate the updated entry using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateEntry(&updateEntry)
	if !validationResult.IsValid {
		validation.ValidationHelperInstance.HandleValidationErrors(c, validationResult, "Please check all required fields")
		return
	}

	// Use service for timeout-aware update
	userEntriesService := services.NewUserEntriesService()
	if err := userEntriesService.UpdateEntryWithTimeout(ctx, entry.ID, &updateEntry); err != nil {
		Response.DatabaseError(c, "update", err)
		return
	}

	// Invalidate user cache since we updated an entry
	if services.UserCache != nil {
		services.UserCache.InvalidateUser(entry.Submitter)
	}

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, fmt.Sprintf("/entries/%d", entry.ID), fmt.Sprintf("Entry '%s' updated successfully.", entry.Name))
	} else {
		Response.Success(c, gin.H{"id": entry.ID, "name": entry.Name}, "Entry updated successfully")
	}
}

// DeleteEntry handles DELETE /entries/:id requests, deletes the entry,
// and removes associated restaurant records if no other entries reference them with timeout support.
func DeleteEntry(c *gin.Context) {
	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var entry models.Entry
	if err := database.WithContext(ctx).Where("id = ?", c.Param("id")).First(&entry).Error; err != nil {
		Response.NotFound(c, "Entry")
		return
	}

	// Store submitter for cache invalidation
	submitter := entry.Submitter

	// Use service for timeout-aware deletion
	userEntriesService := services.NewUserEntriesService()
	if err := userEntriesService.DeleteEntryWithTimeout(ctx, entry.ID); err != nil {
		Response.DatabaseError(c, "delete", err)
		return
	}

	// Invalidate user cache since we deleted an entry
	if services.UserCache != nil {
		services.UserCache.InvalidateUser(submitter)
	}

	// Delete restaurant entry if there's no more entries referencing that restaurant
	DeleteRestaurantIfUnused(c, entry.Name)

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/entries", fmt.Sprintf("Entry '%s' deleted successfully.", entry.Name))
	} else {
		Response.Success(c, nil, fmt.Sprintf("Entry '%s' deleted successfully", entry.Name))
	}
}

package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	if err := db.AutoMigrate(&models.Entry{}, &models.Restaurant{}); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// Create test data
	entries := []models.Entry{
		{
			Name:      "Test Restaurant 1",
			Location:  "N",
			Cuisine:   "Italian",
			Submitter: "testuser",
			Visited:   true,
			Comments:  "Great pasta",
		},
		{
			Name:             "Test Restaurant 2",
			Location:         "S",
			Cuisine:          "Chinese",
			Submitter:        "testuser2",
			Visited:          false,
			LookingForwardTo: "Trying their dumplings",
		},
		{
			Name:      "Pizza Place",
			Location:  "NE",
			Cuisine:   "Italian",
			Submitter: "testuser",
			Visited:   true,
			Ordered:   "Margherita pizza, garlic bread",
		},
	}

	for _, entry := range entries {
		db.Create(&entry)
	}

	restaurants := []models.Restaurant{
		{Name: "Test Restaurant 1", Location: "N", Cuisine: "Italian"},
		{Name: "Test Restaurant 2", Location: "S", Cuisine: "Chinese"},
		{Name: "Pizza Place", Location: "NE", Cuisine: "Italian"},
	}

	for _, restaurant := range restaurants {
		db.Create(&restaurant)
	}

	return db
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Set up test database
	database.DB = setupTestDB()

	// Create search controller
	searchController := NewSearchController()

	// Set up routes
	r.GET("/search", searchController.GetSearch)
	r.GET("/search/entries", searchController.GetSearchEntries)
	r.GET("/search/restaurants", searchController.GetSearchRestaurants)
	r.GET("/search/suggestions", searchController.GetSuggestions)

	return r
}

func TestSearchController_Search(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name                string
		query               string
		expectedStatus      int
		expectedEntries     int
		expectedRestaurants int
	}{
		{
			name:                "Search for Italian cuisine",
			query:               "Italian",
			expectedStatus:      http.StatusOK,
			expectedEntries:     2,
			expectedRestaurants: 2,
		},
		{
			name:                "Search for specific restaurant",
			query:               "Pizza Place",
			expectedStatus:      http.StatusOK,
			expectedEntries:     1,
			expectedRestaurants: 1,
		},
		{
			name:                "Search by author",
			query:               "author:testuser",
			expectedStatus:      http.StatusOK,
			expectedEntries:     2,
			expectedRestaurants: 0,
		},
		{
			name:                "Search wishlist items",
			query:               "wishlist",
			expectedStatus:      http.StatusOK,
			expectedEntries:     1,
			expectedRestaurants: 0,
		},
		{
			name:           "Empty query should return success",
			query:          "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/search?q="+tt.query+"&format=json", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				var entries []interface{}
				var restaurants []interface{}

				if response["entries"] != nil {
					entries = response["entries"].([]interface{})
				}
				if response["restaurants"] != nil {
					restaurants = response["restaurants"].([]interface{})
				}

				assert.Equal(t, tt.expectedEntries, len(entries))
				assert.Equal(t, tt.expectedRestaurants, len(restaurants))
			}
		})
	}
}

func TestSearchController_SearchEntries(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		query          string
		sort           string
		order          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Search entries by cuisine",
			query:          "Italian",
			sort:           "name",
			order:          "asc",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Search entries by author",
			query:          "author:testuser",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Invalid sort field",
			query:          "Italian",
			sort:           "invalid_field",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid sort order",
			query:          "Italian",
			order:          "invalid_order",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			url := "/search/entries?q=" + tt.query + "&format=json"
			if tt.sort != "" {
				url += "&sort=" + tt.sort
			}
			if tt.order != "" {
				url += "&order=" + tt.order
			}

			req, _ := http.NewRequest("GET", url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				entries := response["entries"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(entries))
			}
		})
	}
}

func TestSearchController_SearchRestaurants(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		query          string
		sort           string
		order          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Search restaurants by cuisine",
			query:          "Italian",
			sort:           "name",
			order:          "asc",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Search restaurants by location",
			query:          "location:N",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Sort by entry count",
			query:          "Italian",
			sort:           "entry_count",
			order:          "desc",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			url := "/search/restaurants?q=" + tt.query + "&format=json"
			if tt.sort != "" {
				url += "&sort=" + tt.sort
			}
			if tt.order != "" {
				url += "&order=" + tt.order
			}

			req, _ := http.NewRequest("GET", url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				restaurants := response["restaurants"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(restaurants))
			}
		})
	}
}

func TestSearchController_GetSuggestions(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		minSuggestions int
	}{
		{
			name:           "Get suggestions for 'Test'",
			query:          "Test",
			expectedStatus: http.StatusOK,
			minSuggestions: 1,
		},
		{
			name:           "Get suggestions for 'Pizza'",
			query:          "Pizza",
			expectedStatus: http.StatusOK,
			minSuggestions: 1,
		},
		{
			name:           "Short query returns empty suggestions",
			query:          "T",
			expectedStatus: http.StatusOK,
			minSuggestions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/search/suggestions?q="+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				suggestions := response["suggestions"].([]interface{})
				assert.GreaterOrEqual(t, len(suggestions), tt.minSuggestions)
			}
		})
	}
}

func TestSearchController_Pagination(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		page           string
		limit          string
		expectedStatus int
	}{
		{
			name:           "Valid pagination parameters",
			page:           "1",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid page parameter",
			page:           "invalid",
			limit:          "10",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid limit parameter",
			page:           "1",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Limit too high",
			page:           "1",
			limit:          "200",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			url := "/search?q=Italian&format=json&page=" + tt.page + "&limit=" + tt.limit
			req, _ := http.NewRequest("GET", url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

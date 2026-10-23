package controllers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/helpers"
	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLocationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&models.Restaurant{}, &models.Entry{}, &models.User{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupLocationTestRouter creates a test router with the required template functions registered.
func setupLocationTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	router := gin.New()

	h := &helpers.TemplateHelpers{}
	funcMap := template.FuncMap{
		"FormatBoolean":        h.FormatBoolean,
		"FormatRating":         h.FormatRating,
		"GetRatingDescription": h.GetRatingDescription,
		"GetRatingLabel":       h.GetRatingLabel,
		"UserProfileLink":      h.UserProfileLink,
		"TooltipText":          h.TooltipText,
		"GetRatingOptions":     h.GetRatingOptions,
		"SafeHTML":             h.SafeHTML,
		"Add":                  h.Add,
		"add":                  h.Add,
		"StringToInt":          h.StringToInt,
		"IsEmpty":              h.IsEmpty,
		"FormatDate":           h.FormatDate,
		"FormatDateTime":       h.FormatDateTime,
		"CSRFToken":            h.CSRFToken,
		"CSRFField":            h.CSRFField,
		"CSRFInput":            h.CSRFInput,
		"sub":                  h.Sub,
		"Sub":                  h.Sub,
		"mul":                  h.Mul,
		"Mul":                  h.Mul,
		"default":              h.Default,
	}
	router.SetFuncMap(funcMap)
	router.LoadHTMLGlob("../../templates/**/*")
	return router
}

func TestShowLocationSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := setupLocationTestRouter(t)
	router.GET("/location", controller.ShowLocationSearch)

	req, _ := http.NewRequest("GET", "/location", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSearchNearby_InvalidLatitude(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := setupLocationTestRouter(t)
	router.POST("/location/search", controller.SearchNearby)

	form := url.Values{}
	form.Add("latitude", "invalid")
	form.Add("longitude", "-74.0060")
	form.Add("radius", "10")
	form.Add("search-type", "nearby")

	req, _ := http.NewRequest("POST", "/location/search", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchNearby_InvalidLongitude(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := setupLocationTestRouter(t)
	router.POST("/location/search", controller.SearchNearby)

	form := url.Values{}
	form.Add("latitude", "40.7128")
	form.Add("longitude", "invalid")
	form.Add("radius", "10")
	form.Add("search-type", "nearby")

	req, _ := http.NewRequest("POST", "/location/search", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchNearby_InvalidRadius(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := setupLocationTestRouter(t)
	router.POST("/location/search", controller.SearchNearby)

	tests := []struct {
		name   string
		radius string
	}{
		{"negative radius", "-10"},
		{"zero radius", "0"},
		{"too large radius", "150"},
		{"invalid radius", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("latitude", "40.7128")
			form.Add("longitude", "-74.0060")
			form.Add("radius", tt.radius)
			form.Add("search-type", "nearby")

			req, _ := http.NewRequest("POST", "/location/search", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestSearchNearby_OutOfRangeCoordinates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := setupLocationTestRouter(t)
	router.POST("/location/search", controller.SearchNearby)

	tests := []struct {
		name      string
		latitude  string
		longitude string
	}{
		{"latitude too high", "91", "0"},
		{"latitude too low", "-91", "0"},
		{"longitude too high", "0", "181"},
		{"longitude too low", "0", "-181"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("latitude", tt.latitude)
			form.Add("longitude", tt.longitude)
			form.Add("radius", "10")
			form.Add("search-type", "nearby")

			req, _ := http.NewRequest("POST", "/location/search", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestSearchNearby_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := setupLocationTestRouter(t)
	router.POST("/location/search", controller.SearchNearby)

	form := url.Values{}
	form.Add("latitude", "40.7128")
	form.Add("longitude", "-74.0060")
	form.Add("radius", "10")
	form.Add("search-type", "nearby")

	req, _ := http.NewRequest("POST", "/location/search", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetNearbyRestaurantsAPI_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := gin.New()
	router.GET("/api/location/nearby", controller.GetNearbyRestaurantsAPI)

	tests := []struct {
		name      string
		latitude  string
		longitude string
		radius    string
	}{
		{"invalid latitude", "invalid", "-74.0060", "10"},
		{"invalid longitude", "40.7128", "invalid", "10"},
		{"invalid radius", "40.7128", "-74.0060", "invalid"},
		{"negative radius", "40.7128", "-74.0060", "-10"},
		{"too large radius", "40.7128", "-74.0060", "150"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/location/nearby?latitude=" + tt.latitude + "&longitude=" + tt.longitude + "&radius=" + tt.radius
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestGetNearbyRestaurantsAPI_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := gin.New()
	router.GET("/api/location/nearby", controller.GetNearbyRestaurantsAPI)

	req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=40.7128&longitude=-74.0060&radius=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetNearestWishlistAPI_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := gin.New()
	router.GET("/api/location/wishlist", controller.GetNearestWishlistAPI)

	req, _ := http.NewRequest("GET", "/api/location/wishlist?latitude=40.7128&longitude=-74.0060", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetNearestWishlistAPI_WithAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupLocationTestDB(t)
	controller := NewLocationController(db)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})
	router.GET("/api/location/wishlist", controller.GetNearestWishlistAPI)

	req, _ := http.NewRequest("GET", "/api/location/wishlist?latitude=40.7128&longitude=-74.0060", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

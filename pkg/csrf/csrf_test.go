package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with CSRF middleware
	r := gin.New()
	r.Use(Middleware("test-secret-key", false))

	// Add a test route
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Test GET request (should work without CSRF token)
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestToken(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Test that Token function doesn't panic
	// Note: Without proper CSRF middleware setup, this will return empty string
	token := Token(c)
	assert.IsType(t, "", token)
}

func TestField(t *testing.T) {
	// Test that Field returns the expected field name
	// filippo.io/csrf/gorilla maintains gorilla/csrf API compatibility
	fieldName := Field()
	assert.Equal(t, "gorilla.csrf.Token", fieldName)
}

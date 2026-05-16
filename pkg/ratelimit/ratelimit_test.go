package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig(true, 100, "1m", 10, "1m")

	assert.True(t, config.Enabled)
	assert.Equal(t, 100, config.Requests)
	assert.Equal(t, "1m", config.Period)
	assert.Equal(t, 10, config.AuthRequests)
	assert.Equal(t, "1m", config.AuthPeriod)
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback time.Duration
		expected time.Duration
	}{
		{
			name:     "valid duration",
			input:    "1m",
			fallback: time.Second,
			expected: time.Minute,
		},
		{
			name:     "invalid duration uses fallback",
			input:    "invalid",
			fallback: time.Second,
			expected: time.Second,
		},
		{
			name:     "empty duration uses fallback",
			input:    "",
			fallback: time.Hour,
			expected: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDuration(tt.input, tt.fallback)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneralRateLimitMiddleware_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Enabled:  false,
		Requests: 1,
		Period:   "1m",
	}

	middleware := GeneralRateLimitMiddleware(config)

	// Create test router
	r := gin.New()
	r.Use(middleware)
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Make multiple requests - should all succeed since rate limiting is disabled
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	}
}

func TestGeneralRateLimitMiddleware_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Enabled:  true,
		Requests: 2,
		Period:   "1m",
	}

	middleware := GeneralRateLimitMiddleware(config)

	// Create test router
	r := gin.New()
	r.Use(middleware)
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	}

	// Third request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 429, w.Code)
}

func TestAuthRateLimitMiddleware_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Enabled:      false,
		AuthRequests: 1,
		AuthPeriod:   "1m",
	}

	middleware := AuthRateLimitMiddleware(config)

	// Create test router
	r := gin.New()
	r.Use(middleware)
	r.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "login success"})
	})

	// Make multiple requests - should all succeed since rate limiting is disabled
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	}
}

func TestAuthRateLimitMiddleware_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Enabled:      true,
		AuthRequests: 1,
		AuthPeriod:   "1m",
	}

	middleware := AuthRateLimitMiddleware(config)

	// Create test router
	r := gin.New()
	r.Use(middleware)
	r.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "login success"})
	})

	// First request should succeed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Second request should be rate limited
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 429, w.Code)
}

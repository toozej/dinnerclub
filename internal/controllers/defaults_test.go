package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouterWithSession() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup session store for flash message testing
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("test-session", store))

	return router
}

func TestSetDefaults(t *testing.T) {
	router := setupTestRouterWithSession()

	// Add middleware
	router.Use(SetDefaults("pdx", "test123"))

	// Add test route to check context values
	router.GET("/test", func(c *gin.Context) {
		cityCode := c.MustGet("citycode").(string)
		referralCode := c.MustGet("referralcode").(string)
		c.JSON(http.StatusOK, gin.H{
			"citycode":     cityCode,
			"referralcode": referralCode,
		})
	})

	t.Run("sets city code and referral code in context", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"citycode":"PDX"`)
		assert.Contains(t, w.Body.String(), `"referralcode":"test123"`)
	})

	t.Run("converts city code to uppercase", func(t *testing.T) {
		// Create new router with lowercase city code
		testRouter := setupTestRouterWithSession()
		testRouter.Use(SetDefaults("nyc", "test456"))
		testRouter.GET("/test", func(c *gin.Context) {
			cityCode := c.MustGet("citycode").(string)
			c.JSON(http.StatusOK, gin.H{"citycode": cityCode})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"citycode":"NYC"`)
	})
}

func TestFlashMessage(t *testing.T) {
	router := setupTestRouterWithSession()

	router.POST("/flash", func(c *gin.Context) {
		flashMessage(c, "Test flash message")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/check-flash", func(c *gin.Context) {
		flashes := Flashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes})
	})

	t.Run("flash message is stored and retrieved", func(t *testing.T) {
		// First, set a flash message
		req, _ := http.NewRequest("POST", "/flash", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Get the session cookie from the response
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies, "Should have session cookie")

		// Now check if flash message is retrieved
		req2, _ := http.NewRequest("GET", "/check-flash", nil)
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Contains(t, w2.Body.String(), "Test flash message")
	})
}

func TestFlashes(t *testing.T) {
	router := setupTestRouterWithSession()

	router.GET("/empty-flashes", func(c *gin.Context) {
		flashes := Flashes(c)
		c.JSON(http.StatusOK, gin.H{
			"flashes": flashes,
			"count":   len(flashes),
		})
	})

	t.Run("returns empty slice when no flash messages", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/empty-flashes", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"count":0`)
		assert.Contains(t, w.Body.String(), `"flashes":[]`)
	})
}

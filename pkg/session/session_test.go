package session

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInitSession(t *testing.T) {
	// Test that InitSession returns a valid middleware function
	middleware := InitSession("test-secret-key")
	assert.NotNil(t, middleware, "InitSession should return a middleware function")

	// Test that the middleware can be used with Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)

	router.GET("/test", func(c *gin.Context) {
		// Try to use the session
		session := sessions.Default(c)
		assert.NotNil(t, session, "Session should be available")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFlashMessage_WithoutSessionMiddleware(t *testing.T) {
	// Setup Gin without session middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		// This should not panic even without session middleware
		FlashMessage(c, "test message")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test the endpoint
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not panic and return OK
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetFlashes_WithoutSessionMiddleware(t *testing.T) {
	// Setup Gin without session middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		// This should not panic even without session middleware
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes})
	})

	// Test the endpoint
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not panic and return OK
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFlashMessage_WithSessionMiddleware(t *testing.T) {
	// Setup Gin with session middleware using our InitSession function
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use our InitSession function
	router.Use(InitSession("test-secret-key"))

	router.POST("/add-flash", func(c *gin.Context) {
		FlashMessage(c, "success message")
		c.JSON(http.StatusOK, gin.H{"status": "flash added"})
	})

	router.GET("/get-flashes", func(c *gin.Context) {
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes})
	})

	// Create a cookie jar to maintain session across requests
	jar := httptest.NewRecorder()

	// Add a flash message
	req, _ := http.NewRequest("POST", "/add-flash", nil)
	router.ServeHTTP(jar, req)
	assert.Equal(t, http.StatusOK, jar.Code)

	// Extract cookies from the first response
	cookies := jar.Result().Cookies()

	// Get flash messages with the same session
	w := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/get-flashes", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response to check flashes
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	flashes, ok := response["flashes"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, flashes, 1)
	assert.Equal(t, "success message", flashes[0])
}

func TestGetFlashes_EmptyWhenNoFlashes(t *testing.T) {
	// Setup Gin with session middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add session middleware
	store := cookie.NewStore([]byte("test-secret-key"))
	router.Use(sessions.Sessions("test-session", store))

	router.GET("/get-flashes", func(c *gin.Context) {
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes, "count": len(flashes)})
	})

	// Test getting flashes when none exist
	req, _ := http.NewRequest("GET", "/get-flashes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	count, ok := response["count"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(0), count)
}

func TestFlashMessage_SingleMessage(t *testing.T) {
	// Test adding a single flash message to ensure basic functionality works
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add session middleware
	store := cookie.NewStore([]byte("test-secret-key"))
	router.Use(sessions.Sessions("test-session", store))

	router.POST("/add-flash", func(c *gin.Context) {
		FlashMessage(c, "single test message")
		c.JSON(http.StatusOK, gin.H{"status": "flash added"})
	})

	router.GET("/get-flashes", func(c *gin.Context) {
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes})
	})

	// Create a cookie jar to maintain session across requests
	jar := httptest.NewRecorder()

	// Add a flash message
	req, _ := http.NewRequest("POST", "/add-flash", nil)
	router.ServeHTTP(jar, req)
	assert.Equal(t, http.StatusOK, jar.Code)

	// Extract cookies from the first response
	cookies := jar.Result().Cookies()

	// Get flash messages with the same session
	w := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/get-flashes", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response to check flashes
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	flashes, ok := response["flashes"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, flashes, 1)
	assert.Equal(t, "single test message", flashes[0])
}

func TestFlashMessage_ConsumedAfterReading(t *testing.T) {
	// Setup Gin with session middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add session middleware
	store := cookie.NewStore([]byte("test-secret-key"))
	router.Use(sessions.Sessions("test-session", store))

	router.POST("/add-flash", func(c *gin.Context) {
		FlashMessage(c, "temporary message")
		c.JSON(http.StatusOK, gin.H{"status": "flash added"})
	})

	router.GET("/get-flashes", func(c *gin.Context) {
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes})
	})

	// Create a cookie jar to maintain session across requests
	jar := httptest.NewRecorder()

	// Add a flash message
	req, _ := http.NewRequest("POST", "/add-flash", nil)
	router.ServeHTTP(jar, req)
	assert.Equal(t, http.StatusOK, jar.Code)

	// Extract cookies from the first response
	cookies := jar.Result().Cookies()

	// Get flash messages first time
	w1 := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/get-flashes", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(w1, req)

	assert.Equal(t, http.StatusOK, w1.Code)

	// Parse first response
	var response1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.NoError(t, err)

	flashes1, ok := response1["flashes"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, flashes1, 1)
	assert.Equal(t, "temporary message", flashes1[0])

	// Extract cookies from the second response (session updated)
	cookies = w1.Result().Cookies()

	// Get flash messages second time - should be empty
	w2 := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/get-flashes", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)

	// Parse second response
	var response2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err)

	// Check if flashes field exists and handle both nil and empty slice cases
	if flashesRaw, exists := response2["flashes"]; exists {
		if flashesRaw == nil {
			// nil is equivalent to empty for our purposes
			t.Log("Flashes field is nil (equivalent to empty)")
		} else {
			flashes2, ok := flashesRaw.([]interface{})
			assert.True(t, ok, "Flashes should be an array")
			assert.Len(t, flashes2, 0, "Flashes should be empty after consumption")
		}
	} else {
		// Field doesn't exist, which is also fine for empty flashes
		t.Log("Flashes field doesn't exist (equivalent to empty)")
	}
}

func TestFlashMessage_MultipleMessagesInSeparateRequests(t *testing.T) {
	// Test adding multiple flash messages across separate requests
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add session middleware
	store := cookie.NewStore([]byte("test-secret-key"))
	router.Use(sessions.Sessions("test-session", store))

	router.POST("/add-flash/:message", func(c *gin.Context) {
		message := c.Param("message")
		FlashMessage(c, message)
		c.JSON(http.StatusOK, gin.H{"status": "flash added"})
	})

	router.GET("/get-flashes", func(c *gin.Context) {
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{"flashes": flashes})
	})

	// Start with no cookies
	var cookies []*http.Cookie

	// Add first flash message
	jar1 := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/add-flash/first", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(jar1, req)
	assert.Equal(t, http.StatusOK, jar1.Code)
	cookies = jar1.Result().Cookies()

	// Add second flash message
	jar2 := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/add-flash/second", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(jar2, req)
	assert.Equal(t, http.StatusOK, jar2.Code)
	cookies = jar2.Result().Cookies()

	// Get all flash messages
	w := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/get-flashes", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response to check flashes
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	flashes, ok := response["flashes"].([]interface{})
	assert.True(t, ok)

	// Should have both messages
	assert.Len(t, flashes, 2)
	assert.Contains(t, flashes, "first")
	assert.Contains(t, flashes, "second")
}

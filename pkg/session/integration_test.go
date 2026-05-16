package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestSessionIntegration tests the session package in a more realistic scenario
// similar to how it would be used in the actual application
func TestSessionIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up session middleware using our InitSession function
	router.Use(InitSession("test-secret"))

	// Simulate a controller that adds flash messages on POST and displays them on GET
	router.POST("/action", func(c *gin.Context) {
		// Simulate some action that results in a flash message
		action := c.PostForm("action")
		switch action {
		case "success":
			FlashMessage(c, "Action completed successfully!")
		case "error":
			FlashMessage(c, "An error occurred while processing your request.")
		case "warning":
			FlashMessage(c, "Please review your input and try again.")
		default:
			FlashMessage(c, "Unknown action performed.")
		}

		// Redirect to the display page (in real app this would be c.Redirect)
		c.JSON(http.StatusOK, gin.H{"status": "action processed"})
	})

	router.GET("/display", func(c *gin.Context) {
		// Get flash messages to display to user
		flashes := GetFlashes(c)
		c.JSON(http.StatusOK, gin.H{
			"flashes": flashes,
			"count":   len(flashes),
		})
	})

	// Test success flow
	t.Run("Success Flow", func(t *testing.T) {
		// Start session
		jar := httptest.NewRecorder()

		// Perform action
		req, _ := http.NewRequest("POST", "/action", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.PostForm = map[string][]string{"action": {"success"}}
		router.ServeHTTP(jar, req)
		assert.Equal(t, http.StatusOK, jar.Code)

		// Get cookies and display flashes
		cookies := jar.Result().Cookies()
		w := httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/display", nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Action completed successfully!")
	})

	// Test error flow
	t.Run("Error Flow", func(t *testing.T) {
		// Start session
		jar := httptest.NewRecorder()

		// Perform action that causes error
		req, _ := http.NewRequest("POST", "/action", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.PostForm = map[string][]string{"action": {"error"}}
		router.ServeHTTP(jar, req)
		assert.Equal(t, http.StatusOK, jar.Code)

		// Get cookies and display flashes
		cookies := jar.Result().Cookies()
		w := httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/display", nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "An error occurred")
	})

	// Test that flashes are consumed after display
	t.Run("Flashes Consumed After Display", func(t *testing.T) {
		// Start session
		jar := httptest.NewRecorder()

		// Perform action
		req, _ := http.NewRequest("POST", "/action", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.PostForm = map[string][]string{"action": {"warning"}}
		router.ServeHTTP(jar, req)
		assert.Equal(t, http.StatusOK, jar.Code)

		// Get cookies and display flashes first time
		cookies := jar.Result().Cookies()
		w1 := httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/display", nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		router.ServeHTTP(w1, req)

		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Contains(t, w1.Body.String(), "Please review your input")

		// Display flashes second time - should be empty
		cookies = w1.Result().Cookies()
		w2 := httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/display", nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		router.ServeHTTP(w2, req)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Contains(t, w2.Body.String(), `"count":0`)
	})
}

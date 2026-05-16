package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/helpers"
)

func setupAPITestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup test database
	db := helpers.FreshDb(t)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	// Setup API controller
	apiController := NewReferralProfileAPIController()

	// Setup routes
	api := router.Group("/api")

	referrals := api.Group("/referrals")
	referrals.POST("/generate", apiController.PostGenerateReferral)
	referrals.GET("/stats", apiController.GetReferralStats)

	profile := api.Group("/profile")
	profile.POST("/change-email", apiController.PostChangeEmail)
	profile.POST("/confirm", apiController.PostConfirm)

	return router, db
}

func TestPostGenerateReferral(t *testing.T) {
	router, db := setupAPITestRouter(t)

	// Create test user
	user := models.User{
		Username:     "testuser",
		Password:     "testpassword123",
		Firstname:    "Test",
		Lastname:     "User",
		Email:        "test@example.com",
		ReferralCode: "TESTCODE123",
	}
	require.NoError(t, db.Create(&user).Error)

	t.Run("successful referral generation with custom message", func(t *testing.T) {
		requestBody := GenerateReferralRequest{
			CustomMessage: "Check out this amazing restaurant discovery platform!",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/api/referrals/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GenerateReferralResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Message)
		assert.NotEmpty(t, response.ReferralCode)
		assert.Contains(t, response.Message, "Check out this amazing restaurant discovery platform!")
	})

	t.Run("successful referral generation with default message", func(t *testing.T) {
		requestBody := GenerateReferralRequest{}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/api/referrals/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GenerateReferralResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Message)
		assert.NotEmpty(t, response.ReferralCode)
	})

	t.Run("invalid JSON request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/referrals/generate", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["error"].(string), "Invalid request format")
	})
}

func TestGetReferralStats(t *testing.T) {
	router, db := setupAPITestRouter(t)

	// Create test user
	user := models.User{
		Username:     "testuser",
		Password:     "testpassword123",
		Firstname:    "Test",
		Lastname:     "User",
		Email:        "test@example.com",
		ReferralCode: "TESTCODE123",
	}
	require.NoError(t, db.Create(&user).Error)

	// Create some referred users
	referredUser1 := models.User{
		Username:   "referred1",
		Password:   "password123",
		Firstname:  "Referred",
		Lastname:   "User1",
		Email:      "referred1@example.com",
		ReferredBy: &user.ID,
	}
	require.NoError(t, db.Create(&referredUser1).Error)

	referredUser2 := models.User{
		Username:   "referred2",
		Password:   "password123",
		Firstname:  "Referred",
		Lastname:   "User2",
		Email:      "referred2@example.com",
		ReferredBy: &user.ID,
	}
	require.NoError(t, db.Create(&referredUser2).Error)

	t.Run("successful stats retrieval", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/referrals/stats", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ReferralStatsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, uint(1), response.UserID)
		assert.Equal(t, 2, response.TotalReferrals)
		assert.Equal(t, 2, response.ActiveReferrals)
	})
}

func TestPostChangeEmail(t *testing.T) {
	router, db := setupAPITestRouter(t)

	// Create test user
	user := models.User{
		Username:  "testuser",
		Password:  "testpassword123",
		Firstname: "Test",
		Lastname:  "User",
		Email:     "test@example.com",
	}
	require.NoError(t, db.Create(&user).Error)

	t.Run("successful email change request", func(t *testing.T) {
		requestBody := ChangeEmailRequest{
			NewEmail: "newemail@example.com",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/api/profile/change-email", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ChangeEmailResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Token)
		assert.Contains(t, response.Message, "confirmation sent")

		// Verify profile change record was created
		var profileChange models.ProfileChange
		err = db.Where("user_id = ? AND change_type = ?", user.ID, "email").First(&profileChange).Error
		require.NoError(t, err)
		assert.Equal(t, "newemail@example.com", profileChange.NewValue)
		assert.Equal(t, "test@example.com", profileChange.OldValue)
	})

	t.Run("invalid email format", func(t *testing.T) {
		requestBody := ChangeEmailRequest{
			NewEmail: "invalid-email",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/api/profile/change-email", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
	})

	t.Run("email already in use", func(t *testing.T) {
		// Create another user with the email we want to change to
		existingUser := models.User{
			Username:  "existing",
			Password:  "password123",
			Firstname: "Existing",
			Lastname:  "User",
			Email:     "existing@example.com",
		}
		require.NoError(t, db.Create(&existingUser).Error)

		requestBody := ChangeEmailRequest{
			NewEmail: "existing@example.com",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/api/profile/change-email", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["error"].(string), "already in use")
	})
}

func TestPostConfirm(t *testing.T) {
	router, db := setupAPITestRouter(t)

	// Create test user
	user := models.User{
		Username:  "testuser",
		Password:  "testpassword123",
		Firstname: "Test",
		Lastname:  "User",
		Email:     "test@example.com",
	}
	require.NoError(t, db.Create(&user).Error)

	t.Run("successful confirmation", func(t *testing.T) {
		// First create a profile change request
		requestBody := ChangeEmailRequest{
			NewEmail: "newemail@example.com",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/api/profile/change-email", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var changeResponse ChangeEmailResponse
		err := json.Unmarshal(w.Body.Bytes(), &changeResponse)
		require.NoError(t, err)

		// Now confirm the change
		confirmBody := ConfirmRequest{
			Token: changeResponse.Token,
		}
		confirmJSON, _ := json.Marshal(confirmBody)

		confirmReq, _ := http.NewRequest("POST", "/api/profile/confirm", bytes.NewBuffer(confirmJSON))
		confirmReq.Header.Set("Content-Type", "application/json")

		confirmW := httptest.NewRecorder()
		router.ServeHTTP(confirmW, confirmReq)

		assert.Equal(t, http.StatusOK, confirmW.Code)

		var confirmResponse ConfirmResponse
		err = json.Unmarshal(confirmW.Body.Bytes(), &confirmResponse)
		require.NoError(t, err)

		assert.True(t, confirmResponse.Success)
		assert.Contains(t, confirmResponse.Message, "confirmed successfully")

		// Verify the user's email was actually updated
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "newemail@example.com", updatedUser.Email)
	})

	t.Run("invalid token", func(t *testing.T) {
		confirmBody := ConfirmRequest{
			Token: "invalid-token",
		}
		confirmJSON, _ := json.Marshal(confirmBody)

		req, _ := http.NewRequest("POST", "/api/profile/confirm", bytes.NewBuffer(confirmJSON))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
	})

	t.Run("missing token", func(t *testing.T) {
		confirmBody := ConfirmRequest{}
		confirmJSON, _ := json.Marshal(confirmBody)

		req, _ := http.NewRequest("POST", "/api/profile/confirm", bytes.NewBuffer(confirmJSON))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["error"].(string), "Invalid request format")
	})
}

func TestAPIControllerErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup API controller
	apiController := NewReferralProfileAPIController()

	// Setup routes without authentication middleware to test auth errors
	api := router.Group("/api")

	referrals := api.Group("/referrals")
	referrals.POST("/generate", apiController.PostGenerateReferral)
	referrals.GET("/stats", apiController.GetReferralStats)

	profile := api.Group("/profile")
	profile.POST("/change-email", apiController.PostChangeEmail)

	t.Run("unauthenticated requests", func(t *testing.T) {
		endpoints := []struct {
			method string
			path   string
			body   string
		}{
			{"POST", "/api/referrals/generate", `{}`},
			{"GET", "/api/referrals/stats", ""},
			{"POST", "/api/profile/change-email", `{"new_email": "test@example.com"}`},
		}

		for _, endpoint := range endpoints {
			var req *http.Request
			if endpoint.body != "" {
				req, _ = http.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer([]byte(endpoint.body)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(endpoint.method, endpoint.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Endpoint: %s %s", endpoint.method, endpoint.path)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["error"].(string), "not authenticated")
		}
	})
}

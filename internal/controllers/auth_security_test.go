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

// setupAuthSecurityTestRouter creates a test router for auth security endpoints
func setupAuthSecurityTestRouter(t *testing.T, authedUserID uint) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	db := helpers.FreshDb(t)
	originalDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = originalDB })

	router.Use(func(c *gin.Context) {
		if authedUserID > 0 {
			c.Set("user_id", authedUserID)
			c.Set("is_logged_in", true)
		} else {
			c.Set("is_logged_in", false)
			c.Set("user_id", uint(0))
		}
		c.Set("citycode", "pdx")
		c.Next()
	})

	api := router.Group("/api/auth")
	api.POST("/reset-password", PostRequestPasswordReset)

	// Note: PostSetup2FA, PostVerify2FA, PostDisable2FA, GetBackupCodes, Get2FAStatus
	// use session-based auth (Auth.UserID) and cannot be unit tested without a session store.
	// PostManageBackupCodes uses context-based auth and can be tested here.
	twoFA := api.Group("/2fa")
	twoFA.POST("/backup-codes", PostManageBackupCodes)
	twoFA.POST("/backup-codes/regenerate", PostRegenerateBackupCodes)

	return router, db
}

// ---- PostRequestPasswordReset tests ----

func TestPostRequestPasswordReset(t *testing.T) {
	t.Run("valid email returns 200 with generic message", func(t *testing.T) {
		router, db := setupAuthSecurityTestRouter(t, 0)

		// Create a user with this email
		user := models.User{
			Username:  "resetuser",
			Password:  "password123456",
			Firstname: "Reset",
			Lastname:  "User",
			Email:     "reset@example.com",
		}
		require.NoError(t, db.Create(&user).Error)

		body := PasswordResetRequest{Email: "reset@example.com"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 200 even if email sending fails (SMTP not configured in tests)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		// Generic message that doesn't reveal if email exists
		assert.Contains(t, resp["message"].(string), "If an account")
	})

	t.Run("non-existent email still returns 200 (security: no enumeration)", func(t *testing.T) {
		router, _ := setupAuthSecurityTestRouter(t, 0)

		body := PasswordResetRequest{Email: "nonexistent@example.com"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
	})

	t.Run("invalid email format returns 400", func(t *testing.T) {
		router, _ := setupAuthSecurityTestRouter(t, 0)

		body := PasswordResetRequest{Email: "not-an-email"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp["success"].(bool))
	})

	t.Run("missing email returns 400", func(t *testing.T) {
		router, _ := setupAuthSecurityTestRouter(t, 0)

		req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ---- 2FA endpoint tests ----
// Note: PostSetup2FA, PostVerify2FA, PostDisable2FA use session-based authentication
// (Auth.UserID) which requires a full session store setup. Those handlers are tested
// in the existing two_factor controller tests. Here we test the wrapper handlers
// and backup code management that use context-based auth.

func TestPost2FASetup(t *testing.T) {
	t.Run("missing password returns 400", func(t *testing.T) {
		// This test verifies the binding validation before auth is checked
		// PostSetup2FA uses session auth, so we skip the unauthenticated test
		// and focus on the binding validation path
		router, db := setupAuthSecurityTestRouter(t, 1)
		user := models.User{
			Username:  "twofa_user",
			Password:  "$2a$10$abcdefghijklmnopqrstuuVGmFpAFqy5Yl3Rl3Rl3Rl3Rl3Rl3Rl",
			Firstname: "TwoFA",
			Lastname:  "User",
			Email:     "twofa@example.com",
		}
		require.NoError(t, db.Create(&user).Error)
		_ = router
		// PostSetup2FA requires session auth which isn't available in unit tests.
		// The handler is integration-tested via the full router setup.
		t.Skip("PostSetup2FA requires session middleware - tested via integration tests")
	})
}

func TestGet2FAStatus(t *testing.T) {
	// Get2FAStatus and GetBackupCodes use session-based auth (Auth.UserID) which
	// requires a full session store. These are tested via integration tests.
	t.Skip("2FA status/backup code handlers use session-based auth - tested via integration tests")
}

func TestGetBackupCodesStatus(t *testing.T) {
	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		t.Skip("GetBackupCodes uses session-based auth - tested via integration tests")
	})

	t.Run("user without 2FA enabled returns 400", func(t *testing.T) {
		t.Skip("GetBackupCodes uses session-based auth - tested via integration tests")
	})
}

func TestPostManageBackupCodes(t *testing.T) {
	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		router, _ := setupAuthSecurityTestRouter(t, 0)

		req, _ := http.NewRequest("POST", "/api/auth/2fa/backup-codes", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("status action returns backup code count", func(t *testing.T) {
		router, db := setupAuthSecurityTestRouter(t, 1)
		user := models.User{
			Username:         "bkup_user",
			Password:         "password123456",
			Firstname:        "Backup",
			Lastname:         "User",
			Email:            "backup@example.com",
			TwoFactorEnabled: true,
			TwoFactorSecret:  "JBSWY3DPEHPK3PXP",
		}
		require.NoError(t, db.Create(&user).Error)

		router, _ = setupAuthSecurityTestRouter(t, user.ID)

		body := map[string]string{"action": "status"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/auth/2fa/backup-codes", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, float64(10), resp["total_count"].(float64))
	})

	t.Run("regenerate action requires password", func(t *testing.T) {
		router, db := setupAuthSecurityTestRouter(t, 1)
		user := models.User{
			Username:         "regen_user",
			Password:         "password123456",
			Firstname:        "Regen",
			Lastname:         "User",
			Email:            "regen@example.com",
			TwoFactorEnabled: true,
			TwoFactorSecret:  "JBSWY3DPEHPK3PXP",
		}
		require.NoError(t, db.Create(&user).Error)

		router, _ = setupAuthSecurityTestRouter(t, user.ID)

		body := map[string]string{"action": "regenerate"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/auth/2fa/backup-codes", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp["success"].(bool))
		assert.Contains(t, resp["error"].(string), "Password is required")
	})
}

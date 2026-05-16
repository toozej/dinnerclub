package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// setupSocialLocationTestRouter creates a test router with the social/location controller
// and an optional authenticated user ID.
func setupSocialLocationTestRouter(t *testing.T, authedUserID uint) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	db := helpers.FreshDb(t)
	originalDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = originalDB })

	// Inject auth middleware
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

	ctrl := NewSocialLocationController()

	api := router.Group("/api")
	api.POST("/social/follow/:userID", ctrl.PostFollowUserByID)
	api.GET("/social/feed", ctrl.GetSocialFeed)
	api.GET("/location/nearby", ctrl.GetNearbyRestaurants)
	api.GET("/users/id/:userID/profile", ctrl.GetUserProfileByID)

	return router, db
}

// createSocialTestUser is a helper to create a user in the test DB for social/location tests
func createSocialTestUser(t *testing.T, db *gorm.DB, username, email string) models.User {
	t.Helper()
	user := models.User{
		Username:  username,
		Password:  "password123456",
		Firstname: "Test",
		Lastname:  "User",
		Email:     email,
	}
	require.NoError(t, db.Create(&user).Error)
	return user
}

// ---- PostFollowUserByID tests ----

func TestPostFollowUserByID(t *testing.T) {
	t.Run("follow a user successfully", func(t *testing.T) {
		_, db := setupSocialLocationTestRouter(t, 1)
		follower := createSocialTestUser(t, db, "follower", "follower@example.com")
		target := createSocialTestUser(t, db, "target", "target@example.com")

		// Re-setup with correct follower ID so the DB is shared
		router, db2 := setupSocialLocationTestRouter(t, follower.ID)
		follower2 := createSocialTestUser(t, db2, "follower_b", "follower_b@example.com")
		target2 := createSocialTestUser(t, db2, "target_b", "target_b@example.com")
		_ = follower
		_ = target

		body := FollowRequest{Action: "follow"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/social/follow/%d", target2.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.True(t, resp["following"].(bool))
		_ = follower2
	})

	t.Run("cannot follow yourself", func(t *testing.T) {
		_, db := setupSocialLocationTestRouter(t, 1)
		user := createSocialTestUser(t, db, "selfuser_seed", "self_seed@example.com")

		router, db2 := setupSocialLocationTestRouter(t, user.ID)
		selfUser := createSocialTestUser(t, db2, "selfuser", "self@example.com")

		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/social/follow/%d", selfUser.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp["success"].(bool))
		assert.Contains(t, resp["error"].(string), "Cannot follow yourself")
	})

	t.Run("unauthenticated follow request returns 401", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 0)
		target := createSocialTestUser(t, db, "target2", "target2@example.com")

		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/social/follow/%d", target.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid user ID returns 400", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 1)

		req, _ := http.NewRequest("POST", "/api/social/follow/notanumber", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unfollow a user successfully", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 1)
		follower := createSocialTestUser(t, db, "follower3", "follower3@example.com")
		target := createSocialTestUser(t, db, "target3", "target3@example.com")

		// Create follow relationship directly
		follow := models.UserFollow{FollowerID: follower.ID, FollowingID: target.ID}
		require.NoError(t, db.Create(&follow).Error)

		// Re-setup with follower's ID
		router2, db2 := setupSocialLocationTestRouter(t, follower.ID)
		follower2 := createSocialTestUser(t, db2, "follower3b", "follower3b@example.com")
		target2 := createSocialTestUser(t, db2, "target3b", "target3b@example.com")
		follow2 := models.UserFollow{FollowerID: follower2.ID, FollowingID: target2.ID}
		require.NoError(t, db2.Create(&follow2).Error)
		_ = router

		body := FollowRequest{Action: "unfollow"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/social/follow/%d", target2.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.False(t, resp["following"].(bool))
	})
}

// ---- GetSocialFeed tests ----

func TestGetSocialFeed(t *testing.T) {
	t.Run("returns empty feed for user with no follows", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 1)
		user := createSocialTestUser(t, db, "feeduser", "feeduser@example.com")
		_ = user

		req, _ := http.NewRequest("GET", "/api/social/feed", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, float64(0), resp["count"].(float64))
	})

	t.Run("unauthenticated feed request returns 401", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/social/feed", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("pagination parameters are respected", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 1)
		user := createSocialTestUser(t, db, "feeduser2", "feeduser2@example.com")
		_ = user

		req, _ := http.NewRequest("GET", "/api/social/feed?page=1&limit=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, float64(1), resp["page"].(float64))
		assert.Equal(t, float64(5), resp["limit"].(float64))
	})
}

// ---- GetNearbyRestaurants tests ----

func TestGetNearbyRestaurants(t *testing.T) {
	t.Run("returns results for valid coordinates", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=45.5231&longitude=-122.6765&radius=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		// results may be empty slice or nil when no restaurants exist in test DB
		assert.Equal(t, float64(0), resp["count"].(float64))
		assert.Equal(t, "nearby", resp["search_type"].(string))
	})

	t.Run("missing coordinates returns 400", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp["success"].(bool))
	})

	t.Run("invalid latitude returns 400", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=999&longitude=-122.6765", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid longitude returns 400", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=45.5231&longitude=999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid radius returns 400", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=45.5231&longitude=-122.6765&radius=999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wishlist search requires authentication", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=45.5231&longitude=-122.6765&type=wishlist", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("unvisited search requires authentication", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=45.5231&longitude=-122.6765&type=unvisited", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("wishlist search succeeds when authenticated", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 1)
		user := createSocialTestUser(t, db, "locuser", "locuser@example.com")
		_ = user

		req, _ := http.NewRequest("GET", "/api/location/nearby?latitude=45.5231&longitude=-122.6765&type=wishlist", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "wishlist", resp["search_type"].(string))
	})
}

// ---- GetUserProfileByID tests ----

func TestGetUserProfileByID(t *testing.T) {
	t.Run("returns profile for valid user ID", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 0)
		user := createSocialTestUser(t, db, "profileuser", "profileuser@example.com")

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/users/id/%d/profile", user.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		assert.NotNil(t, resp["profile"])
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/users/id/99999/profile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp["success"].(bool))
	})

	t.Run("returns 400 for invalid user ID", func(t *testing.T) {
		router, _ := setupSocialLocationTestRouter(t, 0)

		req, _ := http.NewRequest("GET", "/api/users/id/notanumber/profile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("includes follow status when viewer is authenticated", func(t *testing.T) {
		router, db := setupSocialLocationTestRouter(t, 1)
		viewer := createSocialTestUser(t, db, "viewer", "viewer@example.com")
		target := createSocialTestUser(t, db, "viewtarget", "viewtarget@example.com")
		_ = viewer

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/users/id/%d/profile", target.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
		profile := resp["profile"].(map[string]interface{})
		// is_following should be present and false (not yet following)
		assert.False(t, profile["is_following"].(bool))
	})
}

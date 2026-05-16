// Package controllers provides HTTP request handlers for the dinnerclub web application.
// This file contains API endpoints for social features (follow/feed) and
// location-based restaurant discovery.
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
)

// SocialLocationController handles social and location-based API endpoints
type SocialLocationController struct {
	socialService   *services.SocialService
	locationService *services.LocationService
}

// NewSocialLocationController creates a new SocialLocationController instance
func NewSocialLocationController() *SocialLocationController {
	return &SocialLocationController{
		socialService:   services.NewSocialService(database.DB),
		locationService: services.NewLocationService(database.DB),
	}
}

// FollowRequest represents the request body for follow/unfollow actions
type FollowRequest struct {
	Action string `json:"action"` // "follow" or "unfollow"; defaults to toggle
}

// PostFollowUserByID handles POST /api/social/follow/:userID
// Follows or unfollows a user by their numeric ID.
func (sc *SocialLocationController) PostFollowUserByID(c *gin.Context) {
	if !authentication.IsLoggedIn(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"success": false,
		})
		return
	}

	followerID, exists := authentication.GetUserID(c)
	if !exists || followerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"success": false,
		})
		return
	}

	// Parse target user ID from URL parameter
	targetIDStr := c.Param("userID")
	targetIDUint64, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}
	targetID := uint(targetIDUint64)

	// Prevent self-following
	if followerID == targetID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot follow yourself",
			"success": false,
		})
		return
	}

	// Parse optional action from body
	var req FollowRequest
	_ = c.ShouldBindJSON(&req)

	// Determine current follow state to support toggle
	isFollowing, err := sc.socialService.IsFollowing(followerID, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check follow status",
			"success": false,
		})
		return
	}

	// Determine action: explicit or toggle
	action := req.Action
	if action == "" {
		if isFollowing {
			action = "unfollow"
		} else {
			action = "follow"
		}
	}

	switch action {
	case "follow":
		if isFollowing {
			c.JSON(http.StatusConflict, gin.H{
				"error":     "Already following this user",
				"following": true,
				"success":   false,
			})
			return
		}
		if err := sc.socialService.FollowUser(followerID, targetID); err != nil {
			log.Warnf("Failed to follow user %d -> %d: %v", followerID, targetID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to follow user",
				"success": false,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Successfully followed user",
			"following": true,
			"success":   true,
		})

	case "unfollow":
		if !isFollowing {
			c.JSON(http.StatusConflict, gin.H{
				"error":     "Not following this user",
				"following": false,
				"success":   false,
			})
			return
		}
		if err := sc.socialService.UnfollowUser(followerID, targetID); err != nil {
			log.Warnf("Failed to unfollow user %d -> %d: %v", followerID, targetID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to unfollow user",
				"success": false,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Successfully unfollowed user",
			"following": false,
			"success":   true,
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid action. Use 'follow' or 'unfollow'",
			"success": false,
		})
	}
}

// GetSocialFeed handles GET /api/social/feed
// Returns a paginated feed of entries from users the authenticated user follows.
func (sc *SocialLocationController) GetSocialFeed(c *gin.Context) {
	if !authentication.IsLoggedIn(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"success": false,
		})
		return
	}

	userID, exists := authentication.GetUserID(c)
	if !exists || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"success": false,
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get feed with author details
	feedEntries, err := sc.socialService.GetFollowingFeedWithDetails(userID, limit, offset)
	if err != nil {
		log.Warnf("Failed to get social feed for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve social feed",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feed":     feedEntries,
		"count":    len(feedEntries),
		"page":     page,
		"limit":    limit,
		"has_more": len(feedEntries) == limit,
		"success":  true,
	})
}

// GetNearbyRestaurants handles GET /api/location/nearby
// Returns restaurants near the provided coordinates within the given radius.
// Query params: latitude, longitude, radius (km, default 10), type (nearby|wishlist|unvisited)
func (sc *SocialLocationController) GetNearbyRestaurants(c *gin.Context) {
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "10")
	searchType := c.DefaultQuery("type", "nearby")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "latitude and longitude query parameters are required",
			"success": false,
		})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid latitude. Must be a number between -90 and 90.",
			"success": false,
		})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil || lon < -180 || lon > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid longitude. Must be a number between -180 and 180.",
			"success": false,
		})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 || radius > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid radius. Must be a number between 1 and 100 km.",
			"success": false,
		})
		return
	}

	userLocation := services.Location{
		Latitude:  lat,
		Longitude: lon,
	}

	var results []services.RestaurantDistance

	switch searchType {
	case "wishlist":
		userID, exists := authentication.GetUserID(c)
		if !exists || userID == 0 || !authentication.IsLoggedIn(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required for wishlist search",
				"success": false,
			})
			return
		}
		results, err = sc.locationService.FindNearestWishlist(userID, userLocation)

	case "unvisited":
		userID, exists := authentication.GetUserID(c)
		if !exists || userID == 0 || !authentication.IsLoggedIn(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required for unvisited search",
				"success": false,
			})
			return
		}
		results, err = sc.locationService.FindNearestUnvisited(userID, userLocation)

	default: // "nearby"
		results, err = sc.locationService.FindNearbyRestaurants(userLocation, radius)
	}

	if err != nil {
		log.Warnf("Location search failed (type=%s, lat=%f, lon=%f): %v", searchType, lat, lon, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Location search failed. Please try again.",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"count":       len(results),
		"latitude":    lat,
		"longitude":   lon,
		"radius_km":   radius,
		"search_type": searchType,
		"success":     true,
	})
}

// GetUserProfileByID handles GET /api/users/:userID/profile
// Returns the enhanced profile for a user identified by numeric ID.
func (sc *SocialLocationController) GetUserProfileByID(c *gin.Context) {
	targetIDStr := c.Param("userID")
	targetIDUint64, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}
	targetID := uint(targetIDUint64)

	// Get optional viewer ID for follow status and privacy checks
	viewerID, _ := authentication.GetUserID(c)

	profile, err := sc.socialService.GetUserProfile(targetID, viewerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
		"success": true,
	})
}

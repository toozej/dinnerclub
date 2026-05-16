// Package routers provides router setup and configuration for the dinnerclub web application.
// It handles middleware setup, static assets, templates, and route configuration.
package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
	"github.com/toozej/dinnerclub/pkg/ratelimit"
)

// SetupPublicRoutes configures all public routes that don't require authentication.
// This includes entry viewing, restaurant browsing, search functionality, and authentication endpoints.
func SetupPublicRoutes(rateLimitConfig *ratelimit.Config) {
	r := ResolveRouter()

	// primary routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/entries/")
	})

	// entries related routes
	entries := r.Group("/entries")
	entries.GET("/", controllers.GetEntries)
	entries.GET("/:id", controllers.GetEntry)

	// restaurants related routes
	restaurants := r.Group("/restaurants")
	restaurants.GET("/", controllers.GetRestaurants)
	restaurants.GET("/name/:name", controllers.GetRestaurantByName)
	restaurants.GET("/:id", controllers.GetRestaurant)

	// search related routes
	searchController := controllers.NewSearchController()
	search := r.Group("/search")
	search.GET("/", searchController.GetSearch)
	search.GET("/entries", searchController.GetSearchEntries)
	search.GET("/restaurants", searchController.GetSearchRestaurants)
	search.GET("/suggestions", searchController.GetSuggestions)

	// API suggestions routes for auto-complete functionality
	api := r.Group("/api")
	suggestions := api.Group("/suggestions")
	suggestions.GET("/cuisine", controllers.GetCuisineSuggestions)
	suggestions.GET("/restaurant-names", controllers.GetRestaurantNameSuggestions)
	suggestions.GET("/meal-service", controllers.GetMealServiceSuggestions)
	suggestions.GET("/dishes", controllers.GetDishSuggestions)
	suggestions.GET("/locations", controllers.GetLocationSuggestions)

	// user pre-authenticated authentication related routes
	preAuth := r.Group("/auth")
	preAuth.Use(controllers.EnsureNotLoggedIn())
	preAuth.Use(ratelimit.AuthRateLimitMiddleware(rateLimitConfig))
	preAuth.GET("/register", controllers.GetRegister)
	preAuth.POST("/register", controllers.PostRegister)
	preAuth.GET("/login", controllers.GetLogin)
	preAuth.POST("/login", controllers.PostLogin)

	// token refresh endpoint (doesn't require EnsureNotLoggedIn since it's for existing tokens)
	// Apply auth rate limiting to refresh endpoint as well
	authRefresh := r.Group("/auth")
	authRefresh.Use(ratelimit.AuthRateLimitMiddleware(rateLimitConfig))
	authRefresh.POST("/refresh", controllers.PostRefreshToken)

	// user profile routes (public)
	profileController := controllers.NewProfileDisplayController()
	users := r.Group("/users")
	users.GET("/:username/profile", profileController.GetUserProfile)
	users.GET("/:username/entries", profileController.GetUserEntries)
	users.GET("/:id", controllers.GetUserByID)
	users.GET("/username/:username", controllers.GetUserByUsername)

	// API routes for profile statistics and privacy
	api.GET("/users/:username/stats", profileController.GetProfileStatistics)
	api.GET("/users/:username/access", profileController.CheckProfileAccess)
	api.GET("/users/visible", profileController.GetVisibleUsers)

	// Password reset API endpoint (public - no auth required, rate limited)
	authAPI := r.Group("/api/auth")
	authAPI.Use(ratelimit.AuthRateLimitMiddleware(rateLimitConfig))
	authAPI.POST("/reset-password", controllers.PostRequestPasswordReset)

	// Password reset page (HTML, public)
	preAuth.GET("/reset-password", controllers.GetPasswordResetPage)

	// Public user profile API endpoint (by numeric ID)
	socialLocationController := controllers.NewSocialLocationController()
	api.GET("/users/id/:userID/profile", socialLocationController.GetUserProfileByID)

	// Public location nearby endpoint (auth optional - wishlist/unvisited require auth)
	api.GET("/location/nearby", socialLocationController.GetNearbyRestaurants)

	// general healthcheck route
	r.GET("/health", controllers.GetHealth)
}

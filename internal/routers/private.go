// Package routers provides router setup and configuration for the dinnerclub web application.
// It handles middleware setup, static assets, templates, and route configuration.
package routers

import (
	"github.com/toozej/dinnerclub/internal/controllers"
	"github.com/toozej/dinnerclub/pkg/ratelimit"
)

// SetupPrivateRoutes configures all authenticated routes that require user login.
// These routes include profile management, entry creation/editing, and user-specific operations.
func SetupPrivateRoutes() {
	SetupPrivateRoutesWithRateLimit(nil)
}

// SetupPrivateRoutesWithRateLimit configures all authenticated routes with optional rate limiting.
func SetupPrivateRoutesWithRateLimit(rateLimitConfig *ratelimit.Config) {
	r := ResolveRouter()

	profile := r.Group("/profile")
	profile.Use(controllers.EnsureLoggedIn())
	profile.GET("/", controllers.GetProfile)
	profile.POST("/update", controllers.PostProfileUpdate)
	profile.POST("/delete", controllers.PostProfileDelete)

	// Enhanced profile functionality
	profileController := controllers.NewProfileDisplayController()
	profile.GET("/enhanced", profileController.GetEnhancedProfile)
	profile.POST("/privacy", profileController.PostUpdatePrivacy)
	profile.POST("/update-bio", profileController.PostUpdateBio)

	// Email change pages (authenticated)
	profile.GET("/change-email", controllers.GetEmailChangePage)
	profile.GET("/confirm-email", controllers.GetEmailConfirmPage)

	// Privacy API endpoints
	profileAPI := r.Group("/api/profile")
	profileAPI.Use(controllers.EnsureLoggedIn())
	profileAPI.GET("/privacy", profileController.GetPrivacySettings)

	// Referral and Profile Management API endpoints
	apiController := controllers.NewReferralProfileAPIController()

	// Referral API endpoints
	referralAPI := r.Group("/api/referrals")
	referralAPI.Use(controllers.EnsureLoggedIn())
	referralAPI.POST("/generate", apiController.PostGenerateReferral)
	referralAPI.GET("/stats", apiController.GetReferralStats)

	// Profile Management API endpoints
	profileAPI.POST("/change-email", apiController.PostChangeEmail)
	profileAPI.POST("/confirm", apiController.PostConfirm)

	// Social actions (HTML form-based follow/unfollow)
	socialActions := r.Group("/users")
	socialActions.Use(controllers.EnsureLoggedIn())
	socialActions.POST("/:username/follow", profileController.PostFollowUser)

	// Social and Location API endpoints
	socialLocationController := controllers.NewSocialLocationController()

	socialAPI := r.Group("/api/social")
	socialAPI.Use(controllers.EnsureLoggedIn())
	socialAPI.POST("/follow/:userID", socialLocationController.PostFollowUserByID)
	socialAPI.GET("/feed", socialLocationController.GetSocialFeed)

	// Note: /api/location/nearby is registered publicly in public.go (auth is optional per search type)

	// 2FA API endpoints (require authentication)
	twoFactorAPI := r.Group("/api/auth/2fa")
	twoFactorAPI.Use(controllers.EnsureLoggedIn())
	twoFactorAPI.POST("/setup", controllers.PostSetup2FA)
	twoFactorAPI.POST("/verify", controllers.PostVerify2FA)
	twoFactorAPI.POST("/backup-codes", controllers.PostManageBackupCodes)
	twoFactorAPI.GET("/backup-codes", controllers.GetBackupCodes)
	twoFactorAPI.POST("/backup-codes/regenerate", controllers.PostRegenerateBackupCodes)
	twoFactorAPI.POST("/disable", controllers.PostDisable2FA)
	twoFactorAPI.GET("/status", controllers.Get2FAStatus)

	// 2FA setup page (HTML)
	auth2FA := r.Group("/auth/2fa")
	auth2FA.Use(controllers.EnsureLoggedIn())
	auth2FA.GET("", controllers.Get2FASetupPage)

	// Social feed and notifications pages (HTML)
	social := r.Group("/social")
	social.Use(controllers.EnsureLoggedIn())
	social.GET("/feed", controllers.GetSocialFeedPage)
	social.GET("/notifications", controllers.GetNotificationsPage)

	// Mark all notifications read
	socialNotifAPI := r.Group("/api/social/notifications")
	socialNotifAPI.Use(controllers.EnsureLoggedIn())
	socialNotifAPI.POST("/mark-all-read", controllers.PostMarkAllNotificationsRead)

	postAuth := r.Group("/")
	postAuth.Use(controllers.EnsureLoggedIn())
	postAuth.POST("/auth/logout", controllers.PostLogout)
	postAuth.GET("/entries/new", controllers.GetEntryForm)
	postAuth.POST("/entries/new", controllers.PostEntry)
	postAuth.GET("/entries/:id/update", controllers.GetEntryUpdateForm)
	postAuth.PUT("/entries/:id", controllers.PutEntry)
	postAuth.POST("/entries/:id/update", controllers.PutEntry) // Backward compatibility for HTML forms
	postAuth.POST("/entries/:id/delete", controllers.DeleteEntry)
	postAuth.GET("/entries/submittedby/:username", controllers.GetEntriesByUsername)
	postAuth.POST("/entries/submittedby/:username", controllers.GetEntriesByUsername)
	postAuth.GET("/status", controllers.GetStatus)
}

// Package controllers provides HTTP request handlers and middleware for the dinnerclub web application.
// It includes authentication, entry management, restaurant operations, search functionality,
// and consistent response handling patterns.
package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/internal/validation"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginCreds represents the login form data structure with validation tags
// for username and password fields.
type LoginCreds struct {
	Username   string `form:"username" json:"username" binding:"required,alphanum,lowercase"`
	Password   string `form:"password" json:"password" binding:"required,min=10"`
	TOTPCode   string `form:"totp_code" json:"totp_code"`     // Optional 2FA code
	BackupCode string `form:"backup_code" json:"backup_code"` // Optional backup code
}

// GetLogin handles GET /auth/login requests and renders the login form template
// with city code and flash messages.
func GetLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/login.html", gin.H{"citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// PostLogin handles POST /auth/login requests, validates credentials, generates JWT tokens,
// and manages user sessions with appropriate response handling.
func PostLogin(c *gin.Context) {
	Auth := authentication.Resolve()

	// validate and bind user input
	var loginData LoginCreds
	if err := c.ShouldBind(&loginData); err != nil {
		Response.BadRequest(c, "Invalid form data", err)
		return
	}

	// Validate login credentials using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateLoginCredentials(loginData.Username, loginData.Password)
	if !validationResult.IsValid {
		var validationErrors []ValidationErrorResponse
		for _, err := range validationResult.Errors {
			validationErrors = append(validationErrors, ValidationErrorResponse{
				Field:   err.Field,
				Message: err.Message,
				Value:   err.Value,
			})
		}
		Response.ValidationError(c, "Invalid login credentials", validationErrors)
		return
	}

	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// get the user record by username from db
	var user models.User
	result := database.WithContext(ctx).Where("username = ?", loginData.Username).First(&user)
	// check if the record not found
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		Response.Unauthorized(c, "The username or password you entered is incorrect")
		return
	}

	// handle database error incase there is any
	if result.Error != nil {
		Response.DatabaseError(c, "find", result.Error)
		return
	}

	// compare the password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		Response.AuthenticationError(c, "invalid_password", err)
		return
	}

	// Check if 2FA is enabled for this user
	if user.TwoFactorEnabled {
		// Initialize TOTP service
		totpService := services.NewTOTPService(database.DB, c.MustGet("citycode").(string)+"DinnerClub")

		// User must provide either TOTP code or backup code
		if loginData.TOTPCode == "" && loginData.BackupCode == "" {
			// Return a specific response indicating 2FA is required
			if Response.IsHTMLRequest(c) {
				Response.RedirectWithMessage(c, "/auth/login?require_2fa=true", "Two-factor authentication required")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":       "Two-factor authentication required",
					"code":        "2FA_REQUIRED",
					"require_2fa": true,
					"user_id":     user.ID,
				})
			}
			return
		}

		// Validate 2FA code
		var validAuth bool
		if loginData.TOTPCode != "" {
			// Validate TOTP code
			validAuth = totpService.ValidateToken(user.TwoFactorSecret, loginData.TOTPCode)
			if !validAuth {
				Response.AuthenticationError(c, "invalid_2fa_code", fmt.Errorf("invalid 2FA code"))
				return
			}
		} else if loginData.BackupCode != "" {
			// Validate backup code
			valid, err := totpService.ValidateBackupCode(user.ID, loginData.BackupCode)
			if err != nil {
				Response.InternalError(c, "Backup code validation failed", err)
				return
			}
			if !valid {
				Response.AuthenticationError(c, "invalid_backup_code", fmt.Errorf("invalid backup code"))
				return
			}
			validAuth = true

			// Check if backup codes are running low
			unusedCount, err := totpService.GetUnusedBackupCodeCount(user.ID)
			if err == nil && unusedCount <= 2 {
				log.Warnf("User %s has only %d backup codes remaining", user.Username, unusedCount)
				// Could send a notification here
			}
		}

		if !validAuth {
			Response.AuthenticationError(c, "2fa_validation_failed", fmt.Errorf("2FA validation failed"))
			return
		}
	}

	// generate the jwt token
	token, err := Auth.AuthService.GenerateToken(user.ID, user.Username)
	if err != nil {
		Response.AuthenticationError(c, "token_generation_failed", err)
		return
	}

	// login the user (store in session for backward compatibility)
	err = Auth.Login(user.ID, c)
	if err != nil {
		Response.InternalError(c, "Session creation failed", err)
		return
	}

	// set user as logged in via Gin context
	c.Set("is_logged_in", true)
	c.Set("user_id", user.ID)
	c.Set("username", user.Username)

	// set the JWT token in a secure HTTP-only cookie for browser-based authentication
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("jwt_token", token, 24*60*60, "/", "", false, true) // 24 hours, HTTP-only

	// also set the token in Authorization header for API usage
	c.Header("Authorization", "Bearer "+token)

	log.Debugf("User %s successfully logged in", user.Username)

	// Preload user entries in background for better performance
	go preloadUserEntries(user.Username)

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/entries/", fmt.Sprintf("User '%s' logged in successfully.", user.Username))
	} else {
		Response.Success(c, gin.H{"token": token}, "Login successful")
	}
}

// PostLogout handles POST /auth/logout requests, clears user sessions and JWT tokens,
// and redirects to the entries page with a success message.
func PostLogout(c *gin.Context) {
	Auth := authentication.Resolve()

	// Clear session data
	err := Auth.Logout(c)
	if err != nil {
		Response.InternalError(c, "Logout failed", err)
		return
	}

	// Clear the JWT token cookie with proper security settings
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("jwt_token", "", -1, "/", "", false, true)

	// Clear any additional cookies that might exist
	c.SetCookie("session", "", -1, "/", "", false, true)

	// Clear context values
	c.Set("is_logged_in", false)
	c.Set("user_id", uint(0))
	c.Set("username", "")

	// Clear Authorization header
	c.Header("Authorization", "")

	log.Debug("User successfully logged out")

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/entries/", "User logged out successfully.")
	} else {
		Response.Success(c, nil, "Logged out successfully")
	}
}

// GetRegister handles GET /auth/register requests and renders the registration form
// template with city code and flash messages.
func GetRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/register.html", gin.H{"citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// PostRegister handles POST /auth/register requests, validates registration data,
// checks referral codes, and creates new user accounts with password hashing.
func PostRegister(c *gin.Context) {
	// bind the input to the user's model
	var user models.User
	if err := c.ShouldBind(&user); err != nil {
		Response.BadRequest(c, "Invalid form data", err)
		return
	}

	// Validate registration data using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateRegistrationData(&user)
	if !validationResult.IsValid {
		var validationErrors []ValidationErrorResponse
		for _, err := range validationResult.Errors {
			validationErrors = append(validationErrors, ValidationErrorResponse{
				Field:   err.Field,
				Message: err.Message,
				Value:   err.Value,
			})
		}
		Response.ValidationError(c, "The registration information you entered is incorrect, please try again.", validationErrors)
		return
	}

	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// check if there is a record with the given username
	res := database.WithContext(ctx).Where("username = ?", user.Username).First(&models.User{})
	if res.Error == nil {
		Response.Conflict(c, "User already signed up.")
		return
	}

	// check the referral code from registration form matches the one from environment variable/config
	referralCode, _ := c.Get("referralcode")
	if user.ReferralCode != referralCode {
		Response.Forbidden(c, "The referral code you entered is incorrect.")
		return
	}

	// hash the password
	hashedPWD, err := hashPassword(user.Password)
	if err != nil {
		Response.InternalError(c, "Password processing failed.", err)
		return
	}

	// use the hashed password
	user.Password = hashedPWD
	// create the db record
	res = database.WithContext(ctx).Create(&user)
	if res.Error != nil {
		Response.DatabaseError(c, "create", res.Error)
		return
	}

	// Handle response based on request type
	if Response.IsHTMLRequest(c) {
		Response.RedirectWithMessage(c, "/auth/login", fmt.Sprintf("New user '%s' registered successfully.", user.Username))
	} else {
		Response.Created(c, gin.H{"username": user.Username}, "Registration successful")
	}
}

// hashPassword hashes passwords
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// JWT middleware instance - will be initialized in main
var jwtMiddleware *authentication.JWTMiddleware

// InitializeJWTMiddleware initializes the JWT middleware with the auth service
// for use in authentication-required routes throughout the application.
func InitializeJWTMiddleware(authService *authentication.AuthService) {
	jwtMiddleware = authentication.NewJWTMiddleware(authService)
}

// SetUserStatus provides optional JWT authentication middleware that extracts user
// information if a valid token is present but doesn't require authentication.
func SetUserStatus() gin.HandlerFunc {
	if jwtMiddleware == nil {
		// Fallback to legacy behavior if middleware not initialized
		return func(c *gin.Context) {
			Auth := authentication.Resolve()
			if authed, err := Auth.UserID(c); err == nil && authed != 0 {
				c.Set("is_logged_in", true)
			} else {
				c.Set("is_logged_in", false)
			}
		}
	}
	return jwtMiddleware.OptionalAuth()
}

// EnsureLoggedIn ensures that a request will be aborted with an error if the user
// is not logged in, providing authentication protection for protected routes.
func EnsureLoggedIn() gin.HandlerFunc {
	if jwtMiddleware == nil {
		// Fallback to legacy behavior if middleware not initialized
		return func(c *gin.Context) {
			loggedInInterface, _ := c.Get("is_logged_in")
			loggedIn := loggedInInterface.(bool)
			if !loggedIn {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		}
	}
	return jwtMiddleware.AuthRequired()
}

// EnsureNotLoggedIn ensures that a request will be aborted with an error if the user
// is already logged in, preventing access to login/register pages for authenticated users.
func EnsureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedIn := authentication.IsLoggedIn(c)
		if loggedIn {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": "Already authenticated",
				"code":  "ALREADY_AUTHENTICATED",
			})
		}
	}
}

// PostRefreshToken handles POST /auth/refresh requests for JWT token refresh
// operations using the configured JWT middleware.
func PostRefreshToken(c *gin.Context) {
	if jwtMiddleware == nil {
		Response.InternalError(c, "Authentication service not available", errors.New("JWT middleware not initialized"))
		return
	}
	jwtMiddleware.RefreshToken()(c)
}

// preloadUserEntries loads user entries in the background during login
// to improve performance when the user navigates to their entries page
func preloadUserEntries(username string) {
	// Check if database is available before proceeding
	if database.DB == nil {
		log.Warn("Database not available for preloading user entries")
		return
	}

	// Add a small delay to avoid race conditions with in-memory databases in tests
	time.Sleep(10 * time.Millisecond)

	// Create a timeout context for the preload operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize services if not already done
	if services.UserCache == nil {
		services.InitUserCache()
	}

	userEntriesService := services.NewUserEntriesService()

	// Fetch user entries with timeout
	entries, err := userEntriesService.PreloadUserEntries(ctx, username)
	if err != nil {
		log.Warnf("Failed to preload entries for user %s: %v", username, err)
		return
	}

	// Cache the entries for quick access
	services.UserCache.SetUserEntries(username, entries)

	log.Debugf("Successfully preloaded %d entries for user %s", len(entries), username)
}

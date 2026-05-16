// Package routers provides router setup and configuration for the dinnerclub web application.
// It handles middleware setup, static assets, templates, and route configuration.
package routers

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/assets"
	"github.com/toozej/dinnerclub/internal/controllers"
	"github.com/toozej/dinnerclub/internal/helpers"
	"github.com/toozej/dinnerclub/pkg/csrf"
	"github.com/toozej/dinnerclub/pkg/ratelimit"
	"github.com/toozej/dinnerclub/templates"
)

// Router is the global Gin router instance.
var Router *gin.Engine

// NewRouter creates and initializes a new Gin router instance.
func NewRouter() *gin.Engine {
	r := gin.Default()
	Router = r
	return Router
}

// ResolveRouter returns the global router instance.
func ResolveRouter() *gin.Engine {
	return Router
}

// SetupRouterDefaults configures default middleware for the router including user status and defaults.
func SetupRouterDefaults(cityCode string, referralCode string) {
	r := ResolveRouter()
	r.Use(controllers.SetUserStatus())
	r.Use(controllers.SetDefaults(cityCode, referralCode))
}

func faviconFS() http.FileSystem {
	sub, err := fs.Sub(&assets.Assets, "favicon.ico")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// SetupStaticAssets configures static asset serving including favicon and general assets.
func SetupStaticAssets() {
	r := ResolveRouter()

	// serve regular static assets under /assets/*
	fs := &assets.Assets
	r.StaticFS("/assets", http.FS(fs))

	// serve favicon.ico at /favicon.ico
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS(".", faviconFS())
	})
}

// createTemplateFuncMap creates a function map with all template helper functions
func createTemplateFuncMap() template.FuncMap {
	h := &helpers.TemplateHelpers{}

	return template.FuncMap{
		"FormatBoolean":        h.FormatBoolean,
		"FormatRating":         h.FormatRating,
		"GetRatingDescription": h.GetRatingDescription,
		"GetRatingLabel":       h.GetRatingLabel,
		"UserProfileLink":      h.UserProfileLink,
		"TooltipText":          h.TooltipText,
		"GetRatingOptions":     h.GetRatingOptions,
		"SafeHTML":             h.SafeHTML,
		"Add":                  h.Add,
		"add":                  h.Add, // lowercase alias for compatibility
		"StringToInt":          h.StringToInt,
		"IsEmpty":              h.IsEmpty,
		"FormatDate":           h.FormatDate,
		"FormatDateTime":       h.FormatDateTime,
		"CSRFToken":            h.CSRFToken,
		"CSRFField":            h.CSRFField,
		"CSRFInput":            h.CSRFInput,
		"sub":                  h.Sub,
		"Sub":                  h.Sub,
		"mul":                  h.Mul,
		"Mul":                  h.Mul,
		"default":              h.Default,
	}
}

// SetupTemplates configures HTML template parsing with helper functions.
func SetupTemplates() {
	r := ResolveRouter()

	// create template helper functions
	funcMap := createTemplateFuncMap()

	// load HTML templates with helper functions from all subdirectories
	tmpl := template.New("").Funcs(funcMap)

	// Parse templates from all subdirectories (one level deep)
	tmpl = template.Must(tmpl.ParseFS(&templates.Templates, "*/*.html"))

	// Parse templates from two levels deep (e.g., social/feed.html)
	// Use a separate parse call so a missing pattern doesn't panic
	if t2, err := tmpl.ParseFS(&templates.Templates, "*/*/*.html"); err == nil {
		tmpl = t2
	}

	r.SetHTMLTemplate(tmpl)
}

// SetupCSRFProtection configures CSRF protection middleware for the router.
func SetupCSRFProtection(csrfSecret string, secure bool) {
	r := ResolveRouter()
	r.Use(csrf.Middleware(csrfSecret, secure))
}

// SetupRateLimiting configures general rate limiting middleware for all routes.
func SetupRateLimiting(rateLimitConfig *ratelimit.Config) {
	r := ResolveRouter()

	// Apply general rate limiting to all routes
	r.Use(ratelimit.GeneralRateLimitMiddleware(rateLimitConfig))
}

// GetAuthRateLimitMiddleware returns the auth rate limiting middleware for use in route groups.
func GetAuthRateLimitMiddleware(rateLimitConfig *ratelimit.Config) gin.HandlerFunc {
	return ratelimit.AuthRateLimitMiddleware(rateLimitConfig)
}

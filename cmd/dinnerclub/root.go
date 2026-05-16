// Package cmd provides the command-line interface and application entry point
// for the dinnerclub web application using Cobra CLI framework.
package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"

	"github.com/toozej/dinnerclub/internal/controllers"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/routers"
	"github.com/toozej/dinnerclub/internal/services"

	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/config"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/man"
	"github.com/toozej/dinnerclub/pkg/ratelimit"
	"github.com/toozej/dinnerclub/pkg/session"
	"github.com/toozej/dinnerclub/pkg/version"
)

func serveGin(rootPath string, sessionSecret string, csrfSecret string, cityCode string, referralCode string, ginMode string, rateLimitConfig *ratelimit.Config) {
	// init Gin router
	r := routers.NewRouter()

	// init sessions
	sessionMiddleware := session.InitSession(sessionSecret)
	r.Use(sessionMiddleware)
	log.Info("session system setup successfully")

	// init authentication service
	sessionStore := session.NewStore(sessionSecret)
	authService := authentication.NewAuthService(sessionSecret, "dinnerclub")
	authentication.New(sessionStore, authService)
	log.Info("authentication system setup successfully")

	// initialize JWT middleware
	controllers.InitializeJWTMiddleware(authService)
	log.Info("JWT middleware initialized successfully")

	// setup CSRF protection (secure=true for production)
	secure := ginMode == "release"
	routers.SetupCSRFProtection(csrfSecret, secure)
	log.Info("CSRF protection setup successfully")

	// setup rate limiting
	routers.SetupRateLimiting(rateLimitConfig)
	log.Info("rate limiting setup successfully")

	// setup router defaults, static assets and templates
	routers.SetupTemplates()
	routers.SetupRouterDefaults(cityCode, referralCode)
	routers.SetupStaticAssets()

	// setup public and private routes
	routers.SetupPublicRoutes(rateLimitConfig)
	routers.SetupPrivateRoutes()
	log.Info("routes setup successfully")

	// start up Gin web server
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Execute initializes and runs the dinnerclub web application with all necessary
// configurations, database connections, middleware, and routing setup.
func Execute() {
	// load application configurations
	if err := config.LoadConfig("./app.env"); err != nil {
		log.Fatalf("invalid application configuration: %s", err)
	}
	c := config.Config

	// set log level based off environment variable
	lvl, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		log.Fatal("parsing log level from environment variable failed")
	}
	log.SetLevel(lvl)
	log.Printf("log level set to %v", lvl)

	command := &cobra.Command{
		Use:   "dinnerclub",
		Short: "dinnerclub",
		Long:  `Main entrypoint into the dinnerclub web application`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugf("Loaded config: %+v\n", c)

			var connString string
			if c.DatabaseURL != "" {
				// Postgres connection string is already in environment as DATABASE_URL
				connString = c.DatabaseURL
			} else {
				// form Postgres connection string from config variables
				connString = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/Los_Angeles", c.PostgresHostname, c.PostgresUser, c.PostgresPassword, c.PostgresDB, c.PostgresPort)
			}
			// connect to Postgres database via Gorm
			database.ConnectDatabase(connString, c.LogLevel)
			// auto-migrate database schema
			models.MigrateSchema()

			// initialize user cache service
			services.InitUserCache()
			log.Info("user cache service initialized successfully")

			// create rate limiting configuration
			rateLimitConfig := ratelimit.NewConfig(
				c.RateLimitEnabled,
				c.RateLimitRequests,
				c.RateLimitPeriod,
				c.RateLimitAuthRequests,
				c.RateLimitAuthPeriod,
			)

			// setup Gin router and serve
			serveGin(".", c.SessionSecret, c.CSRFSecret, c.CityCode, c.ReferralCode, c.GinMode, rateLimitConfig)
		},
	}

	command.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)

	if err := command.Execute(); err != nil {
		log.Fatal(err.Error())
	}

}

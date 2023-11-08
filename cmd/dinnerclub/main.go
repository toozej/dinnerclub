package main

import (
	"fmt"

	"github.com/gocondor/core/jwt"
	"github.com/gocondor/core/sessions"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/routers"

	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/config"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/man"
	"github.com/toozej/dinnerclub/pkg/session"
	"github.com/toozej/dinnerclub/pkg/version"
)

func serveGin(rootPath string, sessionSecret string, cityCode string, referralCode string) {
	// init Gin router
	r := routers.NewRouter()

	// init sessions
	ses := session.InitSession(sessionSecret)
	r.Use(ses)
	log.Info("session system setup successfully")

	// setup router defaults, static assets and templates
	routers.SetupTemplates()
	routers.SetupRouterDefaults(cityCode, referralCode)
	routers.SetupStaticAssets()

	// setup public and private routes
	routers.SetupPublicRoutes()
	routers.SetupPrivateRoutes()
	log.Info("routes setup successfully")

	// init auth
	jwt.New()
	authentication.New(sessions.Resolve(), jwt.Resolve())
	log.Info("authentication system setup successfully")

	// start up Gin web server
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
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
			// TODO remove printing of sensitive env vars
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

			// setup Gin router and serve
			serveGin(".", c.SessionSecret, c.CityCode, c.ReferralCode)
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

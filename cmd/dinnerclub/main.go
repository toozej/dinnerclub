package main

import (
	"fmt"

	"github.com/gocondor/core/jwt"
	"github.com/gocondor/core/sessions"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/routers"

	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/config"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/man"
	"github.com/toozej/dinnerclub/pkg/session"
	"github.com/toozej/dinnerclub/pkg/version"
)

func serveGin(rootPath string, sessionSecret string, cityCode string) {
	// init Gin router
	r := routers.NewRouter()

	// init sessions
	ses := session.InitSession(sessionSecret)
	r.Use(ses)
	log.Info("session system setup successfully")

	// setup router defaults, static assets and templates
	routers.SetupTemplates()
	routers.SetupRouterDefaults(cityCode)
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
	if err := config.LoadConfig("./"); err != nil {
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
			// TODO make viper load config from OS environment variables as well as *.env files
			log.Debugf("Loaded config: %+v\n", c)

			// form variable portions of Postgres connection string from config variables
			conn_string := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.PostgresHostname, c.PostgresUser, c.PostgresPassword, c.PostgresDB, c.PostgresPort)
			// connect to Postgres database via Gorm
			database.ConnectDatabase(conn_string, c.LogLevel)
			// auto-migrate database schema
			models.MigrateSchema()

			// setup Gin router and serve
			serveGin(".", c.SessionSecret, c.CityCode)
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

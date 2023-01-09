package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/toozej/dinnerclub/internal/models"

	"github.com/toozej/dinnerclub/pkg/config"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/man"
	"github.com/toozej/dinnerclub/pkg/version"
)

func migrateSchema() {
	var schemaModels = []interface{}{
		models.Entry{},
		models.User{},
	}

	for m := range schemaModels {
		if err := database.DB.AutoMigrate(&schemaModels[m]); err != nil {
			log.Fatalf("Failed to migrate database schema: %s", err)
		}
	}
	log.Printf("Successfully migrated all database schemas")
}

func setupRouter(rootPath string) *gin.Engine {
	r := gin.Default()

	// load HTML templates
	r.LoadHTMLGlob(rootPath + "/templates/*.html")

	// serve static favicon file from a location relative to main.go directory
	r.StaticFile("/favicon.ico", rootPath+"/assets/favicon.ico")

	// TODO move setting up routes to own function
	// TODO change routes funcs from inline to own functions?
	// primary routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// entries related routes
	entries := r.Group("/entries")
	entries.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// restaurants related routes
	restaurants := r.Group("/restaurants")
	restaurants.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "restaurants.html", nil)
	})

	// profile related routes
	profile := r.Group("/profile")
	profile.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "profile.html", nil)
	})
	profile.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// health and status routes (which are identical)
	// TODO include database connectivity health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}

func main() {
	// load application configurations
	if err := config.LoadConfig("./"); err != nil {
		log.Fatalf("invalid application configuration: %s", err)
	}
	c := config.Config

	command := &cobra.Command{
		Use:   "dinnerclub",
		Short: "dinnerclub",
		Long:  `Main entrypoint into the dinnerclub web application`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO remove printing of sensitive env vars
			// TODO make viper load config from OS environment variables as well as *.env files
			fmt.Printf("Loaded config: %+v\n", c)

			// form variable portions of Postgres connection string from config variables
			conn_string := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.PostgresHostname, c.PostgresUser, c.PostgresPassword, c.PostgresDB, c.PostgresPort)
			// connect to Postgres database via Gorm
			database.ConnectDatabase(conn_string)
			// auto-migrate database schema
			migrateSchema()

			// setup Gin router
			r := setupRouter(".")

			// start up Gin web server
			err := r.Run(":8080")
			if err != nil {
				log.Fatal(err.Error())
			}
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

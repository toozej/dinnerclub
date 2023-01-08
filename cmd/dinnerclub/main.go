package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/toozej/dinnerclub/pkg/config"
	"github.com/toozej/dinnerclub/pkg/man"
	"github.com/toozej/dinnerclub/pkg/version"
)

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
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	command := &cobra.Command{
		Use:   "dinnerclub",
		Short: "dinnerclub",
		Long:  `Main entrypoint into the dinnerclub web application`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO remove printing of sensitive env vars
			// TODO make viper load config from OS environment variables as well as *.env files
			fmt.Printf("%+v\n", config.Config)
			r := setupRouter(".")
			err := r.Run(":8080")
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	command.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)

	if err := command.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

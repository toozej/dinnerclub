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

func setupRouter() *gin.Engine {
	r := gin.Default()

	// serve static favicon file from a location relative to main.go directory
	r.StaticFile("/favicon.ico", "./assets/favicon.ico")

	// routes
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "hello world"})
	})

	// health and status routes (which are identical)
	// TODO include database connectivity health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})

	return r
}

func main() {
	// load application configurations
	if err := config.LoadConfig("./config"); err != nil {
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	command := &cobra.Command{
		Use:   "dinnerclub",
		Short: "dinnerclub",
		Long:  `Main entrypoint into the dinnerclub web application`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Config.ConfigVar)
			r := setupRouter()
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

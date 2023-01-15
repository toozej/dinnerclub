package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
)

func SetupPublicRoutes(rootPath string) {
	r := ResolveRouter()

	// load HTML templates
	r.LoadHTMLGlob(rootPath + "/templates/*.html")

	// serve static favicon file from a location relative to main.go directory
	r.StaticFile("/favicon.ico", rootPath+"/assets/favicon.ico")

	// primary routes
	// TODO change routes funcs from inline to own functions
	// TODO change routes funcs to handle JSON, HTML and XML
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

	// user authentication related routes
	preAuth := r.Group("/auth")
	preAuth.POST("/register", controllers.Register)
	preAuth.POST("/login", controllers.Login)
	preAuth.POST("/logout", controllers.Logout)

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
}

package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
)

func SetupPublicRoutes(rootPath string) {
	r := ResolveRouter()

	// load HTML templates
	r.LoadHTMLGlob(rootPath + "/templates/*/*.html")

	// serve static favicon file from a location relative to main.go directory
	r.StaticFile("/favicon.ico", rootPath+"/assets/favicon.ico")

	// primary routes
	// TODO change routes funcs from inline to own functions
	// TODO change routes funcs to handle JSON, HTML and XML
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/entries/")
	})

	// entries related routes
	entries := r.Group("/entries")
	entries.GET("/", controllers.FindEntries)
	entries.GET("/:id", controllers.FindEntry)

	// restaurants related routes
	restaurants := r.Group("/restaurants")
	// TODO create controllers/restaurants.go with similar FindRestaurants, FindRestaurant as Entries/Entry
	// TODO use restaurants controllers here
	restaurants.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "restaurants/index.html", nil)
	})
	restaurants.GET("/:name", func(c *gin.Context) {
		c.HTML(http.StatusOK, "restaurants/restaurant.html", nil)
	})

	// user authentication related routes
	preAuth := r.Group("/auth")
	preAuth.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth/register.html", nil)
	})
	preAuth.POST("/register", controllers.Register)
	preAuth.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth/login.html", nil)
	})
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

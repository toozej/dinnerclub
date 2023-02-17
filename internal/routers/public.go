package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
)

func SetupPublicRoutes() {
	r := ResolveRouter()

	// primary routes
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
	restaurants.GET("/", controllers.FindRestaurants)
	restaurants.GET("/:id", controllers.FindRestaurant)

	// user pre-authenticated authentication related routes
	preAuth := r.Group("/auth")
	preAuth.Use(controllers.EnsureNotLoggedIn())
	preAuth.GET("/register", controllers.RegisterGet)
	preAuth.POST("/register", controllers.RegisterPost)
	preAuth.GET("/login", controllers.LoginGet)
	preAuth.POST("/login", controllers.LoginPost)

	// general healthcheck route
	r.GET("/health", controllers.HealthGet)
}

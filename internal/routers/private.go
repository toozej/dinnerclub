package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
)

func SetupPrivateRoutes(rootPath string) {
	r := ResolveRouter()

	profile := r.Group("/profile")
	// TODO figure out how to ensure private routes are authenticated with gocondor libs
	// profile.Use(authentication.Resolve())
	profile.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "users/profile.html", nil)
	})
	// TODO write GetProfile controller
	// profile.GET("/", controllers.GetProfile)

	postAuth := r.Group("/")
	// TODO figure out how to ensure private routes are authenticated with gocondor libs
	// postAuth.Use(authentication.Resolve())
	postAuth.GET("/entries/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "entries/new.html", nil)
	})
	postAuth.POST("/entries/new", controllers.CreateEntry)
	postAuth.PATCH("/entries/update", controllers.UpdateEntry)
	postAuth.DELETE("/entries/delete", controllers.DeleteEntry)
}

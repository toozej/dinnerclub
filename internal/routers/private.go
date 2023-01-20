package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
)

// private routes are ones which the user must be authenticated to use
func SetupPrivateRoutes(rootPath string) {
	r := ResolveRouter()

	profile := r.Group("/profile")
	profile.Use(controllers.EnsureLoggedIn())
	profile.GET("/", controllers.GetProfile)

	postAuth := r.Group("/")
	postAuth.Use(controllers.EnsureLoggedIn())
	postAuth.POST("/auth/logout", controllers.Logout)
	postAuth.GET("/entries/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "entries/new.html", gin.H{"citycode": c.MustGet("citycode").(string)})
	})
	postAuth.POST("/entries/new", controllers.CreateEntry)
	postAuth.PATCH("/entries/:id", controllers.UpdateEntry)
	postAuth.DELETE("/entries/:id", controllers.DeleteEntry)
}

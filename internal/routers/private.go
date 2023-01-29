package routers

import (
	"github.com/toozej/dinnerclub/internal/controllers"
)

// private routes are ones which the user must be authenticated to use
func SetupPrivateRoutes() {
	r := ResolveRouter()

	profile := r.Group("/profile")
	profile.Use(controllers.EnsureLoggedIn())
	profile.GET("/", controllers.GetProfile)

	postAuth := r.Group("/")
	postAuth.Use(controllers.EnsureLoggedIn())
	postAuth.POST("/auth/logout", controllers.Logout)
	postAuth.GET("/entries/new", controllers.CreateEntryGet)
	postAuth.POST("/entries/new", controllers.CreateEntryPost)
	postAuth.GET("/entries/:id/update", controllers.UpdateEntryGet)
	postAuth.PATCH("/entries/:id", controllers.UpdateEntryPatch)
	postAuth.DELETE("/entries/:id", controllers.DeleteEntryPost)
}

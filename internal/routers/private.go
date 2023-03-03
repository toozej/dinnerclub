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
	profile.POST("/update", controllers.UpdateProfile)
	profile.POST("/delete", controllers.DeleteUser)

	postAuth := r.Group("/")
	postAuth.Use(controllers.EnsureLoggedIn())
	postAuth.POST("/auth/logout", controllers.Logout)
	postAuth.GET("/entries/new", controllers.CreateEntryGet)
	postAuth.POST("/entries/new", controllers.CreateEntryPost)
	postAuth.GET("/entries/:id/update", controllers.UpdateEntryGet)
	postAuth.POST("/entries/:id/update", controllers.UpdateEntryPost)
	postAuth.POST("/entries/:id/delete", controllers.DeleteEntry)
	postAuth.GET("/entries/submittedby/:username", controllers.FindEntryByUsername)
	postAuth.POST("/entries/submittedby/:username", controllers.FindEntryByUsername)
	postAuth.GET("/status", controllers.StatusGet)
}

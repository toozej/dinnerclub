package routers

import (
	"github.com/toozej/dinnerclub/internal/controllers"
)

func SetupPrivateRoutes(rootPath string) {
	r := ResolveRouter()

	profile := r.Group("/profile")
	// TODO figure out how to ensure private routes are authenticated with gocondor libs
	// profile.Use(authentication.Resolve())
	profile.GET("/", controllers.GetProfile)

	postAuth := r.Group("/")
	// TODO figure out how to ensure private routes are authenticated with gocondor libs
	// postAuth.Use(authentication.Resolve())
	postAuth.POST("/entry", controllers.CreateEntry)
	postAuth.PATCH("/entry", controllers.UpdateEntry)
	postAuth.DELETE("/entry", controllers.DeleteEntry)
}

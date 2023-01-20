package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/controllers"
)

var Router *gin.Engine

// New initiates new router
func NewRouter() *gin.Engine {
	r := gin.Default()
	Router = r
	return Router
}

// Resolve resolves an already initiated router
func ResolveRouter() *gin.Engine {
	return Router
}

func SetupRouterDefaults(cityCode string) {
	r := ResolveRouter()
	r.Use(controllers.SetUserStatus())
	r.Use(controllers.SetDefaults(cityCode))
}

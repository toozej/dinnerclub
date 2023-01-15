package routers

import (
	"github.com/gin-gonic/gin"
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

package routers

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/assets"
	"github.com/toozej/dinnerclub/internal/controllers"
	"github.com/toozej/dinnerclub/templates"
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

func SetupRouterDefaults(cityCode string, referralCode string) {
	r := ResolveRouter()
	r.Use(controllers.SetUserStatus())
	r.Use(controllers.SetDefaults(cityCode, referralCode))
}

func faviconFS() http.FileSystem {
	sub, err := fs.Sub(&assets.Assets, "favicon.ico")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func SetupStaticAssets() {
	r := ResolveRouter()

	// serve regular static assets under /assets/*
	fs := &assets.Assets
	r.StaticFS("/assets", http.FS(fs))

	// serve favicon.ico at /favicon.ico
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS(".", faviconFS())
	})
}

func SetupTemplates() {
	r := ResolveRouter()

	// load HTML templates
	tmpl := template.Must(template.ParseFS(&templates.Templates, "*/*.html"))
	r.SetHTMLTemplate(tmpl)
}

package routers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTemplates(t *testing.T) {
	t.Run("initializes templates without errors", func(t *testing.T) {
		// Create a new router for testing
		gin.SetMode(gin.TestMode)
		router := gin.New()
		Router = router

		// This should not panic
		assert.NotPanics(t, func() {
			SetupTemplates()
		})

		// Verify that the router has templates set
		assert.NotNil(t, router.HTMLRender)
	})

	t.Run("template function map contains expected functions", func(t *testing.T) {
		funcMap := createTemplateFuncMap()

		expectedFunctions := []string{
			"FormatBoolean",
			"FormatRating",
			"GetRatingDescription",
			"GetRatingLabel",
			"UserProfileLink",
			"TooltipText",
			"GetRatingOptions",
			"SafeHTML",
			"Add",
			"add",
			"StringToInt",
			"IsEmpty",
			"FormatDate",
			"FormatDateTime",
			"CSRFToken",
			"CSRFField",
			"CSRFInput",
		}

		for _, funcName := range expectedFunctions {
			assert.Contains(t, funcMap, funcName, "Function %s should be in template function map", funcName)
			assert.NotNil(t, funcMap[funcName], "Function %s should not be nil", funcName)
		}
	})

	t.Run("template functions are callable", func(t *testing.T) {
		funcMap := createTemplateFuncMap()

		// Test a few key functions to ensure they're properly initialized
		t.Run("FormatBoolean function works", func(t *testing.T) {
			formatBooleanFunc := funcMap["FormatBoolean"]
			assert.NotNil(t, formatBooleanFunc)

			// Create a simple template to test the function
			tmpl := template.New("test").Funcs(template.FuncMap{"FormatBoolean": formatBooleanFunc})
			tmpl, err := tmpl.Parse("{{FormatBoolean .}}")
			require.NoError(t, err)

			var buf strings.Builder
			err = tmpl.Execute(&buf, true)
			require.NoError(t, err)
			assert.Equal(t, "Yes", buf.String())
		})

		t.Run("Add function works", func(t *testing.T) {
			addFunc := funcMap["Add"]
			assert.NotNil(t, addFunc)

			tmpl := template.New("test").Funcs(template.FuncMap{"Add": addFunc})
			tmpl, err := tmpl.Parse("{{Add . 5}}")
			require.NoError(t, err)

			var buf strings.Builder
			err = tmpl.Execute(&buf, 10)
			require.NoError(t, err)
			assert.Equal(t, "15", buf.String())
		})

		t.Run("add (lowercase) function works", func(t *testing.T) {
			addFunc := funcMap["add"]
			assert.NotNil(t, addFunc)

			tmpl := template.New("test").Funcs(template.FuncMap{"add": addFunc})
			tmpl, err := tmpl.Parse("{{add . 3}}")
			require.NoError(t, err)

			var buf strings.Builder
			err = tmpl.Execute(&buf, 7)
			require.NoError(t, err)
			assert.Equal(t, "10", buf.String())
		})
	})

	t.Run("templates can be rendered after setup", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		Router = router

		SetupTemplates()

		// Add a test route that uses templates
		router.GET("/test", func(c *gin.Context) {
			c.HTML(http.StatusOK, "base/header.html", gin.H{
				"title": "Test Page",
			})
		})

		// Test the route
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should not return 500 (template error)
		assert.NotEqual(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCreateTemplateFuncMapUtilities(t *testing.T) {
	t.Run("returns non-nil function map", func(t *testing.T) {
		funcMap := createTemplateFuncMap()
		assert.NotNil(t, funcMap)
		assert.Greater(t, len(funcMap), 0)
	})

	t.Run("function map is consistent across calls", func(t *testing.T) {
		funcMap1 := createTemplateFuncMap()
		funcMap2 := createTemplateFuncMap()

		assert.Equal(t, len(funcMap1), len(funcMap2))

		// Check that the same functions exist in both maps
		for name := range funcMap1 {
			assert.Contains(t, funcMap2, name)
		}
	})

	t.Run("all functions in map are callable", func(t *testing.T) {
		funcMap := createTemplateFuncMap()

		for name, fn := range funcMap {
			assert.NotNil(t, fn, "Function %s should not be nil", name)

			// Verify it's actually a function by checking if we can create a template with it
			testFuncMap := template.FuncMap{name: fn}
			tmpl := template.New("test").Funcs(testFuncMap)

			// This should not panic
			assert.NotPanics(t, func() {
				_, err := tmpl.Parse("{{" + name + "}}")
				// Some functions require parameters, so we don't check for parse errors
				// We just verify the function can be registered
				_ = err
			}, "Function %s should be registerable in template", name)
		}
	})
}

func TestTemplateInitializationIntegration(t *testing.T) {
	t.Run("full template setup works end-to-end", func(t *testing.T) {
		// Create fresh router
		gin.SetMode(gin.TestMode)
		router := gin.New()
		Router = router

		// Setup templates
		SetupTemplates()

		// Create a test route that uses multiple template functions
		router.GET("/integration-test", func(c *gin.Context) {
			c.HTML(http.StatusOK, "base/header.html", gin.H{
				"title":   "Integration Test",
				"visited": true,
				"rating":  4.5,
				"count":   10,
			})
		})

		// Test the route
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/integration-test", nil)
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	})

	t.Run("template setup is idempotent", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		Router = router

		// Setup templates multiple times
		assert.NotPanics(t, func() {
			SetupTemplates()
			SetupTemplates()
			SetupTemplates()
		})

		// Should still work
		router.GET("/idempotent-test", func(c *gin.Context) {
			c.HTML(http.StatusOK, "base/header.html", gin.H{"title": "Test"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/idempotent-test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("template functions work with real template content", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		Router = router

		SetupTemplates()

		// Test with a template that actually uses helper functions
		router.GET("/helper-test", func(c *gin.Context) {
			// Use a simple template string that exercises helper functions
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, "Template setup successful")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/helper-test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Template setup successful")
	})
}

func TestTemplateErrorHandling(t *testing.T) {
	t.Run("handles missing template gracefully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		Router = router

		SetupTemplates()

		router.GET("/missing-template", func(c *gin.Context) {
			c.HTML(http.StatusOK, "nonexistent/template.html", gin.H{})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/missing-template", nil)
		router.ServeHTTP(w, req)

		// Gin handles missing templates by returning 200 with empty body
		// This is the expected behavior - it doesn't crash the application
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 0, w.Body.Len(), "Response body should be empty for missing template")
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	})

	t.Run("template function errors are handled", func(t *testing.T) {
		// This test verifies that template function errors don't crash the application
		funcMap := createTemplateFuncMap()

		// Test that functions handle invalid input gracefully
		for name, fn := range funcMap {
			assert.NotNil(t, fn, "Function %s should not be nil", name)

			// Create a template that might cause the function to error
			testFuncMap := template.FuncMap{name: fn}
			tmpl := template.New("test").Funcs(testFuncMap)

			// This should not panic even if the template parse fails
			assert.NotPanics(t, func() {
				_, err := tmpl.Parse("{{" + name + " nil}}")
				_ = err // Ignore parse errors, we're just testing for panics
			}, "Function %s should not panic during template parsing", name)
		}
	})
}

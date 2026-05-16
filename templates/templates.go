// Package templates provides embedded HTML templates for the dinnerclub web application.
// It includes all template files that are embedded into the binary for rendering web pages.
package templates

import "embed"

// Templates contains all HTML template files embedded in the application binary.
//
//go:embed *
var Templates embed.FS

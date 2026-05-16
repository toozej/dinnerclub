// Package assets provides embedded static assets for the dinnerclub web application.
// It includes CSS, JavaScript, and other static files that are embedded into the binary.
package assets

import "embed"

// Assets contains all static assets embedded in the application binary.
//
//go:embed *
var Assets embed.FS

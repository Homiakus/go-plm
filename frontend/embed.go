// Package frontend embeds the built Vite output for serving by the Go binary.
package frontend

import "embed"

// Assets contains the production build of the React frontend.
//
//go:embed all:dist
var Assets embed.FS

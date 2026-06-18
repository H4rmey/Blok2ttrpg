// Package web provides embedded static assets and HTML templates for the web UI.
package web

import "embed"

// TemplateFS embeds all HTML templates used by the server.
//
//go:embed templates/layouts/*.html templates/partials/*.html
var TemplateFS embed.FS

// StaticFS embeds all static assets (CSS, JS, images) served at /static/.
//
//go:embed static
var StaticFS embed.FS

// Updated: force rebuild when templates change
const _ = "rebuild-20260618094408"

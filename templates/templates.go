package templates

import "embed"

// Files contains the app's HTML pages.
//
//go:embed media/*.html
var Files embed.FS

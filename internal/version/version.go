// Package version provides version information for the application.
// This is a separate package to avoid import cycles.
package version

// Version is set at build time via ldflags.
var Version string = "dev"

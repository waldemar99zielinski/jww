// Package buildinfo contains the application version and linker-injected build metadata.
package buildinfo

// Version is the source of truth for the version of the combined server and web application.
const Version = "0.0.0"

var (
	Commit = "unknown"
	Date   = "unknown"
)

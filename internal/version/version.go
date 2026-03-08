package version

// Version is set at build time via ldflags (e.g. -ldflags="-X '...Version=v1.0.0'").
var Version = "dev"

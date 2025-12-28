package main

import (
	_ "embed"
	"runtime/debug"
)

//go:generate sh -c "printf \"%s\" $(git describe --tags --always) > version.txt"

//go:embed version.txt
var version string

var commit string // nolint: gochecknoglobals

var buildDate string // nolint: gochecknoglobals

// nolint: gochecknoinits
// this is for build info like commit hash.
func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				commit = setting.Value[:7]
			case "vcs.modified":
				version += "-dev"
			}
		}
	}
}

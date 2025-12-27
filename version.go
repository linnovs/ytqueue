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
// this is for build info like commit hash and build date.
func init() {
	isDirty := false
	vcsTime := ""

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				commit = setting.Value[:7]
			case "vcs.time":
				vcsTime = setting.Value
			case "vcs.modified":
				version += "-dev"
				isDirty = true
			}
		}
	}

	if !isDirty {
		buildDate = vcsTime
	}
}

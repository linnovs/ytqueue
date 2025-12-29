package main

import (
	"os"
	"path/filepath"
	"strings"
)

func shortenPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("unable to get user home directory")
	}

	var found bool

	path, found = strings.CutPrefix(path, homeDir)
	if found {
		path = filepath.Join("~", path)
	}

	return filepath.Clean(path)
}

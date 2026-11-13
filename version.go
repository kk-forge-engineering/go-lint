// Package golint holds go-lint's release number.
package golint

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// Version is the release number in VERSION.
var Version = strings.TrimSpace(versionFile)

// Package wiring sits under an entrypoint, so its compose.go is no composition root.
package wiring

import _ "example.com/app/features/a/infra" // want `only the composition root, the compose\.go of an entrypoints/<name> package, imports a feature's infra`

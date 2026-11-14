// Package wiring sits under an entrypoint, so its compose.go is no composition root.
package wiring

import _ "example.com/app/features/a/infra" // want `import example\.com/app/features/a, not its inner package example\.com/app/features/a/infra`

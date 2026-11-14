package start

// Only the composition root, compose.go, imports a feature's infra.
import _ "example.com/app/features/a/infra" // want `import example\.com/app/features/a, not its inner package example\.com/app/features/a/infra`

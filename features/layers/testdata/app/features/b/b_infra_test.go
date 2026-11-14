package b_test

import (
	// A test in another module may import a feature's infra, its second surface.
	_ "example.com/app/features/a/infra"
	_ "example.com/app/features/a/infra/sub" // want `import example\.com/app/features/a, not its inner package example\.com/app/features/a/infra/sub`
)

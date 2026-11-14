package b_test

import (
	// A test imports a surface's test kit: the feature's, and its infra's.
	_ "example.com/app/features/a/a_testing"
	_ "example.com/app/features/a/infra/infra_testing"
	_ "example.com/app/features/a/infra/other_testing" // want `import example\.com/app/features/a, not its inner package example\.com/app/features/a/infra/other_testing`
)

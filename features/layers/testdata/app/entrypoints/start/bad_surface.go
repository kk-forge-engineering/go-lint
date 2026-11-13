package start

import (
	_ "example.com/app/features/a/sub" // want `import example\.com/app/features/a, not its inner package example\.com/app/features/a/sub`
)

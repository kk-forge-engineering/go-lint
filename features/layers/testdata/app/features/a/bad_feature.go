package a

import (
	_ "example.com/app/features/b" // want `features may not import features \(example\.com/app/features/b\)`
)

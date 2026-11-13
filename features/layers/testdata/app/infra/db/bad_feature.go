package db

import (
	_ "example.com/app/features/b" // want `infra may not import features \(example\.com/app/features/b\)`
)

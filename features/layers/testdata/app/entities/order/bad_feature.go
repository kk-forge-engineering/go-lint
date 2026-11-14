// No module imports entities/order, so the report lands on the first file by name.
package order // want `no features, orchestration or entities module imports entities/order: move it into the module that uses it, since an entity is what two or more of them share`

import (
	_ "example.com/app/features/b" // want `entities may not import features \(example\.com/app/features/b\)`
)

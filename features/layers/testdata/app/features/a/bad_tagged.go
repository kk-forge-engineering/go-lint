//go:build integration

package a

import _ "example.com/app/infra/db" // want `features may not import infra \(example\.com/app/infra/db\)`

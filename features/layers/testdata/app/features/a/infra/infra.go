// Package infra holds the adapters only features/a needs. It builds on the root infra.
package infra

import (
	_ "example.com/app/infra/db"
	_ "example.com/app/shared"
)

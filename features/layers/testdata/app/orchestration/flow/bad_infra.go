package flow

import (
	_ "example.com/app/infra/db" // want `orchestration may not import infra \(example\.com/app/infra/db\)`
)

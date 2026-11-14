package start

import (
	_ "example.com/app"
	_ "example.com/app/entities/lone"
	_ "example.com/app/entities/user"
	_ "example.com/app/features/a"
	_ "example.com/app/features/a/infra"
	_ "example.com/app/infra/db"
	_ "example.com/app/orchestration/flow"
	_ "example.com/app/shared"
	_ "example.com/app/shared/log"
)

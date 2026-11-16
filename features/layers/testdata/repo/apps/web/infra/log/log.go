package log

import (
	logger "example.com/repo/logger/shared"
	web "example.com/repo/web/shared"
)

// Names takes a value from the `logger` app's shared package and one from this app's own shared package.
// The layers check allows an import of another app's shared package, so this file passes.
var Names = []string{logger.App, web.Name}

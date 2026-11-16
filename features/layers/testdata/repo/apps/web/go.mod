module example.com/repo/web

go 1.26.0

require (
	example.com/repo/api v0.0.0
	example.com/repo/logger v0.0.0
)

replace (
	example.com/repo/api => ../api
	example.com/repo/logger => ../../packages/logger
)

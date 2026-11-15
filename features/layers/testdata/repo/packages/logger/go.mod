module example.com/repo/logger

go 1.26.0

require (
	example.com/repo/api v0.0.0
	example.com/repo/web v0.0.0
)

replace (
	example.com/repo/api => ../../apps/api
	example.com/repo/web => ../../apps/web
)

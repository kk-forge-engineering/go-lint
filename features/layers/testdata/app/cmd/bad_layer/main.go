package main

import (
	_ "example.com/app/features/b" // want `cmd may import only entrypoints \(example\.com/app/features/b\)`

	"example.com/app/entrypoints/bad_layer"
)

func main() { bad_layer.Start() }

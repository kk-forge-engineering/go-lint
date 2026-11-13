package main

import (
	_ "fmt" // want `cmd may import only entrypoints \(fmt\)`

	"example.com/app/entrypoints/bad_import"
)

func main() { bad_import.Start() }

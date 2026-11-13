package main

import "example.com/app/entrypoints/bad_extra"

var mode = "x" // want `cmd/bad_extra may hold only func main with one call, without arguments, into entrypoints/bad_extra`

func main() { bad_extra.Start() }

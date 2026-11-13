package main

import "example.com/app/entrypoints/bad_local"

func main() { run() } // want `cmd/bad_local may hold only func main with one call, without arguments, into entrypoints/bad_local`

func run() { bad_local.Start() } // want `cmd/bad_local may hold only func main with one call, without arguments, into entrypoints/bad_local`

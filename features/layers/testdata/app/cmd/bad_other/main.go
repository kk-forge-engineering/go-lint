package main

import "example.com/app/entrypoints/start"

func main() { start.Start() } // want `cmd/bad_other may hold only func main with one call, without arguments, into entrypoints/bad_other`

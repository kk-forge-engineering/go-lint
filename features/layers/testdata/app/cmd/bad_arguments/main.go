package main

import "example.com/app/entrypoints/bad_arguments"

func main() { bad_arguments.StartWith("x") } // want `cmd/bad_arguments may hold only func main with one call, without arguments, into entrypoints/bad_arguments`

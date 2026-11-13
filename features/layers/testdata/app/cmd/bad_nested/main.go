package main

import "example.com/app/entrypoints/bad_nested"

func main() { bad_nested.New().Run() } // want `cmd/bad_nested may hold only func main with one call, without arguments, into entrypoints/bad_nested`

package main

import "example.com/app/entrypoints/bad_two_calls"

func main() { // want `cmd/bad_two_calls may hold only func main with one call, without arguments, into entrypoints/bad_two_calls`
	bad_two_calls.Start()
	bad_two_calls.Start()
}

package main // want `a launcher goes in cmd/<name>/, beside entrypoints/<name>`

import "example.com/app/entrypoints/start"

func main() { start.Start() }

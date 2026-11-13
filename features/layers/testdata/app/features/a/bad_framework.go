package a

import _ "github.com/gofiber/fiber/v2/middleware" // want `only entrypoints may import the HTTP server framework github\.com/gofiber/fiber/v2/middleware: an HTTP server is an entrypoint.s adapter: move it to entrypoints/<name>/http`

package cache

import (
	_ "github.com/go-chi/chi/v5"            // want `only entrypoints may import the HTTP server framework github\.com/go-chi/chi/v5: an HTTP server is an entrypoint.s adapter: move it to entrypoints/<name>/http`
	_ "github.com/gorilla/mux"              // want `only entrypoints may import the HTTP server framework github\.com/gorilla/mux: an HTTP server is an entrypoint.s adapter: move it to entrypoints/<name>/http`
	_ "github.com/julienschmidt/httprouter" // want `only entrypoints may import the HTTP server framework github\.com/julienschmidt/httprouter: an HTTP server is an entrypoint.s adapter: move it to entrypoints/<name>/http`
	_ "github.com/labstack/echo/v4"         // want `only entrypoints may import the HTTP server framework github\.com/labstack/echo/v4: an HTTP server is an entrypoint.s adapter: move it to entrypoints/<name>/http`
)

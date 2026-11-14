package a

import web "net/http"

// Mux returns a router. The host build leaves this file out, so it has no type information and the server check skips
// it: GOOS=windows go_lint run checks it.
func mux() *web.ServeMux { return web.NewServeMux() }

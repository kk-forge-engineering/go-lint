package a

import web "net/http"

// Mux returns a router.
func Mux() *web.ServeMux { return web.NewServeMux() } // want `net/http.s ServeMux serves HTTP` `net/http.s NewServeMux serves HTTP`

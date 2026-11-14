package dot

import . "net/http"

// Run serves the handler.
func run(handler Handler) error { return ListenAndServe(":1", handler) } // want `net/http.s Handler serves HTTP` `net/http.s ListenAndServe serves HTTP`

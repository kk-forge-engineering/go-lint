package log

import "net/http"

// Start serves the handler in each way net/http offers.
func Start(handler http.Handler, server *http.Server, mux *http.ServeMux) { // want `net/http.s Handler serves HTTP` `net/http.s Server serves HTTP` `net/http.s ServeMux serves HTTP`
	mux.Handle("/", handler)                      // want `net/http.s Handle serves HTTP`
	mux.HandleFunc("/f", nil)                     // want `net/http.s HandleFunc serves HTTP`
	_ = server.ListenAndServe()                   // want `net/http.s ListenAndServe serves HTTP`
	_ = http.ListenAndServeTLS(":0", "", "", nil) // want `net/http.s ListenAndServeTLS serves HTTP`
	_ = http.Serve(nil, nil)                      // want `net/http.s Serve serves HTTP`
	_ = http.ServeTLS(nil, nil, "", "")           // want `net/http.s ServeTLS serves HTTP`
}

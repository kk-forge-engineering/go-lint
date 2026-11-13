package log

import "net/http"

// Middleware logs each request.
func Middleware(next http.Handler) http.Handler { // want `net/http.s Handler serves HTTP` `net/http.s Handler serves HTTP`
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // want `net/http.s HandlerFunc serves HTTP` `net/http.s ResponseWriter serves HTTP`
		next.ServeHTTP(w, r)
	})
}

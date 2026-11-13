package start

import "net/http"

// Serve serves the app.
func Serve() error {
	mux := http.NewServeMux()
	return http.ListenAndServe(":8080", mux)
}

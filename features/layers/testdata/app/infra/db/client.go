package db

import (
	"context"
	"net/http"
)

// Fetch asks the server for its page, as a client.
func Fetch(ctx context.Context, client *http.Client) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	if err != nil {
		return nil, err
	}
	return client.Do(request)
}

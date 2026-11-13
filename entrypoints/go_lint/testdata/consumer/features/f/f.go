package f

import "net/http"

var client = http.DefaultClient

func init() {
	client = &http.Client{}
}

// Fetch asks the server for its page.
func Fetch() error {
	request, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	return response.Body.Close()
}

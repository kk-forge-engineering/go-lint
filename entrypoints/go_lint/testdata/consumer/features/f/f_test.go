package f_test

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestRequest(t *testing.T) {
	if _, err := http.NewRequest(http.MethodGet, "https://example.com", nil); err != nil {
		t.Fatal(err)
	}
}

func TestInsecureClient(t *testing.T) {
	config := &tls.Config{InsecureSkipVerify: true}
	if !config.InsecureSkipVerify {
		t.Fatal("the configuration checks certificates")
	}
}

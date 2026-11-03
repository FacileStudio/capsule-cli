package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeviceLoginFallsBackWithoutExchange(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	token, fallback, err := deviceLogin(server.URL)
	if err != nil || !fallback || token != "" {
		t.Fatalf("want fallback, got token %q fallback %v err %v", token, fallback, err)
	}
}

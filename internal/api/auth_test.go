package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/oidc/exchange" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"token":"tok"}`))
	}))
	defer server.Close()

	token, err := NewClient(server.URL).ExchangeCode(context.Background(), "code")
	if err != nil || token != "tok" {
		t.Fatalf("ExchangeCode = %q, %v", token, err)
	}
}

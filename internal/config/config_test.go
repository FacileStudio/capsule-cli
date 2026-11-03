package config

import "testing"

func TestServerPrecedence(t *testing.T) {
	t.Setenv("CAPSULE_SERVER_URL", "")
	t.Setenv("CAPSULE_URL", "")

	if got := (&Config{}).Server(); got != DefaultServerURL {
		t.Fatalf("no config and no env: got %q, want the default", got)
	}
	if got := (&Config{ServerURL: "https://stored.example/"}).Server(); got != "https://stored.example" {
		t.Fatalf("stored value: got %q", got)
	}
	t.Setenv("CAPSULE_URL", "https://env-url.example")
	if got := (&Config{ServerURL: "https://stored.example"}).Server(); got != "https://env-url.example" {
		t.Fatalf("CAPSULE_URL beats stored: got %q", got)
	}
	t.Setenv("CAPSULE_SERVER_URL", "https://env-server.example")
	if got := (&Config{}).Server(); got != "https://env-server.example" {
		t.Fatalf("CAPSULE_SERVER_URL beats CAPSULE_URL: got %q", got)
	}
}

package activity

import (
	"net/http/httptest"
	"testing"
)

func TestHashIPDeterministicAndPeppered(t *testing.T) {
	p1 := []byte("pepper-one")
	p2 := []byte("pepper-two")

	a := HashIP(p1, "203.0.113.5")
	b := HashIP(p1, "203.0.113.5")
	if a != b {
		t.Fatal("hashing must be deterministic for same pepper+ip")
	}
	if len(a) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(a))
	}
	if HashIP(p2, "203.0.113.5") == a {
		t.Fatal("different pepper must produce different hash")
	}
	if HashIP(p1, "203.0.113.6") == a {
		t.Fatal("different ip must produce different hash")
	}
}

func TestHashIPEmpty(t *testing.T) {
	if HashIP([]byte("p"), "") != "" {
		t.Fatal("empty ip should hash to empty string")
	}
	if HashIP([]byte("p"), "   ") != "" {
		t.Fatal("whitespace-only ip should hash to empty string")
	}
}

func TestClientIPPrefersCloudflareHeader(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.2")
	r.Header.Set("Cf-Connecting-Ip", "203.0.113.9")
	if got := ClientIP(r); got != "203.0.113.9" {
		t.Fatalf("ClientIP = %q, want Cf-Connecting-Ip value", got)
	}
}

func TestClientIPFallsBackToXFF(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.2")
	if got := ClientIP(r); got != "198.51.100.7" {
		t.Fatalf("ClientIP = %q, want first XFF hop", got)
	}
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.4:5678"
	if got := ClientIP(r); got != "192.0.2.4" {
		t.Fatalf("ClientIP = %q, want host from RemoteAddr", got)
	}
}

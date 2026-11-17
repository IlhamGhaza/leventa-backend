package usecase

import (
	"strings"
	"testing"
)

func TestNormalizeWhatsApp(t *testing.T) {
	cases := map[string]string{
		"0812-3456-7890":  "6281234567890",
		"+62 812 3456 78": "62812345678",
		"812345678":       "62812345678",
		"62812345678":     "62812345678",
	}
	for in, want := range cases {
		got, ok := NormalizeWhatsApp(in)
		if !ok || got != want {
			t.Errorf("%q -> %q ok=%v, want %q", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "abc", "123", "+1 202 555 0100"} {
		if _, ok := NormalizeWhatsApp(bad); ok {
			t.Errorf("%q should be invalid", bad)
		}
	}
}

func TestBuildWhatsAppURLEncoding(t *testing.T) {
	price := 2500000000.0
	u := BuildWhatsAppURL("6281234567890", "Rumah & Taman", "PROP-1", &price, "Jakarta Selatan")
	if !strings.HasPrefix(u, "https://wa.me/6281234567890?text=") {
		t.Fatalf("bad prefix: %s", u)
	}
	for _, bad := range []string{" ", "&T", "+"} {
		if strings.Contains(strings.TrimPrefix(u, "https://wa.me/6281234567890?text="), bad) {
			t.Errorf("unencoded %q in %s", bad, u)
		}
	}
	if !strings.Contains(u, "Rp%202.500.000.000") || !strings.Contains(u, "PROP-1") {
		t.Errorf("missing price/code: %s", u)
	}
}

func TestStatusBadgeAndPublicStatus(t *testing.T) {
	for s, want := range map[string]bool{"available": true, "reserved": true, "sold": true, "rented": true, "draft": false, "archived": false, "maintenance": false} {
		if publicStatus(s) != want {
			t.Errorf("publicStatus(%s) != %v", s, want)
		}
	}
}

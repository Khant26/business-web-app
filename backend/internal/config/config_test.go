package config

import (
	"net/http"
	"testing"
)

func TestParseLine(t *testing.T) {
	cases := []struct{ in, k, v string }{
		{"PORT=8080", "PORT", "8080"},
		{"APP_ENV=development            # comment", "APP_ENV", "development"},
		{"SMTP_HOST=                     # comment only", "SMTP_HOST", ""},
		{`SMTP_FROM="Sawyun Ecommerce <no-reply@example.com>"`, "SMTP_FROM", "Sawyun Ecommerce <no-reply@example.com>"},
		{"SMTP_USERNAME=", "SMTP_USERNAME", ""},
	}
	for _, c := range cases {
		k, v, ok := parseLine(c.in)
		if !ok || k != c.k || v != c.v {
			t.Errorf("parseLine(%q) = %q,%q,%v", c.in, k, v, ok)
		}
	}
	if _, _, ok := parseLine("# x"); ok {
		t.Error("comment parsed")
	}
}

func TestSplitList(t *testing.T) {
	got := splitList(" https://a.com/ , ,https://b.vercel.app")
	if len(got) != 2 || got[0] != "https://a.com" || got[1] != "https://b.vercel.app" {
		t.Fatalf("splitList = %q", got)
	}
}

func TestCookieSameSiteNoneNeedsSecure(t *testing.T) {
	t.Setenv("COOKIE_SAMESITE", "none")
	t.Setenv("COOKIE_SECURE", "false")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for SameSite=None without Secure")
	}
	t.Setenv("COOKIE_SECURE", "true")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.CookieSameSite != http.SameSiteNoneMode {
		t.Fatalf("SameSite = %v", c.CookieSameSite)
	}
}

package config

import "testing"

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

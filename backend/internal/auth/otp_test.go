package auth

import (
	"regexp"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	re := regexp.MustCompile(`^\d{6}$`)
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		c, err := GenerateCode()
		if err != nil || !re.MatchString(c) {
			t.Fatalf("bad code %q %v", c, err)
		}
		seen[c] = true
	}
	if len(seen) < 190 {
		t.Fatalf("codes not random enough: %d unique", len(seen))
	}
}

func TestHashVerify(t *testing.T) {
	h := HashCode("secret", "a@b.com", "123456")
	if len(h) != 64 {
		t.Fatal("expected hex sha256")
	}
	if !VerifyCode("secret", "a@b.com", "123456", h) {
		t.Fatal("valid code rejected")
	}
	for _, c := range []struct{ secret, email, code string }{
		{"secret", "a@b.com", "123457"},
		{"other", "a@b.com", "123456"},
		{"secret", "x@b.com", "123456"},
	} {
		if VerifyCode(c.secret, c.email, c.code, h) {
			t.Fatalf("unexpected match %+v", c)
		}
	}
}

func TestToken(t *testing.T) {
	tok, id, err := NewToken()
	if err != nil || len(tok) != 43 || HashToken(tok) != id || len(id) != 64 {
		t.Fatalf("bad token %q %q %v", tok, id, err)
	}
	if ShortID(id) != id[:16] {
		t.Fatal("short id")
	}
}

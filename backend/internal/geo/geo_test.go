package geo

import "testing"

func TestProvinces(t *testing.T) {
	if len(List()) != 77 {
		t.Fatalf("expected 77 provinces, got %d", len(List()))
	}
	for in, want := range map[string]string{"Bangkok": "Bangkok", "chiang mai": "Chiang Mai", "กรุงเทพมหานคร": "Bangkok"} {
		if got, ok := Canonical(in); !ok || got != want {
			t.Errorf("Canonical(%q) = %q,%v", in, got, ok)
		}
	}
	if _, ok := Canonical("Atlantis"); ok {
		t.Error("unknown province accepted")
	}
}

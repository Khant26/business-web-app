package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(5, time.Hour)
	l.now = func() time.Time { return now }

	for i := 0; i < 5; i++ {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	ok, wait := l.Allow("a")
	if ok {
		t.Fatal("6th request should be denied")
	}
	if wait < 11*time.Minute || wait > 13*time.Minute {
		t.Fatalf("unexpected retry-after %v", wait)
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Fatal("other key must be independent")
	}

	now = now.Add(12 * time.Minute)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("token should have refilled after 12 minutes")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Fatal("only one token should have refilled")
	}

	now = now.Add(2 * time.Hour)
	l.Cleanup()
	if l.Len() != 0 {
		t.Fatalf("cleanup should drop full buckets, have %d", l.Len())
	}
}

func TestOnePerMinute(t *testing.T) {
	now := time.Unix(0, 0)
	l := New(1, time.Minute)
	l.now = func() time.Time { return now }
	if ok, _ := l.Allow("x"); !ok {
		t.Fatal("first allowed")
	}
	if ok, w := l.Allow("x"); ok || w != time.Minute {
		t.Fatalf("second denied with 60s, got %v %v", ok, w)
	}
	now = now.Add(61 * time.Second)
	if ok, _ := l.Allow("x"); !ok {
		t.Fatal("allowed after a minute")
	}
}

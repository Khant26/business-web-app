package promptpay

import (
	"strings"
	"testing"
)

func TestCRC16CheckValue(t *testing.T) {
	// Standard CRC-16/CCITT-FALSE check value.
	if got := CRC16([]byte("123456789")); got != 0x29B1 {
		t.Fatalf("CRC16(123456789) = %04X, want 29B1", got)
	}
}

// Vectors from the reference implementation (github.com/dtinth/promptpay-qr, index.test.js).
func TestPayloadKnownVectors(t *testing.T) {
	cases := []struct {
		id     string
		amount int64
		want   string
	}{
		{"0801234567", 0, "00020101021129370016A000000677010111011300668012345675802TH530376463046197"},
		{"080-123-4567", 0, "00020101021129370016A000000677010111011300668012345675802TH530376463046197"},
		{"+66-89-123-4567", 0, "00020101021129370016A000000677010111011300668912345675802TH5303764630429C1"},
		{"1111111111111", 0, "00020101021129370016A000000677010111021311111111111115802TH530376463047B5A"},
		{"0123456789012", 0, "00020101021129370016A000000677010111021301234567890125802TH530376463040CBD"},
		{"012345678901234", 0, "00020101021129390016A00000067701011103150123456789012345802TH530376463049781"},
		{"000-000-0000", 422, "00020101021229370016A000000677010111011300660000000005802TH530376454044.226304E469"},
	}
	for _, c := range cases {
		got, err := Payload(c.id, c.amount)
		if err != nil {
			t.Fatalf("%s: %v", c.id, err)
		}
		if got != c.want {
			t.Errorf("Payload(%q, %d)\n got %s\nwant %s", c.id, c.amount, got, c.want)
		}
	}
}

func TestPayloadAmountFormatting(t *testing.T) {
	got, _ := Payload("0812345678", 5)
	if !strings.Contains(got, "54040.05") {
		t.Fatalf("amount formatting wrong: %s", got)
	}
	got, _ = Payload("0812345678", 123450)
	if !strings.Contains(got, "54071234.50") {
		t.Fatalf("amount formatting wrong: %s", got)
	}
}

func TestInvalidID(t *testing.T) {
	for _, id := range []string{"", "12345", "abc", "081234567"} {
		if _, err := Payload(id, 100); err == nil {
			t.Errorf("expected error for %q", id)
		}
	}
}

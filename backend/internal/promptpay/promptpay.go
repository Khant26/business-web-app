// Package promptpay generates Thai PromptPay EMVCo QR payloads.
package promptpay

import (
	"errors"
	"fmt"
	"strings"
)

const aid = "A000000677010111"

// ErrInvalidID is returned for an unsupported PromptPay ID.
var ErrInvalidID = errors.New("promptpay: id must be a 10-digit phone, 13-digit tax ID or 15-digit e-wallet ID")

func tlv(tag, value string) string { return fmt.Sprintf("%s%02d%s", tag, len(value), value) }

// NormalizeID strips non-digits and validates the ID, returning the account sub-tag and value.
func NormalizeID(id string) (tag, value string, err error) {
	var b strings.Builder
	for _, r := range id {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	switch {
	case len(d) == 10 && d[0] == '0': // mobile phone → 0066 + 9 digits
		return "01", "0066" + d[1:], nil
	case len(d) == 11 && strings.HasPrefix(d, "66"): // already international
		return "01", "00" + d, nil
	case len(d) == 13:
		return "02", d, nil
	case len(d) == 15:
		return "03", d, nil
	}
	return "", "", ErrInvalidID
}

// Payload builds the EMVCo payload. amountSatang > 0 makes a dynamic QR with amount.
func Payload(id string, amountSatang int64) (string, error) {
	tag, val, err := NormalizeID(id)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(tlv("00", "01"))
	if amountSatang > 0 {
		b.WriteString(tlv("01", "12"))
	} else {
		b.WriteString(tlv("01", "11"))
	}
	b.WriteString(tlv("29", tlv("00", aid)+tlv(tag, val)))
	// Tag order follows the reference promptpay-qr implementation (58, 53, 54).
	b.WriteString(tlv("58", "TH"))
	b.WriteString(tlv("53", "764"))
	if amountSatang > 0 {
		b.WriteString(tlv("54", formatAmount(amountSatang)))
	}
	b.WriteString("6304")
	s := b.String()
	return s + fmt.Sprintf("%04X", CRC16([]byte(s))), nil
}

func formatAmount(satang int64) string {
	s := fmt.Sprintf("%d.%02d", satang/100, satang%100)
	return s
}

// CRC16 computes CRC-16/CCITT-FALSE (poly 0x1021, init 0xFFFF).
func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, c := range data {
		crc ^= uint16(c) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

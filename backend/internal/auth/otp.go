package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// OTP parameters.
const (
	OTPTTL         = 5 * time.Minute
	OTPMaxAttempts = 5
)

// GenerateCode returns a uniformly random 6-digit code.
func GenerateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashCode returns HMAC-SHA256(secret, email|code) hex-encoded.
func HashCode(secret, email, code string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(email))
	m.Write([]byte{'|'})
	m.Write([]byte(code))
	return hex.EncodeToString(m.Sum(nil))
}

// VerifyCode checks code against a stored hash in constant time.
func VerifyCode(secret, email, code, hash string) bool {
	return hmac.Equal([]byte(HashCode(secret, email, code)), []byte(hash))
}

// NewToken returns a random 32-byte session token (base64url) and its sha256 hex id.
func NewToken() (token, id string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token), nil
}

// HashToken returns the sha256 hex of a raw session token (the stored session id).
func HashToken(token string) string {
	s := sha256.Sum256([]byte(token))
	return hex.EncodeToString(s[:])
}

// ShortID is the opaque public id of a session (prefix of its stored hash).
func ShortID(sessionID string) string {
	if len(sessionID) > 16 {
		return sessionID[:16]
	}
	return sessionID
}

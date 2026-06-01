package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// sha512Hex returns the hex-encoded SHA-512 digest of s (Midtrans signatures).
func sha512Hex(s string) string {
	sum := sha512.Sum512([]byte(s))
	return hex.EncodeToString(sum[:])
}

// hmacSHA256Hex returns the hex-encoded HMAC-SHA256 of message keyed by key
// (Tripay signatures).
func hmacSHA256Hex(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// basicAuth builds an HTTP Basic auth header value for "username:" style keys.
func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

// secureEqual compares two strings in constant time.
func secureEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// guestEmail produces a deterministic fallback email when the customer did not
// provide one (some gateways require a payer email).
func guestEmail(email, orderNumber string) string {
	if e := strings.TrimSpace(email); e != "" {
		return e
	}
	return "guest-" + strings.ToLower(orderNumber) + "@hotspot.local"
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

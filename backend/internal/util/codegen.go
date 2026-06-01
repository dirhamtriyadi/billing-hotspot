// Package util holds small, dependency-light helpers shared across the backend.
package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// codeAlphabet excludes visually ambiguous characters (0/O, 1/I/L) so printed
// vouchers are easy to read and type at the hotspot login page.
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// GenerateCode returns a cryptographically-random voucher code of the given
// length, optionally prefixed. The prefix is upper-cased.
func GenerateCode(prefix string, length int) (string, error) {
	if length <= 0 {
		length = 8
	}
	max := big.NewInt(int64(len(codeAlphabet)))
	var sb strings.Builder
	sb.WriteString(strings.ToUpper(strings.TrimSpace(prefix)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		sb.WriteByte(codeAlphabet[n.Int64()])
	}
	return sb.String(), nil
}

// OrderNumber returns a human-readable, time-ordered, unique-enough order id,
// e.g. "INV-20260530-7K3QF9".
func OrderNumber() string {
	suffix, _ := GenerateCode("", 6)
	return fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102"), suffix)
}

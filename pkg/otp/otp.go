// Package otp: numeric one-time password generation backed by crypto/rand.
package otp

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

const (
	DefaultLength = 6
	minLength     = 4
	maxLength     = 10
)

type Generator struct {
	length int
	bound  *big.Int
	format string
}

// Compile-time check against the domain contract.
var _ domain.OTPGenerator = (*Generator)(nil)

// NewGenerator: a length outside [4, 10] falls back to DefaultLength.
func NewGenerator(length int) *Generator {
	if length < minLength || length > maxLength {
		length = DefaultLength
	}
	return &Generator{
		length: length,
		bound:  new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil),
		format: fmt.Sprintf("%%0%dd", length),
	}
}

// Generate returns a zero-padded numeric code, e.g. "048213".
func (g *Generator) Generate() (string, error) {
	// crypto/rand, never math/rand: a predictable OTP defeats the whole flow.
	n, err := rand.Int(rand.Reader, g.bound)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf(g.format, n), nil
}

func (g *Generator) Length() int { return g.length }

// Equal compares two codes in constant time so response latency leaks nothing.
func Equal(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

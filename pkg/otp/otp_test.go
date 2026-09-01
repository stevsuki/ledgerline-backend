package otp_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/pkg/otp"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		length     int
		wantLength int
	}{
		{"default length", 6, 6},
		{"custom length", 8, 8},
		{"too short falls back to default", 2, otp.DefaultLength},
		{"too long falls back to default", 20, otp.DefaultLength},
		{"zero falls back to default", 0, otp.DefaultLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, err := otp.NewGenerator(tt.length).Generate()

			require.NoError(t, err)
			assert.Len(t, code, tt.wantLength)
			assert.Regexp(t, regexp.MustCompile(`^[0-9]+$`), code)
		})
	}
}

func TestGenerateKeepsLeadingZeros(t *testing.T) {
	t.Parallel()

	// Roughly 1 in 10 codes starts with a zero, so 500 draws make this reliable.
	g := otp.NewGenerator(6)
	seen := false
	for range 500 {
		code, err := g.Generate()
		require.NoError(t, err)
		require.Len(t, code, 6, "length must stay fixed even when the value is small")
		if code[0] == '0' {
			seen = true
		}
	}
	assert.True(t, seen, "expected at least one code starting with a zero")
}

func TestGenerateIsNotRepeating(t *testing.T) {
	t.Parallel()

	g := otp.NewGenerator(6)
	seen := make(map[string]bool, 200)
	for range 200 {
		code, err := g.Generate()
		require.NoError(t, err)
		seen[code] = true
	}
	// A predictable source would collapse into a handful of values.
	assert.Greater(t, len(seen), 150)
}

func TestEqual(t *testing.T) {
	t.Parallel()

	assert.True(t, otp.Equal("048213", "048213"))
	assert.False(t, otp.Equal("048213", "48213"))
	assert.False(t, otp.Equal("048213", "048214"))
	assert.False(t, otp.Equal("048213", ""))
}

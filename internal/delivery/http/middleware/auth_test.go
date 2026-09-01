package middleware

import "testing"

func TestExtractToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"format standar RFC 6750", "Bearer abc.def.ghi", "abc.def.ghi"},
		{"spasi berlebih dirapikan", "  Bearer   abc.def.ghi  ", "abc.def.ghi"},
		{"a bare token is rejected", "abc.def.ghi", ""},
		{"a lowercase prefix is rejected", "bearer abc.def.ghi", ""},
		{"another scheme is rejected", "Basic YWRtaW46YWRtaW4=", ""},
		{"empty header", "", ""},
		{"only the word Bearer", "Bearer ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := extractToken(tt.header); got != tt.want {
				t.Errorf("extractToken(%q) = %q, mau %q", tt.header, got, tt.want)
			}
		})
	}
}

package auth

import "testing"

func TestRedactToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{name: "short", token: "abcd", want: "***"},
		{name: "long", token: "1234567890abcdef", want: "1234...cdef"},
		{name: "trim", token: "  1234567890abcdef  ", want: "1234...cdef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactToken(tt.token); got != tt.want {
				t.Fatalf("RedactToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeCookieToken(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "value only", input: "abc123", want: "abc123"},
		{name: "named t cookie", input: " t=abc123 ", want: "abc123"},
		{name: "full header rejected", input: "Cookie: t=abc123", wantErr: true},
		{name: "multiple cookies rejected", input: "t=abc123; other=secret", wantErr: true},
		{name: "embedded whitespace rejected", input: "abc 123", wantErr: true},
		{name: "empty rejected", input: " t= ", wantErr: true},
		{name: "unexpected key rejected", input: "other=abc123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCookieToken(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeCookieToken() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeCookieToken() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeCookieToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

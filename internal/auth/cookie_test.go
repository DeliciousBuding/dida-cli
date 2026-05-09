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

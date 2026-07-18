package cli

import "testing"

func TestNormalizeTaskTime(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		tz   string
		want string
	}{
		{"", "Asia/Shanghai", ""},
		{"2026-07-18T20:00:00+08:00", "Asia/Shanghai", "2026-07-18T12:00:00.000+0000"},
		{"2026-07-18T12:00:00.000+0000", "Asia/Shanghai", "2026-07-18T12:00:00.000+0000"},
		{"2026-07-18 20:00", "Asia/Shanghai", "2026-07-18T12:00:00.000+0000"},
		{"2026-07-18", "Asia/Shanghai", "2026-07-17T16:00:00.000+0000"},
	}
	for _, tc := range cases {
		got, err := normalizeTaskTime(tc.in, tc.tz)
		if err != nil {
			t.Fatalf("normalizeTaskTime(%q) err = %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("normalizeTaskTime(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeTaskTimeInvalid(t *testing.T) {
	t.Parallel()
	if _, err := normalizeTaskTime("not-a-time", "Asia/Shanghai"); err == nil {
		t.Fatalf("expected error for invalid time")
	}
}

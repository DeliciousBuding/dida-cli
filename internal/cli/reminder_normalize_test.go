package cli

import "testing"

func TestNormalizeReminder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"30m", "TRIGGER:-PT30M"},
		{"15min", "TRIGGER:-PT15M"},
		{"1h", "TRIGGER:-PT1H"},
		{"1h30m", "TRIGGER:-PT1H30M"},
		{"at-start", "TRIGGER:PT0S"},
		{"0", "TRIGGER:PT0S"},
		{"TRIGGER:-PT30M", "TRIGGER:-PT30M"},
		{"TRIGGER:P0DT9H0M0S", "TRIGGER:P0DT9H0M0S"},
		{"-PT15M", "TRIGGER:-PT15M"},
	}
	for _, tc := range cases {
		got, err := normalizeReminder(tc.in)
		if err != nil {
			t.Fatalf("normalizeReminder(%q) err = %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("normalizeReminder(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeReminderInvalid(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "soon", "TRIGGER:", "xyz"} {
		if _, err := normalizeReminder(in); err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}

func TestNormalizeRemindersDedup(t *testing.T) {
	t.Parallel()
	got, err := normalizeReminders([]string{"30m", "TRIGGER:-PT30M", "15m"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %#v", len(got), got)
	}
}

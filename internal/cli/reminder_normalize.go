package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// humanReminder re matches optional leading '-', then duration chunks.
	// Examples: 30m, 15min, 1h, 1h30m, -15m, 0, at-start
	humanDurationRe = regexp.MustCompile(`(?i)^-?(?:(\d+)\s*h(?:ours?)?)?\s*(?:(\d+)\s*m(?:in(?:utes?)?)?)?\s*(?:(\d+)\s*s(?:ec(?:onds?)?)?)?$`)
)

// normalizeReminder converts a human or TRIGGER reminder into a Dida TRIGGER string.
// "Before due/start" is the default semantic for bare durations (30m => TRIGGER:-PT30M).
func normalizeReminder(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty reminder value")
	}
	upper := strings.ToUpper(value)
	switch strings.ToLower(value) {
	case "at-start", "atstart", "0", "none-delay", "on-time", "ontime":
		return "TRIGGER:PT0S", nil
	}
	if strings.HasPrefix(upper, "TRIGGER:") {
		body := strings.TrimSpace(value[len("TRIGGER:"):])
		if body == "" {
			return "", fmt.Errorf("invalid reminder %q: empty TRIGGER body", value)
		}
		// Accept common ISO-8601 duration forms already used by Dida.
		if !strings.HasPrefix(strings.ToUpper(body), "P") && !strings.HasPrefix(strings.ToUpper(body), "-P") {
			return "", fmt.Errorf("invalid TRIGGER reminder %q: expected ISO-8601 duration after TRIGGER prefix", value)
		}
		return "TRIGGER:" + body, nil
	}
	// Bare ISO duration without TRIGGER prefix.
	if strings.HasPrefix(upper, "P") || strings.HasPrefix(upper, "-P") {
		return "TRIGGER:" + value, nil
	}

	neg := strings.HasPrefix(value, "-")
	raw := strings.TrimSpace(strings.TrimPrefix(value, "-"))
	raw = strings.TrimSpace(strings.TrimPrefix(raw, "+"))
	if raw == "" {
		return "", fmt.Errorf("invalid reminder %q", value)
	}

	m := humanDurationRe.FindStringSubmatch(raw)
	if m == nil {
		return "", fmt.Errorf("invalid reminder %q; use 30m, 15min, 1h, 1h30m, at-start, or TRIGGER:-PT30M", value)
	}
	hours := atoiDefault(m[1])
	mins := atoiDefault(m[2])
	secs := atoiDefault(m[3])
	if hours == 0 && mins == 0 && secs == 0 && raw != "0" && !strings.EqualFold(raw, "0s") && !strings.EqualFold(raw, "0m") {
		// Pattern matched empty groups only — reject bare junk.
		if !regexp.MustCompile(`(?i)^\d`).MatchString(raw) {
			return "", fmt.Errorf("invalid reminder %q", value)
		}
	}
	// Default: duration means "before" => negative TRIGGER offset.
	// Leading '-' on human form is accepted as the same (still before).
	_ = neg
	if hours == 0 && mins == 0 && secs == 0 {
		return "TRIGGER:PT0S", nil
	}
	// Prefer compact PT form for OpenAPI compatibility observed in production.
	return "TRIGGER:-" + formatISODuration(hours, mins, secs), nil
}

func formatISODuration(hours, mins, secs int) string {
	body := "PT"
	if hours > 0 {
		body += fmt.Sprintf("%dH", hours)
	}
	if mins > 0 {
		body += fmt.Sprintf("%dM", mins)
	}
	if secs > 0 || (hours == 0 && mins == 0) {
		body += fmt.Sprintf("%dS", secs)
	}
	return body
}

func normalizeReminders(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, v := range values {
		n, err := normalizeReminder(v)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}

func atoiDefault(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

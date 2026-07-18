package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
)

// didaWireTimeUTC is the timestamp layout observed to work with Web batch/task
// and OpenAPI task writes (UTC with millisecond precision, offset without colon).
const didaWireTimeUTC = "2006-01-02T15:04:05.000+0000"

// normalizeTaskTime converts a user-supplied --start/--due value into the Dida
// wire format in UTC. Empty input is left empty. When the value has no zone,
// it is interpreted in defaultTZ (IANA, e.g. Asia/Shanghai).
func normalizeTaskTime(value string, defaultTZ string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	loc := time.Local
	if strings.TrimSpace(defaultTZ) != "" {
		if loaded, err := time.LoadLocation(defaultTZ); err == nil {
			loc = loaded
		}
	}

	// Zoneless forms first so YYYY-MM-DD and local wall times use defaultTZ.
	localLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range localLayouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil && t.Format(layout) == value {
			return t.UTC().Format(didaWireTimeUTC), nil
		}
	}

	if t, ok := model.ParseDidaTime(value); ok {
		return t.UTC().Format(didaWireTimeUTC), nil
	}

	altLayouts := []string{
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05.000Z07:00",
	}
	for _, layout := range altLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC().Format(didaWireTimeUTC), nil
		}
	}
	return "", fmt.Errorf("invalid time %q; use RFC3339, YYYY-MM-DD[ HH:MM], or Dida wire format (e.g. 2026-07-18T12:00:00.000+0000)", value)
}

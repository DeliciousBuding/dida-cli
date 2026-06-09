package cli

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func parseIntStrict(value string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(value))
}

func parseInt64Strict(value string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(value), 10, 64)
}

func parseFloatStrict(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, fmt.Errorf("value must be finite")
	}
	return parsed, nil
}

func validateIDArg(name string, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s id is required", name)
	}
	if strings.HasPrefix(trimmed, "-") {
		return fmt.Errorf("%s id must not start with '-'", name)
	}
	return nil
}

func parseIDValue(args []string, index int, name string) (string, error) {
	if index >= len(args) {
		return "", fmt.Errorf("%s id is required", name)
	}
	value := args[index]
	if err := validateIDArg(name, value); err != nil {
		return "", err
	}
	return value, nil
}

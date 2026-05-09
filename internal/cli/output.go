package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type envelope struct {
	OK      bool      `json:"ok"`
	Command string    `json:"command"`
	Meta    any       `json:"meta,omitempty"`
	Data    any       `json:"data,omitempty"`
	Error   *cliError `json:"error,omitempty"`
}

type cliError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

func fail(command string, message string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	return failTyped(command, "", message, "", jsonOut, stdout, stderr)
}

func missingAuth(command string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	return failTyped(command, "auth", "missing cookie auth", "run: dida auth login", jsonOut, stdout, stderr)
}

func failTyped(command string, errType string, message string, hint string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if jsonOut {
		_ = writeJSON(stdout, envelope{OK: false, Command: command, Error: &cliError{Type: errType, Message: message, Hint: hint}})
		return 1
	}
	fmt.Fprintf(stderr, "dida: %s\n", message)
	if hint != "" {
		fmt.Fprintf(stderr, "hint: %s\n", hint)
	}
	return 1
}

func writeJSON(w io.Writer, value any) int {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return 1
	}
	return 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

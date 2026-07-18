package cli

import (
	"context"
	"fmt"
	"io"
	"time"
)

func runAccount(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAccountHelp(stdout)
		return 0
	}
	switch args[0] {
	case "whoami":
		return runAccountWhoami(args[1:], jsonOut, stdout, stderr)
	case "verify":
		return runAccountVerify(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("account", fmt.Sprintf("unknown account command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func printAccountHelp(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  dida account whoami [--json]
  dida account verify [--json]

Report and refresh cross-channel account identity bindings.
verify calls Web user status/profile + OpenAPI project list and stores
non-secret fingerprints in ~/.dida-cli/identity.json.
Multi-channel reminder writes require matching identities.`)
}

func runAccountWhoami(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 {
		return failTyped("account whoami", "validation", fmt.Sprintf("unknown flag %q", args[0]), "run: dida account --help", jsonOut, stdout, stderr)
	}
	data := accountWhoami()
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "account whoami", Data: data})
	}
	fmt.Fprintf(stdout, "Identity path: %v\n", data["identity_path"])
	fmt.Fprintf(stdout, "Match: %v (%v)\n", data["identity_match"], data["match_reason"])
	return 0
}

func runAccountVerify(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 {
		return failTyped("account verify", "validation", fmt.Sprintf("unknown flag %q", args[0]), "run: dida account --help", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	data, err := verifyAccountIdentities(ctx)
	if err != nil {
		return failTyped("account verify", "api", err.Error(), "run: dida auth status --verify --json && dida openapi status --json", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "account verify", Data: data})
	}
	fmt.Fprintf(stdout, "Identity path: %v\n", data["identity_path"])
	fmt.Fprintf(stdout, "Match: %v (%v)\n", data["identity_match"], data["match_reason"])
	return 0
}

package cli

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/openapi"
)

func runOpenAPI(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printOpenAPIHelp(stdout)
		return 0
	}
	switch args[0] {
	case "doctor":
		return runOpenAPIDoctor(jsonOut, stdout, stderr)
	case "status":
		return runOpenAPIStatus(jsonOut, stdout, stderr)
	case "logout":
		return runOpenAPILogout(jsonOut, stdout, stderr)
	case "login":
		return runOpenAPILogin(args[1:], jsonOut, stdout, stderr)
	case "auth-url":
		return runOpenAPIAuthURL(args[1:], jsonOut, stdout, stderr)
	case "exchange-code":
		return runOpenAPIExchangeCode(args[1:], jsonOut, stdout, stderr)
	case "listen-callback":
		return runOpenAPIListenCallback(args[1:], jsonOut, stdout, stderr)
	case "project":
		return runOpenAPIProject(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("openapi", fmt.Sprintf("unknown openapi command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runOpenAPIDoctor(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	clientID, err := openapi.ResolveClientID("")
	clientSecret, err2 := openapi.ResolveClientSecret("")
	tokenStatus := openapi.TokenStatus()
	data := map[string]any{
		"client_id_available":     err == nil && clientID != "",
		"client_secret_available": err2 == nil && clientSecret != "",
		"token":                   tokenStatus,
		"base_url":                openapi.DefaultAPIBaseURL,
		"auth_url":                openapi.DefaultAuthBaseURL + "/oauth/authorize",
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi doctor", Data: data})
	}
	fmt.Fprintf(stdout, "OpenAPI client id: %v\n", data["client_id_available"])
	fmt.Fprintf(stdout, "OpenAPI client secret: %v\n", data["client_secret_available"])
	fmt.Fprintf(stdout, "OpenAPI token: %v\n", tokenStatus["available"])
	return 0
}

func runOpenAPIStatus(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	data := map[string]any{"token": openapi.TokenStatus()}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi status", Data: data})
	}
	fmt.Fprintf(stdout, "OpenAPI token available: %v\n", data["token"].(map[string]any)["available"])
	return 0
}

func runOpenAPILogout(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if err := openapi.ClearToken(); err != nil {
		return failTyped("openapi logout", "auth", err.Error(), "", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi logout", Data: map[string]any{"token_cleared": true}})
	}
	fmt.Fprintln(stdout, "OpenAPI token cleared.")
	return 0
}

func runOpenAPILogin(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	redirectURI, scope, state, host, port, timeout, noOpen, err := parseOpenAPILoginFlags(args)
	if err != nil {
		return failTyped("openapi login", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	clientID, err := openapi.ResolveClientID("")
	if err != nil {
		return failTyped("openapi login", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_ID", jsonOut, stdout, stderr)
	}
	clientSecret, err := openapi.ResolveClientSecret("")
	if err != nil {
		return failTyped("openapi login", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_SECRET", jsonOut, stdout, stderr)
	}
	redirectURI = fmt.Sprintf("http://%s:%d/callback", host, port)
	authURL := openapi.AuthorizationURL(clientID, redirectURI, scope, state)
	type callbackResult struct {
		code  string
		state string
	}
	codeCh := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		select {
		case codeCh <- callbackResult{code: r.URL.Query().Get("code"), state: r.URL.Query().Get("state")}:
		default:
		}
		_, _ = w.Write([]byte("DidaCLI OpenAPI callback received. You can return to the terminal."))
	})
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: mux}
	go func() { _ = server.ListenAndServe() }()
	if !noOpen {
		_ = openBrowserURL(authURL)
	}
	if jsonOut {
		_ = writeJSON(stdout, envelope{OK: true, Command: "openapi login", Data: map[string]any{
			"authorization_url": authURL,
			"redirect_uri":      redirectURI,
			"state":             state,
			"scope":             scope,
			"waiting":           true,
		}})
	} else {
		fmt.Fprintln(stdout, "Open this URL in a browser and finish authorization:")
		fmt.Fprintln(stdout, authURL)
	}
	select {
	case result := <-codeCh:
		_ = server.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		token, err := openapi.ExchangeCode(ctx, clientID, clientSecret, result.code, redirectURI, scope)
		if err != nil {
			return failTyped("openapi login", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if err := openapi.SaveToken(token); err != nil {
			return failTyped("openapi login", "auth", err.Error(), "", jsonOut, stdout, stderr)
		}
		data := map[string]any{"saved": true, "state": result.state, "token": openapi.TokenStatus()}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi login", Data: data})
		}
		fmt.Fprintln(stdout, "OpenAPI token saved.")
		return 0
	case <-time.After(timeout):
		_ = server.Close()
		return failTyped("openapi login", "timeout", "timed out waiting for OAuth callback", "", jsonOut, stdout, stderr)
	}
}

func runOpenAPIAuthURL(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	redirectURI, scope, state, err := parseOpenAPIAuthURLFlags(args)
	if err != nil {
		return failTyped("openapi auth-url", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	clientID, err := openapi.ResolveClientID("")
	if err != nil {
		return failTyped("openapi auth-url", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_ID", jsonOut, stdout, stderr)
	}
	url := openapi.AuthorizationURL(clientID, redirectURI, scope, state)
	data := map[string]any{"authorization_url": url, "redirect_uri": redirectURI, "scope": scope, "state": state}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi auth-url", Data: data})
	}
	fmt.Fprintln(stdout, url)
	return 0
}

func runOpenAPIExchangeCode(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	code, redirectURI, scope, err := parseOpenAPIExchangeFlags(args)
	if err != nil {
		return failTyped("openapi exchange-code", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	clientID, err := openapi.ResolveClientID("")
	if err != nil {
		return failTyped("openapi exchange-code", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_ID", jsonOut, stdout, stderr)
	}
	clientSecret, err := openapi.ResolveClientSecret("")
	if err != nil {
		return failTyped("openapi exchange-code", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_SECRET", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := openapi.ExchangeCode(ctx, clientID, clientSecret, code, redirectURI, scope)
	if err != nil {
		return failTyped("openapi exchange-code", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	if err := openapi.SaveToken(token); err != nil {
		return failTyped("openapi exchange-code", "auth", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"saved": true, "token": openapi.TokenStatus()}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi exchange-code", Data: data})
	}
	fmt.Fprintln(stdout, "OpenAPI token saved.")
	return 0
}

func runOpenAPIListenCallback(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	host, port, err := parseOpenAPIListenFlags(args)
	if err != nil {
		return failTyped("openapi listen-callback", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	redirectURI := fmt.Sprintf("http://%s:%d/callback", host, port)
	codeCh := make(chan map[string]string, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		values := map[string]string{
			"code":  r.URL.Query().Get("code"),
			"state": r.URL.Query().Get("state"),
		}
		select {
		case codeCh <- values:
		default:
		}
		_, _ = w.Write([]byte("OpenAPI callback received. You can return to the CLI."))
	})
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: mux}
	go func() { _ = server.ListenAndServe() }()
	select {
	case values := <-codeCh:
		_ = server.Close()
		data := map[string]any{"redirect_uri": redirectURI, "code": values["code"], "state": values["state"]}
		if jsonOut {
			_ = writeJSON(stdout, envelope{OK: true, Command: "openapi listen-callback", Data: data})
			return 0
		}
		fmt.Fprintf(stdout, "Code: %s\nState: %s\n", values["code"], values["state"])
		return 0
	case <-time.After(10 * time.Minute):
		_ = server.Close()
		return failTyped("openapi listen-callback", "timeout", "timed out waiting for OAuth callback", "", jsonOut, stdout, stderr)
	}
}

func runOpenAPIProject(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "list" {
		return failTyped("openapi project", "validation", "usage: dida openapi project list", "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	token, err := openapi.LoadToken()
	if err != nil {
		return failTyped("openapi project list", "auth", err.Error(), "run: dida openapi auth-url and dida openapi exchange-code first", jsonOut, stdout, stderr)
	}
	client := openapi.NewClient(token.AccessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	projects, err := client.Projects(ctx)
	if err != nil {
		return failTyped("openapi project list", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	meta := map[string]any{"count": len(projects)}
	data := map[string]any{"projects": projects}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi project list", Meta: meta, Data: data})
	}
	printMapList(stdout, projects, "openapi projects")
	return 0
}

func parseOpenAPIAuthURLFlags(args []string) (string, string, string, error) {
	redirectURI := "http://127.0.0.1:17890/callback"
	scope := openapi.DefaultScopes
	state := fmt.Sprintf("dida-%d-%d", time.Now().Unix(), rand.Intn(100000))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--redirect-uri":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--redirect-uri requires a value")
			}
			redirectURI = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--scope requires a value")
			}
			scope = args[i+1]
			i++
		case "--state":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--state requires a value")
			}
			state = args[i+1]
			i++
		default:
			return "", "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return redirectURI, scope, state, nil
}

func parseOpenAPIExchangeFlags(args []string) (string, string, string, error) {
	code := ""
	redirectURI := "http://127.0.0.1:17890/callback"
	scope := openapi.DefaultScopes
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--code":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--code requires a value")
			}
			code = args[i+1]
			i++
		case "--redirect-uri":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--redirect-uri requires a value")
			}
			redirectURI = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--scope requires a value")
			}
			scope = args[i+1]
			i++
		default:
			return "", "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if code == "" {
		return "", "", "", fmt.Errorf("missing --code")
	}
	return code, redirectURI, scope, nil
}

func parseOpenAPIListenFlags(args []string) (string, int, error) {
	host := "127.0.0.1"
	port := 17890
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--host requires a value")
			}
			host = args[i+1]
			i++
		case "--port":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--port requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &port); err != nil || port <= 0 {
				return "", 0, fmt.Errorf("--port must be a positive integer")
			}
			i++
		default:
			return "", 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return host, port, nil
}

func parseOpenAPILoginFlags(args []string) (string, string, string, string, int, time.Duration, bool, error) {
	redirectURI := ""
	scope := openapi.DefaultScopes
	state := fmt.Sprintf("dida-%d-%d", time.Now().Unix(), rand.Intn(100000))
	host := "127.0.0.1"
	port := 17890
	timeout := 10 * time.Minute
	noOpen := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--redirect-uri":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--redirect-uri requires a value")
			}
			redirectURI = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--scope requires a value")
			}
			scope = args[i+1]
			i++
		case "--state":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--state requires a value")
			}
			state = args[i+1]
			i++
		case "--host":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--host requires a value")
			}
			host = args[i+1]
			i++
		case "--port":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--port requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &port); err != nil || port <= 0 {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--port must be a positive integer")
			}
			i++
		case "--timeout":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--timeout requires seconds")
			}
			var seconds int
			if _, err := fmt.Sscanf(args[i+1], "%d", &seconds); err != nil || seconds <= 0 {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--timeout must be a positive integer")
			}
			timeout = time.Duration(seconds) * time.Second
			i++
		case "--no-open":
			noOpen = true
		default:
			return "", "", "", "", 0, 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return redirectURI, scope, state, host, port, timeout, noOpen, nil
}

func openBrowserURL(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}

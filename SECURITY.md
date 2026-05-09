# Security Policy

## Supported Versions

The `main` branch is the active development line.

## Reporting a Vulnerability

Please open a private security advisory on GitHub or contact the maintainer privately.

Do not include live Dida365 cookies, bearer tokens, private task exports, or full response dumps in public issues.

## Credential Handling

DidaCLI stores credentials under `~/.dida-cli/` and redacts token previews in CLI output. The repository ignores local env files, private data, and generated binaries.

Agents and scripts should import cookies through stdin or browser capture. They should never ask users to paste secrets into chat.

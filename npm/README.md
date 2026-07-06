# DidaCLI

JSON-first CLI for Dida365 / TickTick task automation.

This npm package installs the `dida` command by downloading the matching platform binary from the project's GitHub Release. The binary verifies release checksums during install.

## Install

```bash
npm install -g @delicious233/dida-cli
```

## Quick Start

```bash
dida version
dida doctor --json
dida auth cookie set --token-stdin
dida task today --compact --json
```

Use `dida --help` and `dida schema list --json` to inspect available commands.

## Links

- Repository: https://github.com/DeliciousBuding/dida-cli
- Documentation: https://deliciousbuding.github.io/dida-cli/
- Releases: https://github.com/DeliciousBuding/dida-cli/releases
- Issues: https://github.com/DeliciousBuding/dida-cli/issues

## Security

Do not paste cookies or tokens into shell history. Prefer `--token-stdin` for local auth setup. DidaCLI redacts known token patterns in command output and release automation verifies package provenance.

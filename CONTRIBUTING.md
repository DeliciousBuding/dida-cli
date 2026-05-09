# Contributing

Thanks for considering a contribution.

## Development

```bash
go test ./...
go build -o bin/dida ./cmd/dida
```

## Pull Request Checklist

- Add or update tests for changed command behavior.
- Keep JSON envelopes stable.
- Do not commit cookies, tokens, raw private response dumps, or private fixtures.
- Document new Web API assumptions in `docs/web-api.md` or `docs/research/api-surfaces.md`.
- Keep raw non-GET writes out of the CLI unless they are wrapped as a named, tested command.

## Command Design

- Prefer resource commands such as `task create` over generic request commands.
- Reads should be directly runnable.
- Normal writes may run directly and should support `--dry-run`.
- Destructive writes must require `--yes`.
- JSON errors should include `type`, `message`, and a useful `hint`.

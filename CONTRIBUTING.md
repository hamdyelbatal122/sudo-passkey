# Contributing to Passkey-Sudo

Thanks for your interest in contributing.

## Development setup

1. Install Go 1.23+
2. Clone the repository
3. Run `make tidy`
4. Run `make test`

## Pull request guidelines

- Keep changes focused and small
- Add tests for new behavior when possible
- Update docs for user-facing changes
- Use clear commit messages

## Local quality checks

```bash
make fmt
make vet
make test
```

## Reporting issues

Open a GitHub issue with:

- OS and version
- Go version
- Reproduction steps
- Logs or error output

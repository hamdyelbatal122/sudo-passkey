# Architecture

Passkey-Sudo consists of four main layers:

1. CLI layer (`internal/cli`)
2. Configuration layer (`internal/config`)
3. WebAuthn flow server (`internal/webauthnserver`)
4. Privileged command gateway (`internal/gate`)

## Execution flow

1. User runs `passkey-sudo run -- <cmd>`
2. CLI loads local config from `~/.config/passkey-sudo/config.json`
3. A local HTTP server is started at configured RP origin (`127.0.0.1` by default)
4. Browser performs WebAuthn assertion using platform or roaming authenticator
5. Server verifies assertion and returns success
6. Gateway executes `sudo -- <cmd>`

## Trust boundaries

- Browser/WebAuthn runtime
- Local challenge server
- Local config and credential metadata
- Sudo subsystem and host policy

## Future extension points

- PAM integration mode
- Audit event export
- Rich policy engine
- Team-managed passkey enrollment

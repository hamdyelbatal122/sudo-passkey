# Passkey-Sudo (The Security Gate)

Passkey-Sudo is a Go CLI that adds a Passkey biometric security gate before privileged commands.
Instead of typing a sudo password for each protected operation, you approve access with your fingerprint/face/passkey from your laptop or mobile device.

## Status

This project is production-ready as a repository template and MVP implementation.

- Passkey registration via WebAuthn
- Passkey verification before sudo execution
- Local-only verification flow on 127.0.0.1
- Clean CLI UX for Linux/macOS
- CI-ready GitHub setup

## Why this exists

sudo protects privileged access with passwords. Password prompts are weak against shoulder-surfing, reuse, and phishing.
Passkey-Sudo enforces modern, phishing-resistant WebAuthn checks before a command is allowed to run.

## How it works

1. passkey-sudo enroll opens a local WebAuthn flow.
2. You register a passkey (Touch ID, Windows Hello, Android/iOS passkey, or hardware key).
3. passkey-sudo run -- COMMAND launches a short authentication challenge.
4. After successful biometric approval, the command is executed with sudo.

## Important sudo note

Passkey-Sudo is a security gate wrapper and does not replace PAM internals.
For passwordless sudo UX, configure sudoers for the exact allowed commands and keep Passkey-Sudo as the biometric policy gate.

See docs/sudoers.example.

## Installation

### From source

```bash
git clone https://github.com/hamdy/passkey-sudo.git
cd passkey-sudo
make build
sudo install -m 0755 bin/passkey-sudo /usr/local/bin/passkey-sudo
```

### Quick local build

```bash
go build -o passkey-sudo ./cmd/passkey-sudo
./passkey-sudo version
```

## Quick start

```bash
passkey-sudo init
passkey-sudo enroll
passkey-sudo check
passkey-sudo run -- systemctl restart nginx
```

## Configuration

Config file path:

- Linux/macOS: ~/.config/passkey-sudo/config.json

Default shape:

```json
{
  "version": 1,
  "rp_id": "localhost",
  "rp_origin": "http://127.0.0.1:14141",
  "rp_display_name": "Passkey-Sudo",
  "username": "local-admin",
  "user_id": "<generated>",
  "credentials": [],
  "allowed_commands": [],
  "sudo_non_interactive": true,
  "open_browser_on_prompt": true
}
```

## Command reference

```text
passkey-sudo init [--rp-id localhost --rp-origin http://127.0.0.1:14141 --rp-name Passkey-Sudo --username local-admin]
passkey-sudo enroll
passkey-sudo check
passkey-sudo run -- <command> [args...]
passkey-sudo version
```

## Security principles

- Local challenge server only (127.0.0.1)
- No remote auth dependency
- No plaintext passwords handled by this tool
- Config file permission set to 0600

## Repository quality

- Conventional Go project layout
- CI workflow (build, test, vet)
- Release automation template
- Community docs (CONTRIBUTING, SECURITY, code of conduct)

## Roadmap

- Native PAM module mode
- Policy profiles per host/user/team
- Enterprise audit log output
- FIDO2 resident-key hardening profiles

## Disclaimer

This tool gates privileged commands but does not guarantee full host hardening.
Use with least-privilege sudoers policy, disk encryption, and endpoint protection.

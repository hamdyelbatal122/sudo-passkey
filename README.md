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

### Recommended: install latest release (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

What this installer does:

- Detects your OS/architecture
- Downloads and installs the latest GitHub Release binary when available
- Falls back to source build if no release asset exists for your platform
- Automatically installs Go only when fallback build is needed and Go is missing

Optional environment overrides:

```bash
PASSKEY_SUDO_INSTALL_DIR=$HOME/.local/bin curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
PASSKEY_SUDO_GO_VERSION=1.23.10 curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

### Install from source manually

```bash
git clone https://github.com/hamdyelbatal122/sudo-passkey.git
cd sudo-passkey
make tidy
make build
sudo install -m 0755 bin/passkey-sudo /usr/local/bin/passkey-sudo
```

### If Go is not installed and you prefer manual setup

Install Go 1.23+ from:

- https://go.dev/doc/install

Then verify:

```bash
go version
```

## Quick start

```bash
passkey-sudo init
passkey-sudo enroll
passkey-sudo check
passkey-sudo run -- systemctl restart nginx
```

## Easy customization

### Manage passkeys

```bash
passkey-sudo passkey add
passkey-sudo passkey list
passkey-sudo passkey remove 1
```

### Manage allowed commands

```bash
passkey-sudo allow add /usr/bin/systemctl
passkey-sudo allow list
passkey-sudo allow remove /usr/bin/systemctl
```

### Manage settings from CLI

```bash
passkey-sudo settings show
passkey-sudo settings set username my-admin
passkey-sudo settings set open-browser false
passkey-sudo settings set sudo-non-interactive true
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
passkey-sudo add-passkey
passkey-sudo passkey <add|list|remove>
passkey-sudo allow <list|add|remove>
passkey-sudo settings <show|set>
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

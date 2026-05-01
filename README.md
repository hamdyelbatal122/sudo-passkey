# Passkey-Sudo (The Security Gate)

Passkey-Sudo is a Go CLI that adds a biometric/passkey gate before privileged commands.
Instead of typing a sudo password every time, you approve with your fingerprint/face/passkey.

## What You Get

- Passkey enrollment via WebAuthn
- Passkey verification before sudo execution
- Local-only verification flow (no remote auth dependency)
- Easy install from latest release
- CLI commands for passkeys, allowlist, and settings

## How It Works

1. `passkey-sudo enroll` starts a local WebAuthn flow.
2. You register a passkey (platform or hardware authenticator).
3. `passkey-sudo run -- <command>` requests biometric approval.
4. On success, command is executed with `sudo`.

## Install (Recommended)

Install latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

Installer behavior:

- Detects OS and architecture
- Downloads latest release binary when available
- Falls back to source build if no matching asset exists
- Installs Go automatically only when fallback build needs it

Optional overrides:

```bash
PASSKEY_SUDO_INSTALL_DIR=$HOME/.local/bin curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
PASSKEY_SUDO_GO_VERSION=1.23.10 curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

## Update To Latest Release

Run the same installer command again:

```bash
curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

## Manual Install From Source

```bash
git clone https://github.com/hamdyelbatal122/sudo-passkey.git
cd sudo-passkey
make tidy
make build
sudo install -m 0755 bin/passkey-sudo /usr/local/bin/passkey-sudo
```

If you prefer installing Go manually first, see [Go installation guide](https://go.dev/doc/install).

## Quick Start

```bash
passkey-sudo init
passkey-sudo enroll
passkey-sudo check
passkey-sudo run -- systemctl restart nginx
```

## Command Reference

```text
passkey-sudo init [--rp-id localhost --rp-origin http://localhost:14141 --rp-name Passkey-Sudo --username local-admin]
passkey-sudo enroll
passkey-sudo add-passkey
passkey-sudo passkey <add|list|remove>
passkey-sudo allow <list|add|remove>
passkey-sudo settings <show|set>
passkey-sudo check
passkey-sudo run -- <command> [args...]
passkey-sudo version
```

## Daily Operations

Manage passkeys:

```bash
passkey-sudo passkey add
passkey-sudo passkey list
passkey-sudo passkey remove 1
```

Manage allowed commands:

```bash
passkey-sudo allow add /usr/bin/systemctl
passkey-sudo allow list
passkey-sudo allow remove /usr/bin/systemctl
```

Manage runtime settings:

```bash
passkey-sudo settings show
passkey-sudo settings set username my-admin
passkey-sudo settings set open-browser false
passkey-sudo settings set sudo-non-interactive true
```

## Mobile Passkey From Laptop (QR)

If you are on a laptop and want to approve passkey from your phone:

1. Ensure laptop and phone are on the same network.
2. Set RP host to laptop LAN IP/hostname (not localhost).
3. Run enrollment/auth command.
4. Scan QR shown on the browser page from phone.

Example:

```bash
passkey-sudo settings set rp-id 192.168.1.10
passkey-sudo settings set rp-origin http://192.168.1.10:14141
passkey-sudo enroll
```

Notes:

- `localhost` works for laptop-only flow.
- For mobile QR flow, `localhost` and loopback IPs are intentionally blocked.

### Reliable Mobile Flow (Trusted HTTPS)

Mobile WebAuthn requires a trusted HTTPS origin. Best workflow:

1. Start app flow on laptop:

```bash
passkey-sudo enroll
```

2. In another terminal, expose local port with ngrok:

```bash
ngrok http 14141
```

3. Re-run `passkey-sudo enroll`.

Passkey-Sudo auto-detects ngrok HTTPS tunnel and uses it as trusted RP origin.
Then QR works on phone without insecure-origin errors.

## Configuration

Default config path:

- Linux/macOS: `~/.config/passkey-sudo/config.json`

Default config:

```json
{
  "version": 1,
  "rp_id": "localhost",
  "rp_origin": "http://localhost:14141",
  "rp_display_name": "Passkey-Sudo",
  "username": "local-admin",
  "user_id": "<generated>",
  "credentials": [],
  "allowed_commands": [],
  "sudo_non_interactive": true,
  "open_browser_on_prompt": true
}
```

## Troubleshooting

### "Error: This is an invalid domain"

Use matching RP settings:

```bash
passkey-sudo settings set rp-id localhost
passkey-sudo settings set rp-origin http://localhost:14141
```

Then re-enroll:

```bash
passkey-sudo enroll
```

### Go not found

The installer auto-installs Go only when source fallback is required.
If you need manual install, use: [Go installation guide](https://go.dev/doc/install).

## Security Notes

- Keep `sudoers` least-privilege and command-restricted
- Keep passkey-enabled browser and OS updated
- Config file is saved with `0600` permissions

See `docs/sudoers.example` for policy guidance.

## Disclaimer

Passkey-Sudo improves privileged command gating but is not a full host-hardening replacement.
Use it with least-privilege policy, endpoint protection, and disk encryption.

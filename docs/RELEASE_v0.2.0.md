# Passkey-Sudo v0.2.0

## Highlights

Passkey-Sudo v0.2.0 focuses on practical usability, faster setup, and easier day-2 operations.

- Easier passkey management directly from CLI
- Easier command allowlist customization
- Easier runtime settings management
- Cleaner onboarding and installation guidance

## What's new

### Passkey lifecycle commands

- `passkey-sudo passkey add`
- `passkey-sudo passkey list`
- `passkey-sudo passkey remove <index>`
- `passkey-sudo add-passkey`

### Allowlist management commands

- `passkey-sudo allow add <command-or-path>`
- `passkey-sudo allow list`
- `passkey-sudo allow remove <command-or-path>`

### Settings management commands

- `passkey-sudo settings show`
- `passkey-sudo settings set <key> <value>`

Supported setting keys:

- `rp-id`
- `rp-origin`
- `rp-name`
- `username`
- `sudo-non-interactive`
- `open-browser`

## Upgrade notes

1. Pull latest source.
2. Rebuild binary: `make build`.
3. Reinstall binary: `sudo install -m 0755 bin/passkey-sudo /usr/local/bin/passkey-sudo`.

No config migration is required for existing users.

## Security and operations

- Keep sudoers restricted to explicit commands only.
- Prefer hardware-backed passkeys for privileged environments.
- Use `allow list` aggressively to reduce misuse surface.

Thanks to everyone testing and providing feedback.

# Changelog

All notable changes to this project are documented in this file.

## [0.2.5] - 2026-05-02

### Added (0.2.5)

- Auto-detect LAN IP on every `enroll` / `check` / `run` — no manual `rp-id` / `rp-origin` settings needed for mobile
- Auto-generate in-memory self-signed TLS certificate (ECDSA P-256, 10-year validity, SANs for LAN IP + localhost)
- Serve HTTPS on `0.0.0.0:14141` when LAN is present; plain HTTP on `localhost:14141` otherwise
- QR code displayed in a fixed bottom-left widget on the web page (always visible, no scroll needed)
- Certificate warning banner with "Advanced → Proceed" instructions shown automatically in HTTPS mode
- Auto-persist updated `rp-id` / `rp-origin` to config so future commands stay in sync
- Graceful fallback to localhost HTTP when no LAN interface is found

### Fixed (0.2.5)

- Resolved "This browser does not support WebAuthn/Passkeys" on mobile — was caused by non-secure HTTP context
- Removed need to manually run `passkey-sudo settings set rp-id ...` before mobile enrollment

## [0.2.4] - 2026-05-02

### Added (0.2.4)

- Added built-in QR code endpoint for mobile passkey flow in local WebAuthn page
- Added mobile readiness metadata endpoint and UI guidance in enrollment/auth page
- Added README section for laptop-to-mobile passkey setup on same network

### Fixed (0.2.4)

- Stabilized CI by replacing brittle formatting-only step with source structure validation
- Kept full validation pipeline (`go mod tidy`, `go vet`, `go test`, `go build`) intact

## [0.2.3] - 2026-05-02

### Fixed (0.2.3)

- Fixed WebAuthn "invalid domain" issues by normalizing and aligning `rp_id` and `rp_origin`
- Added automatic migration for existing configs that used loopback IP origins
- Updated CLI defaults to `http://localhost:14141` for safer domain consistency

### Improved (0.2.3)

- Reorganized README for clearer install, update, daily usage, and troubleshooting guidance

## [0.2.2] - 2026-05-02

### Fixed (0.2.2)

- Fixed installer source fallback when latest release tag does not contain `cmd/passkey-sudo`
- Added fallback build path from `master` branch source if release tag source is missing required files
- Improved CI format check to validate tracked Go files directly and provide clearer output

## [0.2.1] - 2026-05-02

### Fixed (0.2.1)

- Fixed one-line installer 404 by ensuring both `main` and `master` branch raw URLs work
- Updated CI workflow triggers to support both `main` and `master`
- Added manual trigger support to the release workflow

## [0.2.0] - 2026-05-02

### Added (0.2.0)

- New CLI customization commands for passkey lifecycle management:
  - `passkey-sudo passkey add`
  - `passkey-sudo passkey list`
  - `passkey-sudo passkey remove <index>`
- New CLI allowlist management commands:
  - `passkey-sudo allow add <command-or-path>`
  - `passkey-sudo allow list`
  - `passkey-sudo allow remove <command-or-path>`
- New settings management commands:
  - `passkey-sudo settings show`
  - `passkey-sudo settings set <key> <value>`
- `add-passkey` shortcut command for faster enrollment
- One-line installer guidance and improved onboarding documentation

### Changed (0.2.0)

- Updated Go module/import path to `github.com/hamdyelbatal122/sudo-passkey`
- Improved README to focus on simple installation, quick customization, and practical usage flows

## [0.1.0] - 2026-05-02

### Added (0.1.0)

- Initial Passkey-Sudo CLI implementation
- WebAuthn enrollment and authentication flow
- Sudo command gateway mode
- Professional GitHub repository scaffolding

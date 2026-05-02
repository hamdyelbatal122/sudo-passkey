# Changelog

All notable changes to this project are documented in this file.

## [0.2.14] - 2026-05-02

### Fixed (0.2.14)

- Improved handling of NFC enrollment failures on mobile Credential Manager
- Added multi-stage retry path for registration (`normal` -> `mobile-compatible` -> `NFC-compatible cross-platform`)
- Reduced strict create options in NFC retry mode to avoid provider startup failures after scan

## [0.2.13] - 2026-05-02

### Fixed (0.2.13)

- Added retry fallback for Android/mobile `Credential Manager` startup failures during passkey create/get
- Relaxed strict WebAuthn client options on retry to improve NFC/phone compatibility
- Improved failure handling path for the error: "An unknown error occurred while starting the credential manager"

### Improved (0.2.13)

- Kept mobile guidance text general and user-friendly

## [0.2.12] - 2026-05-02

### Fixed (0.2.12)

- Fixed mobile enrollment failure: `error parsing attestation response`
- Corrected client-side WebAuthn binary serialization for typed-array responses with proper byte offsets
- Sent only required WebAuthn response fields for registration/authentication payloads

### Improved (0.2.12)

- Replaced technical mobile readiness text with general user-friendly guidance

## [0.2.11] - 2026-05-02

### Fixed (0.2.11)

- Fixed mobile WebAuthn root cause by requiring trusted HTTPS origin for phone flow
- Added automatic trusted-origin discovery from running ngrok tunnel (`https` tunnel to `:14141`)
- Removed misleading LAN-IP WebAuthn behavior that caused insecure-origin failures on mobile

### Improved (0.2.11)

- Added explicit mobile readiness hints in API/UI when trusted origin is unavailable
- Kept localhost enrollment stable while auto-switching to trusted public origin when available

## [0.2.10] - 2026-05-02

### Fixed (0.2.10)

- Fixed WebAuthn failure when opening LAN IP URL from the same machine/browser
- Added automatic redirect from same-device LAN access to localhost ceremony URL
- Kept LAN URL available for external devices without forcing localhost redirect

### Improved (0.2.10)

- Improved unsupported-origin UI message with clearer secure-origin guidance

## [0.2.9] - 2026-05-02

### Fixed (0.2.9)

- Removed separate mobile helper page from QR flow
- QR now opens the same WebAuthn flow page path (`/`) on phone and laptop
- Updated UI copy to reflect same-page QR behavior

## [0.2.8] - 2026-05-02

### Fixed (0.2.8)

- Restored bottom-left QR visibility on desktop in default localhost mode
- Made LAN URL reachable from mobile by binding helper mode on `0.0.0.0:14141` when LAN IP exists
- Redirected non-local hosts to a dedicated mobile helper page to avoid WebAuthn domain errors

### Improved (0.2.8)

- Kept actual WebAuthn ceremony on trusted localhost origin while still providing mobile-accessible guidance page
- Clarified QR purpose in UI (opens phone helper page on LAN)

## [0.2.7] - 2026-05-02

### Fixed (0.2.7)

- Fixed WebAuthn failures caused by untrusted local TLS certificates in mobile-LAN mode
- Restored stable default flow to trusted localhost secure context (no self-signed cert dependency)
- Prevented domain mismatch errors by keeping default ceremony host aligned with `localhost`

### Improved (0.2.7)

- Added explicit guidance to use browser native "use a phone or tablet" passkey flow for cross-device enrollment
- Kept LAN TLS mode available only as explicit opt-in (`PASSKEY_SUDO_ENABLE_LAN_TLS=1`)

## [0.2.6] - 2026-05-02

### Fixed (0.2.6)

- Fixed WebAuthn error: "The relying party ID is not a registrable domain suffix of, nor equal to the current domain"
- Enforced single host identity for ceremony by using LAN IP as both `rp_id` and `rp_origin` in mobile mode
- Added automatic redirect from `localhost`/loopback requests to active RP host URL to prevent host mismatch
- Kept QR hidden when page is opened from a mobile device

### Improved (0.2.6)

- Auto-persisted active runtime RP settings to config for consistent `enroll`, `check`, and `run` behavior
- Mobile QR target now always matches active RP host used by backend session

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

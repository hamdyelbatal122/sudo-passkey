# Changelog

All notable changes to this project are documented in this file.

## [0.2.1] - 2026-05-02

### Fixed

- Fixed one-line installer 404 by ensuring both `main` and `master` branch raw URLs work
- Updated CI workflow triggers to support both `main` and `master`
- Added manual trigger support to the release workflow

## [0.2.0] - 2026-05-02

### Added

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

### Changed

- Updated Go module/import path to `github.com/hamdyelbatal122/sudo-passkey`
- Improved README to focus on simple installation, quick customization, and practical usage flows

## [0.1.0] - 2026-05-02

### Added

- Initial Passkey-Sudo CLI implementation
- WebAuthn enrollment and authentication flow
- Sudo command gateway mode
- Professional GitHub repository scaffolding

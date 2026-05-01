# Passkey-Sudo v0.2.3

## Patch release

This patch focuses on reliability for passkey enrollment and cleaner user guidance.

## Highlights

- Fixed WebAuthn domain validation failures (`invalid domain`) by normalizing domain settings.
- Added automatic config migration for existing loopback-IP origins.
- Updated default RP origin to `http://localhost:14141` for strict domain consistency.
- Improved README structure with clear sections for install, update, operations, and troubleshooting.

## Why this matters

Users no longer need manual trial-and-error for RP domain settings when enrolling a passkey.
Existing configs are auto-healed when loaded.

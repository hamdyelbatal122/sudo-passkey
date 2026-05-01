# Passkey-Sudo v0.2.4

Release date: 2026-05-02

## Highlights

- Mobile passkey flow via built-in QR code in the local WebAuthn page
- Better user guidance for laptop + phone setup on the same network
- CI stabilization to eliminate fragile formatting-only failures

## Added

- New endpoint: `/qr.png`
  - Generates QR directly from local server URL (no third-party API)
- New endpoint: `/api/meta`
  - Exposes target URL and whether current RP host is mobile-ready
- Updated local web UI
  - Shows QR when mobile setup is valid
  - Shows clear guidance when RP is localhost/loopback

## Fixed

- CI now uses a stable source structure check instead of a brittle format-only gate
- Core validation remains unchanged:
  - `go mod tidy`
  - `go vet ./...`
  - `go test ./...`
  - `go build ./...`

## Mobile Setup Quick Example

```bash
passkey-sudo settings set rp-id 192.168.1.10
passkey-sudo settings set rp-origin http://192.168.1.10:14141
passkey-sudo enroll
```

## Upgrade

```bash
curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

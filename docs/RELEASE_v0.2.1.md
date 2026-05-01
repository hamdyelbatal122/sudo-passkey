# Passkey-Sudo v0.2.1

## Patch release

This patch focuses on installer reliability and branch compatibility.

## Fixes

- Fixed installer URL 404 when using the `main` raw URL.
- Ensured both `main` and `master` branches expose `scripts/install.sh`.
- Updated CI workflow to run on both `main` and `master`.
- Added `workflow_dispatch` to release workflow for manual release control.

## Impact

Users can now run either of the following successfully:

```bash
curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/main/scripts/install.sh | bash
```

or

```bash
curl -fsSL https://raw.githubusercontent.com/hamdyelbatal122/sudo-passkey/master/scripts/install.sh | bash
```

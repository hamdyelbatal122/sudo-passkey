# Passkey-Sudo v0.2.2

## Patch release

This patch fixes installation reliability for environments where release binary assets are missing and source fallback is required.

## Fixes

- Fixed source fallback failure when the latest release tag does not include `cmd/passkey-sudo`.
- Added a second fallback source build path from `master` when tag source is incomplete.
- Improved CI format-check logic to inspect tracked Go files directly and show clear failure output.

## Result

The installer now follows this order:

1. Install prebuilt binary from latest release asset.
2. If unavailable, build from latest release source tag.
3. If tag source is incomplete, build from `master` source.

This ensures a successful installation path for users across more repository states.

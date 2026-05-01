# Threat Model (MVP)

## Assets

- Privileged command execution path
- Passkey credential metadata
- Local configuration integrity

## Primary threats

- Command abuse if allowed command scope is too broad
- Local malware hijacking browser sessions
- Misconfigured sudoers creating over-privilege
- Shoulder surfing / password theft (mitigated by passkeys)

## Mitigations

- Restrict sudoers to explicit command paths
- Restrict `allowed_commands` in config
- Keep flow local-only on loopback
- Use hardware-backed authenticators where possible
- Keep OS/browser patched

## Residual risks

- Fully compromised host can bypass user-space controls
- Browser-level compromise may affect challenge UX

## Operational recommendations

- Pair with endpoint protection and disk encryption
- Use least privilege and command allowlists
- Monitor sudo command logs

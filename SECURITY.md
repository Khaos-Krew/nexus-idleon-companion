# Security and Data Safety

Nexus Idleon Companion is intentionally a local, read-only account viewer.

## Save safety

The application does not edit Idleon's save data. During capture it:

1. Locates the local Idleon LevelDB folder.
2. Copies relevant files to a temporary directory.
3. Reads and decodes the temporary copy.
4. Deletes the temporary copy when processing finishes.

It does not open the live database for writing, inject into the game, read process memory or automate gameplay.

## Local data

Snapshots, goals, roles, timers, custom rules, cached icons and preferences remain under:

```text
%LOCALAPPDATA%\Idleon Account Monitor
```

The historical folder name is retained to preserve data across upgrades.

## Network activity

The application can make outbound HTTPS requests only for user-facing features such as:

- Checking the configured update manifest.
- Downloading a selected verified update.
- Loading and locally caching visual icons.

Raw Idleon save data is not uploaded by the application.

## Update verification

Before an automatic update is staged, the application:

- Requires an HTTPS URL.
- Applies a maximum download size.
- Verifies the published SHA-256 checksum.
- Validates that the result is a 64-bit Windows executable.
- Stages the replacement through a separate updater process.

Users can also select an update executable manually.

## Khaos Nexus integration

The local integration endpoints are read-only and versioned. Website exports can redact character names and omit goals or findings. A future hosted bridge should require authenticated user consent and should never expose the local service directly to the public internet.

## Support diagnostics

Diagnostics are designed to exclude raw snapshots, character names, save payloads and credentials. Users should still review any diagnostic file before sharing it publicly.

## Reporting a vulnerability

Please open a private security advisory in this repository when possible. Avoid posting raw account snapshots, authentication information or private character data in a public issue.

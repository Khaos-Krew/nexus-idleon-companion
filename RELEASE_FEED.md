# Release Feed

Nexus Idleon Companion checks the following public stable feed by default:

```text
https://raw.githubusercontent.com/Khaos-Krew/nexus-idleon-companion/main/releases/stable.json
```

## Manifest format

```json
{
  "version": "0.5.2",
  "channel": "stable",
  "publishedAt": "2026-07-19T00:00:00Z",
  "downloadUrl": "https://github.com/Khaos-Krew/nexus-idleon-companion/releases/download/v0.5.2/Nexus-Idleon-Companion-v0.5.2-x64.exe",
  "sha256": "64 lowercase hexadecimal characters",
  "sizeBytes": 0,
  "releaseNotes": "Release summary",
  "releasePageUrl": "https://github.com/Khaos-Krew/nexus-idleon-companion/releases/tag/v0.5.2",
  "minimumVersion": "0.5.0"
}
```

## Publishing process

The GitHub Actions release workflow:

1. Reconstructs the retained source archive.
2. Verifies the source archive SHA-256 before extraction.
3. Runs `go test ./...` and `go vet ./...`.
4. Builds the native Windows x64 GUI executable.
5. Calculates the executable SHA-256 checksum.
6. Creates the portable ZIP and source archive.
7. Publishes or replaces the GitHub Release assets.
8. Rewrites `releases/stable.json` with the exact published file hash and size.

The workflow will stop before building or publishing when the retained source payload does not match its expected hash.

## Update channels

The current public channel is `stable`. The desktop app also supports a preview channel for future testing releases.

## Compatibility

The local app-data location intentionally remains `%LOCALAPPDATA%\Idleon Account Monitor`, allowing releases under the Nexus Idleon Companion name to retain all existing local history and preferences.

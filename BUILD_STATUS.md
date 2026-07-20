# Build Status

## Release target

- Product: Nexus Idleon Companion
- Version: 0.5.2
- Platform: Windows x64
- Runtime: Native Go executable with a local browser dashboard
- External runtime dependencies: None

## Validation performed before publication

- Go unit tests passed.
- Go race tests passed in the source build environment.
- `go vet` and JavaScript syntax validation passed.
- Full local API workflow passed against a real ten-character snapshot.
- Backup, restore, diagnostics, privacy redaction, update checking and update staging passed.
- Desktop and narrow mobile layouts were browser-tested.
- Source archive rebuilt and passed tests from a clean extraction.
- Windows executable was validated as a PE32+ x64 GUI application.

## GitHub release safeguards

The release workflow independently:

1. Reconstructs the retained source archive.
2. Verifies the source archive SHA-256.
3. Runs Go tests.
4. Runs static analysis.
5. Builds the Windows x64 application.
6. Generates release checksums and packages.
7. Publishes the GitHub Release.
8. Updates the stable release manifest with the actual executable hash and size.

A failed source hash, test, static check or build prevents release publication.

## Remaining platform limitation

The original Windows binaries were cross-compiled in Linux, so they could not be launched directly in that build environment. The GitHub workflow builds on `windows-latest`, providing a native Windows build environment for public releases.

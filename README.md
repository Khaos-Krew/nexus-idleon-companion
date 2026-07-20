# Nexus Idleon Companion

**A read-only Windows account command center for Legends of Idleon.**

Nexus Idleon Companion helps players understand what their account is doing, what each character should work on next, and which account systems need attention. It runs locally, reads a temporary copy of the game's local save database, and never writes back to the game.

## Download

The recommended download is the newest Windows executable from [GitHub Releases](https://github.com/Khaos-Krew/nexus-idleon-companion/releases/latest).

After the first published release, Nexus Idleon Companion can also check this repository's verified stable update feed from inside the app.

## Main features

- Beginner-friendly guided setup and explanations.
- Read-only local Idleon save capture.
- Prioritized Today dashboard.
- Full account review with evidence and recommended actions.
- Individual progression breakdown for every character.
- Suggested character roles and setup coverage.
- Equipment, tools, food, cards, talents and system icons.
- Goals and resource planning.
- Timers and optional browser notifications.
- Custom review rules.
- Snapshot comparisons and progress analytics.
- Twenty-two dedicated Idleon system planners.
- Local backups and validated restoration.
- Privacy-safe support diagnostics.
- Versioned, sanitized Khaos Nexus integration export.
- Manual and verified automatic updates.

## How it works

1. Run the Windows executable.
2. The app opens its private dashboard in your normal browser.
3. Start Idleon and allow the game to synchronize.
4. Return to character selection, then select **Capture Now**.
5. Open **Today**, **Account Review**, or **Character Progress**.

All snapshots and preferences remain on the user's computer under:

```text
%LOCALAPPDATA%\Idleon Account Monitor
```

The older folder name is intentionally retained so upgrading from earlier builds does not lose account history, goals, roles, timers, rules or settings.

## Safety model

Nexus Idleon Companion is intentionally read-only. It does not:

- Edit the Idleon save.
- Inject into the game.
- Read game memory.
- Automate gameplay.
- Send passwords or account credentials anywhere.
- Upload raw account data.

The app copies the live LevelDB files to a temporary location before decoding them. See [SECURITY.md](SECURITY.md) for details.

## Automatic updates

Stable releases are described by:

```text
https://raw.githubusercontent.com/Khaos-Krew/nexus-idleon-companion/main/releases/stable.json
```

Before installing an automatic update, the application:

1. Requires HTTPS.
2. Limits the download size.
3. Verifies the published SHA-256 checksum.
4. Confirms the file is a 64-bit Windows executable.
5. Stages the replacement safely.
6. Restarts without changing local account data.

See [RELEASE_FEED.md](RELEASE_FEED.md) for the release process.

## Build from source

Requirements:

- Go 1.23 or newer.
- Windows for native runtime testing.

```powershell
go test ./...
go vet ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o Nexus-Idleon-Companion.exe .
```

The project uses only the Go standard library at runtime.

## Project status

Current version: **0.5.2**

The application is functional and tested against a real ten-character account snapshot. Idleon changes its internal save structure over time, so unsupported or newly changed fields fail closed rather than generating invented advice.

## Khaos Nexus

The local application already exposes a versioned, privacy-controlled integration export for eventual use inside the Khaos Nexus website. A hosted integration should use authenticated user-controlled synchronization rather than exposing the local reader directly to the internet.

See [INTEGRATION.md](INTEGRATION.md) for the current contract.

## Disclaimer

Nexus Idleon Companion is an independent community project and is not affiliated with or endorsed by LavaFlame2 or the developers of Legends of Idleon. Game names and artwork belong to their respective owners.

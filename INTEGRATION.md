# Khaos Nexus Integration

Nexus Idleon Companion includes a local, read-only integration foundation intended for eventual use by the Khaos Nexus website.

## Current local endpoints

The application exposes versioned local responses for:

- Service health.
- Latest analyzed account summary.
- Sanitized Khaos Nexus export.

The current export schema is:

```text
khaos-nexus.idleon.account.v1
```

## Privacy controls

Exports can:

- Redact character names.
- Omit goals.
- Omit review findings.
- Exclude raw save data.

The local service is not designed to be exposed directly to the internet.

## Recommended hosted architecture

A future Khaos Nexus connection should use an authenticated, user-controlled synchronization bridge:

1. The desktop app creates a sanitized versioned export.
2. The user explicitly links the desktop app to their Khaos Nexus account.
3. A short-lived authorization grant identifies the destination account.
4. The desktop app uploads only the fields selected by the user over HTTPS.
5. Khaos Nexus stores the schema version and import timestamp.
6. Users can revoke synchronization and delete imported data.

## Compatibility principles

- Treat the desktop export schema as versioned and additive.
- Reject unknown breaking schema versions rather than guessing.
- Keep raw Idleon save payloads local.
- Preserve character-name redaction through the full pipeline.
- Never require the Khaos Nexus website to access the local LevelDB directly.
- Keep website imports read-only with respect to the game.

## Release integration

The desktop updater reads the public GitHub manifest at:

```text
https://raw.githubusercontent.com/Khaos-Krew/nexus-idleon-companion/main/releases/stable.json
```

A future Khaos Nexus module can display the latest version and release notes from the same public manifest.

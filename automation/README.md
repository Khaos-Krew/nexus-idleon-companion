# IdleOn Account Agent

Standalone Windows account-assessment and UI-automation agent for Legends of IdleOn. It is intentionally separate from Khaos Nexus.

## What it does

The agent has two layers:

1. **Assessment brain** — detects IdleOn, loads account data, scores progression systems, timers, characters and green stacks, and produces a ranked action list.
2. **Calibrated automation** — executes a routine only when a recommendation explicitly names a routine you configured and calibrated.

Unknown data and unsupported game states fail closed. The tool does not inject into IdleOn, alter saves, read process memory, manipulate packets, or contain anti-cheat evasion.

## Account coverage

The planner currently understands these major systems:

- Stamps, Forge, Anvil, Alchemy, Post Office and Obols
- Refinery, Construction, 3D Printer, Worship, Trapping and Shrines
- Cooking, Breeding, Laboratory and Rift
- Divinity, Sailing, Gaming and Companions
- Sneaking, Farming and Summoning
- Cards, Star Signs, Gear/Tools, Talents and Statues
- Dungeons, Bosses, Green Stacks, Death Note and Tome
- Owl and Roo side progression

Each system has world gating, account-wide value, unlock value, expected effort and a preferred class where one is especially useful (for example Bubo for Alchemy, ES for Worship/green-stack farming, DK for Construction/Death Note).

## Toolbox / Efficiency / local data

All sources are normalized into one `AccountSnapshot` model.

### Public IdleOn Toolbox profile

Toolbox exposes public profiles through its profiles service. Use your public main-character/profile name:

```powershell
.\IdleOn-Automation.exe assess -toolbox "YourMainCharacter"
```

### IdleOn Efficiency export/JSON

```powershell
.\IdleOn-Automation.exe assess -efficiency .\idleon-efficiency.json
```

The adapter is intentionally tolerant: recognizable account/system/character fields are normalized while unknown fields are preserved in raw data rather than guessed.

### Local/companion snapshot

```powershell
.\IdleOn-Automation.exe assess -snapshot .\account.json
```

See `account.example.json` for the normalized format. This is the preferred path as the standalone companion's local reader is expanded because it avoids depending on a third-party site being available.

## Detect the active game

```powershell
.\IdleOn-Automation.exe detect
```

On Windows this reports whether the configured IdleOn window exists and whether it is the foreground window.

## Run the live assessment agent

Assessment only:

```powershell
.\IdleOn-Automation.exe agent -config automation.json -snapshot account.json
```

Use a public Toolbox profile instead:

```powershell
.\IdleOn-Automation.exe agent -config automation.json -toolbox "YourMainCharacter"
```

Allow calibrated recommendations to execute:

```powershell
.\IdleOn-Automation.exe agent -config automation.json -snapshot account.json -execute
```

Execution defaults to **foreground-only**. A recommendation without an attached routine remains advice only. The same recommendation is also rate-limited so the agent cannot hammer the same action every refresh cycle.

Press `F12` for the global emergency stop.

## Recommendation scoring

The planner ranks work using roughly:

`account-wide value × unlock value × deficiency × ease × urgency × readiness ÷ sqrt(time)`

That keeps cheap, broad upgrades ahead of long grinds with weak immediate returns. Ready/capped timers are added as immediate-loss recommendations, and unfinished green stacks are tracked toward the permanent 10M threshold.

## Build

Requires Go 1.23+.

```powershell
cd automation
go test ./...
go vet ./...
go build -trimpath -o IdleOn-Automation.exe .
```

CI also cross-compiles Windows x64 on Linux and produces a native Windows artifact.

## Automation setup

Copy the example configuration:

```powershell
Copy-Item automation.example.json automation.json
```

The sample uses named click points rather than hard-coded screen coordinates.

### Calibrate

Start IdleOn and place it where you normally play, then run:

```powershell
.\IdleOn-Automation.exe calibrate -config automation.json
```

For each requested location, move the mouse there and press `F8`. Coordinates are stored locally in `automation.json`. Recalibrate after changing resolution, UI scale, monitor arrangement, or game-window position.

### Test a routine

```powershell
.\IdleOn-Automation.exe list -config automation.json
.\IdleOn-Automation.exe run -config automation.json -routine focus-bubo
.\IdleOn-Automation.exe run -config automation.json -routine travel-greenstack-1
```

Test every navigation routine manually before enabling `agent -execute`.

## Green-stack rotation

```powershell
.\IdleOn-Automation.exe greenstack -config automation.json -cycles 1
```

The loop skips disabled/completed targets, travels to the next target, farms for the configured duration and advances.

Mark a permanent 10M stack complete:

```powershell
.\IdleOn-Automation.exe done -config automation.json -target "Target Name"
```

Reopen it if needed:

```powershell
.\IdleOn-Automation.exe done -config automation.json -target "Target Name" -reopen
```

## Routine primitives

Focus:

```json
{ "type": "focus" }
```

Click calibrated point:

```json
{ "type": "click", "point": "map_button", "delayAfterMs": 500 }
```

Double-click:

```json
{ "type": "double-click", "point": "some_point" }
```

Key/hotkey:

```json
{ "type": "key", "key": "F" }
{ "type": "key", "key": "CTRL+1" }
```

Wait:

```json
{ "type": "wait", "ms": 1000 }
```

Supported keyboard input currently includes A-Z, 0-9, Ctrl, Shift, Alt, Enter, Space, Tab, Escape, arrows and F1-F12.

## Direction

The next iterations can deepen the local parser so account assessment no longer depends on manual normalized fields, add screen-state verification before clicks, and generate more routine bindings from detected character/class/map context. The architecture already separates data sources, planning and execution so those improvements do not require rewriting the automation engine.

# IdleOn Account Agent

Standalone Windows account-assessment and calibrated UI-automation agent for **Legends of IdleOn**. It is deliberately independent from Khaos Nexus.

## V1 status

V1 is build-complete and CI validated on Windows x64.

It combines:

- IdleOn process/window detection.
- Local account snapshot discovery.
- Public IdleOn Toolbox profile input.
- IdleOn Efficiency JSON/export input.
- Multi-source account-data fusion.
- W1-W7 progression scoring.
- Character/class assignment.
- Ready-timer and permanent green-stack prioritization.
- A local dashboard.
- Persistent assessment/action history.
- Relative screen calibration.
- Guarded Windows mouse/keyboard routines.
- Global F12 emergency stop.

It does **not** edit saves, inject code, read process memory, manipulate packets, or include anti-cheat evasion.

## How the agent works

```text
IdleOn detection
      ↓
Local / Toolbox / Efficiency account data
      ↓
Normalized AccountSnapshot
      ↓
W1-W7 progression planner
      ↓
Ranked objectives + character plan
      ↓
Calibrated routine available?
   ↙             ↘
  no              yes
 advice       guarded execution
```

Assessment is automatic. Gameplay input is only dispatched through routines that exist **and are fully calibrated**.

## Progression coverage

The rule catalog includes roughly fifty account systems and progression categories.

### World 1 / account foundations

- World pushing
- Stamps
- Vault
- Forge
- Anvil / Smithing
- Statues
- Cards
- Star Signs
- Talents
- Gear and tools
- Quests
- Tasks / Merits
- Party Dungeons

### World 2

- Alchemy
- Prisma Bubbles
- Post Office
- Obols

### World 3

- Refinery
- Construction
- 3D Printer
- Worship
- Trapping
- Shrines
- Death Note

### World 4

- Cooking
- Breeding
- Laboratory
- Rift
- Tome
- Killroy

### World 5

- Divinity
- Sailing
- Gaming
- Companions
- The Hole
- Slab

### World 6

- Sneaking
- Farming
- Summoning
- Beanstalk
- Master Classes

### World 7

- Research
- Spelunking
- Coral Reef
- Sushi Station

### Cross-world / recurring

- Green Stacks
- Bosses
- Minibosses
- Colosseum
- Weekly Boss
- Voidwalker
- Owl
- Roo

The catalog also models dependencies. For example, Cooking can be down-ranked when Breeding is the actual blocker, Construction depends on Refinery health, Worship can depend on Construction, and Coral Reef can depend on Spelunking progress.

## Character-aware planning

The planner recognizes common elite-class names and aliases, including:

- Bubo / Bubonic Conjuror
- Arcane Cultist
- DK / Divine Knight
- ES / Elemental Sorcerer
- BB / Blood Berserker
- BM / Beast Master
- SB / Siege Breaker
- VMan / Voidwalker

It selects a named character from the supplied account snapshot, not just a class label. Example preferences include Bubo for Alchemy, DK for Construction/Death Note, ES for Worship/resource farming, BB for Cooking, BM for Breeding, and VMan for Voidwalker progression.

## Account sources

All supported sources normalize into one `AccountSnapshot`.

### Automatic local source

The default source is `-snapshot auto`. The agent searches the existing local companion data area under:

```text
%LOCALAPPDATA%\Idleon Account Monitor
```

for the newest usable JSON account snapshot/export.

### Public IdleOn Toolbox profile

```powershell
.\IdleOn-Account-Agent.exe assess -config automation.json -toolbox "YourMainCharacter"
```

### IdleOn Efficiency JSON/export

```powershell
.\IdleOn-Account-Agent.exe assess -config automation.json -efficiency .\idleon-efficiency.json
```

### Fuse sources

You can combine sources. Local data is kept where it is already strong; missing fields can be filled from Efficiency and/or Toolbox.

```powershell
.\IdleOn-Account-Agent.exe assess `
  -config automation.json `
  -snapshot auto `
  -efficiency .\idleon-efficiency.json `
  -toolbox "YourMainCharacter"
```

Unknown fields are preserved as raw data rather than converted into invented account values.

## First-time setup

### 1. Create your config

Copy the supplied template:

```powershell
Copy-Item automation.example.json automation.json
```

The template includes system entry routines for major W2-W7 systems, class-switch routines, and a green-stack travel example.

### 2. Start IdleOn

Run IdleOn in the window size/layout you normally use.

### 3. Run the doctor

```powershell
.\IdleOn-Account-Agent.exe doctor -config automation.json
```

If you use Toolbox as an additional source:

```powershell
.\IdleOn-Account-Agent.exe doctor -config automation.json -toolbox "YourMainCharacter"
```

Doctor reports:

- whether IdleOn is running/foreground;
- detected window geometry;
- account source, world, characters and parsed systems;
- missing calibration points;
- planner coverage and account-health score.

### 4. Calibrate

```powershell
.\IdleOn-Account-Agent.exe calibrate -config automation.json
```

For each requested point:

1. Put the mouse on the requested IdleOn control/location.
2. Press **F8**.
3. The agent records both absolute and window-relative coordinates.

Points outside the IdleOn window are rejected. Because normalized coordinates are retained, moving/resizing the window is less fragile than raw desktop-only coordinates, though major UI-scale/layout changes should still be recalibrated.

### 5. Assess the account

```powershell
.\IdleOn-Account-Agent.exe assess -config automation.json
```

This prints:

- account stage/world;
- health score;
- parsed-data coverage;
- active character;
- top progression priorities;
- confidence;
- recommended character;
- whether a safe calibrated routine is available;
- a character-by-character work plan.

## Local dashboard

Launch:

```powershell
.\IdleOn-Account-Agent.exe serve -config automation.json
```

Default address:

```text
http://127.0.0.1:17654
```

The server binds to localhost only.

The dashboard includes:

- current game status;
- account health and data coverage;
- top objectives;
- recommended character assignments;
- **Assess Now**;
- **Execute Top Safe Action**, with an explicit confirmation prompt.

## Continuous agent

Assessment-only mode:

```powershell
.\IdleOn-Account-Agent.exe agent -config automation.json
```

Assessment + calibrated execution:

```powershell
.\IdleOn-Account-Agent.exe agent -config automation.json -execute
```

Execution defaults to foreground-only.

The agent also applies:

- per-routine cooldowns;
- a default 15-minute objective cooldown;
- an hourly action budget;
- snapshot-freshness checks;
- full-calibration requirements;
- dependency-aware recommendation ranking;
- UI pixel guards when a routine defines them.

Press **F12** at any point to emergency-stop a routine or continuous agent wait cycle.

## UI-state guards

A routine may require a screen pixel to match an expected color before clicking:

```json
{
  "type": "click",
  "point": "claim_button",
  "expectedRgb": "#D8C56A",
  "tolerance": 25,
  "delayAfterMs": 500
}
```

If the sampled screen state is outside the tolerance, the click does not occur and the routine fails closed.

You can also use a standalone guard:

```json
{
  "type": "assert-pixel",
  "point": "screen_marker",
  "expectedRgb": "#FFFFFF",
  "tolerance": 20
}
```

## Routine primitives

Focus:

```json
{ "type": "focus" }
```

Calibrated click:

```json
{ "type": "click", "point": "map_button", "delayAfterMs": 500 }
```

Double-click:

```json
{ "type": "double-click", "point": "some_point" }
```

Keyboard/hotkey:

```json
{ "type": "key", "key": "F" }
{ "type": "key", "key": "CTRL+1" }
```

Wait:

```json
{ "type": "wait", "ms": 1000 }
```

Repeat a safe primitive up to 25 times:

```json
{ "type": "click", "point": "claim", "repeat": 3 }
```

Routines are capped at 500 top-level steps.

## Green stacks

The dedicated loop remains available:

```powershell
.\IdleOn-Account-Agent.exe greenstack -config automation.json -cycles 1
```

Completed targets are skipped. Mark a permanent 10M stack complete:

```powershell
.\IdleOn-Account-Agent.exe done -config automation.json -target "Target Name"
```

Reopen one if necessary:

```powershell
.\IdleOn-Account-Agent.exe done -config automation.json -target "Target Name" -reopen
```

When an imported snapshot contains green-stack counts, the normal account planner can also rank unfinished stacks automatically.

## Local state

Agent history is stored under:

```text
%LOCALAPPDATA%\IdleOn Account Agent
```

Files include:

- `last-assessment.json`
- `assessment-history.jsonl`
- `actions.jsonl`

## Build from source

Requires Go 1.23+.

```powershell
cd automation
go test ./...
go vet ./...
go build -trimpath -o IdleOn-Account-Agent.exe .
```

CI validates:

- unit tests;
- `go vet`;
- Windows x64 cross-compilation;
- native Windows build;
- artifact upload.

## Important execution boundary

The assessment engine can cover far more of IdleOn than the out-of-box automation routines. A recommendation remains **advice only** until its routine exists and every named UI point has been calibrated. Complex in-game paths still need to be calibrated/tested against your actual account and UI before unattended execution is sensible.

That is intentional: the account brain can expand rapidly without making the input layer guess at unknown screens.

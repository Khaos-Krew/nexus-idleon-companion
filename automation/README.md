# IdleOn Automation V1

Standalone Windows UI automation for Legends of IdleOn. This project is intentionally separate from the Khaos Nexus application and backend.

## Current focus

V1 is built around Bubo green-stack farming:

- Focus the IdleOn window.
- Calibrate named click targets instead of hard-coding screen coordinates.
- Run reusable travel/navigation routines.
- Rotate through configured green-stack farming targets for set durations.
- Mark finished 10M+ green stacks complete so they are skipped permanently.
- Press `F12` at any time for an emergency stop.
- No memory editing, save editing, packet manipulation, injection, or anti-cheat evasion.

The automation only sends normal Windows mouse/keyboard input that you explicitly configure.

## Build

Requires Go 1.23+.

```powershell
cd automation
go test ./...
go build -trimpath -o IdleOn-Automation.exe .
```

## First setup

Copy the example configuration:

```powershell
Copy-Item automation.example.json automation.json
```

Edit `automation.json` and set the green-stack target names and farm durations you want.

The sample travel routines use named points:

- `map_button`
- `greenstack_target_1`
- `greenstack_target_2`
- `greenstack_target_3`
- `travel_confirm`

You can rename them or add as many routines/points as needed.

## Calibrate your screen

Start IdleOn and place it exactly where you normally play. Then run:

```powershell
.\IdleOn-Automation.exe calibrate -config automation.json
```

For each point:

1. Move your mouse to the requested button/location.
2. Press `F8`.
3. The coordinate is stored in `automation.json`.

Press `F12` to abort calibration.

If the game window moves, resolution changes, UI scale changes, or you switch monitor layouts, recalibrate.

## Test a routine

```powershell
.\IdleOn-Automation.exe list -config automation.json
.\IdleOn-Automation.exe run -config automation.json -routine focus-bubo
.\IdleOn-Automation.exe run -config automation.json -routine travel-greenstack-1
```

Test every travel routine manually before running a long cycle.

## Run the Bubo green-stack loop

```powershell
.\IdleOn-Automation.exe greenstack -config automation.json -cycles 1
```

The runner will:

1. Skip disabled or already-completed targets.
2. Execute the next target's configured travel routine.
3. Leave Bubo farming for `farmMinutes`.
4. Move to the next unfinished target.
5. Continue until the requested number of cycles is complete.

`F12` is checked repeatedly during all waits and stops the loop.

## Mark a green stack complete

Once an item has reached the 10M+ green-stack threshold, mark it finished:

```powershell
.\IdleOn-Automation.exe done -config automation.json -target "Bubo Green Stack Target 1"
```

That target will no longer be included in automatic rotations. To reopen it:

```powershell
.\IdleOn-Automation.exe undo-done -config automation.json -target "Bubo Green Stack Target 1"
```

## Configuration format

A routine consists of these step types:

### Focus the game

```json
{ "type": "focus" }
```

### Click a calibrated point

```json
{ "type": "click", "point": "map_button", "delayAfterMs": 500 }
```

### Double click

```json
{ "type": "double-click", "point": "some_point" }
```

### Press a key/hotkey

```json
{ "type": "key", "key": "F" }
{ "type": "key", "key": "CTRL+1" }
```

Supported keys currently include A-Z, 0-9, Ctrl, Shift, Alt, Enter, Space, Tab, Escape, arrows and F1-F12.

### Wait

```json
{ "type": "wait", "ms": 1000 }
```

## Bubo progression direction

The automation layer is deliberately generic. The next progression layer can read the account snapshot and generate the target list automatically, prioritizing missing green stacks and other account-wide progression bottlenecks instead of requiring you to maintain the list by hand.

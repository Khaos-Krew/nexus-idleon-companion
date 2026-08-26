package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func usage() {
	fmt.Println("IdleOn Account Agent")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  detect      Detect IdleOn and its window geometry")
	fmt.Println("  doctor      Validate game, data source, config and calibration")
	fmt.Println("  assess      Rank account-wide progression and character assignments")
	fmt.Println("  agent       Continuously reassess and optionally execute safe routines")
	fmt.Println("  serve       Launch the local account dashboard")
	fmt.Println("  list        Show routines, bindings, calibration and green stacks")
	fmt.Println("  calibrate   Capture named click points relative to the IdleOn window")
	fmt.Println("  run         Run one named calibrated routine")
	fmt.Println("  greenstack  Rotate unfinished configured green-stack targets")
	fmt.Println("  done        Mark or reopen a permanent green-stack target")
	fmt.Println("")
	fmt.Println("Sources can be fused: local companion snapshot + Efficiency JSON + public Toolbox profile.")
	fmt.Println("Emergency stop: F12 by default")
}

func configFlag(fs *flag.FlagSet) *string { return fs.String("config", "automation.json", "automation configuration path") }

func providerFlags(fs *flag.FlagSet) (snapshot, efficiency, toolbox *string) {
	snapshot = fs.String("snapshot", "auto", "local/normalized JSON; auto discovers newest companion snapshot")
	efficiency = fs.String("efficiency", "", "IdleOn Efficiency JSON/export path")
	toolbox = fs.String("toolbox", "", "public IdleOn Toolbox profile/main character")
	return
}

func chooseProvider(snapshot, efficiency, toolbox string) (SnapshotProvider, error) {
	var providers []SnapshotProvider
	local := strings.TrimSpace(snapshot)
	if local == "" || strings.EqualFold(local, "auto") {
		providers = append(providers, AutoLocalProvider{})
	} else if !strings.EqualFold(local, "none") {
		providers = append(providers, FileSnapshotProvider{Path: local, Source: "local-companion-snapshot"})
	}
	if strings.TrimSpace(efficiency) != "" {
		providers = append(providers, FileSnapshotProvider{Path: efficiency, Source: "idleon-efficiency-import"})
	}
	if strings.TrimSpace(toolbox) != "" {
		providers = append(providers, ToolboxProvider{Profile: toolbox})
	}
	if len(providers) == 0 { return nil, fmt.Errorf("no account source selected") }
	if len(providers) == 1 { return providers[0], nil }
	return FusionProvider{Providers: providers}, nil
}

func waitForCapture(input InputDriver, captureKey, stopKey string) (Point, error) {
	fmt.Printf("Move the mouse to the target and press %s. Press %s to abort.\n", captureKey, stopKey)
	for {
		if input.KeyDown(stopKey) { return Point{}, fmt.Errorf("calibration aborted") }
		if input.KeyDown(captureKey) {
			p, err := input.CursorPosition()
			if err != nil { return Point{}, err }
			for input.KeyDown(captureKey) { time.Sleep(40 * time.Millisecond) }
			return p, nil
		}
		time.Sleep(40 * time.Millisecond)
	}
}

func printAssessment(a Assessment) {
	fmt.Printf("IdleOn: running=%v foreground=%v | %s | source=%s | world=%d | health=%d%% | coverage=%d%%\n", a.Game.Running, a.Game.Foreground, a.Stage, a.Source, a.World, a.HealthScore, a.CoveragePercent)
	if a.ActiveCharacter != nil { fmt.Printf("Active: %s (%s) %s\n", a.ActiveCharacter.Name, a.ActiveCharacter.Class, a.ActiveCharacter.Map) }
	for _, warning := range a.Warnings { fmt.Println("WARNING:", warning) }
	fmt.Println("Top priorities:")
	for i, r := range a.Top {
		if i >= 10 { break }
		who := r.Character
		if who == "" { who = r.PreferredClass }
		auto := ""
		if r.Automatable { auto = " {routine:" + r.Routine + "}" }
		fmt.Printf("  %d. %.1f [%d%% conf] %s [%s]%s — %s\n", i+1, r.Score, int(r.Confidence*100), r.Title, who, auto, r.Action)
	}
	if len(a.CharacterPlan) > 0 {
		fmt.Println("Character plan:")
		for i, p := range a.CharacterPlan { if i >= 10 { break }; fmt.Printf("  - %s (%s): %s\n", p.Character, p.Class, p.Role) }
	}
}

func loadAssessment(cfg Config, provider SnapshotProvider) (Assessment, error) {
	snap, err := provider.Load()
	if err != nil { return Assessment{}, err }
	return buildAssessmentWithConfig(DetectGame(cfg.WindowTitle), snap, cfg), nil
}

func cmdDetect(args []string) error {
	fs := flag.NewFlagSet("detect", flag.ContinueOnError)
	title := fs.String("window", "Legends of IdleOn", "IdleOn window title")
	if err := fs.Parse(args); err != nil { return err }
	raw, _ := json.MarshalIndent(DetectGame(*title), "", "  ")
	fmt.Println(string(raw))
	return nil
}

func cmdDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	path := configFlag(fs)
	snapshot, efficiency, toolbox := providerFlags(fs)
	if err := fs.Parse(args); err != nil { return err }
	cfg, err := loadConfig(*path)
	if err != nil { return fmt.Errorf("config: %w", err) }
	provider, err := chooseProvider(*snapshot, *efficiency, *toolbox)
	if err != nil { return err }
	game := DetectGame(cfg.WindowTitle)
	fmt.Printf("Game running: %v | foreground: %v | window: %dx%d at %d,%d\n", game.Running, game.Foreground, game.Width, game.Height, game.X, game.Y)
	snap, err := provider.Load()
	if err != nil { return fmt.Errorf("account source: %w", err) }
	fmt.Printf("Account source: %s | world: %d | characters: %d | systems: %d\n", snap.Source, snap.World, len(snap.Characters), len(snap.Systems))
	var missing []string
	for _, name := range cfg.requiredCalibrationPoints() { if _, ok := cfg.Calibration[name]; !ok { missing = append(missing, name) } }
	sort.Strings(missing)
	fmt.Printf("Routines: %d | calibrated points: %d | missing points: %d\n", len(cfg.Routines), len(cfg.Calibration), len(missing))
	if len(missing) > 0 { fmt.Println("Missing:", strings.Join(missing, ", ")) }
	a := buildAssessmentWithConfig(game, snap, cfg)
	fmt.Printf("Planner coverage: %d%% | health: %d%% | top recommendations: %d\n", a.CoveragePercent, a.HealthScore, len(a.Top))
	return nil
}

func cmdAssess(args []string) error {
	fs := flag.NewFlagSet("assess", flag.ContinueOnError)
	path := configFlag(fs)
	snapshot, efficiency, toolbox := providerFlags(fs)
	out := fs.String("out", "", "optional assessment JSON output")
	if err := fs.Parse(args); err != nil { return err }
	cfg, err := loadConfig(*path); if err != nil { return err }
	provider, err := chooseProvider(*snapshot, *efficiency, *toolbox); if err != nil { return err }
	a, err := loadAssessment(cfg, provider); if err != nil { return err }
	printAssessment(a)
	_ = NewAgentStore("").SaveAssessment(a)
	if *out != "" { raw, _ := json.MarshalIndent(a, "", "  "); return os.WriteFile(*out, append(raw, '\n'), 0o644) }
	return nil
}

func executeRecommendation(cfg Config, runner *Runner, store *AgentStore, rec *Recommendation) error {
	if rec == nil { return nil }
	allowed, reason := store.CanExecute(cfg)
	if !allowed { fmt.Println("Execution paused:", reason); return nil }
	if rec.SwitchRoutine != "" {
		if err := runner.RunRoutine(rec.SwitchRoutine); err != nil { return err }
	}
	fmt.Printf("Executing %s via %s\n", rec.Title, rec.Routine)
	err := runner.RunRoutine(rec.Routine)
	record := ActionRecord{Routine: rec.Routine, Objective: rec.Title, Character: rec.Character, Success: err == nil}
	if err != nil { record.Error = err.Error() }
	_ = store.LogAction(record)
	return err
}

func cmdAgent(args []string) error {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	path := configFlag(fs)
	snapshot, efficiency, toolbox := providerFlags(fs)
	interval := fs.Int("interval", 0, "assessment interval seconds; 0 uses config")
	execute := fs.Bool("execute", false, "allow calibrated routine execution")
	foregroundOnly := fs.Bool("foreground-only", true, "only execute while IdleOn is foreground")
	cycles := fs.Int("cycles", 0, "cycles; 0 runs until F12/Ctrl+C")
	if err := fs.Parse(args); err != nil { return err }
	cfg, err := loadConfig(*path); if err != nil { return err }
	provider, err := chooseProvider(*snapshot, *efficiency, *toolbox); if err != nil { return err }
	input := NewInputDriver()
	runner := NewRunner(cfg, input)
	store := NewAgentStore("")
	if *interval <= 0 { *interval = cfg.Settings.AssessIntervalSec }
	if *interval < 5 { *interval = 5 }
	last := map[string]time.Time{}

	for n := 0; *cycles == 0 || n < *cycles; n++ {
		if input.KeyDown(cfg.EmergencyStopKey) { return fmt.Errorf("emergency stop requested") }
		game := DetectGame(cfg.WindowTitle)
		if !game.Running {
			fmt.Println("IdleOn not running; execution paused.")
		} else {
			snap, loadErr := provider.Load()
			if loadErr != nil {
				fmt.Println("Account source error:", loadErr)
			} else {
				a := buildAssessmentWithConfig(game, snap, cfg)
				printAssessment(a)
				_ = store.SaveAssessment(a)
				if *execute && (!*foregroundOnly || game.Foreground) {
					rec := chooseAutomation(a)
					if rec != nil {
						cooldown := cfg.Settings.DefaultCooldownMinutes
						if rt, ok := cfg.Routines[rec.Routine]; ok && rt.CooldownMinutes > 0 { cooldown = rt.CooldownMinutes }
						if time.Since(last[rec.ID]) >= time.Duration(cooldown)*time.Minute {
							if err := executeRecommendation(cfg, runner, store, rec); err != nil { return err }
							last[rec.ID] = time.Now()
						}
					}
				}
			}
		}
		if *cycles != 0 && n+1 >= *cycles { break }
		if err := runner.sleep(*interval * 1000); err != nil { return err }
	}
	return nil
}

func cmdServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	path := configFlag(fs)
	snapshot, efficiency, toolbox := providerFlags(fs)
	port := fs.Int("port", 17654, "local dashboard port")
	if err := fs.Parse(args); err != nil { return err }
	cfg, err := loadConfig(*path); if err != nil { return err }
	provider, err := chooseProvider(*snapshot, *efficiency, *toolbox); if err != nil { return err }
	return NewDashboard(cfg, provider, NewAgentStore("")).Serve(*port)
}

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	path := configFlag(fs)
	if err := fs.Parse(args); err != nil { return err }
	cfg, err := loadConfig(*path); if err != nil { return err }
	var names []string
	for n := range cfg.Routines { names = append(names, n) }
	sort.Strings(names)
	fmt.Println("Routines:")
	for _, n := range names { fmt.Printf("  - %s: %s\n", n, cfg.Routines[n].Description) }
	fmt.Println("System bindings:")
	var systems []string
	for s := range cfg.SystemRoutines { systems = append(systems, s) }
	sort.Strings(systems)
	for _, s := range systems { fmt.Printf("  - %s -> %s\n", s, cfg.SystemRoutines[s]) }
	points := cfg.requiredCalibrationPoints(); sort.Strings(points)
	fmt.Println("Calibration:")
	for _, n := range points {
		if p, ok := cfg.Calibration[n]; ok { fmt.Printf("  - %s = (%d,%d) normalized(%.4f,%.4f)\n", n, p.X, p.Y, p.NX, p.NY) } else { fmt.Printf("  - %s = NOT SET\n", n) }
	}
	fmt.Println("Green stacks:")
	for _, t := range cfg.GreenstackTargets { state := "disabled"; if t.Enabled { state = "enabled" }; if t.Completed { state = "complete" }; fmt.Printf("  - %s [%s] %dm via %s\n", t.Name, state, t.FarmMinutes, t.TravelRoutine) }
	return nil
}

func cmdCalibrate(args []string) error {
	fs := flag.NewFlagSet("calibrate", flag.ContinueOnError)
	path := configFlag(fs)
	if err := fs.Parse(args); err != nil { return err }
	cfg, err := loadConfig(*path); if err != nil { return err }
	input := NewInputDriver()
	game := DetectGame(cfg.WindowTitle)
	if !game.Running { return fmt.Errorf("start IdleOn before calibration") }
	if game.Width <= 0 || game.Height <= 0 { return fmt.Errorf("could not read IdleOn window geometry") }
	points := cfg.requiredCalibrationPoints(); sort.Strings(points)
	if len(points) == 0 { return fmt.Errorf("no named click points are referenced") }
	for _, name := range points {
		fmt.Printf("\nCalibrating: %s\n", name)
		p, err := waitForCapture(input, cfg.CapturePointKey, cfg.EmergencyStopKey); if err != nil { return err }
		p.NX = float64(p.X-game.X) / float64(game.Width)
		p.NY = float64(p.Y-game.Y) / float64(game.Height)
		if p.NX < 0 || p.NX > 1 || p.NY < 0 || p.NY > 1 { return fmt.Errorf("point %s is outside IdleOn window", name) }
		cfg.Calibration[name] = p
		fmt.Printf("Saved %s = (%d,%d), normalized (%.4f,%.4f)\n", name, p.X, p.Y, p.NX, p.NY)
		if err := saveConfig(*path, cfg); err != nil { return err }
	}
	fmt.Println("Calibration complete.")
	return nil
}

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError); path := configFlag(fs); routine := fs.String("routine", "", "routine name")
	if err := fs.Parse(args); err != nil { return err }; if *routine == "" { return fmt.Errorf("-routine is required") }
	cfg, err := loadConfig(*path); if err != nil { return err }; return NewRunner(cfg, NewInputDriver()).RunRoutine(*routine)
}

func cmdGreenstack(args []string) error {
	fs := flag.NewFlagSet("greenstack", flag.ContinueOnError); path := configFlag(fs); cycles := fs.Int("cycles", 1, "full target cycles")
	if err := fs.Parse(args); err != nil { return err }; cfg, err := loadConfig(*path); if err != nil { return err }; return NewRunner(cfg, NewInputDriver()).RunGreenstackLoop(*cycles)
}

func cmdDone(args []string) error {
	fs := flag.NewFlagSet("done", flag.ContinueOnError); path := configFlag(fs); target := fs.String("target", "", "green-stack target name"); reopen := fs.Bool("reopen", false, "mark target incomplete again")
	if err := fs.Parse(args); err != nil { return err }; if strings.TrimSpace(*target) == "" { return fmt.Errorf("-target is required") }
	cfg, err := loadConfig(*path); if err != nil { return err }; if !cfg.markTargetCompleted(*target, !*reopen) { return fmt.Errorf("green-stack target not found: %s", *target) }
	if err := saveConfig(*path, cfg); err != nil { return err }; state := "complete"; if *reopen { state = "open" }; fmt.Printf("Marked %s as %s.\n", *target, state); return nil
}

func main() {
	if len(os.Args) < 2 { usage(); os.Exit(2) }
	var err error
	switch os.Args[1] {
	case "detect": err = cmdDetect(os.Args[2:])
	case "doctor": err = cmdDoctor(os.Args[2:])
	case "assess": err = cmdAssess(os.Args[2:])
	case "agent": err = cmdAgent(os.Args[2:])
	case "serve": err = cmdServe(os.Args[2:])
	case "list": err = cmdList(os.Args[2:])
	case "calibrate": err = cmdCalibrate(os.Args[2:])
	case "run": err = cmdRun(os.Args[2:])
	case "greenstack": err = cmdGreenstack(os.Args[2:])
	case "done": err = cmdDone(os.Args[2:])
	case "help", "-h", "--help": usage(); return
	default: usage(); err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil { fmt.Fprintln(os.Stderr, "ERROR:", err); os.Exit(1) }
}

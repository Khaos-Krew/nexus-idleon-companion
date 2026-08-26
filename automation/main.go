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
    fmt.Println("  detect      Detect whether IdleOn is running and foreground")
    fmt.Println("  assess      Load account data and rank all major progression systems")
    fmt.Println("  agent       Continuously detect the active game, reassess and optionally execute calibrated routines")
    fmt.Println("  list        Show routines, calibration points and green-stack targets")
    fmt.Println("  calibrate   Capture all named click points with F8")
    fmt.Println("  run         Run one named routine")
    fmt.Println("  greenstack  Rotate through enabled green-stack targets")
    fmt.Println("  done        Mark a permanent green-stack target complete")
    fmt.Println("")
    fmt.Println("Data sources: local/companion JSON, IdleOn Efficiency JSON export, or public IdleOn Toolbox profile")
    fmt.Println("Emergency stop: F12 by default")
}

func configFlag(fs *flag.FlagSet) *string {
    return fs.String("config", "automation.json", "path to automation configuration")
}

func waitForCapture(input InputDriver, captureKey, stopKey string) (Point, error) {
    fmt.Printf("Move the mouse to the target and press %s. Press %s to abort.\n", captureKey, stopKey)
    for {
        if input.KeyDown(stopKey) { return Point{}, fmt.Errorf("calibration aborted") }
        if input.KeyDown(captureKey) {
            point, err := input.CursorPosition(); if err != nil { return Point{}, err }
            for input.KeyDown(captureKey) { time.Sleep(40 * time.Millisecond) }
            return point, nil
        }
        time.Sleep(40 * time.Millisecond)
    }
}

func cmdDetect(args []string) error {
    fs := flag.NewFlagSet("detect", flag.ContinueOnError)
    title := fs.String("window", "Legends of IdleOn", "IdleOn window title")
    if err := fs.Parse(args); err != nil { return err }
    raw, _ := json.MarshalIndent(DetectGame(*title), "", "  ")
    fmt.Println(string(raw))
    return nil
}

func providerFlags(fs *flag.FlagSet) (snapshot, efficiency, toolbox *string) {
    snapshot = fs.String("snapshot", "", "path to normalized/local companion account JSON")
    efficiency = fs.String("efficiency", "", "path to IdleOn Efficiency JSON/export")
    toolbox = fs.String("toolbox", "", "public IdleOn Toolbox profile/main character")
    return
}

func chooseProvider(snapshot, efficiency, toolbox string) (SnapshotProvider, error) {
    if strings.TrimSpace(toolbox) != "" { return ToolboxProvider{Profile: toolbox}, nil }
    if strings.TrimSpace(efficiency) != "" { return FileSnapshotProvider{Path: efficiency, Source: "idleon-efficiency-import"}, nil }
    if strings.TrimSpace(snapshot) != "" { return FileSnapshotProvider{Path: snapshot, Source: "local-companion-snapshot"}, nil }
    return nil, fmt.Errorf("choose one account source: -snapshot, -efficiency, or -toolbox")
}

func printAssessment(a Assessment) {
    fmt.Printf("IdleOn: running=%v foreground=%v | source=%s | account=%s | world=%d | health=%d%%\n", a.Game.Running, a.Game.Foreground, a.Source, a.AccountName, a.World, a.HealthScore)
    if a.ActiveCharacter != nil { fmt.Printf("Active: %s (%s) %s\n", a.ActiveCharacter.Name, a.ActiveCharacter.Class, a.ActiveCharacter.Map) }
    fmt.Println("Top priorities:")
    for i, r := range a.Top {
        if i >= 8 { break }
        who := ""; if r.PreferredClass != "" { who = " [" + r.PreferredClass + "]" }
        auto := ""; if r.Automatable { auto = " {routine:" + r.Routine + "}" }
        fmt.Printf("  %d. %.1f %s%s%s — %s\n", i+1, r.Score, r.Title, who, auto, r.Action)
    }
}

func cmdAssess(args []string) error {
    fs := flag.NewFlagSet("assess", flag.ContinueOnError)
    title := fs.String("window", "Legends of IdleOn", "IdleOn window title")
    snapshot, efficiency, toolbox := providerFlags(fs)
    out := fs.String("out", "", "optional assessment JSON output path")
    if err := fs.Parse(args); err != nil { return err }
    provider, err := chooseProvider(*snapshot, *efficiency, *toolbox); if err != nil { return err }
    snap, err := provider.Load(); if err != nil { return err }
    assessment := buildAssessment(DetectGame(*title), snap)
    printAssessment(assessment)
    if *out != "" {
        raw, _ := json.MarshalIndent(assessment, "", "  ")
        if err := os.WriteFile(*out, append(raw, '\n'), 0o644); err != nil { return err }
    }
    return nil
}

func cmdAgent(args []string) error {
    fs := flag.NewFlagSet("agent", flag.ContinueOnError)
    configPath := configFlag(fs)
    snapshot, efficiency, toolbox := providerFlags(fs)
    interval := fs.Int("interval", 60, "assessment interval in seconds")
    foregroundOnly := fs.Bool("foreground-only", true, "only execute routines while IdleOn is foreground")
    execute := fs.Bool("execute", false, "allow calibrated routine execution")
    cycles := fs.Int("cycles", 0, "assessment cycles; 0 means run until F12/Ctrl+C")
    if err := fs.Parse(args); err != nil { return err }
    cfg, err := loadConfig(*configPath); if err != nil { return err }
    provider, err := chooseProvider(*snapshot, *efficiency, *toolbox); if err != nil { return err }
    input := NewInputDriver(); runner := NewRunner(cfg, input)
    if *interval < 5 { *interval = 5 }
    lastExecuted := ""; lastExecutedAt := time.Time{}

    for n := 0; *cycles == 0 || n < *cycles; n++ {
        if input.KeyDown(cfg.EmergencyStopKey) { return fmt.Errorf("emergency stop requested") }
        game := DetectGame(cfg.WindowTitle)
        if !game.Running {
            fmt.Println("IdleOn is not running; assessment/execution paused.")
        } else {
            snap, loadErr := provider.Load()
            if loadErr != nil { fmt.Println("Account source error:", loadErr) } else {
                assessment := buildAssessment(game, snap)
                printAssessment(assessment)
                if *execute && (!*foregroundOnly || game.Foreground) {
                    if rec := chooseAutomation(assessment); rec != nil {
                        cooldownMet := time.Since(lastExecutedAt) >= 15*time.Minute
                        if rec.ID != lastExecuted || cooldownMet {
                            fmt.Printf("Executing calibrated recommendation: %s via %s\n", rec.Title, rec.Routine)
                            if err := runner.RunRoutine(rec.Routine); err != nil { return err }
                            lastExecuted, lastExecutedAt = rec.ID, time.Now()
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

func cmdList(args []string) error {
    fs := flag.NewFlagSet("list", flag.ContinueOnError)
    path := configFlag(fs)
    if err := fs.Parse(args); err != nil { return err }
    cfg, err := loadConfig(*path); if err != nil { return err }
    routineNames := make([]string, 0, len(cfg.Routines)); for name := range cfg.Routines { routineNames = append(routineNames, name) }; sort.Strings(routineNames)
    fmt.Println("Routines:"); for _, name := range routineNames { fmt.Printf("  - %s: %s\n", name, cfg.Routines[name].Description) }
    points := cfg.requiredCalibrationPoints(); sort.Strings(points); fmt.Println("Calibration points:")
    for _, name := range points { if p, ok := cfg.Calibration[name]; ok { fmt.Printf("  - %s = (%d,%d)\n", name,p.X,p.Y) } else { fmt.Printf("  - %s = NOT SET\n", name) } }
    fmt.Println("Green-stack targets:")
    for _, target := range cfg.GreenstackTargets { state := "disabled"; if target.Enabled { state="enabled" }; if target.Completed { state="complete" }; fmt.Printf("  - %s [%s] %dm via %s\n", target.Name,state,target.FarmMinutes,target.TravelRoutine) }
    return nil
}

func cmdCalibrate(args []string) error {
    fs := flag.NewFlagSet("calibrate", flag.ContinueOnError); path := configFlag(fs); if err := fs.Parse(args); err != nil { return err }
    cfg, err := loadConfig(*path); if err != nil { return err }; input := NewInputDriver(); points := cfg.requiredCalibrationPoints(); sort.Strings(points)
    if len(points)==0 { return fmt.Errorf("no named click points are referenced by any routine") }
    for _, name := range points { fmt.Printf("\nCalibrating: %s\n",name); point,err:=waitForCapture(input,cfg.CapturePointKey,cfg.EmergencyStopKey); if err!=nil{return err}; cfg.Calibration[name]=point; fmt.Printf("Saved %s = (%d,%d)\n",name,point.X,point.Y); if err:=saveConfig(*path,cfg);err!=nil{return err} }
    fmt.Println("Calibration complete."); return nil
}

func cmdRun(args []string) error {
    fs:=flag.NewFlagSet("run",flag.ContinueOnError); path:=configFlag(fs); routine:=fs.String("routine","","routine name"); if err:=fs.Parse(args);err!=nil{return err}; if *routine==""{return fmt.Errorf("-routine is required")}; cfg,err:=loadConfig(*path);if err!=nil{return err}; return NewRunner(cfg,NewInputDriver()).RunRoutine(*routine)
}

func cmdGreenstack(args []string) error {
    fs:=flag.NewFlagSet("greenstack",flag.ContinueOnError); path:=configFlag(fs); cycles:=fs.Int("cycles",1,"number of full target cycles");if err:=fs.Parse(args);err!=nil{return err};cfg,err:=loadConfig(*path);if err!=nil{return err};return NewRunner(cfg,NewInputDriver()).RunGreenstackLoop(*cycles)
}

func cmdDone(args []string) error {
    fs:=flag.NewFlagSet("done",flag.ContinueOnError); path:=configFlag(fs);target:=fs.String("target","","green-stack target name");reopen:=fs.Bool("reopen",false,"mark target incomplete again");if err:=fs.Parse(args);err!=nil{return err};if strings.TrimSpace(*target)==""{return fmt.Errorf("-target is required")};cfg,err:=loadConfig(*path);if err!=nil{return err};if !cfg.markTargetCompleted(*target,!*reopen){return fmt.Errorf("green-stack target not found: %s",*target)};if err:=saveConfig(*path,cfg);err!=nil{return err};state:="complete";if *reopen{state="open"};fmt.Printf("Marked %s as %s.\n",*target,state);return nil
}

func main() {
    if len(os.Args)<2 { usage(); os.Exit(2) }
    var err error
    switch os.Args[1] {
    case "detect": err=cmdDetect(os.Args[2:])
    case "assess": err=cmdAssess(os.Args[2:])
    case "agent": err=cmdAgent(os.Args[2:])
    case "list": err=cmdList(os.Args[2:])
    case "calibrate": err=cmdCalibrate(os.Args[2:])
    case "run": err=cmdRun(os.Args[2:])
    case "greenstack": err=cmdGreenstack(os.Args[2:])
    case "done": err=cmdDone(os.Args[2:])
    case "help","-h","--help": usage(); return
    default: usage(); err=fmt.Errorf("unknown command %q",os.Args[1])
    }
    if err!=nil { fmt.Fprintln(os.Stderr,"ERROR:",err); os.Exit(1) }
}

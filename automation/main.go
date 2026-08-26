package main

import (
    "flag"
    "fmt"
    "os"
    "sort"
    "time"
)

func usage() {
    fmt.Println("IdleOn Automation V1")
    fmt.Println("")
    fmt.Println("Commands:")
    fmt.Println("  list       Show routines, calibration points and green-stack targets")
    fmt.Println("  calibrate  Capture all named click points with F8")
    fmt.Println("  run        Run one named routine")
    fmt.Println("  greenstack Rotate through enabled unfinished Bubo green-stack targets")
    fmt.Println("  done       Mark a 10M+ green stack as permanently completed")
    fmt.Println("  undo-done  Reopen a completed green-stack target")
    fmt.Println("")
    fmt.Println("Emergency stop: F12 by default")
}

func configFlag(fs *flag.FlagSet) *string {
    return fs.String("config", "automation.json", "path to automation configuration")
}

func waitForCapture(input InputDriver, captureKey, stopKey string) (Point, error) {
    fmt.Printf("Move the mouse to the target and press %s. Press %s to abort.\n", captureKey, stopKey)
    for {
        if input.KeyDown(stopKey) {
            return Point{}, fmt.Errorf("calibration aborted")
        }
        if input.KeyDown(captureKey) {
            point, err := input.CursorPosition()
            if err != nil {
                return Point{}, err
            }
            for input.KeyDown(captureKey) {
                time.Sleep(40 * time.Millisecond)
            }
            return point, nil
        }
        time.Sleep(40 * time.Millisecond)
    }
}

func cmdList(args []string) error {
    fs := flag.NewFlagSet("list", flag.ContinueOnError)
    path := configFlag(fs)
    if err := fs.Parse(args); err != nil { return err }
    cfg, err := loadConfig(*path)
    if err != nil { return err }

    routineNames := make([]string, 0, len(cfg.Routines))
    for name := range cfg.Routines { routineNames = append(routineNames, name) }
    sort.Strings(routineNames)
    fmt.Println("Routines:")
    for _, name := range routineNames {
        fmt.Printf("  - %s: %s\n", name, cfg.Routines[name].Description)
    }

    points := cfg.requiredCalibrationPoints()
    sort.Strings(points)
    fmt.Println("Calibration points:")
    for _, name := range points {
        if p, ok := cfg.Calibration[name]; ok {
            fmt.Printf("  - %s = (%d,%d)\n", name, p.X, p.Y)
        } else {
            fmt.Printf("  - %s = NOT SET\n", name)
        }
    }

    fmt.Println("Green-stack targets:")
    for _, target := range cfg.GreenstackTargets {
        state := "disabled"
        if target.Completed {
            state = "completed"
        } else if target.Enabled {
            state = "enabled"
        }
        fmt.Printf("  - %s [%s] %dm via %s\n", target.Name, state, target.FarmMinutes, target.TravelRoutine)
    }
    return nil
}

func cmdCalibrate(args []string) error {
    fs := flag.NewFlagSet("calibrate", flag.ContinueOnError)
    path := configFlag(fs)
    if err := fs.Parse(args); err != nil { return err }
    cfg, err := loadConfig(*path)
    if err != nil { return err }
    input := NewInputDriver()

    points := cfg.requiredCalibrationPoints()
    sort.Strings(points)
    if len(points) == 0 {
        return fmt.Errorf("no named click points are referenced by any routine")
    }
    for _, name := range points {
        fmt.Printf("\nCalibrating: %s\n", name)
        point, err := waitForCapture(input, cfg.CapturePointKey, cfg.EmergencyStopKey)
        if err != nil { return err }
        cfg.Calibration[name] = point
        fmt.Printf("Saved %s = (%d,%d)\n", name, point.X, point.Y)
        if err := saveConfig(*path, cfg); err != nil { return err }
    }
    fmt.Println("Calibration complete.")
    return nil
}

func cmdRun(args []string) error {
    fs := flag.NewFlagSet("run", flag.ContinueOnError)
    path := configFlag(fs)
    routine := fs.String("routine", "", "routine name")
    if err := fs.Parse(args); err != nil { return err }
    if *routine == "" { return fmt.Errorf("-routine is required") }
    cfg, err := loadConfig(*path)
    if err != nil { return err }
    runner := NewRunner(cfg, NewInputDriver())
    return runner.RunRoutine(*routine)
}

func cmdGreenstack(args []string) error {
    fs := flag.NewFlagSet("greenstack", flag.ContinueOnError)
    path := configFlag(fs)
    cycles := fs.Int("cycles", 1, "number of full target cycles")
    if err := fs.Parse(args); err != nil { return err }
    cfg, err := loadConfig(*path)
    if err != nil { return err }
    runner := NewRunner(cfg, NewInputDriver())
    return runner.RunGreenstackLoop(*cycles)
}

func cmdTargetCompletion(args []string, completed bool) error {
    name := "done"
    if !completed { name = "undo-done" }
    fs := flag.NewFlagSet(name, flag.ContinueOnError)
    path := configFlag(fs)
    target := fs.String("target", "", "green-stack target name")
    if err := fs.Parse(args); err != nil { return err }
    if *target == "" { return fmt.Errorf("-target is required") }
    cfg, err := loadConfig(*path)
    if err != nil { return err }
    if !cfg.markTargetCompleted(*target, completed) {
        return fmt.Errorf("green-stack target not found: %s", *target)
    }
    if err := saveConfig(*path, cfg); err != nil { return err }
    if completed {
        fmt.Printf("Marked %s as permanently greenstacked. It will be skipped from now on.\n", *target)
    } else {
        fmt.Printf("Reopened %s. It can be farmed again.\n", *target)
    }
    return nil
}

func main() {
    if len(os.Args) < 2 {
        usage()
        os.Exit(2)
    }
    var err error
    switch os.Args[1] {
    case "list": err = cmdList(os.Args[2:])
    case "calibrate": err = cmdCalibrate(os.Args[2:])
    case "run": err = cmdRun(os.Args[2:])
    case "greenstack": err = cmdGreenstack(os.Args[2:])
    case "done": err = cmdTargetCompletion(os.Args[2:], true)
    case "undo-done": err = cmdTargetCompletion(os.Args[2:], false)
    case "help", "-h", "--help": usage(); return
    default:
        usage()
        err = fmt.Errorf("unknown command %q", os.Args[1])
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, "ERROR:", err)
        os.Exit(1)
    }
}

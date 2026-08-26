package main

import (
    "errors"
    "fmt"
    "strings"
    "time"
)

type Runner struct {
    cfg     Config
    input   InputDriver
    stopKey string
}

func NewRunner(cfg Config, input InputDriver) *Runner {
    return &Runner{cfg: cfg, input: input, stopKey: strings.ToUpper(cfg.EmergencyStopKey)}
}

func (r *Runner) stopped() bool {
    return r.input.KeyDown(r.stopKey)
}

func (r *Runner) sleep(ms int) error {
    if ms <= 0 {
        return nil
    }
    deadline := time.Now().Add(time.Duration(ms) * time.Millisecond)
    for time.Now().Before(deadline) {
        if r.stopped() {
            return errors.New("emergency stop requested")
        }
        remaining := time.Until(deadline)
        if remaining > 100*time.Millisecond {
            remaining = 100 * time.Millisecond
        }
        time.Sleep(remaining)
    }
    return nil
}

func (r *Runner) RunRoutine(name string) error {
    routine, ok := r.cfg.Routines[name]
    if !ok {
        return fmt.Errorf("unknown routine %q", name)
    }
    fmt.Printf("Running routine: %s\n", name)
    for i, step := range routine.Steps {
        if r.stopped() {
            return errors.New("emergency stop requested")
        }
        switch strings.ToLower(strings.TrimSpace(step.Type)) {
        case "focus":
            if err := r.input.FocusWindow(r.cfg.WindowTitle); err != nil {
                return fmt.Errorf("step %d focus: %w", i+1, err)
            }
        case "click":
            point, err := r.cfg.resolvePoint(step)
            if err != nil {
                return fmt.Errorf("step %d click: %w", i+1, err)
            }
            if err := r.input.Click(point.X, point.Y, false); err != nil {
                return fmt.Errorf("step %d click: %w", i+1, err)
            }
        case "double-click":
            point, err := r.cfg.resolvePoint(step)
            if err != nil {
                return fmt.Errorf("step %d double-click: %w", i+1, err)
            }
            if err := r.input.Click(point.X, point.Y, true); err != nil {
                return fmt.Errorf("step %d double-click: %w", i+1, err)
            }
        case "key":
            if err := r.input.PressKey(step.Key); err != nil {
                return fmt.Errorf("step %d key: %w", i+1, err)
            }
        case "wait":
            if err := r.sleep(step.MS); err != nil {
                return err
            }
        default:
            return fmt.Errorf("step %d has unsupported type %q", i+1, step.Type)
        }
        if step.DelayAfterMS > 0 {
            if err := r.sleep(step.DelayAfterMS); err != nil {
                return err
            }
        }
    }
    return nil
}

func (r *Runner) RunGreenstackLoop(cycles int) error {
    targets := make([]GreenstackTarget, 0, len(r.cfg.GreenstackTargets))
    for _, target := range r.cfg.GreenstackTargets {
        if target.Enabled && !target.Completed {
            targets = append(targets, target)
        }
    }
    if len(targets) == 0 {
        return errors.New("no enabled, unfinished greenstackTargets in config")
    }
    if cycles < 1 {
        cycles = 1
    }

    for cycle := 1; cycle <= cycles; cycle++ {
        fmt.Printf("Green-stack cycle %d/%d\n", cycle, cycles)
        for _, target := range targets {
            if r.stopped() {
                return errors.New("emergency stop requested")
            }
            fmt.Printf("Target: %s\n", target.Name)
            if target.TravelRoutine != "" {
                if err := r.RunRoutine(target.TravelRoutine); err != nil {
                    return fmt.Errorf("travel to %s: %w", target.Name, err)
                }
            }
            minutes := target.FarmMinutes
            if minutes < 1 {
                minutes = 1
            }
            fmt.Printf("Farming %s for %d minute(s). Press %s to stop.\n", target.Name, minutes, r.stopKey)
            if err := r.sleep(minutes * 60 * 1000); err != nil {
                return err
            }
        }
    }
    return nil
}

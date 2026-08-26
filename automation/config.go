package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "strings"
)

type Point struct {
    X int `json:"x"`
    Y int `json:"y"`
}

type Step struct {
    Type         string `json:"type"`
    Point        string `json:"point,omitempty"`
    X            int    `json:"x,omitempty"`
    Y            int    `json:"y,omitempty"`
    Key          string `json:"key,omitempty"`
    MS           int    `json:"ms,omitempty"`
    DelayAfterMS int    `json:"delayAfterMs,omitempty"`
}

type Routine struct {
    Description string `json:"description,omitempty"`
    Steps       []Step `json:"steps"`
}

type GreenstackTarget struct {
    Name          string `json:"name"`
    FarmMinutes   int    `json:"farmMinutes"`
    TravelRoutine string `json:"travelRoutine"`
    Enabled       bool   `json:"enabled"`
}

type Config struct {
    WindowTitle       string                      `json:"windowTitle"`
    EmergencyStopKey  string                      `json:"emergencyStopKey"`
    CapturePointKey   string                      `json:"capturePointKey"`
    Calibration       map[string]Point            `json:"calibration"`
    Routines          map[string]Routine          `json:"routines"`
    GreenstackTargets []GreenstackTarget          `json:"greenstackTargets"`
}

func defaultConfig() Config {
    return Config{
        WindowTitle:      "Legends of IdleOn",
        EmergencyStopKey: "F12",
        CapturePointKey:  "F8",
        Calibration:      map[string]Point{},
        Routines:         map[string]Routine{},
    }
}

func loadConfig(path string) (Config, error) {
    raw, err := os.ReadFile(path)
    if err != nil {
        return Config{}, err
    }
    cfg := defaultConfig()
    if err := json.Unmarshal(raw, &cfg); err != nil {
        return Config{}, err
    }
    if cfg.Calibration == nil {
        cfg.Calibration = map[string]Point{}
    }
    if cfg.Routines == nil {
        cfg.Routines = map[string]Routine{}
    }
    if strings.TrimSpace(cfg.WindowTitle) == "" {
        return Config{}, errors.New("windowTitle is required")
    }
    if strings.TrimSpace(cfg.EmergencyStopKey) == "" {
        cfg.EmergencyStopKey = "F12"
    }
    if strings.TrimSpace(cfg.CapturePointKey) == "" {
        cfg.CapturePointKey = "F8"
    }
    return cfg, nil
}

func saveConfig(path string, cfg Config) error {
    raw, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    raw = append(raw, '\n')
    return os.WriteFile(path, raw, 0o644)
}

func (c Config) resolvePoint(step Step) (Point, error) {
    if step.Point != "" {
        p, ok := c.Calibration[step.Point]
        if !ok {
            return Point{}, fmt.Errorf("missing calibration point %q", step.Point)
        }
        return p, nil
    }
    if step.X == 0 && step.Y == 0 {
        return Point{}, errors.New("click step needs point or x/y")
    }
    return Point{X: step.X, Y: step.Y}, nil
}

func (c Config) requiredCalibrationPoints() []string {
    seen := map[string]bool{}
    var out []string
    for _, routine := range c.Routines {
        for _, step := range routine.Steps {
            if step.Point != "" && !seen[step.Point] {
                seen[step.Point] = true
                out = append(out, step.Point)
            }
        }
    }
    return out
}

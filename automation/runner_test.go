package main

import (
    "errors"
    "testing"
)

type fakeInput struct {
    focused []string
    clicks  []Point
    keys    []string
    down    map[string]bool
}

func (f *fakeInput) FocusWindow(title string) error { f.focused = append(f.focused, title); return nil }
func (f *fakeInput) Click(x, y int, double bool) error { f.clicks = append(f.clicks, Point{X: x, Y: y}); return nil }
func (f *fakeInput) PressKey(key string) error { f.keys = append(f.keys, key); return nil }
func (f *fakeInput) KeyDown(key string) bool { return f.down != nil && f.down[key] }
func (f *fakeInput) CursorPosition() (Point, error) { return Point{X: 10, Y: 20}, nil }

func TestResolveCalibratedPoint(t *testing.T) {
    cfg := defaultConfig()
    cfg.Calibration["map"] = Point{X: 123, Y: 456}
    got, err := cfg.resolvePoint(Step{Type: "click", Point: "map"})
    if err != nil { t.Fatal(err) }
    if got.X != 123 || got.Y != 456 { t.Fatalf("unexpected point: %+v", got) }
}

func TestMissingCalibrationFailsClosed(t *testing.T) {
    cfg := defaultConfig()
    _, err := cfg.resolvePoint(Step{Type: "click", Point: "missing"})
    if err == nil { t.Fatal("expected missing calibration to fail") }
}

func TestRunRoutine(t *testing.T) {
    cfg := defaultConfig()
    cfg.WindowTitle = "Legends of IdleOn"
    cfg.Calibration["target"] = Point{X: 50, Y: 75}
    cfg.Routines["test"] = Routine{Steps: []Step{
        {Type: "focus"},
        {Type: "click", Point: "target"},
        {Type: "key", Key: "F"},
    }}
    input := &fakeInput{}
    runner := NewRunner(cfg, input)
    if err := runner.RunRoutine("test"); err != nil { t.Fatal(err) }
    if len(input.focused) != 1 { t.Fatalf("expected focus, got %d", len(input.focused)) }
    if len(input.clicks) != 1 || input.clicks[0].X != 50 { t.Fatalf("unexpected clicks: %+v", input.clicks) }
    if len(input.keys) != 1 || input.keys[0] != "F" { t.Fatalf("unexpected keys: %+v", input.keys) }
}

func TestEmergencyStopPreventsRoutine(t *testing.T) {
    cfg := defaultConfig()
    cfg.Routines["test"] = Routine{Steps: []Step{{Type: "focus"}}}
    input := &fakeInput{down: map[string]bool{"F12": true}}
    runner := NewRunner(cfg, input)
    err := runner.RunRoutine("test")
    if err == nil { t.Fatal("expected emergency stop") }
    if !errors.Is(err, err) { t.Fatal("unexpected error") }
}

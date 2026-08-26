package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Point struct {
	X  int     `json:"x"`
	Y  int     `json:"y"`
	NX float64 `json:"nx,omitempty"`
	NY float64 `json:"ny,omitempty"`
}

type Step struct {
	Type         string `json:"type"`
	Point        string `json:"point,omitempty"`
	X            int    `json:"x,omitempty"`
	Y            int    `json:"y,omitempty"`
	Key          string `json:"key,omitempty"`
	MS           int    `json:"ms,omitempty"`
	DelayAfterMS int    `json:"delayAfterMs,omitempty"`
	Repeat       int    `json:"repeat,omitempty"`
	ExpectedRGB  string `json:"expectedRgb,omitempty"`
	Tolerance    int    `json:"tolerance,omitempty"`
}

type Routine struct {
	Description        string `json:"description,omitempty"`
	Steps              []Step `json:"steps"`
	RequiresForeground bool   `json:"requiresForeground,omitempty"`
	CooldownMinutes    int    `json:"cooldownMinutes,omitempty"`
}

type AgentSettings struct {
	AssessIntervalSec      int `json:"assessIntervalSec,omitempty"`
	MaxSnapshotAgeSec      int `json:"maxSnapshotAgeSec,omitempty"`
	DefaultCooldownMinutes int `json:"defaultCooldownMinutes,omitempty"`
	MaxActionsPerHour      int `json:"maxActionsPerHour,omitempty"`
}

type GreenstackTarget struct {
	Name          string `json:"name"`
	FarmMinutes   int    `json:"farmMinutes"`
	TravelRoutine string `json:"travelRoutine"`
	Enabled       bool   `json:"enabled"`
	Completed     bool   `json:"completed"`
}

type Config struct {
	WindowTitle       string                      `json:"windowTitle"`
	EmergencyStopKey  string                      `json:"emergencyStopKey"`
	CapturePointKey   string                      `json:"capturePointKey"`
	Calibration       map[string]Point            `json:"calibration"`
	Routines          map[string]Routine          `json:"routines"`
	SystemRoutines    map[string]string           `json:"systemRoutines,omitempty"`
	ClassSwitch       map[string]string           `json:"classSwitch,omitempty"`
	Settings          AgentSettings               `json:"settings,omitempty"`
	GreenstackTargets []GreenstackTarget          `json:"greenstackTargets"`
}

func defaultConfig() Config {
	return Config{
		WindowTitle:      "Legends of IdleOn",
		EmergencyStopKey: "F12",
		CapturePointKey:  "F8",
		Calibration:      map[string]Point{},
		Routines:         map[string]Routine{},
		SystemRoutines:   map[string]string{},
		ClassSwitch:      map[string]string{},
		Settings: AgentSettings{AssessIntervalSec:60,MaxSnapshotAgeSec:600,DefaultCooldownMinutes:15,MaxActionsPerHour:12},
	}
}

func loadConfig(path string) (Config,error) {
	raw,err:=os.ReadFile(path);if err!=nil{return Config{},err};cfg:=defaultConfig();if err:=json.Unmarshal(raw,&cfg);err!=nil{return Config{},err}
	if cfg.Calibration==nil{cfg.Calibration=map[string]Point{}};if cfg.Routines==nil{cfg.Routines=map[string]Routine{}};if cfg.SystemRoutines==nil{cfg.SystemRoutines=map[string]string{}};if cfg.ClassSwitch==nil{cfg.ClassSwitch=map[string]string{}}
	if strings.TrimSpace(cfg.WindowTitle)==""{return Config{},errors.New("windowTitle is required")};if strings.TrimSpace(cfg.EmergencyStopKey)==""{cfg.EmergencyStopKey="F12"};if strings.TrimSpace(cfg.CapturePointKey)==""{cfg.CapturePointKey="F8"}
	if cfg.Settings.AssessIntervalSec<5{cfg.Settings.AssessIntervalSec=60};if cfg.Settings.MaxSnapshotAgeSec<30{cfg.Settings.MaxSnapshotAgeSec=600};if cfg.Settings.DefaultCooldownMinutes<1{cfg.Settings.DefaultCooldownMinutes=15};if cfg.Settings.MaxActionsPerHour<1{cfg.Settings.MaxActionsPerHour=12};return cfg,nil
}

func saveConfig(path string,cfg Config)error{raw,err:=json.MarshalIndent(cfg,"","  ");if err!=nil{return err};return os.WriteFile(path,append(raw,'\n'),0o644)}

func (c Config) resolvePoint(step Step,game GameState)(Point,error){
	if step.Point!=""{p,ok:=c.Calibration[step.Point];if !ok{return Point{},fmt.Errorf("missing calibration point %q",step.Point)};if p.NX>0||p.NY>0{if game.Width<=0||game.Height<=0{return Point{},errors.New("game window rectangle unavailable for relative calibration")};return Point{X:game.X+int(p.NX*float64(game.Width)),Y:game.Y+int(p.NY*float64(game.Height)),NX:p.NX,NY:p.NY},nil};return p,nil}
	if step.X==0&&step.Y==0{return Point{},errors.New("click step needs point or x/y")};return Point{X:step.X,Y:step.Y},nil
}

func (c Config) requiredCalibrationPoints()[]string{seen:=map[string]bool{};var out []string;for _,routine:=range c.Routines{for _,step:=range routine.Steps{if step.Point!=""&&!seen[step.Point]{seen[step.Point]=true;out=append(out,step.Point)}}};return out}

func (c Config) routineReady(name string) bool {
	routine,ok:=c.Routines[name];if !ok||len(routine.Steps)==0{return false}
	for _,step:=range routine.Steps{if step.Point!=""{if _,ok:=c.Calibration[step.Point];!ok{return false}}}
	return true
}

func (c Config) routineForSystem(system,fallback string)string{if v:=strings.TrimSpace(c.SystemRoutines[strings.ToLower(system)]);v!=""{if _,ok:=c.Routines[v];ok{return v}};if fallback!=""{if _,ok:=c.Routines[fallback];ok{return fallback}};return ""}
func (c Config) switchRoutineForClass(class string)string{wanted:=normalizeClass(class);for key,routine:=range c.ClassSwitch{if normalizeClass(key)==wanted{return strings.TrimSpace(routine)}};return ""}
func (c *Config) markTargetCompleted(name string,completed bool)bool{wanted:=strings.TrimSpace(name);for i:=range c.GreenstackTargets{if strings.EqualFold(strings.TrimSpace(c.GreenstackTargets[i].Name),wanted){c.GreenstackTargets[i].Completed=completed;return true}};return false}

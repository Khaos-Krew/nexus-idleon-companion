package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Runner struct {
	cfg     Config
	input   InputDriver
	stopKey string
	detect  func(string) GameState
}

func NewRunner(cfg Config, input InputDriver) *Runner { return &Runner{cfg:cfg,input:input,stopKey:strings.ToUpper(cfg.EmergencyStopKey),detect:DetectGame} }
func (r *Runner) stopped() bool { return r.input.KeyDown(r.stopKey) }

func (r *Runner) sleep(ms int) error {
	if ms<=0{return nil};deadline:=time.Now().Add(time.Duration(ms)*time.Millisecond)
	for time.Now().Before(deadline){if r.stopped(){return errors.New("emergency stop requested")};remaining:=time.Until(deadline);if remaining>100*time.Millisecond{remaining=100*time.Millisecond};time.Sleep(remaining)};return nil
}

func parseRGB(value string) (RGB,error){v:=strings.TrimSpace(strings.TrimPrefix(value,"#"));if len(v)!=6{return RGB{},fmt.Errorf("RGB guard must be six hex digits")};n,err:=strconv.ParseUint(v,16,32);if err!=nil{return RGB{},err};return RGB{R:int((n>>16)&0xff),G:int((n>>8)&0xff),B:int(n&0xff)},nil}
func colorDistance(a,b RGB)float64{dr,dg,db:=float64(a.R-b.R),float64(a.G-b.G),float64(a.B-b.B);return math.Sqrt(dr*dr+dg*dg+db*db)}

func (r *Runner) guardPixel(step Step,game GameState)error{if strings.TrimSpace(step.ExpectedRGB)==""{return nil};p,err:=r.cfg.resolvePoint(step,game);if err!=nil{return err};actual,err:=r.input.PixelColor(p.X,p.Y);if err!=nil{return err};expected,err:=parseRGB(step.ExpectedRGB);if err!=nil{return err};tolerance:=step.Tolerance;if tolerance<=0{tolerance=30};if colorDistance(actual,expected)>float64(tolerance){return fmt.Errorf("UI guard failed at (%d,%d): expected #%02X%02X%02X ±%d, got #%02X%02X%02X",p.X,p.Y,expected.R,expected.G,expected.B,tolerance,actual.R,actual.G,actual.B)};return nil}

func (r *Runner) RunRoutine(name string) error {
	routine,ok:=r.cfg.Routines[name];if !ok{return fmt.Errorf("unknown routine %q",name)}
	game:=r.detect(r.cfg.WindowTitle);if !game.Running{return errors.New("IdleOn is not running")};if routine.RequiresForeground&&!game.Foreground{return errors.New("routine requires IdleOn to be foreground")};if len(routine.Steps)>500{return errors.New("routine exceeds 500-step safety limit")}
	fmt.Printf("Running routine: %s\n",name)
	for i,step:=range routine.Steps{if r.stopped(){return errors.New("emergency stop requested")};game=r.detect(r.cfg.WindowTitle);if !game.Running{return errors.New("IdleOn closed during routine")};repeats:=step.Repeat;if repeats<1{repeats=1};if repeats>25{return fmt.Errorf("step %d repeat exceeds 25",i+1)};for rep:=0;rep<repeats;rep++{if r.stopped(){return errors.New("emergency stop requested")};typeName:=strings.ToLower(strings.TrimSpace(step.Type));switch typeName{case"focus":if err:=r.input.FocusWindow(r.cfg.WindowTitle);err!=nil{return fmt.Errorf("step %d focus: %w",i+1,err)};case"assert-pixel","guard":if err:=r.guardPixel(step,game);err!=nil{return fmt.Errorf("step %d guard: %w",i+1,err)};case"click","double-click":if err:=r.guardPixel(step,game);err!=nil{return fmt.Errorf("step %d pre-click guard: %w",i+1,err)};point,err:=r.cfg.resolvePoint(step,game);if err!=nil{return fmt.Errorf("step %d click: %w",i+1,err)};if err:=r.input.Click(point.X,point.Y,typeName=="double-click");err!=nil{return fmt.Errorf("step %d click: %w",i+1,err)};case"key":if err:=r.input.PressKey(step.Key);err!=nil{return fmt.Errorf("step %d key: %w",i+1,err)};case"wait":if err:=r.sleep(step.MS);err!=nil{return err};default:return fmt.Errorf("step %d has unsupported type %q",i+1,step.Type)};if step.DelayAfterMS>0{if err:=r.sleep(step.DelayAfterMS);err!=nil{return err}}}};return nil
}

func (r *Runner) RunGreenstackLoop(cycles int) error {targets:=make([]GreenstackTarget,0,len(r.cfg.GreenstackTargets));for _,target:=range r.cfg.GreenstackTargets{if target.Enabled&&!target.Completed{targets=append(targets,target)}};if len(targets)==0{return errors.New("no enabled, unfinished greenstackTargets in config")};if cycles<1{cycles=1};for cycle:=1;cycle<=cycles;cycle++{fmt.Printf("Green-stack cycle %d/%d\n",cycle,cycles);for _,target:=range targets{if r.stopped(){return errors.New("emergency stop requested")};fmt.Printf("Target: %s\n",target.Name);if target.TravelRoutine!=""{if err:=r.RunRoutine(target.TravelRoutine);err!=nil{return fmt.Errorf("travel to %s: %w",target.Name,err)}};minutes:=target.FarmMinutes;if minutes<1{minutes=1};fmt.Printf("Farming %s for %d minute(s). Press %s to stop.\n",target.Name,minutes,r.stopKey);if err:=r.sleep(minutes*60*1000);err!=nil{return err}}};return nil}

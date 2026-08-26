package main

import "testing"

type fakeInput struct {
	focused []string
	clicks  []Point
	keys    []string
	down    map[string]bool
	pixel   RGB
	rect    Rect
}
func (f *fakeInput) FocusWindow(title string)error{f.focused=append(f.focused,title);return nil}
func (f *fakeInput) Click(x,y int,double bool)error{f.clicks=append(f.clicks,Point{X:x,Y:y});return nil}
func (f *fakeInput) PressKey(key string)error{f.keys=append(f.keys,key);return nil}
func (f *fakeInput) KeyDown(key string)bool{return f.down!=nil&&f.down[key]}
func (f *fakeInput) CursorPosition()(Point,error){return Point{X:10,Y:20},nil}
func (f *fakeInput) WindowRect(string)(Rect,error){if f.rect.Width==0{return Rect{X:100,Y:100,Width:1000,Height:800},nil};return f.rect,nil}
func (f *fakeInput) PixelColor(int,int)(RGB,error){return f.pixel,nil}

func testGame()GameState{return GameState{Running:true,Foreground:true,X:100,Y:100,Width:1000,Height:800}}

func TestResolveCalibratedPoint(t *testing.T){cfg:=defaultConfig();cfg.Calibration["map"]=Point{X:123,Y:456};got,err:=cfg.resolvePoint(Step{Type:"click",Point:"map"},testGame());if err!=nil{t.Fatal(err)};if got.X!=123||got.Y!=456{t.Fatalf("unexpected point: %+v",got)}}
func TestResolveRelativePoint(t *testing.T){cfg:=defaultConfig();cfg.Calibration["map"]=Point{NX:.5,NY:.25};got,err:=cfg.resolvePoint(Step{Point:"map"},testGame());if err!=nil{t.Fatal(err)};if got.X!=600||got.Y!=300{t.Fatalf("unexpected relative point: %+v",got)}}
func TestMissingCalibrationFailsClosed(t *testing.T){cfg:=defaultConfig();_,err:=cfg.resolvePoint(Step{Type:"click",Point:"missing"},testGame());if err==nil{t.Fatal("expected missing calibration to fail")}}

func TestRunRoutine(t *testing.T){cfg:=defaultConfig();cfg.Calibration["target"]=Point{X:50,Y:75};cfg.Routines["test"]=Routine{Steps:[]Step{{Type:"focus"},{Type:"click",Point:"target"},{Type:"key",Key:"F"}}};input:=&fakeInput{};runner:=NewRunner(cfg,input);runner.detect=func(string)GameState{return testGame()};if err:=runner.RunRoutine("test");err!=nil{t.Fatal(err)};if len(input.focused)!=1{t.Fatalf("expected focus")};if len(input.clicks)!=1||input.clicks[0].X!=50{t.Fatalf("unexpected clicks: %+v",input.clicks)};if len(input.keys)!=1||input.keys[0]!="F"{t.Fatalf("unexpected keys: %+v",input.keys)}}

func TestPixelGuardFailsClosed(t *testing.T){cfg:=defaultConfig();cfg.Calibration["target"]=Point{X:50,Y:75};cfg.Routines["test"]=Routine{Steps:[]Step{{Type:"click",Point:"target",ExpectedRGB:"#FF0000",Tolerance:5}}};input:=&fakeInput{pixel:RGB{R:0,G:0,B:0}};runner:=NewRunner(cfg,input);runner.detect=func(string)GameState{return testGame()};if err:=runner.RunRoutine("test");err==nil{t.Fatal("expected pixel guard failure")};if len(input.clicks)!=0{t.Fatal("guarded click should not execute")}}

func TestEmergencyStopPreventsRoutine(t *testing.T){cfg:=defaultConfig();cfg.Routines["test"]=Routine{Steps:[]Step{{Type:"focus"}}};input:=&fakeInput{down:map[string]bool{"F12":true}};runner:=NewRunner(cfg,input);runner.detect=func(string)GameState{return testGame()};if err:=runner.RunRoutine("test");err==nil{t.Fatal("expected emergency stop")}}

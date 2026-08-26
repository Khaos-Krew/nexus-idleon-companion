package main

import (
	"testing"
	"time"
)

func TestAssessmentRanksWeakAccountSystems(t *testing.T) {
	snap := AccountSnapshot{
		Source: "test", World: 6, CapturedAt: time.Now(),
		Systems: map[string]SystemState{
			"alchemy": {Progress: .20, Ready: true},
			"stamps": {Progress: .80, Ready: true},
			"sailing": {Progress: .60, Ready: true},
			"talents": {Progress: .95, Ready: true},
		},
	}
	a := buildAssessment(GameState{Running:true,Foreground:true}, snap)
	if len(a.Top)==0 { t.Fatal("expected recommendations") }
	if a.Top[0].System != "alchemy" { t.Fatalf("expected alchemy first, got %s", a.Top[0].System) }
}

func TestAssessmentAddsReadyTimersAndGreenStacks(t *testing.T) {
	snap := AccountSnapshot{Source:"test",World:6,Systems:map[string]SystemState{},Timers:[]TimerState{{Name:"Worship",Ready:true,Routine:"claim-worship"}},GreenStacks:[]GreenStackState{{Name:"Copper Ore",Count:5_000_000,Routine:"farm-copper"},{Name:"Oak Logs",Count:10_000_000}}}
	a:=buildAssessment(GameState{Running:true},snap)
	if len(a.TimersReady)!=1 { t.Fatalf("expected one timer recommendation") }
	if len(a.GreenStacks)!=1 { t.Fatalf("expected one unfinished green stack") }
	if !a.GreenStacks[0].Automatable { t.Fatal("expected routine-bound green stack to be automatable") }
}

func TestNormalizeGenericSnapshotAcceptsEfficiencyStyleJSON(t *testing.T) {
	generic:=map[string]any{"accountName":"Kirito","world":float64(6),"alchemy":map[string]any{"progress":float64(.25)},"characters":[]any{map[string]any{"name":"Bubo","class":"Bubo","level":float64(500),"active":true}}}
	s:=normalizeGenericSnapshot(generic,"idleon-efficiency-import")
	if s.AccountName!="Kirito" || s.World!=6 { t.Fatalf("unexpected normalized snapshot: %+v",s) }
	if _,ok:=s.Systems["alchemy"];!ok { t.Fatal("expected alchemy system") }
	if len(s.Characters)!=1 || !s.Characters[0].Active { t.Fatal("expected active character") }
}

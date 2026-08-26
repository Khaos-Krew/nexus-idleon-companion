package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func clamp01(v float64) float64 { if v < 0 { return 0 }; if v > 1 { return 1 }; return v }

func activeCharacter(snapshot AccountSnapshot) *CharacterSnapshot {
	for i := range snapshot.Characters { if snapshot.Characters[i].Active { return &snapshot.Characters[i] } }
	return nil
}

func scoreSystem(id string, state SystemState, world int) (Recommendation, bool) {
	rule, ok := idleonSystems[id]
	if !ok || world > 0 && world < rule.MinWorld { return Recommendation{}, false }
	progress := clamp01(state.Progress)
	deficiency := 1 - progress
	accountWide := rule.AccountWide; if state.AccountWide > 0 { accountWide = state.AccountWide }
	unlock := rule.UnlockValue; if state.UnlockValue > 0 { unlock = state.UnlockValue }
	ease := rule.Ease; if state.Ease > 0 { ease = state.Ease }
	hours := rule.Hours; if state.Hours > 0 { hours = state.Hours }
	urgency := state.Urgency; if urgency <= 0 { urgency = 1 }
	readiness := 1.0; if !state.Ready && state.Progress > 0 { readiness = .45 }
	score := accountWide * unlock * (.15 + deficiency) * ease * urgency * readiness / math.Sqrt(math.Max(hours, .25))
	preferred := state.PreferredClass; if preferred == "" { preferred = rule.PreferredClass }
	routine := state.Routine
	return Recommendation{
		ID: "system:" + id, Category: "system", Title: rule.Label, System: id, Score: math.Round(score*100)/100,
		Reason: fmt.Sprintf("%s is roughly %.0f%% complete against the supplied target state.", rule.Label, progress*100),
		PreferredClass: preferred, Routine: routine, Action: rule.Action, Automatable: routine != "",
	}, true
}

func buildAssessment(game GameState, snapshot AccountSnapshot) Assessment {
	world := snapshot.World; if world < 1 { world = 1 }
	var recs []Recommendation
	progressTotal := 0.0; progressCount := 0
	for id, state := range snapshot.Systems {
		if rec, ok := scoreSystem(strings.ToLower(id), state, world); ok {
			recs = append(recs, rec); progressTotal += clamp01(state.Progress); progressCount++
		}
	}

	var timers []Recommendation
	for _, timer := range snapshot.Timers {
		if !timer.Ready { continue }
		score := timer.Priority; if score <= 0 { score = 40 }
		r := Recommendation{ID: "timer:"+timer.Name, Category: "timer", Title: timer.Name+" ready", Score: score, Reason: "A capped or ready timer is immediate lost progression if ignored.", Routine: timer.Routine, Automatable: timer.Routine != "", Action: "Claim/refresh this timer now."}
		timers = append(timers, r); recs = append(recs, r)
	}

	var greens []Recommendation
	for _, gs := range snapshot.GreenStacks {
		if gs.Complete || gs.Count >= 10000000 { continue }
		remaining := 1 - clamp01(gs.Count/10000000)
		score := 55 * (.25 + remaining)
		preferred := gs.BestClass; if preferred == "" { preferred = "ES" }
		r := Recommendation{ID: "greenstack:"+gs.Name, Category: "greenstack", Title: "Green stack: "+gs.Name, Score: score, Reason: fmt.Sprintf("%.2fM / 10M stored; once completed it stays permanently credited.", gs.Count/1000000), PreferredClass: preferred, Routine: gs.Routine, Automatable: gs.Routine != "", Action: "Farm until the permanent 10M green-stack threshold is reached."}
		greens = append(greens, r); recs = append(recs, r)
	}

	sort.Slice(recs, func(i,j int) bool { return recs[i].Score > recs[j].Score })
	sort.Slice(timers, func(i,j int) bool { return timers[i].Score > timers[j].Score })
	sort.Slice(greens, func(i,j int) bool { return greens[i].Score > greens[j].Score })
	quick := make([]Recommendation,0)
	for _, r := range recs { if r.Score >= 25 && (r.Category == "timer" || strings.Contains(strings.ToLower(r.Action), "claim") || strings.Contains(strings.ToLower(r.Action), "spend")) { quick = append(quick,r); if len(quick)>=8 { break } } }
	if len(recs) > 12 { recs = recs[:12] }
	if len(greens) > 8 { greens = greens[:8] }
	health := 0; if progressCount > 0 { health = int(math.Round(progressTotal/float64(progressCount)*100)) }
	return Assessment{GeneratedAt: time.Now(), Game: game, Source: snapshot.Source, AccountName: snapshot.AccountName, World: snapshot.World, HealthScore: health, ActiveCharacter: activeCharacter(snapshot), Top: recs, QuickWins: quick, TimersReady: timers, GreenStacks: greens}
}

func chooseAutomation(assessment Assessment) *Recommendation {
	for i := range assessment.Top {
		r := &assessment.Top[i]
		if r.Automatable && r.Routine != "" { return r }
	}
	return nil
}

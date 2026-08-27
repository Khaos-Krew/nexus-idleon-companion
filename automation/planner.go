package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func clamp01(v float64) float64 { if v<0{return 0};if v>1{return 1};return v }
func activeCharacter(snapshot AccountSnapshot)*CharacterSnapshot{for i:=range snapshot.Characters{if snapshot.Characters[i].Active{return &snapshot.Characters[i]}};return nil}
func stageForWorld(world int) string { switch{case world>=7:return "World 7 / current endgame";case world==6:return "World 6";case world==5:return "World 5";case world==4:return "World 4";case world==3:return "World 3";case world==2:return "World 2";default:return "World 1"} }
func unlockedRuleCount(world int)int{n:=0;for _,r:=range idleonSystems{if world>=r.MinWorld{n++}};return n}

func dependencyPenalty(rule SystemRule,snapshot AccountSnapshot)(float64,[]string){penalty:=1.0;var blockers []string;for _,dep:=range rule.Dependencies{state,ok:=snapshot.Systems[dep];if !ok||state.DetectedOnly{continue};if state.Progress<.35{penalty*=.72;blockers=append(blockers,dep)}else if state.Progress<.55{penalty*=.9;blockers=append(blockers,dep)}};return penalty,blockers}

func scoreSystem(id string,state SystemState,world int,snapshot AccountSnapshot,cfg Config)(Recommendation,bool){
	rule,ok:=idleonSystems[id];if !ok||(world>0&&world<rule.MinWorld){return Recommendation{},false}
	progress:=clamp01(state.Progress);if state.Target>0{progress=clamp01(state.Current/state.Target)};deficiency:=1-progress
	accountWide:=rule.AccountWide;if state.AccountWide>0{accountWide=state.AccountWide};unlock:=rule.UnlockValue;if state.UnlockValue>0{unlock=state.UnlockValue};ease:=rule.Ease;if state.Ease>0{ease=state.Ease};hours:=rule.Hours;if state.Hours>0{hours=state.Hours};urgency:=state.Urgency;if urgency<=0{urgency=1};readiness:=1.0;if !state.Ready&&state.Progress>0{readiness=.55}
	depPenalty,blockers:=dependencyPenalty(rule,snapshot);score:=accountWide*unlock*(.15+deficiency)*ease*urgency*readiness*depPenalty/math.Sqrt(math.Max(hours,.25))
	preferred:=state.PreferredClass;if preferred==""{preferred=rule.PreferredClass};routine:=state.Routine;if routine==""{routine=cfg.routineForSystem(id,rule.Routine)}
	best:=findBestCharacter(snapshot,preferred);character:="";switchRoutine:="";if best!=nil{character=best.Name;if preferred!=""&&!best.Active{candidate:=cfg.switchRoutineForClass(preferred);if cfg.routineReady(candidate){switchRoutine=candidate}}}
	confidence:=state.Confidence
	if confidence<=0 { confidence=.65;if state.Evidence!=""||state.Current>0||state.Target>0{confidence=.9}else if state.Progress>0{confidence=.8} }
	if state.DetectedOnly { confidence=math.Min(confidence,.45); score*=.60 }
	if len(blockers)>0{confidence*=.9}
	reason:=fmt.Sprintf("%s is roughly %.0f%% complete against the supplied target state.",rule.Label,progress*100)
	if state.DetectedOnly { reason="This system is present in the raw account data, but this build does not yet claim an exact completion percentage for it." }
	if state.Evidence!=""{reason+=" Evidence: "+state.Evidence}
	if len(blockers)>0{reason+=fmt.Sprintf(" Blockers currently include %s.",strings.Join(blockers,", "))}
	automatable:=cfg.routineReady(routine) && !state.DetectedOnly && confidence>=.80
	return Recommendation{ID:"system:"+id,Category:"system",Title:rule.Label,System:id,Score:math.Round(score*100)/100,Confidence:math.Round(confidence*100)/100,Reason:reason,PreferredClass:preferred,Character:character,Routine:routine,SwitchRoutine:switchRoutine,Action:rule.Action,Automatable:automatable,EstimatedMinutes:int(math.Max(1,hours*60)),Dependencies:rule.Dependencies},true
}

func buildCharacterPlan(snapshot AccountSnapshot,recs []Recommendation)[]CharacterPlan{plans:=map[string]CharacterPlan{};for _,rec:=range recs{if rec.Character==""{continue};class:="";for _,c:=range snapshot.Characters{if c.Name==rec.Character{class=c.Class;break}};current,ok:=plans[rec.Character];priority:=int(math.Round(rec.Score));if !ok||priority>current.Priority{plans[rec.Character]=CharacterPlan{Character:rec.Character,Class:class,Role:rec.Title,Reason:rec.Action,Priority:priority}}};for _,c:=range snapshot.Characters{if _,ok:=plans[c.Name];!ok{plans[c.Name]=CharacterPlan{Character:c.Name,Class:c.Class,Role:classRole(c),Reason:"Maintain this character in its strongest account-support role until a higher-priority objective needs it.",Priority:10}}};out:=make([]CharacterPlan,0,len(plans));for _,p:=range plans{out=append(out,p)};sort.Slice(out,func(i,j int)bool{return out[i].Priority>out[j].Priority});return out}

func buildSystemCoverage(snapshot AccountSnapshot, world int) []SystemCoverage {
	ids:=make([]string,0,len(idleonSystems));for id:=range idleonSystems{ids=append(ids,id)};sort.Strings(ids)
	out:=make([]SystemCoverage,0,len(ids))
	for _,id:=range ids{
		rule:=idleonSystems[id]
		item:=SystemCoverage{ID:id,Label:rule.Label,World:rule.MinWorld,Status:"missing",Confidence:0}
		if world<rule.MinWorld{item.Status="not-unlocked";out=append(out,item);continue}
		if state,ok:=snapshot.Systems[id];ok{
			item.Evidence=state.Evidence
			item.Confidence=int(math.Round(state.Confidence*100))
			if item.Confidence==0{item.Confidence=65}
			if state.DetectedOnly{item.Status="detected";if item.Confidence>45{item.Confidence=45}}else{item.Status="parsed"}
		}
		out=append(out,item)
	}
	return out
}

func buildAssessment(game GameState,snapshot AccountSnapshot)Assessment{return buildAssessmentWithConfig(game,snapshot,defaultConfig())}

func buildAssessmentWithConfig(game GameState,snapshot AccountSnapshot,cfg Config)Assessment{
	world:=snapshot.World;if world<1{world=1};var recs,timers,greens []Recommendation;progressTotal:=0.0;progressCount:=0;parsedCount:=0
	for id,state:=range snapshot.Systems{if rec,ok:=scoreSystem(strings.ToLower(id),state,world,snapshot,cfg);ok{recs=append(recs,rec);if !state.DetectedOnly{progressTotal+=clamp01(state.Progress);progressCount++;parsedCount++}}}
	for _,timer:=range snapshot.Timers{if !timer.Ready{continue};score:=timer.Priority;if score<=0{score=40};ready:=cfg.routineReady(timer.Routine);r:=Recommendation{ID:"timer:"+timer.Name,Category:"timer",Title:timer.Name+" ready",Score:score,Confidence:.95,Reason:"A capped or ready timer can waste future progression if ignored.",Routine:timer.Routine,Automatable:ready,Action:"Claim or refresh this timer now.",EstimatedMinutes:2};timers=append(timers,r);recs=append(recs,r)}
	for _,gs:=range snapshot.GreenStacks{if gs.Complete||gs.Count>=10000000{continue};remaining:=1-clamp01(gs.Count/10000000);score:=55*(.25+remaining);preferred:=gs.BestClass;if preferred==""{preferred="ES"};best:=findBestCharacter(snapshot,preferred);character:="";if best!=nil{character=best.Name};r:=Recommendation{ID:"greenstack:"+gs.Name,Category:"greenstack",Title:"Green stack: "+gs.Name,Score:score,Confidence:.9,Reason:fmt.Sprintf("%.2fM / 10M stored; completion is permanent once credited.",gs.Count/1000000),PreferredClass:preferred,Character:character,Routine:gs.Routine,Automatable:cfg.routineReady(gs.Routine),Action:"Farm until the permanent 10M green-stack threshold is reached."};greens=append(greens,r);recs=append(recs,r)}
	sort.Slice(recs,func(i,j int)bool{return recs[i].Score>recs[j].Score});sort.Slice(timers,func(i,j int)bool{return timers[i].Score>timers[j].Score});sort.Slice(greens,func(i,j int)bool{return greens[i].Score>greens[j].Score})
	quick:=[]Recommendation{};for _,r:=range recs{if r.Confidence>=.65&&r.Score>=25&&(r.Category=="timer"||r.EstimatedMinutes<=15){quick=append(quick,r);if len(quick)>=8{break}}};allRecs:=append([]Recommendation(nil),recs...);if len(recs)>15{recs=recs[:15]};if len(greens)>10{greens=greens[:10]}
	health:=0;if progressCount>0{health=int(math.Round(progressTotal/float64(progressCount)*100))};unlocked:=unlockedRuleCount(world);coverage:=0;if unlocked>0{coverage=int(math.Round(float64(len(snapshot.Systems))/float64(unlocked)*100));if coverage>100{coverage=100}}
	age:=int64(0);warnings:=append([]string{},snapshot.Warnings...);if !snapshot.CapturedAt.IsZero(){age=int64(time.Since(snapshot.CapturedAt).Seconds());if age<0{age=0};if age>int64(cfg.Settings.MaxSnapshotAgeSec){warnings=append(warnings,fmt.Sprintf("Account snapshot is %d seconds old; refresh it before relying on time-sensitive advice.",age))}};if coverage<35{warnings=append(warnings,"Account-data coverage is limited; recommendations are lower confidence until more systems are recognized.")};if parsedCount==0&&len(snapshot.Systems)>0{warnings=append(warnings,"Systems were detected in raw data, but none currently have high-confidence derived progression values. Use the coverage report while parser depth is expanded.")}
	diag:=SourceDiagnostic{Source:snapshot.Source,Schema:snapshot.Schema,Loaded:true,CapturedAt:snapshot.CapturedAt,Characters:len(snapshot.Characters),Systems:len(snapshot.Systems),RawKeys:len(snapshot.Raw),Warnings:snapshot.Warnings};diag.Fresh=age<=int64(cfg.Settings.MaxSnapshotAgeSec)||age==0
	return Assessment{GeneratedAt:time.Now(),Game:game,Source:snapshot.Source,Schema:snapshot.Schema,AccountName:snapshot.AccountName,World:snapshot.World,Stage:stageForWorld(world),HealthScore:health,CoveragePercent:coverage,SnapshotAgeSec:age,ActiveCharacter:activeCharacter(snapshot),Top:recs,QuickWins:quick,TimersReady:timers,GreenStacks:greens,CharacterPlan:buildCharacterPlan(snapshot,allRecs),SystemCoverage:buildSystemCoverage(snapshot,world),SourceDiagnostics:[]SourceDiagnostic{diag},Warnings:warnings}
}

func chooseAutomation(assessment Assessment)*Recommendation{if !assessment.Game.Running||assessment.SnapshotAgeSec>600{return nil};for i:=range assessment.Top{r:=&assessment.Top[i];if r.Automatable&&r.Routine!=""&&r.Confidence>=.8{return r}};return nil}

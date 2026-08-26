package main

import (
	"errors"
	"strings"
	"time"
)

type FusionProvider struct { Providers []SnapshotProvider }
func (p FusionProvider) Name() string { return "fused-account-sources" }

func mergeSnapshots(base AccountSnapshot, incoming AccountSnapshot) AccountSnapshot {
	if base.Systems==nil{base.Systems=map[string]SystemState{}}
	if base.AccountName==""{base.AccountName=incoming.AccountName}
	if incoming.World>base.World{base.World=incoming.World}
	if len(base.Characters)==0&&len(incoming.Characters)>0{base.Characters=incoming.Characters}
	for id,state:=range incoming.Systems{
		old,ok:=base.Systems[id]
		if !ok || (old.Progress==0&&state.Progress>0) || (old.Evidence==""&&state.Evidence!="") { base.Systems[id]=state }
	}
	if len(base.Timers)==0&&len(incoming.Timers)>0{base.Timers=incoming.Timers}
	if len(base.GreenStacks)==0&&len(incoming.GreenStacks)>0{base.GreenStacks=incoming.GreenStacks}
	if base.Raw==nil&&incoming.Raw!=nil{base.Raw=incoming.Raw}
	if base.CapturedAt.IsZero()||(!incoming.CapturedAt.IsZero()&&incoming.CapturedAt.After(base.CapturedAt)){base.CapturedAt=incoming.CapturedAt}
	if base.Source==""{base.Source=incoming.Source}else if incoming.Source!=""&&!strings.Contains(base.Source,incoming.Source){base.Source+="+"+incoming.Source}
	return base
}

func (p FusionProvider) Load()(AccountSnapshot,error){
	var merged AccountSnapshot; var loaded int; var lastErr error
	for _,provider:=range p.Providers{if provider==nil{continue};snap,err:=provider.Load();if err!=nil{lastErr=err;continue};if loaded==0{merged=snap}else{merged=mergeSnapshots(merged,snap)};loaded++}
	if loaded==0{if lastErr!=nil{return AccountSnapshot{},lastErr};return AccountSnapshot{},errors.New("no account source could be loaded")}
	if merged.CapturedAt.IsZero(){merged.CapturedAt=time.Now()};return merged,nil
}

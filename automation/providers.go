package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const toolboxProfilesURL = "https://profiles.idleontoolbox.workers.dev/api/profiles/"

type SnapshotProvider interface { Name() string; Load() (AccountSnapshot,error) }

type FileSnapshotProvider struct { Path string; Source string }
func (p FileSnapshotProvider) Name()string{if p.Source!=""{return p.Source};return "json-file"}
func (p FileSnapshotProvider) Load()(AccountSnapshot,error){raw,err:=os.ReadFile(p.Path);if err!=nil{return AccountSnapshot{},err};snap,err:=decodeFlexibleSnapshot(raw,p.Name());if err==nil{if info,e:=os.Stat(p.Path);e==nil{snap.CapturedAt=info.ModTime()}};return snap,err}

type ToolboxProvider struct { Profile string; Client *http.Client }
func (p ToolboxProvider) Name()string{return "idleon-toolbox-public-profile"}
func (p ToolboxProvider) Load()(AccountSnapshot,error){profile:=strings.TrimSpace(p.Profile);if profile==""{return AccountSnapshot{},errors.New("toolbox profile is required")};client:=p.Client;if client==nil{client=&http.Client{Timeout:12*time.Second}};endpoint:=toolboxProfilesURL+"?profile="+url.QueryEscape(profile);resp,err:=client.Get(endpoint);if err!=nil{return AccountSnapshot{},err};defer resp.Body.Close();if resp.StatusCode!=http.StatusOK{return AccountSnapshot{},fmt.Errorf("toolbox profile returned HTTP %d",resp.StatusCode)};raw,err:=io.ReadAll(io.LimitReader(resp.Body,8<<20));if err!=nil{return AccountSnapshot{},err};return decodeFlexibleSnapshot(raw,p.Name())}

func decodeFlexibleSnapshot(raw []byte,source string)(AccountSnapshot,error){
	var direct AccountSnapshot
	if err:=json.Unmarshal(raw,&direct);err==nil&&(len(direct.Systems)>0||len(direct.Characters)>0||direct.World>0){direct.Source=source;if direct.CapturedAt.IsZero(){direct.CapturedAt=time.Now()};if direct.Systems==nil{direct.Systems=map[string]SystemState{}};return direct,nil}
	var generic map[string]any;if err:=json.Unmarshal(raw,&generic);err!=nil{return AccountSnapshot{},err};return normalizeGenericSnapshot(generic,source),nil
}

var systemAliases=map[string][]string{
	"worldpush":{"worldProgress","highestWorld","portalProgress"},"stamps":{"stamp","stamps"},"vault":{"vault"},"forge":{"forge"},"anvil":{"anvil","smithing"},"statues":{"statue","statues"},"cards":{"cards","cardSets"},"constellations":{"starSigns","constellations"},"talents":{"talents","talentBooks"},"gear":{"gear","equipment","tools"},"quests":{"quests"},"tasks":{"tasks","merits"},"dungeons":{"dungeons","partyDungeons"},
	"alchemy":{"alchemy","bubbles","vials"},"prisma":{"prisma","prismaBubbles"},"postoffice":{"postOffice","postoffice"},"obols":{"obols"},
	"refinery":{"refinery","salts"},"construction":{"construction"},"printer":{"printer","3dPrinter","samples"},"worship":{"worship","souls"},"trapping":{"trapping","traps"},"shrines":{"shrines"},"deathnote":{"deathNote","deathnote"},
	"cooking":{"cooking","meals"},"breeding":{"breeding","pets"},"lab":{"lab","laboratory"},"rift":{"rift"},"tome":{"tome"},"killroy":{"killroy"},
	"divinity":{"divinity","gods"},"sailing":{"sailing","artifacts"},"gaming":{"gaming"},"companions":{"companions"},"hole":{"hole","theHole"},"slab":{"slab"},
	"sneaking":{"sneaking"},"farming":{"farming","crops"},"summoning":{"summoning"},"beanstalk":{"beanstalk"},"masterclasses":{"masterClasses","masterclasses"},
	"research":{"research","observations","researchPoints"},"spelunking":{"spelunking","tunnels","amber"},"coralreef":{"coralReef","coralreef","fishies"},"sushistation":{"sushiStation","sushistation","sushi"},
	"greenstacks":{"greenStacks","greenstacks"},"bosses":{"bosses"},"minibosses":{"minibosses","miniBosses"},"colosseum":{"colosseum"},"weeklyboss":{"weeklyBoss"},"vman":{"voidwalker","vman"},"owl":{"owl"},"roo":{"roo"},
}

func normalizeGenericSnapshot(generic map[string]any,source string)AccountSnapshot{
	s:=AccountSnapshot{Source:source,Systems:map[string]SystemState{},Raw:generic,CapturedAt:time.Now()}
	if v,ok:=firstString(generic,"accountName","mainChar","profileName","name");ok{s.AccountName=v}
	if v,ok:=firstNumber(generic,"world","highestWorld","worldProgress");ok{s.World=int(v)}
	for id:=range idleonSystems{
		keys:=append([]string{id},systemAliases[id]...)
		for _,key:=range keys{if value,ok:=findNumericDeep(generic,key,6);ok{progress:=value;if progress>1&&progress<=100{progress/=100};if progress>1{progress=1};if progress<0{progress=0};s.Systems[id]=SystemState{Progress:progress,Ready:true,Evidence:"Imported from "+source+" field "+key};break}}
	}
	if chars:=findArrayDeep(generic,[]string{"characters","players","chars"},5);len(chars)>0{for idx,rawChar:=range chars{m,ok:=rawChar.(map[string]any);if !ok{continue};c:=CharacterSnapshot{};c.Name,_=firstString(m,"name","characterName","playerName");c.Class,_=firstString(m,"class","className","eliteClass");if lvl,ok:=firstNumber(m,"level","classLevel");ok{c.Level=int(lvl)};if active,ok:=m["active"].(bool);ok{c.Active=active};c.Map,_=firstString(m,"map","mapName","currentMap");if c.Name==""{c.Name=fmt.Sprintf("Character %d",idx+1)};s.Characters=append(s.Characters,c)}}
	s.GreenStacks=extractGreenStacks(generic)
	s.Timers=extractTimers(generic)
	return s
}

func extractGreenStacks(generic map[string]any)[]GreenStackState{arr:=findArrayDeep(generic,[]string{"greenStacks","greenstacks"},6);var out []GreenStackState;for _,v:=range arr{m,ok:=v.(map[string]any);if !ok{continue};g:=GreenStackState{};g.Name,_=firstString(m,"name","item","itemName");g.Count,_=firstNumber(m,"count","amount","quantity");if done,ok:=m["complete"].(bool);ok{g.Complete=done};if g.Count>=10000000{g.Complete=true};g.Map,_=firstString(m,"map","location");g.Routine,_=firstString(m,"routine");g.BestClass,_=firstString(m,"bestClass","class");if g.Name!=""{out=append(out,g)}};return out}
func extractTimers(generic map[string]any)[]TimerState{arr:=findArrayDeep(generic,[]string{"timers","alerts","readyTimers"},5);var out []TimerState;for _,v:=range arr{m,ok:=v.(map[string]any);if !ok{continue};t:=TimerState{};t.Name,_=firstString(m,"name","title","system");if ready,ok:=m["ready"].(bool);ok{t.Ready=ready};t.Priority,_=firstNumber(m,"priority","score");t.Routine,_=firstString(m,"routine");if t.Name!=""{out=append(out,t)}};return out}

func firstString(m map[string]any,keys ...string)(string,bool){for _,key:=range keys{for k,v:=range m{if !strings.EqualFold(k,key){continue};if t,ok:=v.(string);ok&&strings.TrimSpace(t)!=""{return t,true}}};return "",false}
func firstNumber(m map[string]any,keys ...string)(float64,bool){for _,key:=range keys{for k,v:=range m{if !strings.EqualFold(k,key){continue};switch t:=v.(type){case float64:return t,true;case json.Number:n,err:=t.Float64();if err==nil{return n,true}}}};return 0,false}
func findNumericDeep(value any,wanted string,depth int)(float64,bool){if depth<0{return 0,false};switch t:=value.(type){case map[string]any:for k,v:=range t{if strings.EqualFold(k,wanted){if n,ok:=v.(float64);ok{return n,true};if sub,ok:=v.(map[string]any);ok{if n,ok:=firstNumber(sub,"progress","score","percent","completion","level");ok{return n,true}}}};for _,v:=range t{if n,ok:=findNumericDeep(v,wanted,depth-1);ok{return n,true}};case []any:for _,v:=range t{if n,ok:=findNumericDeep(v,wanted,depth-1);ok{return n,true}}};return 0,false}
func findArrayDeep(value any,wanted []string,depth int)[]any{if depth<0{return nil};switch t:=value.(type){case map[string]any:for k,v:=range t{for _,w:=range wanted{if strings.EqualFold(k,w){if arr,ok:=v.([]any);ok{return arr}}}};for _,v:=range t{if arr:=findArrayDeep(v,wanted,depth-1);len(arr)>0{return arr}};case []any:for _,v:=range t{if arr:=findArrayDeep(v,wanted,depth-1);len(arr)>0{return arr}}};return nil}

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Dashboard struct {
	cfg      Config
	provider SnapshotProvider
	store    *AgentStore
	input    InputDriver
	mu       sync.Mutex
	last     Assessment
}

func NewDashboard(cfg Config, provider SnapshotProvider, store *AgentStore) *Dashboard { return &Dashboard{cfg:cfg,provider:provider,store:store,input:NewInputDriver()} }

func (d *Dashboard) assess() (Assessment,error) {
	snap,err:=d.provider.Load();if err!=nil{return Assessment{},err}
	a:=buildAssessmentWithConfig(DetectGame(d.cfg.WindowTitle),snap,d.cfg)
	d.mu.Lock();d.last=a;d.mu.Unlock();_ = d.store.SaveAssessment(a);return a,nil
}

func (d *Dashboard) executeTop() error {
	a,err:=d.assess();if err!=nil{return err}
	if !a.Game.Running{return fmt.Errorf("IdleOn is not running")}
	if !a.Game.Foreground{return fmt.Errorf("bring IdleOn to the foreground before execution")}
	if len(a.Warnings)>0&&a.SnapshotAgeSec>int64(d.cfg.Settings.MaxSnapshotAgeSec){return fmt.Errorf("snapshot is stale; capture fresh account data first")}
	allowed,reason:=d.store.CanExecute(d.cfg);if !allowed{return fmt.Errorf(reason)}
	rec:=chooseAutomation(a);if rec==nil{return fmt.Errorf("no safe calibrated recommendation is currently executable")}
	runner:=NewRunner(d.cfg,d.input)
	if rec.SwitchRoutine!=""{if err:=runner.RunRoutine(rec.SwitchRoutine);err!=nil{return err}}
	err=runner.RunRoutine(rec.Routine)
	record:=ActionRecord{Routine:rec.Routine,Objective:rec.Title,Character:rec.Character,Success:err==nil};if err!=nil{record.Error=err.Error()};_ = d.store.LogAction(record);return err
}

func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(v)}

const dashboardHTML=`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>IdleOn Account Agent</title><style>body{font-family:system-ui;margin:0;background:#101214;color:#eee}main{max-width:1100px;margin:auto;padding:24px}.card{background:#191d21;border:1px solid #30363d;border-radius:12px;padding:18px;margin:12px 0}button{padding:10px 14px;margin:4px;border-radius:8px;border:0;cursor:pointer}table{width:100%;border-collapse:collapse}td,th{padding:8px;border-bottom:1px solid #30363d;text-align:left}.muted{color:#9da7b1}.good{color:#67d391}.warn{color:#f1c75b}</style></head><body><main><h1>IdleOn Account Agent</h1><div class="card"><button onclick="assess()">Assess Now</button><button onclick="executeTop()">Execute Top Safe Action</button><span id="msg" class="muted"></span></div><div id="view" class="card">Loading…</div></main><script>async function assess(){msg.textContent='';let r=await fetch('/api/assess',{method:'POST'});let j=await r.json();if(!r.ok){msg.textContent=j.error||'Assessment failed';return}render(j)}async function executeTop(){if(!confirm('Execute the highest-ranked calibrated action? F12 remains the emergency stop.'))return;let r=await fetch('/api/execute-top',{method:'POST'});let j=await r.json();msg.textContent=r.ok?'Action completed':(j.error||'Action failed');await assess()}function render(a){let rows=(a.top||[]).map((r,i)=>'<tr><td>'+(i+1)+'</td><td>'+esc(r.title)+'</td><td>'+Number(r.score).toFixed(1)+'</td><td>'+esc(r.character||r.preferredClass||'')+'</td><td>'+esc(r.action||'')+'</td><td>'+(r.automatable?'✓':'')+'</td></tr>').join('');let chars=(a.characterPlan||[]).map(c=>'<li><b>'+esc(c.character)+'</b> ('+esc(c.class||'')+'): '+esc(c.role)+'</li>').join('');view.innerHTML='<h2>'+esc(a.accountName||'Account')+'</h2><p>Stage: <b>'+esc(a.stage||'')+'</b> · Health: <b>'+a.healthScore+'%</b> · Data coverage: <b>'+a.coveragePercent+'%</b> · Game: <span class="'+(a.game.running?'good':'warn')+'">'+(a.game.running?'Running':'Not running')+'</span></p><h3>Top priorities</h3><table><thead><tr><th>#</th><th>Objective</th><th>Score</th><th>Character</th><th>Action</th><th>Auto</th></tr></thead><tbody>'+rows+'</tbody></table><h3>Character plan</h3><ul>'+chars+'</ul>'}function esc(s){return String(s??'').replace(/[&<>"']/g,m=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[m]))}assess();</script></body></html>`

func (d *Dashboard) Handler() http.Handler {
	mux:=http.NewServeMux()
	page:=template.Must(template.New("dashboard").Parse(dashboardHTML))
	mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html; charset=utf-8");_ = page.Execute(w,nil)})
	mux.HandleFunc("/api/status",func(w http.ResponseWriter,r *http.Request){d.mu.Lock();last:=d.last;d.mu.Unlock();writeJSON(w,200,map[string]any{"game":DetectGame(d.cfg.WindowTitle),"last":last})})
	mux.HandleFunc("/api/assess",func(w http.ResponseWriter,r *http.Request){if r.Method!="POST"&&r.Method!="GET"{writeJSON(w,405,map[string]string{"error":"method not allowed"});return};a,err:=d.assess();if err!=nil{writeJSON(w,500,map[string]string{"error":err.Error()});return};writeJSON(w,200,a)})
	mux.HandleFunc("/api/execute-top",func(w http.ResponseWriter,r *http.Request){if r.Method!="POST"{writeJSON(w,405,map[string]string{"error":"method not allowed"});return};if err:=d.executeTop();err!=nil{writeJSON(w,409,map[string]string{"error":err.Error()});return};writeJSON(w,200,map[string]bool{"ok":true})})
	mux.HandleFunc("/api/run-routine",func(w http.ResponseWriter,r *http.Request){if r.Method!="POST"{writeJSON(w,405,map[string]string{"error":"method not allowed"});return};name:=strings.TrimSpace(r.URL.Query().Get("name"));if name==""{writeJSON(w,400,map[string]string{"error":"name is required"});return};if _,ok:=d.cfg.Routines[name];!ok{writeJSON(w,404,map[string]string{"error":"unknown routine"});return};if err:=NewRunner(d.cfg,d.input).RunRoutine(name);err!=nil{writeJSON(w,409,map[string]string{"error":err.Error()});return};_ = d.store.LogAction(ActionRecord{Routine:name,Success:true});writeJSON(w,200,map[string]bool{"ok":true})})
	return mux
}

func (d *Dashboard) Serve(port int) error {
	if port<1024||port>65535{port=17654}
	addr:="127.0.0.1:"+strconv.Itoa(port)
	ln,err:=net.Listen("tcp",addr);if err!=nil{return err}
	fmt.Printf("IdleOn Account Agent dashboard: http://%s\n",addr)
	server:=&http.Server{Handler:d.Handler(),ReadHeaderTimeout:5*time.Second}
	return server.Serve(ln)
}

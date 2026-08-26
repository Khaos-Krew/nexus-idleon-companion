package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ActionRecord struct {
	At        time.Time `json:"at"`
	Routine   string    `json:"routine"`
	Objective string    `json:"objective,omitempty"`
	Character string    `json:"character,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

type AgentStore struct { Dir string; mu sync.Mutex }

func defaultAgentDataDir() string {
	if base:=strings.TrimSpace(os.Getenv("LOCALAPPDATA"));base!=""{return filepath.Join(base,"IdleOn Account Agent")}
	if home,err:=os.UserHomeDir();err==nil{return filepath.Join(home,".idleon-account-agent")}
	return "idleon-agent-data"
}

func NewAgentStore(dir string)*AgentStore{if strings.TrimSpace(dir)==""{dir=defaultAgentDataDir()};return &AgentStore{Dir:dir}}
func (s *AgentStore) ensure()error{return os.MkdirAll(s.Dir,0o755)}

func (s *AgentStore) SaveAssessment(a Assessment) error {
	s.mu.Lock();defer s.mu.Unlock();if err:=s.ensure();err!=nil{return err}
	raw,err:=json.MarshalIndent(a,"","  ");if err!=nil{return err}
	if err:=os.WriteFile(filepath.Join(s.Dir,"last-assessment.json"),append(raw,'\n'),0o644);err!=nil{return err}
	compact,_:=json.Marshal(a);f,err:=os.OpenFile(filepath.Join(s.Dir,"assessment-history.jsonl"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0o644);if err!=nil{return err};defer f.Close();_,err=f.Write(append(compact,'\n'));return err
}

func (s *AgentStore) LoadAssessment()(Assessment,error){raw,err:=os.ReadFile(filepath.Join(s.Dir,"last-assessment.json"));if err!=nil{return Assessment{},err};var a Assessment;err=json.Unmarshal(raw,&a);return a,err}

func (s *AgentStore) LogAction(record ActionRecord) error {
	s.mu.Lock();defer s.mu.Unlock();if err:=s.ensure();err!=nil{return err};if record.At.IsZero(){record.At=time.Now()};raw,_:=json.Marshal(record);f,err:=os.OpenFile(filepath.Join(s.Dir,"actions.jsonl"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0o644);if err!=nil{return err};defer f.Close();_,err=f.Write(append(raw,'\n'));return err
}

func (s *AgentStore) RecentActions(since time.Time)([]ActionRecord,error){raw,err:=os.ReadFile(filepath.Join(s.Dir,"actions.jsonl"));if errors.Is(err,os.ErrNotExist){return nil,nil};if err!=nil{return nil,err};var out []ActionRecord;for _,line:=range strings.Split(string(raw),"\n"){if strings.TrimSpace(line)==""{continue};var r ActionRecord;if json.Unmarshal([]byte(line),&r)==nil&&r.At.After(since){out=append(out,r)}};return out,nil}

func (s *AgentStore) CanExecute(cfg Config)(bool,string){actions,err:=s.RecentActions(time.Now().Add(-time.Hour));if err!=nil{return false,"could not read action history"};if len(actions)>=cfg.Settings.MaxActionsPerHour{return false,"hourly action budget reached"};return true,""}

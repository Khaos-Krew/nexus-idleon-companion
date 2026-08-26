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

type SnapshotProvider interface {
	Name() string
	Load() (AccountSnapshot, error)
}

type FileSnapshotProvider struct {
	Path   string
	Source string
}

func (p FileSnapshotProvider) Name() string {
	if p.Source != "" { return p.Source }
	return "json-file"
}

func (p FileSnapshotProvider) Load() (AccountSnapshot, error) {
	raw, err := os.ReadFile(p.Path)
	if err != nil { return AccountSnapshot{}, err }
	return decodeFlexibleSnapshot(raw, p.Name())
}

type ToolboxProvider struct {
	Profile string
	Client  *http.Client
}

func (p ToolboxProvider) Name() string { return "idleon-toolbox-public-profile" }

func (p ToolboxProvider) Load() (AccountSnapshot, error) {
	profile := strings.TrimSpace(p.Profile)
	if profile == "" { return AccountSnapshot{}, errors.New("toolbox profile is required") }
	client := p.Client
	if client == nil { client = &http.Client{Timeout: 12 * time.Second} }
	endpoint := toolboxProfilesURL + "?profile=" + url.QueryEscape(profile)
	resp, err := client.Get(endpoint)
	if err != nil { return AccountSnapshot{}, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return AccountSnapshot{}, fmt.Errorf("toolbox profile returned HTTP %d", resp.StatusCode) }
	limited := io.LimitReader(resp.Body, 8<<20)
	raw, err := io.ReadAll(limited)
	if err != nil { return AccountSnapshot{}, err }
	return decodeFlexibleSnapshot(raw, p.Name())
}

func decodeFlexibleSnapshot(raw []byte, source string) (AccountSnapshot, error) {
	var direct AccountSnapshot
	if err := json.Unmarshal(raw, &direct); err == nil && (len(direct.Systems) > 0 || len(direct.Characters) > 0 || direct.World > 0) {
		direct.Source = source
		if direct.CapturedAt.IsZero() { direct.CapturedAt = time.Now() }
		if direct.Systems == nil { direct.Systems = map[string]SystemState{} }
		return direct, nil
	}

	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil { return AccountSnapshot{}, err }
	return normalizeGenericSnapshot(generic, source), nil
}

func normalizeGenericSnapshot(generic map[string]any, source string) AccountSnapshot {
	s := AccountSnapshot{Source: source, Systems: map[string]SystemState{}, Raw: generic, CapturedAt: time.Now()}

	if value, ok := firstString(generic, "accountName", "mainChar", "profileName", "name"); ok { s.AccountName = value }
	if value, ok := firstNumber(generic, "world", "highestWorld", "worldProgress"); ok { s.World = int(value) }

	for id := range idleonSystems {
		if value, ok := findNumericDeep(generic, id, 4); ok {
			progress := value
			if progress > 1 && progress <= 100 { progress /= 100 }
			if progress > 1 { progress = 1 }
			if progress < 0 { progress = 0 }
			s.Systems[id] = SystemState{Progress: progress, Ready: true}
		}
	}

	if chars, ok := generic["characters"].([]any); ok {
		for _, rawChar := range chars {
			m, ok := rawChar.(map[string]any); if !ok { continue }
			c := CharacterSnapshot{}
			c.Name, _ = firstString(m, "name", "characterName", "playerName")
			c.Class, _ = firstString(m, "class", "className", "eliteClass")
			if lvl, ok := firstNumber(m, "level", "classLevel"); ok { c.Level = int(lvl) }
			if active, ok := m["active"].(bool); ok { c.Active = active }
			c.Map, _ = firstString(m, "map", "mapName", "currentMap")
			s.Characters = append(s.Characters, c)
		}
	}

	return s
}

func firstString(m map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case string: if strings.TrimSpace(t) != "" { return t, true }
			}
		}
	}
	return "", false
}

func firstNumber(m map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case float64: return t, true
			case json.Number: n, err := t.Float64(); if err == nil { return n, true }
			}
		}
	}
	return 0, false
}

func findNumericDeep(value any, wanted string, depth int) (float64, bool) {
	if depth < 0 { return 0, false }
	switch t := value.(type) {
	case map[string]any:
		for k, v := range t {
			if strings.EqualFold(k, wanted) {
				if n, ok := v.(float64); ok { return n, true }
				if sub, ok := v.(map[string]any); ok {
					if n, ok := firstNumber(sub, "progress", "score", "percent", "completion"); ok { return n, true }
				}
			}
		}
		for _, v := range t { if n, ok := findNumericDeep(v, wanted, depth-1); ok { return n, true } }
	case []any:
		for _, v := range t { if n, ok := findNumericDeep(v, wanted, depth-1); ok { return n, true } }
	}
	return 0, false
}

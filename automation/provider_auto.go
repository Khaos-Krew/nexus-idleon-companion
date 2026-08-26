package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AutoLocalProvider struct{}

func (AutoLocalProvider) Name() string { return "auto-local-companion" }

func (AutoLocalProvider) Load() (AccountSnapshot, error) {
	base := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if base == "" { return AccountSnapshot{}, errors.New("LOCALAPPDATA is unavailable; specify -snapshot, -efficiency, or -toolbox") }
	root := filepath.Join(base, "Idleon Account Monitor")
	candidates := make([]string, 0)
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() { return nil }
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".json") { return nil }
		if strings.Contains(name, "snapshot") || strings.Contains(name, "capture") || strings.Contains(name, "account") || strings.Contains(name, "export") {
			candidates = append(candidates, path)
		}
		return nil
	})
	if len(candidates) == 0 { return AccountSnapshot{}, errors.New("no local companion account snapshot found; capture the account or specify another source") }
	sort.Slice(candidates, func(i,j int) bool {
		a,_:=os.Stat(candidates[i]); b,_:=os.Stat(candidates[j]);
		if a==nil { return false }; if b==nil { return true }
		return a.ModTime().After(b.ModTime())
	})
	return FileSnapshotProvider{Path:candidates[0],Source:"auto-local-companion"}.Load()
}

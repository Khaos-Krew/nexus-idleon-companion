//go:build !windows

package main

import "time"

func DetectGame(windowTitle string) GameState {
	return GameState{CheckedAt: time.Now()}
}

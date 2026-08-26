//go:build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procIsWindow            = user32.NewProc("IsWindow")
)

func windowText(hwnd uintptr) string {
	buf := make([]uint16, 512)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func DetectGame(windowTitle string) GameState {
	state := GameState{CheckedAt: time.Now()}
	ptr, err := utf16Ptr(windowTitle)
	if err != nil { return state }
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	if hwnd != 0 {
		valid, _, _ := procIsWindow.Call(hwnd)
		state.Running = valid != 0
		state.WindowTitle = windowText(hwnd)
		var r winRect
		if ok,_,_:=procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ok!=0 {
			state.X=int(r.Left); state.Y=int(r.Top); state.Width=int(r.Right-r.Left); state.Height=int(r.Bottom-r.Top)
		}
	}
	fg, _, _ := procGetForegroundWindow.Call()
	state.Foreground = state.Running && fg == hwnd
	return state
}

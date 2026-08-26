//go:build windows

package main

import (
    "errors"
    "fmt"
    "strings"
    "syscall"
    "time"
    "unsafe"
)

var (
    user32              = syscall.NewLazyDLL("user32.dll")
    procFindWindowW     = user32.NewProc("FindWindowW")
    procSetForeground   = user32.NewProc("SetForegroundWindow")
    procSetCursorPos    = user32.NewProc("SetCursorPos")
    procMouseEvent      = user32.NewProc("mouse_event")
    procGetAsyncKey     = user32.NewProc("GetAsyncKeyState")
    procGetCursorPos    = user32.NewProc("GetCursorPos")
    procKeybdEvent      = user32.NewProc("keybd_event")
)

const (
    mouseLeftDown = 0x0002
    mouseLeftUp   = 0x0004
    keyUpFlag     = 0x0002
)

type winPoint struct {
    X int32
    Y int32
}

type WindowsInput struct{}

func NewInputDriver() InputDriver { return WindowsInput{} }

func utf16Ptr(value string) (*uint16, error) {
    return syscall.UTF16PtrFromString(value)
}

func (WindowsInput) FocusWindow(title string) error {
    ptr, err := utf16Ptr(title)
    if err != nil {
        return err
    }
    hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
    if hwnd == 0 {
        return fmt.Errorf("window not found: %s", title)
    }
    ok, _, _ := procSetForeground.Call(hwnd)
    if ok == 0 {
        return errors.New("could not focus IdleOn window")
    }
    time.Sleep(250 * time.Millisecond)
    return nil
}

func (WindowsInput) Click(x, y int, double bool) error {
    ok, _, _ := procSetCursorPos.Call(uintptr(x), uintptr(y))
    if ok == 0 {
        return errors.New("could not move mouse cursor")
    }
    clickOnce := func() {
        procMouseEvent.Call(mouseLeftDown, 0, 0, 0, 0)
        procMouseEvent.Call(mouseLeftUp, 0, 0, 0, 0)
    }
    clickOnce()
    if double {
        time.Sleep(80 * time.Millisecond)
        clickOnce()
    }
    return nil
}

func normalizeKey(key string) string {
    return strings.ToUpper(strings.TrimSpace(key))
}

func vkCode(key string) (byte, bool) {
    key = normalizeKey(key)
    if len(key) == 1 {
        c := key[0]
        if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
            return c, true
        }
    }
    table := map[string]byte{
        "CTRL": 0x11, "CONTROL": 0x11, "SHIFT": 0x10, "ALT": 0x12,
        "ENTER": 0x0D, "SPACE": 0x20, "TAB": 0x09, "ESC": 0x1B, "ESCAPE": 0x1B,
        "UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
        "F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73, "F5": 0x74, "F6": 0x75,
        "F7": 0x76, "F8": 0x77, "F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B,
    }
    code, ok := table[key]
    return code, ok
}

func keyEvent(code byte, up bool) {
    flags := uintptr(0)
    if up {
        flags = keyUpFlag
    }
    procKeybdEvent.Call(uintptr(code), 0, flags, 0)
}

func (WindowsInput) PressKey(key string) error {
    parts := strings.Split(key, "+")
    codes := make([]byte, 0, len(parts))
    for _, part := range parts {
        code, ok := vkCode(part)
        if !ok {
            return fmt.Errorf("unsupported key %q", part)
        }
        codes = append(codes, code)
    }
    for _, code := range codes {
        keyEvent(code, false)
        time.Sleep(20 * time.Millisecond)
    }
    for i := len(codes) - 1; i >= 0; i-- {
        keyEvent(codes[i], true)
        time.Sleep(20 * time.Millisecond)
    }
    return nil
}

func (WindowsInput) KeyDown(key string) bool {
    code, ok := vkCode(key)
    if !ok {
        return false
    }
    state, _, _ := procGetAsyncKey.Call(uintptr(code))
    return int16(state&0xFFFF) < 0
}

func (WindowsInput) CursorPosition() (Point, error) {
    var p winPoint
    ok, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
    if ok == 0 {
        return Point{}, errors.New("could not read mouse cursor position")
    }
    return Point{X: int(p.X), Y: int(p.Y)}, nil
}

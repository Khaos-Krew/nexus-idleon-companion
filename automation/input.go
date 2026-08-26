package main

type InputDriver interface {
    FocusWindow(title string) error
    Click(x, y int, double bool) error
    PressKey(key string) error
    KeyDown(key string) bool
    CursorPosition() (Point, error)
}

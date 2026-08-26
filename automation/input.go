package main

type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type RGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

type InputDriver interface {
	FocusWindow(title string) error
	Click(x, y int, double bool) error
	PressKey(key string) error
	KeyDown(key string) bool
	CursorPosition() (Point, error)
	WindowRect(title string) (Rect, error)
	PixelColor(x, y int) (RGB, error)
}

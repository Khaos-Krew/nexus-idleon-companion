//go:build !windows

package main

import "errors"

type unsupportedInput struct{}

func NewInputDriver() InputDriver { return unsupportedInput{} }
func (unsupportedInput) FocusWindow(string) error { return errors.New("live automation requires Windows") }
func (unsupportedInput) Click(int, int, bool) error { return errors.New("live automation requires Windows") }
func (unsupportedInput) PressKey(string) error { return errors.New("live automation requires Windows") }
func (unsupportedInput) KeyDown(string) bool { return false }
func (unsupportedInput) CursorPosition() (Point, error) { return Point{}, errors.New("live automation requires Windows") }
func (unsupportedInput) WindowRect(string) (Rect, error) { return Rect{}, errors.New("live automation requires Windows") }
func (unsupportedInput) PixelColor(int, int) (RGB, error) { return RGB{}, errors.New("live automation requires Windows") }

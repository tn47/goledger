package api

import "fmt"

import "github.com/prataprc/color"

var YellowFn = color.New(color.FgYellow).FormatFunc()
var RedFn = color.New(color.FgRed).FormatFunc()

func Color(fg color.Attribute, s string) fmt.Formatter {
	switch fg {
	case color.FgYellow:
		return YellowFn(s)
	case color.FgRed:
		return RedFn(s)
	}
	panic("impossible situation")
}

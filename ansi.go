// Package ansilog is a minimal helper to print colored text.
// See https://misc.flogisoft.com/bash/tip_colors_and_formatting
// and: https://en.wikipedia.org/wiki/ANSI_escape_code#Colors
package ansilog

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

type ConsoleColorsModeEnum int

const (
	ConsoleColorsModeAuto ConsoleColorsModeEnum = iota
	ConsoleColorsModeDisabled
	ConsoleColorsModeEnabled
)

// Color ANSI codes ----------------------------------------------------------------------------------------------------

type color string

const (
	// from gin
	defaultFG color = "39"

	// fixme: test
	green   color = "97;42m"
	white   color = "90;47m"
	yellow  color = "90;43m"
	red     color = "97;41m"
	blue    color = "97;44m"
	magenta color = "97;45m"
	cyan    color = "97;46m"

	black color = "97m" // inverted with white
	//white        color = "30m" // inverted with black
	//red          color = "31m"
	//green        color = "32m"
	//yellow       color = "33m"
	//blue         color = "34m"
	//magenta      color = "35m"
	//cyan         color = "36m"
	lightGrey    color = "37m"
	darkGrey     color = "90m"
	lightRed     color = "91m"
	lightGreen   color = "92m"
	lightYellow  color = "93m"
	lightBlue    color = "94m"
	lightMagenta color = "95m"
	lightCyan    color = "96m"

	esc   = "\033["
	clear = "\033[0m"
)

func IsTerm(out io.Writer) bool {
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		return false
	}
	return true
}

type painter func(interface{}) string

// NewPainter is a PainterFunc which return a painter that can be stored and reused.
func NewPainter(color color) painter {
	return func(arg interface{}) string {
		return colored(fmt.Sprint(arg), color)
	}
}

// NewDynamicPainter is a PainterFunc which return a painter that can be stored and reused.
// It also takes a func to determine if it must use colors or not.
func NewDynamicPainter(color color, mustPaint func() bool) painter {
	return func(arg interface{}) string {
		if mustPaint != nil {
			if !mustPaint() {
				return fmt.Sprint(arg)
			}
		}
		return colored(fmt.Sprint(arg), color)
	}
}

// Black return the argument as a color escaped string
func Black(arg interface{}) string {
	return colored(fmt.Sprint(arg), black)
}

// Red return the argument as a color escaped string
func Red(arg interface{}) string {
	return colored(fmt.Sprint(arg), red)
}

// Green return the argument as a color escaped string
func Green(arg interface{}) string {
	return colored(fmt.Sprint(arg), green)
}

// Yellow return the argument as a color escaped string
func Yellow(arg interface{}) string {
	return colored(fmt.Sprint(arg), yellow)
}

// Blue return the argument as a color escaped string
func Blue(arg interface{}) string {
	return colored(fmt.Sprint(arg), blue)
}

// Magenta return the argument as a color escaped string
func Magenta(arg interface{}) string {
	return colored(fmt.Sprint(arg), magenta)
}

// Cyan return the argument as a color escaped string
func Cyan(arg interface{}) string {
	return colored(fmt.Sprint(arg), cyan)
}

// LightGrey return the argument as a color escaped string
func LightGrey(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightGrey)
}

// DarkGrey return the argument as a color escaped string
func DarkGrey(arg interface{}) string {
	return colored(fmt.Sprint(arg), darkGrey)
}

// LightRed return the argument as a color escaped string
func LightRed(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightRed)
}

// LightGreen return the argument as a color escaped string
func LightGreen(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightGreen)
}

// LightYellow return the argument as a color escaped string
func LightYellow(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightYellow)
}

// LightBlue return the argument as a color escaped string
func LightBlue(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightBlue)
}

// LightMagenta return the argument as a color escaped string
func LightMagenta(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightMagenta)
}

// LightCyan return the argument as a color escaped string
func LightCyan(arg interface{}) string {
	return colored(fmt.Sprint(arg), lightCyan)
}

// White return the argument as a color escaped string
func White(arg interface{}) string {
	return colored(fmt.Sprint(arg), white)
}

// colored return the ANSI colored string.
func colored(arg string, color color) string {
	if len(color) > 0 {
		return fmt.Sprintf(esc+"%s%s"+clear, color, arg)
	}
	return fmt.Sprintf("%s", arg)
}

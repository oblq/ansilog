// Package ansilog is a minimal helper to print colored text.
// See https://misc.flogisoft.com/bash/tip_colors_and_formatting
// and: https://en.wikipedia.org/wiki/ANSI_escape_code#Colors
package ansilog

import (
	"fmt"
)

type color string

// Color ANSI codes
const (
	//defaultFG color = "39"

	black        color = "97" // inverted with white
	red          color = "31"
	green        color = "32"
	yellow       color = "33"
	blue         color = "34"
	magenta      color = "35"
	cyan         color = "36"
	lightGrey    color = "37"
	darkGrey     color = "90"
	lightRed     color = "91"
	lightGreen   color = "92"
	lightYellow  color = "93"
	lightBlue    color = "94"
	lightMagenta color = "95"
	lightCyan    color = "96"
	white        color = "30" // inverted with black

	esc   = "\033["
	clear = "\033[0m"
)

type painter func(interface{}) string

func dynamicPainter(color color) painter {
	return func(arg interface{}) string {
		return colorize(fmt.Sprint(arg), color)
	}
}

// Black return the argument as a color escaped string
func Black(arg interface{}) string {
	return colorize(fmt.Sprint(arg), black)
}

// Red return the argument as a color escaped string
func Red(arg interface{}) string {
	return colorize(fmt.Sprint(arg), red)
}

// Green return the argument as a color escaped string
func Green(arg interface{}) string {
	return colorize(fmt.Sprint(arg), green)
}

// Yellow return the argument as a color escaped string
func Yellow(arg interface{}) string {
	return colorize(fmt.Sprint(arg), yellow)
}

// Blue return the argument as a color escaped string
func Blue(arg interface{}) string {
	return colorize(fmt.Sprint(arg), blue)
}

// Magenta return the argument as a color escaped string
func Magenta(arg interface{}) string {
	return colorize(fmt.Sprint(arg), magenta)
}

// Cyan return the argument as a color escaped string
func Cyan(arg interface{}) string {
	return colorize(fmt.Sprint(arg), cyan)
}

// LightGrey return the argument as a color escaped string
func LightGrey(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightGrey)
}

// DarkGrey return the argument as a color escaped string
func DarkGrey(arg interface{}) string {
	return colorize(fmt.Sprint(arg), darkGrey)
}

// LightRed return the argument as a color escaped string
func LightRed(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightRed)
}

// LightGreen return the argument as a color escaped string
func LightGreen(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightGreen)
}

// LightYellow return the argument as a color escaped string
func LightYellow(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightYellow)
}

// LightBlue return the argument as a color escaped string
func LightBlue(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightBlue)
}

// LightMagenta return the argument as a color escaped string
func LightMagenta(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightMagenta)
}

// LightCyan return the argument as a color escaped string
func LightCyan(arg interface{}) string {
	return colorize(fmt.Sprint(arg), lightCyan)
}

// White return the argument as a color escaped string
func White(arg interface{}) string {
	return colorize(fmt.Sprint(arg), white)
}

// colored return the ANSI colored formatted string.
func colorize(arg string, color color) string {
	coloredFormat := "%v"
	if len(color) > 0 {
		coloredFormat = esc + "%vm%v" + clear
		return fmt.Sprintf(coloredFormat, color, arg)
	}
	return fmt.Sprintf(coloredFormat, arg)
}

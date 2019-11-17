package ansilog

import (
	"fmt"
	"testing"
)

func Test_ansilog(t *testing.T) {
	fmt.Println(Magenta("Magenta"))
	fmt.Println(LightMagenta("LightMagenta"))

	fmt.Println(Red("Red"))
	fmt.Println(LightRed("LightRed"))

	fmt.Println(Yellow("Yellow"))
	fmt.Println(LightYellow("LightYellow"))

	fmt.Println(Green("Green"))
	fmt.Println(LightGreen("LightGreen"))

	fmt.Println(Blue("Blue"))
	fmt.Println(LightBlue("LightBlue"))

	fmt.Println(Cyan("Cyan"))
	fmt.Println(LightCyan("LightCyan"))

	fmt.Println(White("White"))
	fmt.Println(LightGrey("LightGrey"))
	fmt.Println(DarkGrey("DarkGrey"))
	fmt.Println(Black("Black"))
}

package ui

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	Reset       = "\033[0m"
	Bold        = "\033[1m"
	Dim         = "\033[2m"
	Italic      = "\033[3m"
	Underline   = "\033[4m"
	ClearScreen = "\033[H\033[2J\033[3J"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"

	Red     = "\033[1;31m"
	Green   = "\033[1;32m"
	Yellow  = "\033[1;33m"
	Blue    = "\033[1;34m"
	Magenta = "\033[1;35m"
	Cyan    = "\033[1;36m"
	White   = "\033[1;37m"
	Gray    = "\033[38;5;244m"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func VisualLen(s string) int {
	return len([]rune(StripANSI(s)))
}

func PadRight(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

func PadLeft(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	return strings.Repeat(" ", width-vis) + s
}

func PadCenter(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	diff := width - vis
	left := diff / 2
	right := diff - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func ClearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	fmt.Print(ClearScreen)
}

func Rgb(r, g, b int, text string) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s%s", r, g, b, text, Reset)
}

func NeonGradient(text string, offset float64) string {
	var sb strings.Builder
	for i, r := range text {
		t := float64(i)*0.15 + offset
		red := int(127 + 127*math.Sin(t))
		green := int(127 + 127*math.Sin(t+2.0))
		blue := int(200 + 55*math.Cos(t))
		sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c", red, green, blue, r))
	}
	sb.WriteString(Reset)
	return sb.String()
}

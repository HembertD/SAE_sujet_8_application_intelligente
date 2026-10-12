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

// Supprime les séquences d'échappement ANSI d'une chaîne.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Calcule la longueur visible d'une chaîne sans les codes ANSI.
func VisualLen(s string) int {
	return len([]rune(StripANSI(s)))
}

// Aligne une chaîne à gauche en complétant à droite.
func PadRight(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

// Aligne une chaîne à droite en complétant à gauche.
func PadLeft(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	return strings.Repeat(" ", width-vis) + s
}

// Centre une chaîne en complétant des deux côtés.
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

// Efface complètement le terminal et son historique de défilement.
func ClearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	fmt.Print(ClearScreen)
}

// Colore un texte avec une couleur TrueColor RGB.
func Rgb(r, g, b int, text string) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s%s", r, g, b, text, Reset)
}

// Applique un dégradé néon animé à un texte.
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

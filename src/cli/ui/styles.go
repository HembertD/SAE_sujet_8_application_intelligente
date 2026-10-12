package ui

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ANSI Styling Constants
const (
	Reset       = "\033[0m"
	Bold        = "\033[1m"
	Dim         = "\033[2m"
	Italic      = "\033[3m"
	Underline   = "\033[4m"
	ClearScreen = "\033[H\033[2J\033[3J"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"

	// Couleurs standard vives
	Red     = "\033[1;31m"
	Green   = "\033[1;32m"
	Yellow  = "\033[1;33m"
	Blue    = "\033[1;34m"
	Magenta = "\033[1;35m"
	Cyan    = "\033[1;36m"
	White   = "\033[1;37m"
	Gray    = "\033[38;5;244m"
)

// Regex pour isoler et supprimer les séquences ANSI lors du calcul de longueur visuelle
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI retire toutes les séquences d'échappement ANSI d'une chaîne
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// VisualLen renvoie la largeur visuelle exacte dans le terminal (en nombre de runes visibles)
func VisualLen(s string) int {
	return len([]rune(StripANSI(s)))
}

// PadRight complète la chaîne avec des espaces à droite pour atteindre 'width'
func PadRight(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

// PadLeft complète la chaîne avec des espaces à gauche pour atteindre 'width'
func PadLeft(s string, width int) string {
	vis := VisualLen(s)
	if vis >= width {
		return s
	}
	return strings.Repeat(" ", width-vis) + s
}

// PadCenter centre la chaîne dans une largeur donnée
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

// ClearTerminal efface complètement l'écran et le tampon d'historique (scrollback)
func ClearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	fmt.Print(ClearScreen)
}

// Rgb colore un texte avec une couleur TrueColor 24-bit
func Rgb(r, g, b int, text string) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s%s", r, g, b, text, Reset)
}

// NeonGradient applique un dégradé animé néon (Cyan -> Magenta -> Violet)
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

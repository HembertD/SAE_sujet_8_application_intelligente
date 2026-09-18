package views

import (
	"bufio"
	"fmt"
	"strings"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/ui"
)

// ShowMainMenu affiche le menu principal et retourne le choix de l'utilisateur
func ShowMainMenu(reader *bufio.Reader, client *bridge.BackendClient, status *models.RepoStatus) string {
	ui.ClearTerminal()
	ui.PrintBanner(status)

	width := 72
	var menuLines []string
	menuLines = append(menuLines, "")
	menuLines = append(menuLines, fmt.Sprintf("  %s[1]%s 📝 %sGénérer un message de commit%s  (IA Conventional Commits)", ui.Yellow+ui.Bold, ui.Reset, ui.Bold, ui.Reset))
	menuLines = append(menuLines, fmt.Sprintf("  %s[2]%s 📦 %sGénérer les Release Notes%s     (Synthèse entre tags)", ui.Yellow+ui.Bold, ui.Reset, ui.Bold, ui.Reset))
	menuLines = append(menuLines, fmt.Sprintf("  %s[3]%s ⚙️  %sConfiguration%s                  (Modèle, URL Ollama, Langue)", ui.Yellow+ui.Bold, ui.Reset, ui.Bold, ui.Reset))
	menuLines = append(menuLines, fmt.Sprintf("  %s[4]%s 🚪 %sQuitter%s", ui.Yellow+ui.Bold, ui.Reset, ui.Bold, ui.Reset))
	menuLines = append(menuLines, "")

	if client.UseMockMode {
		menuLines = append(menuLines, ui.Dim+"  ℹ Mode simulation actif (le backend Python se connectera automatiquement)"+ui.Reset)
	}

	ui.PrintCard("Menu Principal", menuLines, width, ui.Cyan)

	fmt.Printf("\n  %s👉 Choisissez une option (1-4 ou q) :%s ", ui.Bold+ui.Yellow, ui.Reset)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "q"
	}
	return strings.TrimSpace(input)
}

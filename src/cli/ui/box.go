package ui

import (
	"fmt"
	"strings"

	"sae-git-cli/models"
)

// PrintBanner affiche la bannière stylisée principale en haut du TUI
func PrintBanner(repoStatus *models.RepoStatus) {
	width := 72
	fmt.Println(Cyan + "╭" + strings.Repeat("─", width-2) + "╮" + Reset)
	fmt.Printf("%s│%s│%s\n", Cyan, PadCenter(Bold+Yellow+"⚡ SMART COMMIT & RELEASE NOTES GENERATOR"+Reset, width-2), Cyan)
	fmt.Printf("%s│%s│%s\n", Cyan, PadCenter(Dim+"Application Intelligente - SAÉ Sujet 8"+Reset, width-2), Cyan)
	fmt.Println(Cyan + "├" + strings.Repeat("─", width-2) + "┤" + Reset)

	if repoStatus != nil && repoStatus.IsGitRepo {
		pathText := fmt.Sprintf("📁 Dépôt : %s", repoStatus.RepoPath)
		fmt.Printf("%s│  %s│%s\n", Cyan, PadRight(pathText, width-4), Cyan)

		stagedColor := Green
		if repoStatus.StagedCount == 0 {
			stagedColor = Yellow
		}
		statusText := fmt.Sprintf("🌿 Branche : %s%s%s  |  Staged : %s%d%s  |  Non-staged : %d",
			Bold+Green, repoStatus.Branch, Reset,
			stagedColor+Bold, repoStatus.StagedCount, Reset,
			repoStatus.UnstagedCount,
		)
		fmt.Printf("%s│  %s│%s\n", Cyan, PadRight(statusText, width-4), Cyan)
	} else {
		errText := Red + "⚠️ Aucun dépôt Git détecté dans le répertoire courant" + Reset
		fmt.Printf("%s│  %s│%s\n", Cyan, PadRight(errText, width-4), Cyan)
	}

	fmt.Println(Cyan + "╰" + strings.Repeat("─", width-2) + "╯" + Reset)
	fmt.Println()
}

// PrintCard affiche un panneau encadré avec titre et lignes de contenu
func PrintCard(title string, lines []string, width int, borderColor string) {
	if borderColor == "" {
		borderColor = Cyan
	}
	titleFormatted := ""
	if title != "" {
		titleFormatted = " " + Bold + title + Reset + " "
	}
	topBarLen := width - 2 - VisualLen(titleFormatted)
	if topBarLen < 0 {
		topBarLen = 0
	}

	fmt.Println(borderColor + "╭─" + Reset + titleFormatted + borderColor + strings.Repeat("─", topBarLen) + "╮" + Reset)
	for _, line := range lines {
		fmt.Printf("%s│%s %s %s│%s\n", borderColor, Reset, PadRight(line, width-4), borderColor, Reset)
	}
	fmt.Println(borderColor + "╰" + strings.Repeat("─", width-2) + "╯" + Reset)
}

// PrintDiffTable affiche les fichiers modifiés avec coloration de statut
func PrintDiffTable(files []models.DiffFile, width int) {
	if len(files) == 0 {
		PrintWarning("Aucun fichier indexé (staged) pour le moment.")
		return
	}

	var lines []string
	lines = append(lines, Bold+"Fichiers détectés dans le diff indexé :"+Reset)
	lines = append(lines, Dim+strings.Repeat("─", width-6)+Reset)

	for _, f := range files {
		statusColor := Yellow
		switch f.Status {
		case "A":
			statusColor = Green
		case "D":
			statusColor = Red
		case "M":
			statusColor = Cyan
		}

		diffStats := fmt.Sprintf("(+%d / -%d)", f.Added, f.Removed)
		if f.Binary {
			diffStats = Yellow + "[Binaire]" + Reset
		}

		warning := ""
		if f.HasSecretsMasked {
			warning = " " + Red + Bold + "🔒 [Secret Masqué]" + Reset
		}

		fileLine := fmt.Sprintf(" %s[%s]%s %-34s %s%s",
			statusColor+Bold, f.Status, Reset,
			f.Path,
			diffStats,
			warning,
		)
		lines = append(lines, fileLine)
	}

	PrintCard("Modifications Staged", lines, width, Blue)
}

// PrintCommitCard affiche le message de commit structuré proposé par l'IA
func PrintCommitCard(typeStr, scope, subject, body string, width int) {
	var lines []string

	isErr := strings.ToLower(typeStr) == "erreur" || strings.ToLower(typeStr) == "error" || (typeStr == "" && subject == "")

	borderColor := Green
	cardTitle := "Proposition Conventional Commit"
	headerColor := Bold + Green

	if isErr {
		borderColor = Red
		cardTitle = "Erreur de Génération"
		headerColor = Bold + Red
		if typeStr == "" && subject == "" {
			typeStr = "erreur"
			subject = "aucune réponse du serveur (timeout ou échec IA)"
		}
	}

	header := fmt.Sprintf("%s%s%s", headerColor, typeStr, Reset)
	if scope != "" {
		header = fmt.Sprintf("%s(%s%s%s)", header, Yellow, scope, Reset)
	}
	header = fmt.Sprintf("%s: %s%s%s", header, Bold+White, subject, Reset)

	lines = append(lines, "")
	lines = append(lines, "  "+header)
	lines = append(lines, "")

	if body != "" {
		lines = append(lines, Dim+strings.Repeat("─", width-6)+Reset)
		lines = append(lines, "")
		for _, bLine := range strings.Split(body, "\n") {
			lines = append(lines, "  "+bLine)
		}
		lines = append(lines, "")
	}

	PrintCard(cardTitle, lines, width, borderColor)
}

// PrintSuccess affiche un message de succès
func PrintSuccess(msg string) {
	fmt.Printf("\n  %s✔%s %s%s%s\n\n", Green, Reset, Bold+Green, msg, Reset)
}

// PrintError affiche une alerte d'erreur
func PrintError(msg string) {
	fmt.Printf("\n  %s✖ Erreur :%s %s\n\n", Red+Bold, Reset, msg)
}

// PrintWarning affiche un avertissement
func PrintWarning(msg string) {
	fmt.Printf("\n  %s⚠️ Avertissement :%s %s\n\n", Yellow+Bold, Reset, msg)
}

// PrintInfo affiche une information
func PrintInfo(msg string) {
	fmt.Printf("  %sℹ%s %s\n", Cyan+Bold, Reset, msg)
}

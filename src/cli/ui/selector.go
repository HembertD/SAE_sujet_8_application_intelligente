package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"sae-git-cli/models"
)

// SelectCommit affiche un sélecteur interactif au clavier pour choisir un commit dans l'historique complet.
//
// Touches :
//   - Flèches Haut/Bas ou j/k : déplacer le curseur
//   - Page Up / Page Down : défilement par bloc
//   - Entrée : valider la sélection
//   - Échap ou q : annuler et fermer le sélecteur
func SelectCommit(commits []models.CommitItem, title string, defaultIndex int) (models.CommitItem, bool, error) {
	if len(commits) == 0 {
		return models.CommitItem{}, true, fmt.Errorf("aucun commit disponible dans l'historique")
	}

	// Vérification si stdin est un terminal interactif
	fileInfo, err := os.Stdin.Stat()
	isTerminal := (err == nil) && ((fileInfo.Mode() & os.ModeCharDevice) != 0)

	if !isTerminal {
		// En environnement non interactif (tests automatisés, redirection)
		idx := defaultIndex
		if idx < 0 || idx >= len(commits) {
			idx = 0
		}
		return commits[idx], false, nil
	}

	// Bascule du terminal en mode cbreak / non-canonique sans echo
	rawCmd := exec.Command("stty", "-echo", "-icanon", "min", "1")
	rawCmd.Stdin = os.Stdin
	if err := rawCmd.Run(); err != nil {
		idx := defaultIndex
		if idx < 0 || idx >= len(commits) {
			idx = 0
		}
		return commits[idx], false, nil
	}

	// Restauration garantie du terminal à la sortie
	defer func() {
		restoreCmd := exec.Command("stty", "echo", "icanon")
		restoreCmd.Stdin = os.Stdin
		_ = restoreCmd.Run()
		fmt.Print(ShowCursor)
	}()

	fmt.Print(HideCursor)

	selectedIndex := defaultIndex
	if selectedIndex < 0 {
		selectedIndex = 0
	}
	if selectedIndex >= len(commits) {
		selectedIndex = len(commits) - 1
	}

	viewportSize := 10
	if len(commits) < viewportSize {
		viewportSize = len(commits)
	}

	width := 72
	totalLinesRendered := 0

	for {
		// Calcul de la fenêtre déroulante (scrollOffset)
		scrollOffset := 0
		if selectedIndex >= viewportSize {
			scrollOffset = selectedIndex - viewportSize + 1
		}
		if scrollOffset+viewportSize > len(commits) {
			scrollOffset = len(commits) - viewportSize
		}
		if scrollOffset < 0 {
			scrollOffset = 0
		}

		// Préparation des lignes d'affichage
		var lines []string

		// Titre et instructions
		topBarLen := width - VisualLen(title) - 5
		if topBarLen < 0 {
			topBarLen = 0
		}
		lines = append(lines, Cyan+"╭─ "+Bold+title+Reset+Cyan+" "+strings.Repeat("─", topBarLen)+"╮"+Reset)
		instruction := "Naviguer : [↑/↓]  •  Valider : [Entrée]  •  Annuler : [q/Échap]"
		lines = append(lines, fmt.Sprintf("%s│%s  %s%s│%s", Cyan, Dim, PadRight(instruction, width-4), Reset+Cyan, Reset))
		lines = append(lines, Cyan+"├"+strings.Repeat("─", width-2)+"┤"+Reset)

		// Indicateur de défilement vers le haut
		if scrollOffset > 0 {
			upText := fmt.Sprintf("▲  %d commit(s) plus récents au-dessus", scrollOffset)
			lines = append(lines, fmt.Sprintf("%s│%s %s %s│%s", Cyan, Yellow, PadCenter(upText, width-4), Cyan, Reset))
		} else {
			lines = append(lines, fmt.Sprintf("%s│%s %s %s│%s", Cyan, Dim, PadCenter("─ Début de l'historique (HEAD) ─", width-4), Cyan, Reset))
		}

		// Affichage des commits visibles dans le viewport
		for i := 0; i < viewportSize; i++ {
			cIdx := scrollOffset + i
			if cIdx >= len(commits) {
				break
			}
			commit := commits[cIdx]

			isSelected := (cIdx == selectedIndex)
			cursor := "  "
			itemColor := Reset
			shaColor := Cyan
			if isSelected {
				cursor = Bold + Yellow + "❯ " + Reset
				itemColor = Bold + White
				shaColor = Bold + Yellow
			}

			tagBadge := ""
			if commit.Tag != "" {
				tagBadge = Magenta + "(" + commit.Tag + ") " + Reset
			}

			// Formatage de la ligne du commit
			shaStr := fmt.Sprintf("%s[%s]%s", shaColor, commit.Sha, Reset)
			// Troncature du message
			maxMsgLen := width - 30 - VisualLen(tagBadge)
			if maxMsgLen < 15 {
				maxMsgLen = 15
			}
			msg := commit.Message
			if len([]rune(msg)) > maxMsgLen {
				msg = string([]rune(msg)[:maxMsgLen-3]) + "..."
			}

			dateStr := Dim + commit.Date + Reset
			rawItem := fmt.Sprintf("%s%s %s%s%s", cursor, shaStr, tagBadge, itemColor+msg+Reset, " "+dateStr)
			lines = append(lines, fmt.Sprintf("%s│%s %s %s│%s", Cyan, Reset, PadRight(rawItem, width-4), Cyan, Reset))
		}

		// Indicateur de défilement vers le bas
		remainingBelow := len(commits) - (scrollOffset + viewportSize)
		if remainingBelow > 0 {
			downText := fmt.Sprintf("▼  %d commit(s) plus anciens en-dessous", remainingBelow)
			lines = append(lines, fmt.Sprintf("%s│%s %s %s│%s", Cyan, Yellow, PadCenter(downText, width-4), Cyan, Reset))
		} else {
			lines = append(lines, fmt.Sprintf("%s│%s %s %s│%s", Cyan, Dim, PadCenter("─ Fin de l'historique ─", width-4), Cyan, Reset))
		}

		// Barre de statut inférieure
		lines = append(lines, Cyan+"├"+strings.Repeat("─", width-2)+"┤"+Reset)
		selectedCommit := commits[selectedIndex]
		statusText := fmt.Sprintf("Commit %d/%d  |  SHA: %s  |  Auteur: %s",
			selectedIndex+1, len(commits), selectedCommit.Sha, selectedCommit.Author)
		lines = append(lines, fmt.Sprintf("%s│%s  %s%s│%s", Cyan, Dim, PadRight(statusText, width-4), Reset+Cyan, Reset))
		lines = append(lines, Cyan+"╰"+strings.Repeat("─", width-2)+"╯"+Reset)

		// Si c'est un re-rendu, on remonte le curseur pour écraser l'affichage précédent
		if totalLinesRendered > 0 {
			fmt.Printf("\r\033[%dA", totalLinesRendered)
		}

		for _, l := range lines {
			fmt.Println("\r" + l + "\033[K")
		}
		totalLinesRendered = len(lines)

		// Lecture de la touche saisie
		buf := make([]byte, 16)
		n, readErr := os.Stdin.Read(buf)
		if readErr != nil || n == 0 {
			return models.CommitItem{}, true, readErr
		}

		// Validation (Entrée)
		if buf[0] == '\r' || buf[0] == '\n' {
			// Effacement propre du sélecteur
			fmt.Printf("\r\033[%dA\033[J", totalLinesRendered)
			return commits[selectedIndex], false, nil
		}

		// Annulation ('q', 'Q', Escape seul, Ctrl+C)
		if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 3 || (n == 1 && buf[0] == 27) {
			fmt.Printf("\r\033[%dA\033[J", totalLinesRendered)
			return models.CommitItem{}, true, nil
		}

		// Navigation clavier
		if n == 1 {
			switch buf[0] {
			case 'k', 'K':
				selectedIndex--
			case 'j', 'J':
				selectedIndex++
			case 'g':
				selectedIndex = 0
			case 'G':
				selectedIndex = len(commits) - 1
			}
		} else if n >= 3 && buf[0] == 27 && buf[1] == '[' {
			switch buf[2] {
			case 'A': // Flèche Haut
				selectedIndex--
			case 'B': // Flèche Bas
				selectedIndex++
			case '5': // Page Up
				selectedIndex -= viewportSize
			case '6': // Page Down
				selectedIndex += viewportSize
			case 'H': // Home
				selectedIndex = 0
			case 'F': // End
				selectedIndex = len(commits) - 1
			}
		}

		// Clamping
		if selectedIndex < 0 {
			selectedIndex = 0
		}
		if selectedIndex >= len(commits) {
			selectedIndex = len(commits) - 1
		}
	}
}

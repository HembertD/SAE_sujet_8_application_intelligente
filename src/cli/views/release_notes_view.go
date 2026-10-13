package views

import (
	"bufio"
	"fmt"
	"strings"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/ui"
)

// ShowReleaseNotesFlow orchestre l'écran de génération de changelog / release notes
func ShowReleaseNotesFlow(reader *bufio.Reader, client bridge.BackendClient, status *models.RepoStatus) {
	ui.ClearTerminal()
	ui.PrintBanner(status)
	width := 72

	fmt.Println(ui.Bold + ui.Cyan + "📦 Génération de Release Notes (Changelog)" + ui.Reset)
	fmt.Println(ui.Dim + "Synthèse automatique de l'historique Git catégorisée par le LLM." + ui.Reset)
	fmt.Println()

	// Saisie des bornes de révision
	fmt.Printf("  %s• Tag ou révision de départ [v0.1.0] :%s ", ui.Bold, ui.Reset)
	fromTag, _ := reader.ReadString('\n')
	fromTag = strings.TrimSpace(fromTag)
	if fromTag == "" {
		fromTag = "v0.1.0"
	}

	fmt.Printf("  %s• Tag ou révision d'arrivée [HEAD]   :%s ", ui.Bold, ui.Reset)
	toTag, _ := reader.ReadString('\n')
	toTag = strings.TrimSpace(toTag)
	if toTag == "" {
		toTag = "HEAD"
	}

	fmt.Println()
	spinner := ui.NewSpinner(fmt.Sprintf("Extraction des commits (%s ➔ %s) et rédaction par le LLM...", fromTag, toTag))
	spinner.Start()
	notes, err := client.GetReleaseNotes(fromTag, toTag)
	spinner.Stop("")

	isError := (err != nil) || (notes != nil && (!notes.Success || notes.Markdown == ""))
	if isError {
		errMsg := "délai d'attente dépassé (timeout) ou aucune réponse du serveur LLM"
		if err != nil {
			errMsg = err.Error()
		} else if notes != nil && notes.Error != "" {
			errMsg = notes.Error
		}

		ui.PrintError(fmt.Sprintf("Impossible de générer les Release Notes : %s", errMsg))
		notes = &models.ReleaseNotesResponse{
			Success:  false,
			Error:    errMsg,
			FromTag:  fromTag,
			ToTag:    toTag,
			Markdown: fmt.Sprintf("# Erreur\n\nImpossible de générer les Release Notes : %s\n\nVérifiez l'état du serveur Ollama dans le menu [3] Configuration.", errMsg),
		}
	}

	fmt.Println()
	var lines []string
	for _, l := range strings.Split(notes.Markdown, "\n") {
		lines = append(lines, l)
	}
	cardColor := ui.Magenta
	if isError {
		cardColor = ui.Red
	}
	ui.PrintCard(fmt.Sprintf("Release Notes (%s ➔ %s)", fromTag, toTag), lines, width, cardColor)
	fmt.Println()

	// Choix utilisateur
	fmt.Println(ui.Bold + "  Options :" + ui.Reset)
	fmt.Printf("    %s[S]%s Sauvegarder dans un fichier RELEASE_NOTES.md\n", ui.Green+ui.Bold, ui.Reset)
	fmt.Printf("    %s[R]%s Retourner au menu principal\n\n", ui.Yellow+ui.Bold, ui.Reset)

	fmt.Printf("  %s👉 Votre choix (S/R) :%s ", ui.Bold+ui.Yellow, ui.Reset)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	choice = strings.ToUpper(strings.TrimSpace(choice))

	if choice == "S" {
		if !notes.Success {
			ui.PrintError("Impossible d'enregistrer des Release Notes en erreur.")
			waitForEnter(reader)
			return
		}
		saveErr := client.SaveMarkdownFile("RELEASE_NOTES.md", notes.Markdown)
		if saveErr != nil {
			ui.PrintError(fmt.Sprintf("Erreur lors de l'enregistrement du fichier : %v", saveErr))
		} else {
			ui.PrintSuccess("Le fichier RELEASE_NOTES.md a été enregistré avec succès à la racine du dépôt !")
		}
		waitForEnter(reader)
	}
}

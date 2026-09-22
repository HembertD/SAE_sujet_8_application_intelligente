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

	if err != nil {
		ui.PrintError(fmt.Sprintf("Impossible de générer les Release Notes : %v", err))
		waitForEnter(reader)
		return
	}

	fmt.Println()
	var lines []string
	for _, l := range strings.Split(notes.Markdown, "\n") {
		lines = append(lines, l)
	}
	ui.PrintCard(fmt.Sprintf("Release Notes (%s ➔ %s)", fromTag, toTag), lines, width, ui.Magenta)
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
		saveErr := client.SaveMarkdownFile("RELEASE_NOTES.md", notes.Markdown)
		if saveErr != nil {
			ui.PrintError(fmt.Sprintf("Erreur lors de l'enregistrement du fichier : %v", saveErr))
		} else {
			ui.PrintSuccess("Le fichier RELEASE_NOTES.md a été enregistré avec succès à la racine du dépôt !")
		}
		waitForEnter(reader)
	}
}

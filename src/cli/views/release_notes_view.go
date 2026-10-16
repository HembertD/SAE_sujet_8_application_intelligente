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

		ui.PrintErrorCard("Erreur de Génération des Release Notes", errMsg,
			"Vérifiez que le serveur Ollama est démarré et accessible (menu [3] Configuration)",
			fmt.Sprintf("Vérifiez que les révisions '%s' et '%s' existent dans l'historique Git", fromTag, toTag),
			"Vous pouvez retourner au menu principal avec [R]",
		)

		if notes == nil {
			notes = &models.ReleaseNotesResponse{
				Success: false,
				Error:   errMsg,
				FromTag: fromTag,
				ToTag:   toTag,
			}
		}
	} else {
		fmt.Println()
		var lines []string
		for _, l := range strings.Split(notes.Markdown, "\n") {
			lines = append(lines, l)
		}
		ui.PrintCard(fmt.Sprintf("Release Notes (%s ➔ %s)", fromTag, toTag), lines, width, ui.Magenta)
		fmt.Println()
	}

	// Choix utilisateur
	for {
		fmt.Println(ui.Bold + "  Options :" + ui.Reset)
		if !isError {
			fmt.Printf("    %s[S]%s Sauvegarder dans un fichier RELEASE_NOTES.md\n", ui.Green+ui.Bold, ui.Reset)
		}
		fmt.Printf("    %s[R]%s Retourner au menu principal\n\n", ui.Yellow+ui.Bold, ui.Reset)

		promptChoices := "S/R"
		if isError {
			promptChoices = "R"
		}
		prompt := fmt.Sprintf("  %s👉 Votre choix (%s) :%s ", ui.Bold+ui.Yellow, promptChoices, ui.Reset)
		choice, err := ReadChoice(reader, prompt)
		if err != nil {
			return
		}

		if choice == "S" {
			if isError || notes == nil || !notes.Success || notes.Markdown == "" {
				ui.PrintErrorCard("Sauvegarde Impossible", "Impossible d'enregistrer des Release Notes en erreur.",
					"Assurez-vous qu'une génération valide a été effectuée avant d'enregistrer",
				)
				continue
			}
			saveErr := client.SaveMarkdownFile("RELEASE_NOTES.md", notes.Markdown)
			if saveErr != nil {
				ui.PrintErrorCard("Échec d'Enregistrement", fmt.Sprintf("Erreur lors de l'enregistrement de RELEASE_NOTES.md : %v", saveErr))
			} else {
				ui.PrintSuccess("Le fichier RELEASE_NOTES.md a été enregistré avec succès à la racine du dépôt !")
			}
			waitForEnter(reader)
			return
		} else if choice == "R" {
			return
		} else {
			if isError {
				ui.PrintWarning("Choix non reconnu, appuyez sur R pour retourner au menu principal.")
			} else {
				ui.PrintWarning("Choix non reconnu, veuillez choisir parmi S ou R.")
			}
		}
	}
}

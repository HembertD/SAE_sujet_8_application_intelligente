package views

import (
	"bufio"
	"fmt"
	"strings"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/ui"
)

// ShowReleaseNotesFlow orchestre l'écran de génération de changelog / release notes avec sélection interactive des commits
func ShowReleaseNotesFlow(reader *bufio.Reader, client bridge.BackendClient, status *models.RepoStatus) {
	ui.ClearTerminal()
	ui.PrintBanner(status)
	width := 72

	fmt.Println(ui.Bold + ui.Cyan + "📦 Génération de Release Notes (Changelog)" + ui.Reset)
	fmt.Println(ui.Dim + "Synthèse automatique de l'historique Git catégorisée par le LLM." + ui.Reset)
	fmt.Println()

	// 1. Récupération de l'historique complet des commits
	spinnerCommits := ui.NewSpinner("Chargement de l'historique complet des commits...")
	spinnerCommits.Start()
	commitsResp, errCommits := client.GetRecentCommits(0)
	spinnerCommits.Stop("")

	if errCommits != nil || commitsResp == nil || !commitsResp.Success || len(commitsResp.Commits) == 0 {
		errMsg := "Aucun commit trouvé dans ce dépôt Git."
		if errCommits != nil {
			errMsg = errCommits.Error()
		} else if commitsResp != nil && commitsResp.Error != "" {
			errMsg = commitsResp.Error
		}

		ui.PrintErrorCard("Historique Git Inaccessible", errMsg,
			"Assurez-vous que le dépôt Git contient au moins un commit",
			"Vous pouvez retourner au menu principal avec [R]",
		)
		waitForEnter(reader)
		return
	}

	commits := commitsResp.Commits

	// 2. Sélection interactive du commit de DÉPART
	fmt.Println(ui.Bold + "  Étape 1/2 : Sélection du point de départ" + ui.Reset)
	startCommit, cancelled, err := ui.SelectCommit(commits, "Sélection du commit de DÉPART", 0)
	if cancelled || err != nil {
		ui.PrintInfo("Opération annulée par l'utilisateur.")
		waitForEnter(reader)
		return
	}

	startTagBadge := ""
	if startCommit.Tag != "" {
		startTagBadge = fmt.Sprintf(" %s(%s)%s", ui.Magenta, startCommit.Tag, ui.Reset)
	}
	fmt.Printf("  %s✓ Point de départ sélectionné :%s %s[%s]%s%s %s\n\n",
		ui.Green+ui.Bold, ui.Reset, ui.Cyan+ui.Bold, startCommit.Sha, ui.Reset, startTagBadge, startCommit.Message)

	// 3. Sélection interactive du commit d'ARRIVÉE (pré-positionné sur le commit le plus récent / HEAD)
	fmt.Println(ui.Bold + "  Étape 2/2 : Sélection du point d'arrivée" + ui.Reset)
	arrivalCommit, cancelled, err := ui.SelectCommit(commits, "Sélection du commit d'ARRIVÉE", 0)
	if cancelled || err != nil {
		ui.PrintInfo("Opération annulée par l'utilisateur.")
		waitForEnter(reader)
		return
	}

	arrivalTagBadge := ""
	if arrivalCommit.Tag != "" {
		arrivalTagBadge = fmt.Sprintf(" %s(%s)%s", ui.Magenta, arrivalCommit.Tag, ui.Reset)
	}
	fmt.Printf("  %s✓ Point d'arrivée sélectionné :%s %s[%s]%s%s %s\n\n",
		ui.Green+ui.Bold, ui.Reset, ui.Cyan+ui.Bold, arrivalCommit.Sha, ui.Reset, arrivalTagBadge, arrivalCommit.Message)

	// 4. Génération des Release Notes par le LLM
	fromRef := startCommit.Sha
	toRef := arrivalCommit.Sha

	spinner := ui.NewSpinner(fmt.Sprintf("Extraction des commits (%s ➔ %s) et rédaction par le LLM...", fromRef, toRef))
	spinner.Start()
	notes, err := client.GetReleaseNotes(fromRef, toRef)
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
			fmt.Sprintf("Vérifiez la validité de l'intervalle '%s ➔ %s'", fromRef, toRef),
			"Vous pouvez retourner au menu principal avec [R]",
		)

		if notes == nil {
			notes = &models.ReleaseNotesResponse{
				Success: false,
				Error:   errMsg,
				FromTag: fromRef,
				ToTag:   toRef,
			}
		}

		ui.NotifyError(fmt.Sprintf("Échec de la génération des release notes : %s", errMsg))
	} else {
		ui.NotifySuccess("Release notes générées avec succès !")
		fmt.Println()
		var lines []string
		for _, l := range strings.Split(notes.Markdown, "\n") {
			lines = append(lines, l)
		}
		ui.PrintCard(fmt.Sprintf("Release Notes (%s ➔ %s)", fromRef, toRef), lines, width, ui.Magenta)
		fmt.Println()
	}

	// 5. Options utilisateur (Sauvegarde / Retour)
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

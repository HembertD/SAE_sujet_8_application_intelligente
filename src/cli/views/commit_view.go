package views

import (
	"bufio"
	"fmt"
	"strings"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/ui"
)

// ShowCommitFlow orchestre l'écran interactif de génération de commit
func ShowCommitFlow(reader *bufio.Reader, client bridge.BackendClient, status *models.RepoStatus) {
	ui.ClearTerminal()
	ui.PrintBanner(status)
	width := 72

	fmt.Println(ui.Bold + ui.Cyan + "📝 [ÉTAPE 1/4] Analyse des modifications locales" + ui.Reset)
	fmt.Println()

	// 1. Récupération du diff depuis le backend Python
	spinnerDiff := ui.NewSpinner("Récupération de l'état Git et filtrage des secrets...")
	spinnerDiff.Start()
	diffResp, err := client.GetDiff()
	spinnerDiff.Stop("")

	if err != nil {
		ui.PrintErrorCard("Analyse Git", fmt.Sprintf("Échec de récupération du diff : %v", err), "Vérifiez que le répertoire cible est un dépôt Git valide.")
		waitForEnter(reader)
		return
	}

	// 2. Gestion de l'absence de fichiers staged
	if len(diffResp.Files) == 0 {
		ui.PrintWarning("Aucune modification n'est actuellement indexée (staged).")
		fmt.Printf("  %s👉 Voulez-vous indexer automatiquement toutes les modifications (git add -A) ? [O/n] :%s ", ui.Bold+ui.Yellow, ui.Reset)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))

		if ans == "o" || ans == "oui" || ans == "" {
			spinnerStage := ui.NewSpinner("Indexation des fichiers en cours...")
			spinnerStage.Start()
			res, stageErr := client.StageAll()
			spinnerStage.Stop("")

			if stageErr != nil {
				ui.PrintErrorCard("Indexation Git", fmt.Sprintf("Erreur lors de l'indexation (git add -A) : %v", stageErr))
				waitForEnter(reader)
				return
			}
			ui.PrintSuccess(res.Message)

			// Re-télécharger le diff actualisé
			diffResp, _ = client.GetDiff()
		} else {
			ui.PrintInfo("Opération annulée. Indexez d'abord vos fichiers puis relancez le menu.")
			waitForEnter(reader)
			return
		}
	}

	// 3. Affichage du tableau des fichiers modifiés
	ui.PrintDiffTable(diffResp.Files, width)
	fmt.Println()

	// 4. Génération initiale du message de commit avec le LLM
	proposal, isError := generateCommit(client, "")

	// 5. Boucle d'interaction et de décision utilisateur (Event Loop)
	for {
		if !isError {
			fmt.Println()
			ui.PrintCommitCard(proposal.Type, proposal.Scope, proposal.Subject, proposal.Body, width)
			fmt.Println()
		}

		// Affichage des actions disponibles
		fmt.Println(ui.Bold + "  Actions disponibles :" + ui.Reset)
		if !isError {
			fmt.Printf("    %s[V]%s Valider et appliquer le commit\n", ui.Green+ui.Bold, ui.Reset)
		}
		fmt.Printf("    %s[R]%s Refuser et commenter pour régénérer avec l'IA\n", ui.Yellow+ui.Bold, ui.Reset)
		fmt.Printf("    %s[M]%s Modifier manuellement le message avant de commiter\n", ui.Cyan+ui.Bold, ui.Reset)
		fmt.Printf("    %s[A]%s Annuler et revenir au menu principal\n\n", ui.Red+ui.Bold, ui.Reset)

		promptChoices := "V/R/M/A"
		if isError {
			promptChoices = "R/M/A"
		}
		prompt := fmt.Sprintf("  %s👉 Votre choix (%s) :%s ", ui.Bold+ui.Yellow, promptChoices, ui.Reset)
		choice, err := ReadChoice(reader, prompt)
		if err != nil {
			return
		}

		switch choice {
		case "V":
			if isError || proposal == nil || !proposal.Success || strings.ToLower(proposal.Type) == "erreur" || proposal.Subject == "" {
				ui.PrintErrorCard("Validation Impossible", "Impossible de valider un message de commit en erreur ou incomplet.",
					"Modifiez manuellement le message avec l'option [M]",
					"Ou relancez la génération avec une consigne avec [R]",
				)
				continue
			}

			finalMessage := formatCommit(proposal.Type, proposal.Scope, proposal.Subject, proposal.Body)
			spinnerCommit := ui.NewSpinner("Application du commit via Git...")
			spinnerCommit.Start()
			res, errCommit := client.ApplyCommit(finalMessage)
			spinnerCommit.Stop("")

			if errCommit != nil {
				ui.PrintErrorCard("Échec du Commit", fmt.Sprintf("Impossible d'appliquer le commit Git : %v", errCommit))
				waitForEnter(reader)
				return
			}

			ui.PrintSuccess(fmt.Sprintf("Commit créé avec succès ! [SHA: %s%s%s]", ui.Bold+ui.Green, res.Sha, ui.Reset))

			// Proposition de Push
			fmt.Printf("  %s👉 Voulez-vous pousser vos modifications vers le serveur distant (git push) ? [o/N] :%s ", ui.Bold+ui.Yellow, ui.Reset)
			pushAns, _ := reader.ReadString('\n')
			pushAns = strings.ToLower(strings.TrimSpace(pushAns))

			if pushAns == "o" || pushAns == "oui" {
				spinnerPush := ui.NewSpinner("Envoi des modifications en cours (git push)...")
				spinnerPush.Start()
				resPush, errPush := client.Push()
				spinnerPush.Stop("")

				if errPush != nil {
					ui.PrintErrorCard("Échec du Push", fmt.Sprintf("Impossible de pousser vers le dépôt distant : %v", errPush),
						"Vérifiez votre connexion réseau et vos droits d'accès sur le dépôt distant.",
					)
				} else {
					ui.PrintSuccess(resPush.Message)
				}
			}

			waitForEnter(reader)
			return

		case "R":
			fmt.Printf("\n  %s💬 Donnez votre consigne à l'IA pour ajuster le commit :%s\n  > ", ui.Bold+ui.Yellow, ui.Reset)
			fb, _ := reader.ReadString('\n')
			feedback := strings.TrimSpace(fb)
			if feedback == "" {
				feedback = "Corrige et améliore la précision du message."
			}
			proposal, isError = generateCommit(client, feedback)

		case "M":
			isError = promptManualEdit(reader, proposal)

		case "A":
			ui.PrintInfo("Opération abandonnée. Aucun commit n'a été créé.")
			waitForEnter(reader)
			return

		default:
			if isError {
				ui.PrintWarning("Choix non reconnu, veuillez choisir parmi R, M ou A.")
			} else {
				ui.PrintWarning("Choix non reconnu, veuillez choisir parmi V, R, M ou A.")
			}
		}
	}
}

// generateCommit encapsule l'appel au LLM avec spinner et gestion standardisée des erreurs
func generateCommit(client bridge.BackendClient, feedback string) (*models.CommitProposal, bool) {
	actionLabel := "Génération du message de commit avec le LLM..."
	if feedback != "" {
		actionLabel = fmt.Sprintf("Régénération avec consigne : \"%s\"...", feedback)
	}

	spinnerLLM := ui.NewSpinner(actionLabel)
	spinnerLLM.Start()
	proposal, errGen := client.GenerateCommit(feedback)
	spinnerLLM.Stop("")

	isError := (errGen != nil) || (proposal != nil && (!proposal.Success || proposal.Subject == "" || strings.ToLower(proposal.Type) == "erreur"))
	if isError {
		errMsg := "délai d'attente dépassé (timeout) ou aucune réponse du serveur LLM"
		if errGen != nil {
			errMsg = errGen.Error()
		} else if proposal != nil && proposal.Error != "" {
			errMsg = proposal.Error
		}

		ui.PrintErrorCard("Erreur de Génération du Commit", errMsg,
			"Vérifiez que le serveur Ollama est démarré et accessible (menu [3] Configuration)",
			"Réessayez avec une consigne [R] ou saisissez manuellement votre message [M]",
			"Vous pouvez aussi annuler l'opération et revenir au menu principal [A]",
		)

		if proposal == nil {
			proposal = &models.CommitProposal{
				Success: false,
				Type:    "feat",
				Scope:   "",
				Subject: "",
				Body:    "",
				Error:   errMsg,
			}
		}
	}

	return proposal, isError
}

// promptManualEdit guide l'utilisateur pour éditer manuellement les champs du message de commit
func promptManualEdit(reader *bufio.Reader, proposal *models.CommitProposal) bool {
	fmt.Println(ui.Bold + "\n  ✏️ Modification manuelle :" + ui.Reset)

	defaultType := proposal.Type
	if strings.ToLower(defaultType) == "erreur" {
		defaultType = "feat"
	}
	fmt.Printf("  Type [%s] : ", defaultType)
	newType, _ := reader.ReadString('\n')
	newType = strings.TrimSpace(newType)
	if newType != "" {
		proposal.Type = newType
	} else {
		proposal.Type = defaultType
	}

	fmt.Printf("  Scope [%s] : ", proposal.Scope)
	newScope, _ := reader.ReadString('\n')
	newScope = strings.TrimSpace(newScope)
	if newScope != "" {
		proposal.Scope = newScope
	}

	defaultSubject := proposal.Subject
	if strings.Contains(strings.ToLower(defaultSubject), "aucune réponse") || strings.Contains(strings.ToLower(defaultSubject), "erreur") {
		defaultSubject = ""
	}
	if defaultSubject != "" {
		fmt.Printf("  Sujet [%s] : ", defaultSubject)
	} else {
		fmt.Print("  Sujet : ")
	}
	newSubject, _ := reader.ReadString('\n')
	newSubject = strings.TrimSpace(newSubject)
	if newSubject != "" {
		proposal.Subject = newSubject
	} else if defaultSubject != "" {
		proposal.Subject = defaultSubject
	}

	fmt.Printf("  Description / Body [%s] : ", proposal.Body)
	newBody, _ := reader.ReadString('\n')
	newBody = strings.TrimSpace(newBody)
	if newBody != "" {
		proposal.Body = newBody
	}

	if proposal.Type != "erreur" && proposal.Subject != "" {
		proposal.Success = true
		return false
	}
	return true
}

func formatCommit(typeStr, scope, subject, body string) string {
	msg := typeStr
	if scope != "" {
		msg += "(" + scope + ")"
	}
	msg += ": " + subject
	if body != "" {
		msg += "\n\n" + body
	}
	return msg
}

func waitForEnter(reader *bufio.Reader) {
	fmt.Printf("\n  %s[Appuyez sur Entrée pour continuer...]%s", ui.Dim, ui.Reset)
	_, _ = reader.ReadString('\n')
}

// ReadChoice lit un choix utilisateur et absorbe les appuis sur Entrée à vide sur place sans saut de ligne
func ReadChoice(reader *bufio.Reader, prompt string) (string, error) {
	parts := strings.Split(prompt, "\n")
	prefix := ""
	lastLine := prompt
	if len(parts) > 1 {
		prefix = strings.Join(parts[:len(parts)-1], "\n") + "\n"
		lastLine = parts[len(parts)-1]
		fmt.Print(prefix)
	}
	fmt.Print(lastLine)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		trimmed := strings.ToUpper(strings.TrimSpace(input))
		if trimmed == "" {
			// L'utilisateur a appuyé sur Entrée à vide : on remonte le curseur et on réécrit la ligne du prompt sur place
			fmt.Print("\033[1A\r" + lastLine + "\033[K")
			continue
		}
		return trimmed, nil
	}
}

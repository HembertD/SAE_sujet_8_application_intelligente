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
		ui.PrintError(fmt.Sprintf("Échec de récupération du diff : %v", err))
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
				ui.PrintError(fmt.Sprintf("Erreur lors de l'indexation : %v", stageErr))
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

	// 4. Boucle de génération et de raffinement avec le LLM
	feedback := ""
	for {
		actionLabel := "Génération du message de commit avec le LLM..."
		if feedback != "" {
			actionLabel = fmt.Sprintf("Régénération avec consigne : \"%s\"...", feedback)
		}

		spinnerLLM := ui.NewSpinner(actionLabel)
		spinnerLLM.Start()
		proposal, errGen := client.GenerateCommit(feedback)
		spinnerLLM.Stop("")

		if errGen != nil {
			ui.PrintError(fmt.Sprintf("Erreur lors de la génération IA : %v", errGen))
			waitForEnter(reader)
			return
		}

		fmt.Println()
		ui.PrintCommitCard(proposal.Type, proposal.Scope, proposal.Subject, proposal.Body, width)
		fmt.Println()

		// 5. Menu de décision utilisateur
		fmt.Println(ui.Bold + "  Actions disponibles :" + ui.Reset)
		fmt.Printf("    %s[V]%s Valider et appliquer le commit\n", ui.Green+ui.Bold, ui.Reset)
		fmt.Printf("    %s[R]%s Refuser et commenter pour régénérer avec l'IA\n", ui.Yellow+ui.Bold, ui.Reset)
		fmt.Printf("    %s[M]%s Modifier manuellement le message avant de commiter\n", ui.Cyan+ui.Bold, ui.Reset)
		fmt.Printf("    %s[A]%s Annuler et revenir au menu principal\n\n", ui.Red+ui.Bold, ui.Reset)

		fmt.Printf("  %s👉 Votre choix (V/R/M/A) :%s ", ui.Bold+ui.Yellow, ui.Reset)
		choice, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "V":
			finalMessage := formatCommit(proposal.Type, proposal.Scope, proposal.Subject, proposal.Body)
			spinnerCommit := ui.NewSpinner("Application du commit via Git...")
			spinnerCommit.Start()
			res, errCommit := client.ApplyCommit(finalMessage)
			spinnerCommit.Stop("")

			if errCommit != nil {
				ui.PrintError(fmt.Sprintf("Échec du commit : %v", errCommit))
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
					ui.PrintError(fmt.Sprintf("Échec du push : %v", errPush))
				} else {
					ui.PrintSuccess(resPush.Message)
				}
			}

			waitForEnter(reader)
			return

		case "R":
			fmt.Printf("\n  %s💬 Donnez votre consigne à l'IA pour ajuster le commit :%s\n  > ", ui.Bold+ui.Yellow, ui.Reset)
			fb, _ := reader.ReadString('\n')
			feedback = strings.TrimSpace(fb)
			if feedback == "" {
				feedback = "Corrige et améliore la précision du message."
			}
			// Boucle à nouveau avec le feedback !

		case "M":
			fmt.Println(ui.Bold + "\n  ✏️ Modification manuelle :" + ui.Reset)
			fmt.Printf("  Type [%s] : ", proposal.Type)
			newType, _ := reader.ReadString('\n')
			newType = strings.TrimSpace(newType)
			if newType != "" {
				proposal.Type = newType
			}

			fmt.Printf("  Scope [%s] : ", proposal.Scope)
			newScope, _ := reader.ReadString('\n')
			newScope = strings.TrimSpace(newScope)
			if newScope != "" {
				proposal.Scope = newScope
			}

			fmt.Printf("  Sujet [%s] : ", proposal.Subject)
			newSubject, _ := reader.ReadString('\n')
			newSubject = strings.TrimSpace(newSubject)
			if newSubject != "" {
				proposal.Subject = newSubject
			}

			fmt.Printf("  Description / Body [%s] : ", proposal.Body)
			newBody, _ := reader.ReadString('\n')
			newBody = strings.TrimSpace(newBody)
			if newBody != "" {
				proposal.Body = newBody
			}

			feedback = "" // Reset feedback et réaffiche la carte avec les modifications

		case "A":
			ui.PrintInfo("Opération abandonnée. Aucun commit n'a été créé.")
			waitForEnter(reader)
			return

		default:
			ui.PrintWarning("Choix non reconnu, veuillez choisir parmi V, R, M ou A.")
		}
	}
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

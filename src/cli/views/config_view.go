package views

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/ui"
)

// ShowConfigFlow orchestre l'écran de visualisation et modification de la configuration
func ShowConfigFlow(reader *bufio.Reader, client bridge.BackendClient, status *models.RepoStatus) {
	for {
		ui.ClearTerminal()
		ui.PrintBanner(status)
		width := 72

		cfg, err := client.GetConfig()
		if err != nil {
			ui.PrintError(fmt.Sprintf("Impossible de charger la configuration : %v", err))
			waitForEnter(reader)
			return
		}

		var lines []string
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  • %sURL Serveur Ollama%s : %s%s%s", ui.Bold, ui.Reset, ui.Cyan, cfg.OllamaBaseURL, ui.Reset))
		lines = append(lines, fmt.Sprintf("  • %sModèle LLM%s        : %s%s%s", ui.Bold, ui.Reset, ui.Yellow, cfg.OllamaModel, ui.Reset))
		lines = append(lines, fmt.Sprintf("  • %sTimeout requête%s   : %.0fs", ui.Bold, ui.Reset, cfg.TimeoutS))
		lines = append(lines, fmt.Sprintf("  • %sLangue%s            : %s", ui.Bold, ui.Reset, cfg.Language))

		mockColor := ui.Green
		mockStatus := "Désactivé"
		if cfg.MockInterface {
			mockColor = ui.Yellow
			mockStatus = "Activé"
		}
		lines = append(lines, fmt.Sprintf("  • %sMode démo%s          : %s%s%s", ui.Bold, ui.Reset, mockColor, mockStatus, ui.Reset))
		lines = append(lines, "")
		lines = append(lines, ui.Dim+"  (Toute modification est répercutée dans le .env sans recompilation)"+ui.Reset)
		lines = append(lines, "")

		ui.PrintCard("Configuration Active", lines, width, ui.Yellow)
		fmt.Println()

		fmt.Println(ui.Bold + "  Actions disponibles :" + ui.Reset)
		fmt.Printf("    %s[T]%s 📡 Tester la connexion Ollama (Ping & modèles disponibles)\n", ui.Cyan+ui.Bold, ui.Reset)
		fmt.Printf("    %s[M]%s ✏️  Modifier les paramètres et sauvegarder dans le .env\n", ui.Green+ui.Bold, ui.Reset)
		fmt.Printf("    %s[R]%s 🚪 Retour au menu principal\n\n", ui.Red+ui.Bold, ui.Reset)

		fmt.Printf("  %s👉 Votre choix (T/M/R) :%s ", ui.Bold+ui.Yellow, ui.Reset)
		choice, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		choice = strings.ToUpper(strings.TrimSpace(choice))

		switch choice {
		case "T":
			spinner := ui.NewSpinner(fmt.Sprintf("Ping de %s...", cfg.OllamaBaseURL))
			spinner.Start()
			ping, pingErr := client.PingOllama()
			spinner.Stop("")

			if pingErr != nil || ping == nil || !ping.Reachable {
				errMsg := "délai d'attente dépassé (timeout) ou serveur injoignable"
				if pingErr != nil {
					errMsg = pingErr.Error()
				} else if ping != nil && ping.Error != "" {
					errMsg = ping.Error
				}
				ui.PrintError(fmt.Sprintf("Erreur : Serveur injoignable (%s)", errMsg))
			} else {
				ui.PrintSuccess(fmt.Sprintf("Connexion réussie ! Latence : %d ms", ping.LatencyMs))
				if len(ping.InstalledModels) > 0 {
					fmt.Printf("    %sModèles détectés sur le serveur :%s\n", ui.Bold, ui.Reset)
					for _, m := range ping.InstalledModels {
						fmt.Printf("     - %s%s%s\n", ui.Cyan, m, ui.Reset)
					}
					fmt.Println()
				}
			}
			waitForEnter(reader)

		case "M":
			fmt.Println(ui.Bold + "\n  ✏️ Modification des paramètres (laisser vide pour conserver) :" + ui.Reset)

			fmt.Printf("  URL Serveur [%s] : ", cfg.OllamaBaseURL)
			newURL, _ := reader.ReadString('\n')
			newURL = strings.TrimSpace(newURL)
			if newURL != "" {
				cfg.OllamaBaseURL = newURL
			}

			fmt.Printf("  Modèle [%s] : ", cfg.OllamaModel)
			newModel, _ := reader.ReadString('\n')
			newModel = strings.TrimSpace(newModel)
			if newModel != "" {
				cfg.OllamaModel = newModel
			}

			fmt.Printf("  Timeout (s) [%.0f] : ", cfg.TimeoutS)
			newTimeoutStr, _ := reader.ReadString('\n')
			newTimeoutStr = strings.TrimSpace(newTimeoutStr)
			if newTimeoutStr != "" {
				if t, errConv := strconv.ParseFloat(newTimeoutStr, 64); errConv == nil {
					cfg.TimeoutS = t
				}
			}

			fmt.Printf("  Langue (fr/en) [%s] : ", cfg.Language)
			newLang, _ := reader.ReadString('\n')
			newLang = strings.TrimSpace(newLang)
			if newLang != "" {
				cfg.Language = newLang
			}

			fmt.Printf("  Mode démo (true/false) [%t] : ", cfg.MockInterface)
			newMockStr, _ := reader.ReadString('\n')
			newMockStr = strings.TrimSpace(newMockStr)
			if newMockStr != "" {
				lower := strings.ToLower(newMockStr)
				cfg.MockInterface = (lower == "true" || lower == "1" || lower == "yes" || lower == "oui")
			}

			spinnerSave := ui.NewSpinner("Sauvegarde des paramètres dans le .env...")
			spinnerSave.Start()
			res, saveErr := client.SaveConfig(cfg.OllamaBaseURL, cfg.OllamaModel, cfg.TimeoutS, cfg.Language, cfg.MockInterface)
			spinnerSave.Stop("")

			if saveErr != nil {
				ui.PrintError(fmt.Sprintf("Échec de la sauvegarde : %v", saveErr))
			} else {
				ui.PrintSuccess(res.Message)
			}
			waitForEnter(reader)

		case "R", "":
			return

		default:
			ui.PrintWarning("Option non reconnue.")
			waitForEnter(reader)
		}
	}
}

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
func ShowConfigFlow(reader *bufio.Reader, client bridge.BackendClient, status *models.RepoStatus, isDev ...bool) {
	devMode := false
	if len(isDev) > 0 {
		devMode = isDev[0]
	}

	for {
		ui.ClearTerminal()
		ui.PrintBanner(status)
		width := 72

		cfg, err := client.GetConfig()
		if err != nil {
			ui.PrintErrorCard("Configuration Inaccessible", fmt.Sprintf("Impossible de charger la configuration : %v", err))
			waitForEnter(reader)
			return
		}

		var lines []string
		lines = append(lines, "")
		displayBaseURL := cfg.OllamaBaseURL
		if cfg.DockerLLM {
			displayBaseURL = bridge.DockerOllamaBaseURL
		}
		lines = append(lines, fmt.Sprintf("  • %sURL Serveur Ollama%s : %s%s%s", ui.Bold, ui.Reset, ui.Cyan, displayBaseURL, ui.Reset))
		lines = append(lines, fmt.Sprintf("  • %sModèle LLM%s        : %s%s%s", ui.Bold, ui.Reset, ui.Yellow, cfg.OllamaModel, ui.Reset))
		lines = append(lines, fmt.Sprintf("  • %sTimeout requête%s   : %.0fs", ui.Bold, ui.Reset, cfg.TimeoutS))

		dockerLLMColor := ui.Green
		dockerLLMStatus := "Désactivé"
		if cfg.DockerLLM {
			dockerLLMColor = ui.Yellow
			dockerLLMStatus = "Activé"
		}
		lines = append(lines, fmt.Sprintf("  • %sLLM via Docker%s      : %s%s%s", ui.Bold, ui.Reset, dockerLLMColor, dockerLLMStatus, ui.Reset))

		notifColor := ui.Green
		notifStatus := "Désactivé"
		if cfg.EnableNotifications {
			notifColor = ui.Yellow
			notifStatus = "Activé"
		}
		lines = append(lines, fmt.Sprintf("  • %sNotifications IA%s  : %s%s%s", ui.Bold, ui.Reset, notifColor, notifStatus, ui.Reset))

		if devMode {
			mockInterfaceColor := ui.Green
			mockInterfaceStatus := "Désactivé"
			if cfg.MockInterface {
				mockInterfaceColor = ui.Yellow
				mockInterfaceStatus = "Activé"
			}
			lines = append(lines, fmt.Sprintf("  • %sMock Interface%s    : %s%s%s", ui.Bold, ui.Reset, mockInterfaceColor, mockInterfaceStatus, ui.Reset))

			mockAIColor := ui.Green
			mockAIStatus := "Désactivé"
			if cfg.MockAI {
				mockAIColor = ui.Yellow
				mockAIStatus = "Activé"
			}
			lines = append(lines, fmt.Sprintf("  • %sMock IA%s           : %s%s%s", ui.Bold, ui.Reset, mockAIColor, mockAIStatus, ui.Reset))
			lines = append(lines, "")
			lines = append(lines, ui.Green+"  🛠️  Mode Développeur actif (--dev)"+ui.Reset)
		}
		lines = append(lines, "")

		ui.PrintCard("Configuration Active", lines, width, ui.Yellow)
		fmt.Println()

		fmt.Println(ui.Bold + "  Actions disponibles :" + ui.Reset)
		fmt.Printf("    %s[T]%s 📡 Tester la connexion Ollama (Ping & modèles disponibles)\n", ui.Cyan+ui.Bold, ui.Reset)
		fmt.Printf("    %s[M]%s ✏️  Modifier les paramètres et sauvegarder dans le .env\n", ui.Green+ui.Bold, ui.Reset)
		fmt.Printf("    %s[R]%s 🚪 Retour au menu principal\n\n", ui.Red+ui.Bold, ui.Reset)

		prompt := fmt.Sprintf("  %s👉 Votre choix (T/M/R) :%s ", ui.Bold+ui.Yellow, ui.Reset)
		choice, err := ReadChoice(reader, prompt)
		if err != nil {
			return
		}

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
				ui.PrintErrorCard("Diagnostic Connexion Ollama", fmt.Sprintf("Le serveur Ollama à l'adresse %s est injoignable.", cfg.OllamaBaseURL),
					fmt.Sprintf("Détail technique : %s", errMsg),
					"Vérifiez l'URL du serveur et le port avec l'option [M] Modifier",
					"Assurez-vous qu'Ollama est bien démarré sur la machine hôte",
				)
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

			if cfg.DockerLLM {
				fmt.Printf("  %s(LLM via Docker actif : le backend utilisera %s)%s\n", ui.Dim, bridge.DockerOllamaBaseURL, ui.Reset)
			}
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

			fmt.Printf("  Notifications IA (true/false) [%t] : ", cfg.EnableNotifications)
			newNotifStr, _ := reader.ReadString('\n')
			newNotifStr = strings.TrimSpace(newNotifStr)
			if newNotifStr != "" {
				lower := strings.ToLower(newNotifStr)
				cfg.EnableNotifications = (lower == "true" || lower == "1" || lower == "yes" || lower == "oui")
				ui.SetNotificationsEnabled(cfg.EnableNotifications)
				_ = bridge.WriteNotificationsSetting("", cfg.EnableNotifications)
			}

			if devMode {
				fmt.Printf("  Mock Interface (true/false) [%t] : ", cfg.MockInterface)
				newMockInterfaceStr, _ := reader.ReadString('\n')
				newMockInterfaceStr = strings.TrimSpace(newMockInterfaceStr)
				if newMockInterfaceStr != "" {
					lower := strings.ToLower(newMockInterfaceStr)
					cfg.MockInterface = (lower == "true" || lower == "1" || lower == "yes" || lower == "oui")
				}

				fmt.Printf("  Mock IA (true/false) [%t] : ", cfg.MockAI)
				newMockAIStr, _ := reader.ReadString('\n')
				newMockAIStr = strings.TrimSpace(newMockAIStr)
				if newMockAIStr != "" {
					lower := strings.ToLower(newMockAIStr)
					cfg.MockAI = (lower == "true" || lower == "1" || lower == "yes" || lower == "oui")
				}
			}

			fmt.Printf("  LLM via Docker (true/false) [%t] : ", cfg.DockerLLM)
			newDockerLLMStr, _ := reader.ReadString('\n')
			newDockerLLMStr = strings.TrimSpace(newDockerLLMStr)
			if newDockerLLMStr != "" {
				lower := strings.ToLower(newDockerLLMStr)
				cfg.DockerLLM = (lower == "true" || lower == "1" || lower == "yes" || lower == "oui")
			}

			spinnerSave := ui.NewSpinner("Sauvegarde des paramètres dans le .env...")
			spinnerSave.Start()
			res, saveErr := client.SaveConfigWithDockerLLM(cfg.OllamaBaseURL, cfg.OllamaModel, cfg.TimeoutS, cfg.MockInterface, cfg.DockerLLM, cfg.MockAI)
			spinnerSave.Stop("")

			if saveErr != nil {
				ui.PrintErrorCard("Sauvegarde Impossible", fmt.Sprintf("Échec de la sauvegarde dans le fichier .env : %v", saveErr))
			} else {
				ui.PrintSuccess(res.Message)

				// Gestion du conteneur Ollama Docker selon l'option USE_DOCKER_LLM
				if cfg.DockerLLM {
					fmt.Println()
					fmt.Print(ui.Bold + "  🐳 Le serveur Ollama via Docker est activé." + ui.Reset)
					fmt.Printf("  %s👉 Démarrer le conteneur maintenant ? (o/N) :%s ", ui.Bold+ui.Yellow, ui.Reset)
					startAns, _ := reader.ReadString('\n')
					startAns = strings.ToLower(strings.TrimSpace(startAns))
					if startAns == "o" || startAns == "oui" {
						spinnerDocker := ui.NewSpinner("Lancement du conteneur Ollama via docker compose...")
						spinnerDocker.Start()
						resDocker, dockerErr := client.StartDockerOllama()
						spinnerDocker.Stop("")
						if dockerErr != nil || !resDocker.Success {
							ui.PrintErrorCard("Erreur Docker", fmt.Sprintf("Impossible de démarrer le conteneur Ollama : %v", dockerErr))
						} else {
							ui.PrintSuccess(resDocker.Message)
						}
					}
				} else {
					fmt.Println()
					fmt.Print(ui.Bold + "  🐳 Le serveur Ollama via Docker est désactivé." + ui.Reset)
					fmt.Printf("  %s👉 Arrêter le conteneur Ollama si actif ? (o/N) :%s ", ui.Bold+ui.Yellow, ui.Reset)
					stopAns, _ := reader.ReadString('\n')
					stopAns = strings.ToLower(strings.TrimSpace(stopAns))
					if stopAns == "o" || stopAns == "oui" {
						spinnerDocker := ui.NewSpinner("Arrêt du conteneur Ollama via docker compose...")
						spinnerDocker.Start()
						resDocker, dockerErr := client.StopDockerOllama()
						spinnerDocker.Stop("")
						if dockerErr != nil || !resDocker.Success {
							ui.PrintErrorCard("Erreur Docker", fmt.Sprintf("Impossible d'arrêter le conteneur Ollama : %v", dockerErr))
						} else {
							ui.PrintSuccess(resDocker.Message)
						}
					}
				}
			}
			waitForEnter(reader)

		case "R":
			return

		default:
			ui.PrintWarning("Option non reconnue.")
			waitForEnter(reader)
		}
	}
}

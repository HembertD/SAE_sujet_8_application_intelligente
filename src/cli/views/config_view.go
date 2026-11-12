package views

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		dockerStatus, dockerStatusErr := client.DockerOllamaStatus()

		var lines []string
		lines = append(lines, "")
		displayBaseURL := bridge.EffectiveOllamaBaseURL(cfg)
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

		dockerContainerColor := ui.Dim
		dockerContainerStatus := "Arrêté"
		if dockerStatusErr != nil || dockerStatus == nil || !dockerStatus.Success {
			dockerContainerColor = ui.Red
			dockerContainerStatus = "Indisponible"
		} else if dockerStatus.Running {
			dockerContainerColor = ui.Green
			dockerContainerStatus = "En cours"
		}
		lines = append(lines, fmt.Sprintf("  • %sConteneur Docker%s   : %s%s%s", ui.Bold, ui.Reset, dockerContainerColor, dockerContainerStatus, ui.Reset))

		notifColor := ui.Green
		notifStatus := "Désactivé"
		if cfg.EnableNotifications {
			notifColor = ui.Yellow
			notifStatus = "Activé"
		}
		lines = append(lines, fmt.Sprintf("  • %sNotifications IA%s  : %s%s%s", ui.Bold, ui.Reset, notifColor, notifStatus, ui.Reset))

		pyInterpDisplay := "Non détecté"
		if cfg.PythonBin != "" {
			pyInterpDisplay = fmt.Sprintf("%s (PYTHON_BIN)", cfg.PythonBin)
		} else if cfg.ResolvedPythonBin != "" {
			appRoot := bridge.FindAppRoot()
			displayPath := cfg.ResolvedPythonBin
			if strings.HasPrefix(displayPath, appRoot) {
				if rel, errRel := filepath.Rel(appRoot, displayPath); errRel == nil {
					displayPath = "./" + rel
				}
			}
			pyInterpDisplay = fmt.Sprintf("%s (auto .venv)", displayPath)
		}
		lines = append(lines, fmt.Sprintf("  • %sInterpréteur Python%s: %s%s%s", ui.Bold, ui.Reset, ui.Cyan, pyInterpDisplay, ui.Reset))

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
		fmt.Printf("    %s[P]%s 🐍 Vérification de l'environnement Python (dépendances & intégrité)\n", ui.Yellow+ui.Bold, ui.Reset)
		fmt.Printf("    %s[M]%s ✏️  Modifier les paramètres et sauvegarder dans le .env\n", ui.Green+ui.Bold, ui.Reset)
		fmt.Printf("    %s[D]%s 🐳 Activer/Désactiver LLM Docker (démarre/arrête le conteneur)\n", ui.Magenta+ui.Bold, ui.Reset)
		fmt.Printf("    %s[R]%s 🚪 Retour au menu principal\n\n", ui.Red+ui.Bold, ui.Reset)

		prompt := fmt.Sprintf("  %s👉 Votre choix (T/P/M/D/R) :%s ", ui.Bold+ui.Yellow, ui.Reset)
		choice, err := ReadChoice(reader, prompt)
		if err != nil {
			return
		}

		switch choice {
		case "T":
			spinner := ui.NewSpinner(fmt.Sprintf("Ping de %s...", displayBaseURL))
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
				ui.PrintErrorCard("Diagnostic Connexion Ollama", fmt.Sprintf("Le serveur Ollama à l'adresse %s est injoignable.", displayBaseURL),
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

		case "P":
			spinner := ui.NewSpinner("Vérification approfondie de l'environnement Python et des dépendances...")
			spinner.Start()
			report, errVerify := bridge.VerifyPythonEnvironment("")
			spinner.Stop("")

			if errVerify != nil || (report != nil && !report.Success) {
				mainMsg := "Environnement Python invalide."
				var hints []string
				if errVerify != nil {
					mainMsg = fmt.Sprintf("Environnement Python invalide : %s", errVerify.Error())
				} else if report != nil {
					if len(report.MissingDeps) > 0 {
						mainMsg = fmt.Sprintf("Environnement Python invalide : dépendance(s) manquante(s) : %s", strings.Join(report.MissingDeps, ", "))
					} else if report.AppLoadError != "" {
						mainMsg = fmt.Sprintf("Environnement Python invalide : %s", report.AppLoadError)
					}
				}
				hints = append(hints, "Pour installer les dépendances requises : python scripts/setup_python_env.py")
				hints = append(hints, "Pour modifier l'interpréteur ciblé : option [M] Modifier ou définir PYTHON_BIN dans le .env")
				ui.PrintErrorCard("Environnement Python Invalide", mainMsg, hints...)
			} else {
				ui.PrintSuccess(fmt.Sprintf("Environnement Python opérationnel et complet (Python %s)", report.Version))
				fmt.Printf("    %sBinaire exécutable :%s %s%s%s\n", ui.Bold, ui.Reset, ui.Cyan, report.Executable, ui.Reset)
				fmt.Printf("    %sMode d'attribution :%s %s%s%s\n", ui.Bold, ui.Reset, ui.Yellow, report.Mode, ui.Reset)
				fmt.Printf("    %sDépendances requises :%s %sToutes installées (GitPython, PyYAML)%s\n", ui.Bold, ui.Reset, ui.Green, ui.Reset)
				fmt.Printf("    %sChargement du backend :%s %sSuccès (git_commit_release_notes_generator)%s\n\n", ui.Bold, ui.Reset, ui.Green, ui.Reset)
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

			currentPyBinDisplay := cfg.PythonBin
			if currentPyBinDisplay == "" {
				currentPyBinDisplay = "auto"
			}
			fmt.Printf("  Interpréteur Python (PYTHON_BIN, 'auto' pour .venv) [%s] : ", currentPyBinDisplay)
			newPyBin, _ := reader.ReadString('\n')
			newPyBin = strings.TrimSpace(newPyBin)
			pyBinUpdated := false
			if newPyBin != "" {
				isAuto := strings.ToLower(newPyBin) == "auto" || strings.ToLower(newPyBin) == "clear" || strings.ToLower(newPyBin) == "vide"
				candidateBin := newPyBin
				if isAuto {
					candidateBin = ""
				}

				// Tolérance utilisateur : si le script d'activation 'activate' a été saisi
				if !isAuto && filepath.Base(candidateBin) == "activate" {
					companion := filepath.Join(filepath.Dir(candidateBin), "python")
					if runtime.GOOS == "windows" {
						companion = filepath.Join(filepath.Dir(candidateBin), "python.exe")
					}
					if fi, errStat := os.Stat(companion); errStat == nil && !fi.IsDir() {
						candidateBin = companion
					}
				}

				// Vérification préalable systématique avant acceptation de la modification
				spinnerCheck := ui.NewSpinner("Vérification de l'interpréteur Python candidat...")
				spinnerCheck.Start()
				report, errCheck := bridge.VerifyPythonCandidate("", candidateBin, isAuto)
				spinnerCheck.Stop("")

				if errCheck != nil || (report != nil && !report.Success) {
					mainMsg := "Environnement Python invalide."
					if errCheck != nil {
						mainMsg = fmt.Sprintf("Environnement Python invalide : %s", errCheck.Error())
					} else if report != nil && len(report.MissingDeps) > 0 {
						mainMsg = fmt.Sprintf("Environnement Python invalide : dépendance(s) manquante(s) : %s", strings.Join(report.MissingDeps, ", "))
					} else if report != nil && report.AppLoadError != "" {
						mainMsg = fmt.Sprintf("Environnement Python invalide : %s", report.AppLoadError)
					}
					ui.PrintErrorCard("Modification Python Refusée", mainMsg,
						"L'interpréteur candidat ne respecte pas les prérequis applicatifs.",
						"La modification est annulée : l'ancienne configuration Python est conservée sans modification.",
					)
				} else {
					ui.PrintSuccess(fmt.Sprintf("Interpréteur candidat validé (Python %s) : prérequis satisfaits.", report.Version))
					if isAuto {
						if cfg.PythonBin != "" {
							cfg.PythonBin = ""
							pyBinUpdated = true
						}
					} else if candidateBin != cfg.PythonBin {
						cfg.PythonBin = candidateBin
						pyBinUpdated = true
					}
				}
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
			if saveErr == nil && pyBinUpdated {
				_ = bridge.WritePythonBinSetting("", cfg.PythonBin)
			}
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
						if dockerErr != nil || resDocker == nil || !resDocker.Success {
							detail := fmt.Sprintf("%v", dockerErr)
							if resDocker != nil && resDocker.Error != "" {
								detail = resDocker.Error
							}
							ui.PrintErrorCard("Erreur Docker", fmt.Sprintf("Impossible de démarrer le conteneur Ollama : %s", detail))
						} else {
							ui.PrintSuccess(resDocker.Message)
							ui.PrintInfo(fmt.Sprintf("Premier lancement : le téléchargement du modèle %s peut prendre plusieurs minutes. Suivez la progression avec : docker compose logs -f ollama", cfg.OllamaModel))
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
						if dockerErr != nil || resDocker == nil || !resDocker.Success {
							detail := fmt.Sprintf("%v", dockerErr)
							if resDocker != nil && resDocker.Error != "" {
								detail = resDocker.Error
							}
							ui.PrintErrorCard("Erreur Docker", fmt.Sprintf("Impossible d'arrêter le conteneur Ollama : %s", detail))
						} else {
							ui.PrintSuccess(resDocker.Message)
						}
					}
				}
			}
			waitForEnter(reader)

		case "D":
			toggleDockerLLM(reader, client, cfg)
			waitForEnter(reader)

		case "R":
			return

		default:
			ui.PrintWarning("Option non reconnue.")
			waitForEnter(reader)
		}
	}
}

// toggleDockerLLM active ou désactive le mode Docker et synchronise l'état du
// conteneur Ollama sans passer par l'écran de modification [M].
func toggleDockerLLM(reader *bufio.Reader, client bridge.BackendClient, cfg *models.ConfigResponse) {
	targetEnabled := !cfg.DockerLLM

	fmt.Println()
	if targetEnabled {
		fmt.Print(ui.Bold + "  🐳 Activer LLM via Docker et démarrer le conteneur Ollama ? (o/N) : " + ui.Reset)
	} else {
		fmt.Print(ui.Bold + "  🐳 Désactiver LLM via Docker et arrêter le conteneur Ollama ? (o/N) : " + ui.Reset)
	}
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "o" && answer != "oui" {
		return
	}

	status, statusErr := client.DockerOllamaStatus()
	statusKnown := statusErr == nil && status != nil && status.Success

	if targetEnabled {
		if !statusKnown {
			printDockerUnavailable(status, statusErr)
			return
		}
		if !status.Running {
			spinner := ui.NewSpinner("Lancement du conteneur Ollama via docker compose...")
			spinner.Start()
			res, dockerErr := client.StartDockerOllama()
			spinner.Stop("")
			if dockerErr != nil || res == nil || !res.Success {
				ui.PrintErrorCard("Erreur Docker", fmt.Sprintf("Impossible de démarrer le conteneur Ollama : %s", dockerErrorDetail(res, dockerErr)))
				return
			}
			ui.PrintSuccess(res.Message)
		}
	} else {
		if statusKnown && status.Running {
			spinner := ui.NewSpinner("Arrêt du conteneur Ollama via docker compose...")
			spinner.Start()
			res, dockerErr := client.StopDockerOllama()
			spinner.Stop("")
			if dockerErr != nil || res == nil || !res.Success {
				ui.PrintErrorCard("Erreur Docker", fmt.Sprintf("Impossible d'arrêter le conteneur Ollama : %s", dockerErrorDetail(res, dockerErr)))
				ui.PrintWarning("Le mode LLM via Docker reste activé car le conteneur n'a pas pu être arrêté.")
				return
			}
			ui.PrintSuccess(res.Message)
		} else if !statusKnown {
			ui.PrintWarning("État du conteneur Docker indisponible : désactivation du mode sans arrêt du conteneur.")
		}
	}

	resSave, saveErr := client.SaveConfigWithDockerLLM(cfg.OllamaBaseURL, cfg.OllamaModel, cfg.TimeoutS, cfg.MockInterface, targetEnabled, cfg.MockAI)
	if saveErr != nil || resSave == nil {
		detail := "réponse vide du backend"
		if saveErr != nil {
			detail = saveErr.Error()
		}
		ui.PrintErrorCard("Sauvegarde Impossible", fmt.Sprintf("Le conteneur a été mis à jour mais la configuration n'a pas pu être sauvegardée : %s", detail))
		return
	}
	ui.PrintSuccess(resSave.Message)
	if targetEnabled {
		ui.PrintInfo(fmt.Sprintf("Premier lancement : le téléchargement du modèle %s peut prendre plusieurs minutes. Suivez la progression avec : docker compose logs -f ollama", cfg.OllamaModel))
	}
}

// printDockerUnavailable affiche un diagnostic lorsque l'état du conteneur ne
// peut pas être interrogé (Docker arrêté, binaire absent, compose manquant...).
func printDockerUnavailable(status *models.DockerStatusResponse, statusErr error) {
	detail := "Docker est peut-être arrêté ou non installé."
	if statusErr != nil {
		detail = statusErr.Error()
	} else if status != nil && status.Error != "" {
		detail = status.Error
	}
	ui.PrintErrorCard("Docker Indisponible",
		"Impossible de vérifier l'état du conteneur Ollama.",
		fmt.Sprintf("Détail technique : %s", detail),
		"Vérifiez que Docker est installé et démarré : docker info",
		"Vérifiez la présence de docker-compose.yml à la racine du projet",
	)
}

// dockerErrorDetail extrait le message d'erreur le plus utile d'une action Docker.
func dockerErrorDetail(res *models.ActionResult, err error) string {
	if res != nil && res.Error != "" {
		return res.Error
	}
	if err != nil {
		return err.Error()
	}
	return "erreur inconnue"
}

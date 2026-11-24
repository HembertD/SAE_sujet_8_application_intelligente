package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"sae-git-cli/bridge"
	"sae-git-cli/flags"
	"sae-git-cli/ui"
	"sae-git-cli/views"
)

func main() {
	// 1. Interception prioritaire et immédiate de l'aide et de la version.
	// Cela garantit un code de retour 0 sans charger Python, ni TUI, ni Git.
	if ui.IsVersionRequested(os.Args[1:]) {
		ui.PrintVersion(os.Stdout)
		os.Exit(0)
	}
	if ui.IsHelpRequested(os.Args[1:]) {
		ui.PrintHelp(os.Stdout)
		os.Exit(0)
	}

	// 2. Options de la ligne de commande standard
	flag.Usage = func() {
		ui.PrintHelp(os.Stdout)
	}

	devFlag := flags.Dev
	mockInterfaceFlag := flags.MockInterface
	mockIAFlag := flags.MockIA
	dockerLLMFlag := flags.DockerLLM
	repoFlag := flags.Repo
	appFlag := flags.AppDir
	pythonBinFlag := flags.PythonBin
	flag.Parse()

	// Si un interpréteur Python est explicitement passé en argument CLI, il surcharge l'environnement
	if *pythonBinFlag != "" {
		_ = os.Setenv("PYTHON_BIN", *pythonBinFlag)
	}

	// Gestion rigoureuse des signaux OS pour restaurer le terminal (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print(ui.ShowCursor + ui.Reset + "\n\n  👋 Arrêt de Gitelligence. À bientôt !\n\n")
		os.Exit(0)
	}()
	defer fmt.Print(ui.ShowCursor + ui.Reset + "\n")

	// En CLI, les flags de mock nécessitent impérativement le mode --dev
	if !*devFlag && (*mockIAFlag || *mockInterfaceFlag) {
		flagName := "--mockIA"
		if *mockInterfaceFlag && *mockIAFlag {
			flagName = "--mockIA et --mockInterface"
		} else if *mockInterfaceFlag {
			flagName = "--mockInterface"
		}
		ui.PrintErrorCard("Mode Développeur Requis",
			fmt.Sprintf("L'utilisation de %s nécessite obligatoirement le flag --dev.", flagName),
			fmt.Sprintf("Exemple d'exécution : ./git-generator --dev %s", flagName),
		)
		os.Exit(1)
	}

	forceMockInterface := *mockInterfaceFlag
	forceMockIA := *mockIAFlag
	forceDockerLLM := *dockerLLMFlag

	// Détermination du répertoire racine du dépôt cible à analyser
	absRepoPath, err := filepath.Abs(*repoFlag)
	if err != nil {
		absRepoPath = "."
	}

	// Détermination de la racine de l'application (où réside le .env applicatif)
	appRoot := bridge.FindAppRoot(*appFlag)

	// Initialisation des notifications de bureau selon la configuration .env
	ui.SetNotificationsEnabled(bridge.ReadEnableNotificationsFromEnv(appRoot))

	// Vérification automatique et approfondie de l'environnement Python à chaque lancement.
	// Inutile en mode mock-interface : Python n'y est jamais appelé.
	mockInterfaceActive := forceMockInterface || bridge.ReadMockSettingFromEnv(appRoot)
	if !mockInterfaceActive {
		spinnerPy := ui.NewSpinner("Vérification de l'environnement Python et des dépendances...")
		spinnerPy.Start()
		report, pyErr := bridge.VerifyPythonEnvironment(appRoot)
		spinnerPy.Stop("")

		if pyErr != nil || (report != nil && !report.Success) {
			mainMsg := "Environnement Python invalide."
			var hints []string
			if pyErr != nil {
				mainMsg = fmt.Sprintf("Environnement Python invalide : %s", pyErr.Error())
			} else if report != nil {
				if len(report.MissingDeps) > 0 {
					mainMsg = fmt.Sprintf("Environnement Python invalide : dépendance(s) manquante(s) : %s", strings.Join(report.MissingDeps, ", "))
				} else if report.AppLoadError != "" {
					mainMsg = fmt.Sprintf("Environnement Python invalide : %s", report.AppLoadError)
				}
			}
			hints = append(hints, "Pour initialiser l'environnement et installer les dépendances : python scripts/setup_python_env.py")
			hints = append(hints, "Ou définissez PYTHON_BIN=/chemin/vers/votre/python dans votre .env")

			ui.PrintErrorCard("Environnement Python Invalide", mainMsg, hints...)
			os.Exit(1)
		}

		ui.PrintSuccess(fmt.Sprintf("Environnement Python validé (%s, Python %s) : dépendances et backend opérationnels.", report.Mode, report.Version))
	}

	// Initialisation du client de pont (le .env reste prioritaire pour activer le mock même sans flag CLI)
	client := bridge.NewBackendClientWithAppRoot(absRepoPath, appRoot, forceMockInterface, forceMockIA, forceDockerLLM)
	reader := bufio.NewReader(os.Stdin)

	// Boucle principale du TUI
	for {
		// Récupération de l'état du dépôt auprès du backend Python
		status, errStatus := client.GetRepoStatus()
		if errStatus != nil {
			status = nil
		}

		choice := views.ShowMainMenu(reader, client, status)

		switch choice {
		case "1":
			views.ShowCommitFlow(reader, client, status)
		case "2":
			views.ShowReleaseNotesFlow(reader, client, status)
		case "3":
			views.ShowConfigFlow(reader, client, status, *devFlag)
		case "4", "q", "Q":
			ui.ClearTerminal()
			fmt.Printf("\n  %s👋 Merci d'avoir utilisé Gitelligence ! À bientôt.%s\n\n", ui.Bold+ui.Cyan, ui.Reset)
			return
		default:
			// Option non reconnue, retour direct au menu
		}
	}
}

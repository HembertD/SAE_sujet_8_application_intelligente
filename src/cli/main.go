package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"sae-git-cli/bridge"
	"sae-git-cli/ui"
	"sae-git-cli/views"
)

func main() {
	// Options de la ligne de commande
	devFlag := flag.Bool("dev", false, "Active le mode développement (débloque les options de mock et leur modification)")
	mockInterfaceFlag := flag.Bool("mockInterface", false, "Force la simulation complète de l'interface (mode démo sans Git ni Python, requiert --dev)")
	mockIAFlag := flag.Bool("mockIA", false, "Force la simulation de l'IA hors-ligne sans serveur Ollama (requiert --dev)")
	dockerLLMFlag := flag.Bool("docker-llm", false, "Force l'usage du serveur Ollama local via Docker")
	repoFlag := flag.String("repo", ".", "Chemin du dépôt Git à analyser")
	appFlag := flag.String("app-dir", "", "Chemin vers l'application Gitelligence (dossier contenant le .env applicatif)")
	flag.Parse()

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

	// Échec immédiat et explicite si aucun environnement Python valide n'est
	// configuré (ni PYTHON_BIN, ni .venv/ créé par scripts/setup_python_env.py).
	// Inutile en mode mock-interface : Python n'y est jamais appelé.
	mockInterfaceActive := forceMockInterface || bridge.ReadMockSettingFromEnv(appRoot)
	if !mockInterfaceActive {
		if pyErr := bridge.CheckPythonEnvironment(appRoot); pyErr != nil {
			ui.PrintErrorCard("Environnement Python non initialisé", pyErr.Error())
			os.Exit(1)
		}
	}

	// Initialisation du client de pont (le .env reste prioritaire pour activer le mock même sans flag CLI)
	client := bridge.NewBackendClientWithAppRoot(absRepoPath, appRoot, forceMockInterface, forceMockIA, forceDockerLLM)
	reader := bufio.NewReader(os.Stdin)

	// Boucle principale du TUI
	for {
		// Récupération de l'état du dépôt auprès du backend Python
		status, errStatus := client.GetRepoStatus()
		if errStatus != nil {
			// En cas d'erreur inattendue, on crée un statut dégradé
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

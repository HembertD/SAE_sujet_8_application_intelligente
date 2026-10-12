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

// Point d'entrée initialisant l'interface et pilotant le menu principal.
func main() {
	mockInterfaceFlag := flag.Bool("mockInterface", false, "Force la simulation complète de l'interface (mode démo sans Git ni Python)")
	mockIAFlag := flag.Bool("mockIA", false, "Force la simulation de l'IA hors-ligne sans serveur Ollama")
	repoFlag := flag.String("repo", ".", "Chemin du dépôt Git à analyser")
	appFlag := flag.String("app-dir", "", "Chemin vers l'application Smart Commit (dossier contenant le .env applicatif)")
	flag.Parse()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print(ui.ShowCursor + ui.Reset + "\n\n  👋 Arrêt de l'application. À bientôt !\n\n")
		os.Exit(0)
	}()
	defer fmt.Print(ui.ShowCursor + ui.Reset + "\n")

	absRepoPath, err := filepath.Abs(*repoFlag)
	if err != nil {
		absRepoPath = "."
	}

	appRoot := bridge.FindAppRoot(*appFlag)

	client := bridge.NewBackendClientWithAppRoot(absRepoPath, appRoot, *mockInterfaceFlag, *mockIAFlag)
	reader := bufio.NewReader(os.Stdin)

	for {
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
			views.ShowConfigFlow(reader, client, status)
		case "4", "q", "Q":
			ui.ClearTerminal()
			fmt.Printf("\n  %s👋 Merci d'avoir utilisé Smart Commit Generator ! À bientôt.%s\n\n", ui.Bold+ui.Cyan, ui.Reset)
			return
		default:
		}
	}
}

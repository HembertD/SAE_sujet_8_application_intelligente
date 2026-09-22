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
	demoFlag := flag.Bool("demo", false, "Force le mode simulation/démo pour tester le TUI de manière autonome")
	repoFlag := flag.String("repo", ".", "Chemin du dépôt Git à analyser")
	flag.Parse()

	// Gestion rigoureuse des signaux OS pour restaurer le terminal (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print(ui.ShowCursor + ui.Reset + "\n\n  👋 Arrêt de l'application. À bientôt !\n\n")
		os.Exit(0)
	}()
	defer fmt.Print(ui.ShowCursor + ui.Reset + "\n")

	// Détermination du répertoire racine du dépôt
	absRepoPath, err := filepath.Abs(*repoFlag)
	if err != nil {
		absRepoPath = "."
	}

	// Initialisation du client de pont (aucune logique métier en Go)
	client := bridge.NewBackendClient(absRepoPath, *demoFlag)
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
			views.ShowConfigFlow(reader, client, status)
		case "4", "q", "Q":
			ui.ClearTerminal()
			fmt.Printf("\n  %s👋 Merci d'avoir utilisé Smart Commit Generator ! À bientôt.%s\n\n", ui.Bold+ui.Cyan, ui.Reset)
			return
		default:
			// Option non reconnue, retour direct au menu
		}
	}
}

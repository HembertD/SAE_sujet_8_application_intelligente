package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"

	"sae-git-cli/bridge"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Options de la ligne de commande
	devFlag := flag.Bool("dev", false, "Active le mode développement (débloque les options de mock et de simulation)")
	mockInterfaceFlag := flag.Bool("mockInterface", false, "Active le mockage complet de l'interface (mode autonome sans Git ni Python, requiert --dev)")
	mockIAFlag := flag.Bool("mockIA", false, "Active le mockage de l'IA (sans serveur Ollama, requiert --dev)")
	dockerLLMFlag := flag.Bool("docker-llm", false, "Force l'utilisation du conteneur Ollama local piloté par Docker Compose")
	repoFlag := flag.String("repo", ".", "Chemin du dépôt Git à analyser")
	appFlag := flag.String("app-dir", "", "Chemin vers l'application Gitelligence (dossier contenant le .env applicatif)")
	pythonBinFlag := flag.String("python-bin", "", "Chemin explicite vers l'interpréteur Python à utiliser (surcharge le .env et .venv)")
	flag.Parse()

	// Si un interpréteur Python est explicitement passé en argument, il surcharge l'environnement
	if *pythonBinFlag != "" {
		_ = os.Setenv("PYTHON_BIN", *pythonBinFlag)
	}

	// Contrôle de cohérence : les flags de mock nécessitent impérativement le mode --dev
	if (*mockInterfaceFlag || *mockIAFlag) && !*devFlag {
		fmt.Fprintf(os.Stderr, "Erreur : les options --mockInterface et --mockIA requièrent impérativement le flag --dev.\n")
		os.Exit(1)
	}

	repoRoot := *repoFlag
	appRoot := bridge.FindAppRoot(*appFlag)

	// Vérification de l'environnement Python au lancement (exactement comme le CLI)
	mockInterfaceActive := *mockInterfaceFlag || bridge.ReadMockSettingFromEnv(appRoot)
	if !mockInterfaceActive {
		report, pyErr := bridge.VerifyPythonEnvironment(appRoot)
		if pyErr != nil || (report != nil && !report.Success) {
			if pyErr != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Attention (démarrage GUI) : environnement Python non validé : %s\n", pyErr.Error())
			} else if report != nil {
				if len(report.MissingDeps) > 0 {
					fmt.Fprintf(os.Stderr, "⚠️ Attention (démarrage GUI) : dépendance(s) Python manquante(s) : %v\n", report.MissingDeps)
				} else if report.AppLoadError != "" {
					fmt.Fprintf(os.Stderr, "⚠️ Attention (démarrage GUI) : erreur module backend : %s\n", report.AppLoadError)
				}
			}
			fmt.Fprintf(os.Stderr, "ℹ Vous pouvez ajuster ou vérifier votre configuration dans l'onglet 'Configuration' de la GUI.\n")
		} else if report != nil {
			fmt.Printf("✔ Environnement Python validé (%s, Python %s) : dépendances et backend opérationnels.\n", report.Mode, report.Version)
		}
	}

	// Instanciation de l'orchestrateur d'application Wails
	app := NewApp(repoRoot, appRoot, *devFlag, *mockInterfaceFlag, *mockIAFlag, *dockerLLMFlag)

	// Lancement de la fenêtre native Wails
	err := wails.Run(&options.App{
		Title:     "Gitelligence — Assistant Git Intelligent",
		Width:     1240,
		Height:    840,
		MinWidth:  980,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 13, B: 22, A: 255}, // #090D16
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: *devFlag,
		},
		Linux: &linux.Options{
			ProgramName: "git-gui",
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur Wails : %v\n", err)
		os.Exit(1)
	}
}

package ui

import (
	"flag"
	"fmt"
	"io"
	"strings"

	_ "sae-git-cli/flags"
	"sae-git-cli/version"
)

// Constantes d'identification de l'application (références centralisées vers le package version)
const (
	AppVersion = version.Number
	AppName    = version.AppName
	BinaryName = version.BinaryName
)

// boolFlag est l'interface interne de la bibliothèque standard flag pour les options booléennes
type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

// isBoolFlag détermine dynamiquement si un drapeau flag.Flag est de type booléen
func isBoolFlag(f *flag.Flag) bool {
	if bf, ok := f.Value.(boolFlag); ok {
		return bf.IsBoolFlag()
	}
	return false
}

// formatFlagSignature calcule le nom affiché et la valeur par défaut d'un drapeau pour l'aide
func formatFlagSignature(f *flag.Flag) (string, string) {
	name := "--" + f.Name
	if !isBoolFlag(f) {
		if strings.Contains(strings.ToLower(f.Name), "dir") ||
			strings.Contains(strings.ToLower(f.Name), "repo") ||
			strings.Contains(strings.ToLower(f.Name), "path") {
			name += " <chemin>"
		} else {
			name += " <valeur>"
		}
	}

	defStr := ""
	if f.Name == "app-dir" && f.DefValue == "" {
		defStr = fmt.Sprintf(" %s(défaut : auto-détecté)%s", Dim, Reset)
	} else if isBoolFlag(f) {
		defStr = fmt.Sprintf(" %s(défaut : false)%s", Dim, Reset)
	} else if f.DefValue != "" {
		defStr = fmt.Sprintf(" %s(défaut : %s)%s", Dim, f.DefValue, Reset)
	}

	return name, defStr
}

// EnvVarInfo décrit une variable du fichier .env
type EnvVarInfo struct {
	Name        string
	Default     string
	Description string
}

// EnvVariables liste les variables d'environnement configurables via le .env.
// NOTE AUX DÉVELOPPEURS : Si vous ajoutez une nouvelle variable de configuration dans le
// fichier .env ou .env.exemple, pensez à ajouter une entrée dans ce tableau pour qu'elle
// apparaisse dans la documentation de ./git-generator --help.
var EnvVariables = []EnvVarInfo{
	{
		Name:        "OLLAMA_BASE_URL",
		Default:     "http://10.22.28.190:11434",
		Description: "URL de base de l'instance du serveur Ollama",
	},
	{
		Name:        "OLLAMA_MODEL",
		Default:     "gemma4:26b",
		Description: "Modèle de langage utilisé par défaut",
	},
	{
		Name:        "OLLAMA_TIMEOUT_S",
		Default:     "300",
		Description: "Délai d'attente maximum des requêtes LLM en secondes",
	},
	{
		Name:        "OLLAMA_NUM_CTX",
		Default:     "32768",
		Description: "Taille de la fenêtre de contexte Ollama en tokens",
	},
	{
		Name:        "ENABLE_NOTIFICATIONS",
		Default:     "true",
		Description: "Notifications système de bureau en fin d'inférence",
	},
	{
		Name:        "PYTHON_BIN",
		Default:     "auto-détecté",
		Description: "Chemin de l'exécutable Python du backend",
	},
	{
		Name:        "MOCK_INTERFACE",
		Default:     "false",
		Description: "Mockage de l'interface pour le développement",
	},
	{
		Name:        "MOCK_AI",
		Default:     "false",
		Description: "Mockage de l'IA pour le développement sans serveur Ollama",
	},
}

// GetVersionText retourne le texte de version simple
func GetVersionText() string {
	return version.String()
}

// PrintVersion affiche la version du programme sur la sortie fournie
func PrintVersion(w io.Writer) {
	fmt.Fprint(w, GetVersionText())
}

// GetGeneralHelp génère une aide courte et directe au format standard Linux
func GetGeneralHelp() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s%s (%s) v%s%s - Assistant Git intelligent\n\n", Bold+Cyan, AppName, BinaryName, AppVersion, Reset))

	sb.WriteString(Bold + "Utilisation :" + Reset + "  " + BinaryName + " [options...]\n\n")

	sb.WriteString(Bold + "Options :" + Reset + "\n")

	// 1. Drapeaux standards d'aide et de version affichés en tête
	sb.WriteString(fmt.Sprintf("    %s%-22s%s %s\n", Bold+Cyan, "-h, --help", Reset, "Afficher cette aide et quitter"))
	sb.WriteString(fmt.Sprintf("    %s%-22s%s %s\n", Bold+Cyan, "-v, --version", Reset, "Afficher la version et quitter"))

	// 2. Découverte et rendu dynamique de tous les drapeaux enregistrés dans flag.CommandLine
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		// Ignorer les drapeaux internes du test runner Go et les doublons d'aide/version
		if strings.HasPrefix(f.Name, "test.") || f.Name == "h" || f.Name == "help" || f.Name == "v" || f.Name == "version" {
			return
		}

		flagSig, defStr := formatFlagSignature(f)
		sb.WriteString(fmt.Sprintf("    %s%-22s%s %s%s\n", Bold+Cyan, flagSig, Reset, f.Usage, defStr))
	})
	sb.WriteString("\n")

	sb.WriteString(Bold + "Initialiser l'environnement Python (.venv) :" + Reset + "\n")
	sb.WriteString(fmt.Sprintf("    %s%-36s%s %s\n", Bold+Green, "python scripts/setup_python_env.py", Reset, "Créer le .venv et installer toutes les dépendances"))
	sb.WriteString(fmt.Sprintf("    %s%-36s%s %s\n", Bold+Green, "source .venv/bin/activate", Reset, "Activer l'environnement (.venv\\Scripts\\activate sous Windows)"))
	sb.WriteString("\n")

	sb.WriteString(Bold + "Variables d'environnement (.env) :" + Reset + "\n")
	for _, env := range EnvVariables {
		sb.WriteString(fmt.Sprintf("    %s%-22s%s %s %s(défaut : %s)%s\n", Bold+Magenta, env.Name, Reset, env.Description, Dim, env.Default, Reset))
	}

	return sb.String()
}

// PrintHelp affiche l'aide sur l'écrivain w (os.Stdout)
func PrintHelp(w io.Writer) {
	fmt.Fprintln(w, GetGeneralHelp())
}

// isValuedFlag vérifie dynamiquement auprès de flag.CommandLine si un argument attend une valeur
func isValuedFlag(arg string) bool {
	cleanName := strings.TrimLeft(arg, "-")
	f := flag.CommandLine.Lookup(cleanName)
	if f != nil {
		return !isBoolFlag(f)
	}
	return false
}

// isPrefixedValuedFlag vérifie dynamiquement si un argument contient un drapeau avec sa valeur liée (ex: --repo=valeur)
func isPrefixedValuedFlag(arg string) bool {
	if !strings.HasPrefix(arg, "-") || !strings.Contains(arg, "=") {
		return false
	}
	parts := strings.SplitN(arg, "=", 2)
	cleanName := strings.TrimLeft(parts[0], "-")
	f := flag.CommandLine.Lookup(cleanName)
	if f != nil {
		return !isBoolFlag(f)
	}
	return false
}

// IsHelpRequested analyse les arguments de la ligne de commande et détermine si l'aide a été demandée
func IsHelpRequested(args []string) bool {
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if isValuedFlag(arg) {
			skipNext = true
			continue
		}
		if isPrefixedValuedFlag(arg) {
			continue
		}
		if arg == "-h" || arg == "--help" || arg == "-help" || arg == "help" {
			return true
		}
	}
	return false
}

// IsVersionRequested analyse les arguments pour détecter la demande de version
func IsVersionRequested(args []string) bool {
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if isValuedFlag(arg) {
			skipNext = true
			continue
		}
		if isPrefixedValuedFlag(arg) {
			continue
		}
		if arg == "-v" || arg == "--version" || arg == "-version" {
			return true
		}
	}
	return false
}

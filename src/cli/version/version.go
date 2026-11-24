package version

import "fmt"

// Constantes globales d'identification et de version de l'application
const (
	Number     = "1.0.0"
	AppName    = "Gitelligence"
	BinaryName = "git-generator"
)

// String renvoie l'en-tête de version complet formaté
func String() string {
	return fmt.Sprintf("%s (%s) version %s\n", AppName, BinaryName, Number)
}

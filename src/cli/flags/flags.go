package flags

import "flag"

// Options globales de la ligne de commande de Gitelligence.
// Tout drapeau déclaré ici est automatiquement enregistré dans le registre standard flag.CommandLine
// et découvert dynamiquement par le système d'aide (--help).
var (
	Dev           = flag.Bool("dev", false, "Active le mode développement (débloque les options de mock pour développer l'application)")
	MockInterface = flag.Bool("mockInterface", false, "Active le mockage complet de l'interface pour le développement (mode sans Git ni Python, requiert --dev)")
	MockIA        = flag.Bool("mockIA", false, "Active le mockage de l'IA pour le développement (sans serveur Ollama, requiert --dev)")
	DockerLLM     = flag.Bool("docker-llm", false, "Force l'usage du serveur Ollama local via Docker")
	Repo          = flag.String("repo", ".", "Dépôt Git cible à analyser")
	AppDir        = flag.String("app-dir", "", "Racine de l'application Gitelligence contenant le .env")
	PythonBin     = flag.String("python-bin", "", "Chemin explicite vers l'interpréteur Python à utiliser (surcharge le .env et .venv)")
)

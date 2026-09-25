package bridge

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sae-git-cli/models"
)

// isAppRoot vérifie si un répertoire correspond à la racine de l'application Smart Commit.
func isAppRoot(dir string) bool {
	if dir == "" {
		return false
	}
	clean := filepath.Clean(dir)
	base := filepath.Base(clean)
	if base == "src" || base == "cli" {
		return false
	}
	if isTestDir(clean) {
		return true
	}
	checkPath := filepath.Join(clean, "src", "git_commit_release_notes_generator")
	stat, err := os.Stat(checkPath)
	return err == nil && stat.IsDir()
}

// isTestDir détermine si un chemin correspond à un répertoire temporaire de test (ex: t.TempDir()).
func isTestDir(dir string) bool {
	if dir == "" {
		return false
	}
	tmp := os.TempDir()
	return strings.HasPrefix(dir, tmp) || strings.Contains(dir, "Test")
}

// FindAppRoot localise la racine de l'application Smart Commit (où vit le code et le .env applicatif).
func FindAppRoot(hints ...string) string {
	for _, h := range hints {
		if h != "" && isAppRoot(h) {
			if abs, err := filepath.Abs(h); err == nil {
				return abs
			}
			return h
		}
	}

	if envRoot := os.Getenv("SMART_COMMIT_APP_ROOT"); envRoot != "" && isAppRoot(envRoot) {
		if abs, err := filepath.Abs(envRoot); err == nil {
			return abs
		}
		return envRoot
	}

	// 1. Recherche en remontant depuis l'emplacement du code source Go (développement, go run, tests)
	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		curr := filepath.Dir(sourceFile)
		for {
			if isAppRoot(curr) {
				return curr
			}
			parent := filepath.Dir(curr)
			if parent == curr || parent == "" {
				break
			}
			curr = parent
		}
	}

	// 1. Recherche en remontant depuis l'exécutable Go
	if exe, err := os.Executable(); err == nil {
		if realExe, errEval := filepath.EvalSymlinks(exe); errEval == nil {
			exe = realExe
		}
		curr := filepath.Dir(exe)
		for {
			if isAppRoot(curr) {
				return curr
			}
			parent := filepath.Dir(curr)
			if parent == curr || parent == "" {
				break
			}
			curr = parent
		}
	}

	// 2. Recherche en remontant depuis le répertoire de travail courant
	if cwd, err := os.Getwd(); err == nil {
		curr := cwd
		for {
			if isAppRoot(curr) {
				return curr
			}
			parent := filepath.Dir(curr)
			if parent == curr || parent == "" {
				break
			}
			curr = parent
		}
	}

	// 3. Repli sur le premier hint non vide, ou "."
	for _, h := range hints {
		if h != "" {
			return h
		}
	}
	return "."
}

// FindAppEnvPath renvoie l'emplacement du fichier .env STRICTEMENT à la racine de l'application.
// Il ne cherche JAMAIS dans le dossier src/ ni dans le dépôt cible.
func FindAppEnvPath(appRoot string) string {
	if appRoot == "" {
		appRoot = FindAppRoot()
	}

	clean := filepath.Clean(appRoot)
	for filepath.Base(clean) == "src" || filepath.Base(clean) == "cli" {
		clean = filepath.Dir(clean)
	}

	return filepath.Join(clean, ".env")
}

// FindEnvPath est un alias conservé pour la compatibilité avec le code existant et les tests.
func FindEnvPath(appRoot string) string {
	return FindAppEnvPath(appRoot)
}

// ParseBoolFromString convertit les représentations usuelles de booléens ("true", "1", "yes", "oui") en bool.
func ParseBoolFromString(v string) bool {
	lower := strings.ToLower(strings.TrimSpace(v))
	return lower == "true" || lower == "1" || lower == "yes" || lower == "oui"
}

// ReadMockSettingFromEnv lit la variable MOCK_INTERFACE dans l'environnement ou le fichier .env applicatif.
func ReadMockSettingFromEnv(appRoot string) bool {
	// 1. Variables d'environnement système directes prioritaires
	if envVal := os.Getenv("MOCK_INTERFACE"); envVal != "" {
		return ParseBoolFromString(envVal)
	}
	if envVal := os.Getenv("MOCKAGE_INTERFACE"); envVal != "" {
		return ParseBoolFromString(envVal)
	}

	// 2. Lecture du fichier .env applicatif
	envPath := FindAppEnvPath(appRoot)
	data, err := os.ReadFile(envPath)
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			if k == "MOCK_INTERFACE" || k == "MOCKAGE_INTERFACE" {
				return ParseBoolFromString(v)
			}
		}
	}
	return false
}

// LoadConfigFromEnv lit la configuration depuis le fichier .env applicatif et renvoie une structure ConfigResponse.
func LoadConfigFromEnv(appRoot string, isMockActive bool) *models.ConfigResponse {
	cfg := &models.ConfigResponse{
		Success:       true,
		OllamaBaseURL: "http://10.22.28.190:11434",
		OllamaModel:   "gemma4:26b",
		TimeoutS:      300.0,
		Language:      "fr",
		MockInterface: isMockActive,
	}

	envPath := FindAppEnvPath(appRoot)
	// Sécurité : purge tout fichier .env résiduel dans src/
	_ = os.Remove(filepath.Join(filepath.Dir(envPath), "src", ".env"))
	_ = os.Remove(filepath.Join(filepath.Dir(envPath), "src", ".env.exemple"))

	data, err := os.ReadFile(envPath)
	if err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				switch k {
				case "OLLAMA_BASE_URL":
					if v != "" {
						cfg.OllamaBaseURL = v
					}
				case "OLLAMA_MODEL":
					if v != "" {
						cfg.OllamaModel = v
					}
				case "OLLAMA_TIMEOUT_S":
					var t float64
					if _, errParse := fmt.Sscanf(v, "%f", &t); errParse == nil {
						cfg.TimeoutS = t
					}
				case "APP_LANGUAGE":
					if v != "" {
						cfg.Language = v
					}
				case "MOCK_INTERFACE", "MOCKAGE_INTERFACE":
					cfg.MockInterface = ParseBoolFromString(v)
				}
			}
		}
	}

	return cfg
}

// WriteEnvFile réécrit le fichier .env STRICTEMENT à la racine du projet.
func WriteEnvFile(envPath string, baseURL, model string, timeout float64, lang string, mockInterface bool) error {
	cleanPath := filepath.Clean(envPath)
	dir := filepath.Dir(cleanPath)
	for filepath.Base(dir) == "src" || filepath.Base(dir) == "cli" {
		dir = filepath.Dir(dir)
		cleanPath = filepath.Join(dir, ".env")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`# ==============================================================================
# Configuration Générale 
# ==============================================================================

# Configuration du serveur Ollama IUT
OLLAMA_BASE_URL=%s
OLLAMA_MODEL=%s
OLLAMA_TIMEOUT_S=%.0f

# Langue de l'application (fr / en)
APP_LANGUAGE=%s

# Mode hors-ligne / démo (true / false)
MOCK_INTERFACE=%t
`, baseURL, model, timeout, lang, mockInterface)

	if err := os.WriteFile(cleanPath, []byte(content), 0644); err != nil {
		return err
	}

	// Suppression préventive de tout ancien src/.env pour ne garder STRICTEMENT que la racine du projet
	_ = os.Remove(filepath.Join(dir, "src", ".env"))
	_ = os.Remove(filepath.Join(dir, "src", ".env.exemple"))

	return nil
}

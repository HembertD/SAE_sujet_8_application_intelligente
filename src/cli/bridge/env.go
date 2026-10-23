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

// isAppRoot vérifie si un répertoire correspond à la racine de l'application Gitelligence.
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

// FindAppRoot localise la racine de l'application Gitelligence (où vit le code et le .env applicatif).
func FindAppRoot(hints ...string) string {
	for _, h := range hints {
		if h != "" && isAppRoot(h) {
			if abs, err := filepath.Abs(h); err == nil {
				return abs
			}
			return h
		}
	}

	if envRoot := os.Getenv("GITELLIGENCE_APP_ROOT"); envRoot != "" && isAppRoot(envRoot) {
		if abs, err := filepath.Abs(envRoot); err == nil {
			return abs
		}
		return envRoot
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

// ReadDockerLLMSettingFromEnv lit la variable USE_DOCKER_LLM dans l'environnement ou le fichier .env applicatif.
func ReadDockerLLMSettingFromEnv(appRoot string) bool {
	if envVal := os.Getenv("USE_DOCKER_LLM"); envVal != "" {
		return ParseBoolFromString(envVal)
	}

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
			if k == "USE_DOCKER_LLM" {
				return ParseBoolFromString(v)
			}
		}
	}
	return false
}

// ReadMockAISettingFromEnv lit la variable MOCK_AI dans l'environnement ou le fichier .env applicatif.
func ReadMockAISettingFromEnv(appRoot string) bool {
	if envVal := os.Getenv("MOCK_AI"); envVal != "" {
		return ParseBoolFromString(envVal)
	}

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
			if k == "MOCK_AI" {
				return ParseBoolFromString(v)
			}
		}
	}
	return false
}

// ReadPythonBinFromEnv lit la variable PYTHON_BIN dans l'environnement système ou le fichier .env applicatif.
// Retourne une chaîne vide si rien n'est configuré explicitement (détection automatique à faire par l'appelant).
func ReadPythonBinFromEnv(appRoot string) string {
	if envVal := os.Getenv("PYTHON_BIN"); envVal != "" {
		return envVal
	}

	envPath := FindAppEnvPath(appRoot)
	data, err := os.ReadFile(envPath)
	if err != nil {
		return ""
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
			if k == "PYTHON_BIN" && v != "" {
				return v
			}
		}
	}
	return ""
}

// ReadEnableNotificationsFromEnv lit la variable ENABLE_NOTIFICATIONS dans l'environnement ou le fichier .env applicatif.
// Par défaut, renvoie true si le paramètre n'est pas explicité.
func ReadEnableNotificationsFromEnv(appRoot string) bool {
	if envVal := os.Getenv("ENABLE_NOTIFICATIONS"); envVal != "" {
		return ParseBoolFromString(envVal)
	}

	envPath := FindAppEnvPath(appRoot)
	data, err := os.ReadFile(envPath)
	if err != nil {
		return true
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
			if k == "ENABLE_NOTIFICATIONS" {
				return ParseBoolFromString(v)
			}
		}
	}
	return true
}

// WriteNotificationsSetting met à jour la variable ENABLE_NOTIFICATIONS dans le fichier .env applicatif.
func WriteNotificationsSetting(appRoot string, enabled bool) error {
	cleanPath := FindAppEnvPath(appRoot)
	dir := filepath.Dir(cleanPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	found := false
	var newLines []string
	if existing, errRead := os.ReadFile(cleanPath); errRead == nil {
		scanner := bufio.NewScanner(bytes.NewReader(existing))
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "ENABLE_NOTIFICATIONS=") {
				newLines = append(newLines, fmt.Sprintf("ENABLE_NOTIFICATIONS=%t", enabled))
				found = true
			} else {
				newLines = append(newLines, line)
			}
		}
	}

	if !found {
		newLines = append(newLines, fmt.Sprintf("\n# Notifications de fin d'inférence (true / false)\nENABLE_NOTIFICATIONS=%t", enabled))
	}

	content := strings.Join(newLines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return os.WriteFile(cleanPath, []byte(content), 0644)
}

// LoadConfigFromEnv lit la configuration depuis le fichier .env applicatif et renvoie une structure ConfigResponse.
func LoadConfigFromEnv(appRoot string, isMockActive bool) *models.ConfigResponse {
	cfg := &models.ConfigResponse{
		Success:             true,
		OllamaBaseURL:       "http://10.22.28.190:11434",
		OllamaModel:         "gemma4:26b",
		TimeoutS:            300.0,
		MockInterface:       isMockActive,
		MockAI:              ReadMockAISettingFromEnv(appRoot),
		DockerLLM:           ReadDockerLLMSettingFromEnv(appRoot),
		EnableNotifications: ReadEnableNotificationsFromEnv(appRoot),
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
				case "MOCK_INTERFACE", "MOCKAGE_INTERFACE":
					if os.Getenv("MOCK_INTERFACE") == "" && os.Getenv("MOCKAGE_INTERFACE") == "" {
						cfg.MockInterface = ParseBoolFromString(v)
					}
				case "MOCK_AI":
					if os.Getenv("MOCK_AI") == "" {
						cfg.MockAI = ParseBoolFromString(v)
					}
				case "USE_DOCKER_LLM":
					if os.Getenv("USE_DOCKER_LLM") == "" {
						cfg.DockerLLM = ParseBoolFromString(v)
					}
				case "ENABLE_NOTIFICATIONS":
					if os.Getenv("ENABLE_NOTIFICATIONS") == "" {
						cfg.EnableNotifications = ParseBoolFromString(v)
					}
				}
			}
		}
	}

	return cfg
}

// WriteEnvFile réécrit le fichier .env STRICTEMENT à la racine du projet.
func WriteEnvFile(envPath string, baseURL, model string, timeout float64, mockInterface bool, mockAIArgs ...bool) error {
	return WriteEnvFileWithDockerLLM(envPath, baseURL, model, timeout, mockInterface, false, mockAIArgs...)
}

// WriteEnvFileWithDockerLLM réécrit le fichier .env STRICTEMENT à la racine du projet, avec prise en compte de USE_DOCKER_LLM.
func WriteEnvFileWithDockerLLM(envPath string, baseURL, model string, timeout float64, mockInterface bool, dockerLLM bool, mockAIArgs ...bool) error {
	cleanPath := filepath.Clean(envPath)
	dir := filepath.Dir(cleanPath)
	for filepath.Base(dir) == "src" || filepath.Base(dir) == "cli" {
		dir = filepath.Dir(dir)
		cleanPath = filepath.Join(dir, ".env")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	mockAI := ReadMockAISettingFromEnv(dir)
	if len(mockAIArgs) > 0 {
		mockAI = mockAIArgs[0]
	}

	// Préserve les paramètres non gérés par cet écran de configuration (ex:
	// OLLAMA_NUM_CTX, PYTHON_BIN) : sans ça, ils disparaissaient
	// silencieusement du .env à chaque sauvegarde depuis l'interface.
	managedKeys := map[string]bool{
		"OLLAMA_BASE_URL": true, "OLLAMA_MODEL": true, "OLLAMA_TIMEOUT_S": true,
		"MOCK_INTERFACE": true, "MOCKAGE_INTERFACE": true, "MOCK_AI": true,
		"USE_DOCKER_LLM": true,
	}
	var preserved []string
	if existing, errRead := os.ReadFile(cleanPath); errRead == nil {
		scanner := bufio.NewScanner(bytes.NewReader(existing))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && !managedKeys[strings.TrimSpace(parts[0])] {
				preserved = append(preserved, line)
			}
		}
	}

	content := fmt.Sprintf(`# ==============================================================================
# Configuration Générale
# ==============================================================================

# Configuration du serveur Ollama IUT
OLLAMA_BASE_URL=%s
OLLAMA_MODEL=%s
OLLAMA_TIMEOUT_S=%.0f

# Mode hors-ligne / démo interface (true / false)
MOCK_INTERFACE=%t

# Simulation de l'IA hors-ligne (true / false)
MOCK_AI=%t

# Utilise le serveur Ollama local via Docker (true / false)
USE_DOCKER_LLM=%t
`, baseURL, model, timeout, mockInterface, mockAI, dockerLLM)

	if len(preserved) > 0 {
		content += "\n# Autres paramètres (préservés automatiquement)\n" + strings.Join(preserved, "\n") + "\n"
	}

	if err := os.WriteFile(cleanPath, []byte(content), 0644); err != nil {
		return err
	}

	// Suppression préventive de tout ancien src/.env pour ne garder STRICTEMENT que la racine du projet
	_ = os.Remove(filepath.Join(dir, "src", ".env"))
	_ = os.Remove(filepath.Join(dir, "src", ".env.exemple"))

	return nil
}

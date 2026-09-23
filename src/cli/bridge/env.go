package bridge

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sae-git-cli/models"
)

// FindEnvPath recherche l'emplacement du fichier .env à partir de plusieurs chemins candidats.
func FindEnvPath(repoRoot string) string {
	candidatePaths := []string{
		filepath.Join(repoRoot, "src", ".env"),
		filepath.Join(repoRoot, ".env"),
		filepath.Join(repoRoot, "SAE_sujet_8_application_intelligente", "src", ".env"),
		filepath.Join(".", "src", ".env"),
		filepath.Join(".", ".env"),
		filepath.Join(".", "SAE_sujet_8_application_intelligente", "src", ".env"),
		filepath.Join("..", "src", ".env"),
		filepath.Join("..", ".env"),
	}
	for _, p := range candidatePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(repoRoot, "src", ".env")
}

// ParseBoolFromString convertit les représentations usuelles de booléens ("true", "1", "yes", "oui") en bool.
func ParseBoolFromString(v string) bool {
	lower := strings.ToLower(strings.TrimSpace(v))
	return lower == "true" || lower == "1" || lower == "yes" || lower == "oui"
}

// ReadMockSettingFromEnv lit la variable MOCK_INTERFACE dans l'environnement ou le fichier .env.
func ReadMockSettingFromEnv(repoRoot string) bool {
	// 1. Variables d'environnement système directes prioritaires
	if envVal := os.Getenv("MOCK_INTERFACE"); envVal != "" {
		return ParseBoolFromString(envVal)
	}
	if envVal := os.Getenv("MOCKAGE_INTERFACE"); envVal != "" {
		return ParseBoolFromString(envVal)
	}

	// 2. Lecture du fichier .env
	envPath := FindEnvPath(repoRoot)
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

// LoadConfigFromEnv lit la configuration depuis le fichier .env et renvoie une structure ConfigResponse.
func LoadConfigFromEnv(repoRoot string, isMockActive bool) *models.ConfigResponse {
	cfg := &models.ConfigResponse{
		Success:       true,
		OllamaBaseURL: "http://10.22.28.190:11434",
		OllamaModel:   "gemma4:26b",
		TimeoutS:      60.0,
		Language:      "fr",
		MockInterface: isMockActive,
	}

	envPath := FindEnvPath(repoRoot)
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
					cfg.OllamaBaseURL = v
				case "OLLAMA_MODEL":
					cfg.OllamaModel = v
				case "OLLAMA_TIMEOUT_S":
					var t float64
					if _, errParse := fmt.Sscanf(v, "%f", &t); errParse == nil {
						cfg.TimeoutS = t
					}
				case "APP_LANGUAGE":
					cfg.Language = v
				case "MOCK_INTERFACE", "MOCKAGE_INTERFACE":
					cfg.MockInterface = ParseBoolFromString(v)
				}
			}
		}
	}

	return cfg
}

// WriteEnvFile réécrit le fichier .env avec les nouveaux paramètres.
func WriteEnvFile(envPath string, baseURL, model string, timeout float64, lang string, mockInterface bool) error {
	dir := filepath.Dir(envPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`# ==============================================================================
# Configuration Générale - SAÉ Sujet 8 (Git Commit & Release Notes Generator)
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

	return os.WriteFile(envPath, []byte(content), 0644)
}

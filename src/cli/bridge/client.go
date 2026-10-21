package bridge

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"sae-git-cli/models"
)

// BackendClient définit l'interface publique d'accès aux services du back-end.
// Les vues front-end dépendent exclusivement de cette interface.
type BackendClient interface {
	GetRepoStatus() (*models.RepoStatus, error)
	GetDiff() (*models.DiffResponse, error)
	StageAll() (*models.ActionResult, error)
	GenerateCommit(feedback string) (*models.CommitProposal, error)
	ApplyCommit(message string) (*models.ActionResult, error)
	Push() (*models.ActionResult, error)
	GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error)
	GetConfig() (*models.ConfigResponse, error)
	SaveConfig(baseURL, model string, timeout float64, mockInterface bool, mockAI ...bool) (*models.ActionResult, error)
	SaveConfigWithDockerLLM(baseURL, model string, timeout float64, mockInterface bool, dockerLLM bool, mockAI ...bool) (*models.ActionResult, error)
	PingOllama() (*models.PingResponse, error)
	SaveMarkdownFile(filename, content string) error
	IsMock() bool
	IsDockerLLM() bool
	StartDockerOllama() (*models.ActionResult, error)
	StopDockerOllama() (*models.ActionResult, error)
}

// BridgeClient coordonne l'accès au backend en déléguant à PythonClient ou MockClient.
type BridgeClient struct {
	repoRoot           string
	appRoot            string
	forceMockInterface bool
	forceMockIA        bool
	forceDockerLLM     bool
	isMock             bool
	isDockerLLM        bool
	realClient         *PythonClient
	mockClient         *MockClient
}

// NewBackendClient instancie le client de pontage vers le backend.
func NewBackendClient(repoRoot string, forceMockInterface bool, forceMockIA bool) BackendClient {
	appRoot := FindAppRoot(repoRoot)
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	return NewBackendClientWithAppRoot(repoRoot, appRoot, forceMockInterface, forceMockIA)
}

// NewBackendClientWithAppRoot instancie le client en séparant le dépôt cible et la racine de l'application.
func NewBackendClientWithAppRoot(repoRoot, appRoot string, forceMockInterface bool, forceMockIA bool, forceDockerLLM ...bool) BackendClient {
	if appRoot == "" {
		if isTestDir(repoRoot) {
			appRoot = repoRoot
		} else {
			appRoot = FindAppRoot(repoRoot)
		}
	}
	mockActive := forceMockInterface || ReadMockSettingFromEnv(appRoot)
	mockIAActive := forceMockIA
	if mockIAActive {
		_ = os.Setenv("MOCK_AI", "true")
	}
	dockerLLMActive := false
	if len(forceDockerLLM) > 0 && forceDockerLLM[0] {
		dockerLLMActive = true
		_ = os.Setenv("USE_DOCKER_LLM", "true")
	} else {
		dockerLLMActive = ReadDockerLLMSettingFromEnv(appRoot)
	}

	return &BridgeClient{
		repoRoot:           repoRoot,
		appRoot:            appRoot,
		forceMockInterface: forceMockInterface,
		forceMockIA:        mockIAActive,
		forceDockerLLM:     dockerLLMActive,
		isMock:             mockActive,
		isDockerLLM:        dockerLLMActive,
		realClient:         NewPythonClientWithAppRoot(repoRoot, appRoot),
		mockClient:         NewMockClientWithAppRoot(repoRoot, appRoot),
	}
}

// activeDelegate renvoie l'implémentation active.
func (b *BridgeClient) activeDelegate() BackendClient {
	if b.IsMock() {
		return b.mockClient
	}
	return b.realClient
}

// IsMock indique si le mode simulation de l'interface est actif.
func (b *BridgeClient) IsMock() bool {
	if b.forceMockInterface {
		return true
	}
	return ReadMockSettingFromEnv(b.appRoot)
}

// IsDockerLLM indique si l'usage du serveur Ollama via Docker est activé.
func (b *BridgeClient) IsDockerLLM() bool {
	if b.forceDockerLLM {
		return true
	}
	return ReadDockerLLMSettingFromEnv(b.appRoot)
}

// GetRepoStatus délègue la récupération de l'état du dépôt au client actif
func (b *BridgeClient) GetRepoStatus() (*models.RepoStatus, error) {
	return b.activeDelegate().GetRepoStatus()
}

// GetDiff délègue l'analyse du diff au client actif
func (b *BridgeClient) GetDiff() (*models.DiffResponse, error) {
	return b.activeDelegate().GetDiff()
}

// StageAll délègue l'indexation au client actif
func (b *BridgeClient) StageAll() (*models.ActionResult, error) {
	return b.activeDelegate().StageAll()
}

// GenerateCommit délègue l'inférence du message de commit au client actif
func (b *BridgeClient) GenerateCommit(feedback string) (*models.CommitProposal, error) {
	return b.activeDelegate().GenerateCommit(feedback)
}

// ApplyCommit délègue l'exécution du commit au client actif
func (b *BridgeClient) ApplyCommit(message string) (*models.ActionResult, error) {
	return b.activeDelegate().ApplyCommit(message)
}

// Push délègue la poussée des commits au client actif
func (b *BridgeClient) Push() (*models.ActionResult, error) {
	return b.activeDelegate().Push()
}

// GetReleaseNotes délègue la génération des release notes au client actif
func (b *BridgeClient) GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error) {
	return b.activeDelegate().GetReleaseNotes(fromTag, toTag)
}

// PingOllama délègue le ping du serveur Ollama au client actif
func (b *BridgeClient) PingOllama() (*models.PingResponse, error) {
	return b.activeDelegate().PingOllama()
}

// SaveMarkdownFile délègue la sauvegarde locale au client actif
func (b *BridgeClient) SaveMarkdownFile(filename, content string) error {
	return b.activeDelegate().SaveMarkdownFile(filename, content)
}

// GetConfig lit la configuration depuis le fichier .env de l'application
func (b *BridgeClient) GetConfig() (*models.ConfigResponse, error) {
	cfg := LoadConfigFromEnv(b.appRoot, b.IsMock())
	if b.forceMockIA {
		cfg.MockAI = true
	}
	cfg.DockerLLM = b.IsDockerLLM()
	return cfg, nil
}

// SaveConfig persiste la nouvelle configuration dans le .env applicatif et actualise l'état du client.
func (b *BridgeClient) SaveConfig(baseURL, model string, timeout float64, mockInterface bool, mockAI ...bool) (*models.ActionResult, error) {
	return b.SaveConfigWithDockerLLM(baseURL, model, timeout, mockInterface, b.IsDockerLLM(), mockAI...)
}

// SaveConfigWithDockerLLM persiste la configuration avec l'option USE_DOCKER_LLM explicite.
func (b *BridgeClient) SaveConfigWithDockerLLM(baseURL, model string, timeout float64, mockInterface bool, dockerLLM bool, mockAI ...bool) (*models.ActionResult, error) {
	envPath := FindAppEnvPath(b.appRoot)
	if err := WriteEnvFileWithDockerLLM(envPath, baseURL, model, timeout, mockInterface, dockerLLM, mockAI...); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env applicatif : %w", err)
	}

	// Met à jour dynamiquement le mode si pas forcé par le flag --mockInterface
	if !b.forceMockInterface {
		b.isMock = mockInterface
	}
	b.isDockerLLM = dockerLLM

	// Si en mode réel, tenter de notifier le backend Python
	if !b.IsMock() {
		_, _ = b.realClient.runPythonCommand(
			"--action", "save-config",
			"--base-url", baseURL,
			"--model", model,
			"--timeout", fmt.Sprintf("%f", timeout),
		)
	}

	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Configuration sauvegardée dans %s (USE_DOCKER_LLM=%t).", filepath.Base(envPath), dockerLLM),
	}, nil
}

// StartDockerOllama lance le conteneur Ollama via docker compose et affiche une progression.
func (b *BridgeClient) StartDockerOllama() (*models.ActionResult, error) {
	appRoot := b.appRoot
	if appRoot == "" {
		appRoot = FindAppRoot(b.repoRoot)
	}
	composeFile := filepath.Join(appRoot, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return nil, fmt.Errorf("fichier docker-compose.yml introuvable à la racine de l'application : %s", composeFile)
	}

	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &models.ActionResult{
			Success: false,
			Error:   fmt.Sprintf("échec du lancement du conteneur Ollama : %s", string(output)),
		}, err
	}

	b.isDockerLLM = true
	_ = os.Setenv("USE_DOCKER_LLM", "true")

	return &models.ActionResult{
		Success: true,
		Message: "Conteneur Ollama démarré avec succès via docker compose.",
	}, nil
}

// StopDockerOllama arrête le conteneur Ollama via docker compose.
func (b *BridgeClient) StopDockerOllama() (*models.ActionResult, error) {
	appRoot := b.appRoot
	if appRoot == "" {
		appRoot = FindAppRoot(b.repoRoot)
	}
	composeFile := filepath.Join(appRoot, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return nil, fmt.Errorf("fichier docker-compose.yml introuvable à la racine de l'application : %s", composeFile)
	}

	cmd := exec.Command("docker", "compose", "-f", composeFile, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &models.ActionResult{
			Success: false,
			Error:   fmt.Sprintf("échec de l'arrêt du conteneur Ollama : %s", string(output)),
		}, err
	}

	b.isDockerLLM = false
	_ = os.Setenv("USE_DOCKER_LLM", "false")

	return &models.ActionResult{
		Success: true,
		Message: "Conteneur Ollama arrêté.",
	}, nil
}

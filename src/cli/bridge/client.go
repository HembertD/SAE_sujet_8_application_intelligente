package bridge

import (
	"fmt"
	"os"
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
	SaveConfig(baseURL, model string, timeout float64, lang string, mockInterface bool, mockAI ...bool) (*models.ActionResult, error)
	PingOllama() (*models.PingResponse, error)
	SaveMarkdownFile(filename, content string) error
	IsMock() bool
}

// BridgeClient coordonne l'accès au backend en déléguant à PythonClient ou MockClient.
type BridgeClient struct {
	repoRoot           string
	appRoot            string
	forceMockInterface bool
	forceMockIA        bool
	isMock             bool
	realClient         *PythonClient
	mockClient         *MockClient
}

// NewBackendClient instancie le client de pontage vers le backend.
func NewBackendClient(repoRoot string, forceMockInterface bool, forceMockIA ...bool) BackendClient {
	appRoot := FindAppRoot(repoRoot)
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	return NewBackendClientWithAppRoot(repoRoot, appRoot, forceMockInterface, forceMockIA...)
}

// NewBackendClientWithAppRoot instancie le client en séparant le dépôt cible et la racine de l'application.
func NewBackendClientWithAppRoot(repoRoot, appRoot string, forceMockInterface bool, forceMockIA ...bool) BackendClient {
	if appRoot == "" {
		if isTestDir(repoRoot) {
			appRoot = repoRoot
		} else {
			appRoot = FindAppRoot(repoRoot)
		}
	}
	mockActive := forceMockInterface || ReadMockSettingFromEnv(appRoot)
	mockIAActive := false
	if len(forceMockIA) > 0 && forceMockIA[0] {
		mockIAActive = true
		_ = os.Setenv("MOCK_AI", "true")
	}

	return &BridgeClient{
		repoRoot:           repoRoot,
		appRoot:            appRoot,
		forceMockInterface: forceMockInterface,
		forceMockIA:        mockIAActive,
		isMock:             mockActive,
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
	return cfg, nil
}

// SaveConfig persiste la nouvelle configuration dans le .env applicatif et actualise l'état du client.
func (b *BridgeClient) SaveConfig(baseURL, model string, timeout float64, lang string, mockInterface bool, mockAI ...bool) (*models.ActionResult, error) {
	envPath := FindAppEnvPath(b.appRoot)
	if err := WriteEnvFile(envPath, baseURL, model, timeout, lang, mockInterface, mockAI...); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env applicatif : %w", err)
	}

	// Met à jour dynamiquement le mode si pas forcé par le flag --mockInterface
	if !b.forceMockInterface {
		b.isMock = mockInterface
	}

	// Si en mode réel, tenter de notifier le backend Python
	if !b.IsMock() {
		_, _ = b.realClient.runPythonCommand(
			"--action", "save-config",
			"--base-url", baseURL,
			"--model", model,
			"--timeout", fmt.Sprintf("%f", timeout),
			"--language", lang,
		)
	}

	mockAIVal := ReadMockAISettingFromEnv(b.appRoot)
	if len(mockAI) > 0 {
		mockAIVal = mockAI[0]
	}

	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Configuration sauvegardée dans %s (MOCK_INTERFACE=%t, MOCK_AI=%t).", filepath.Base(envPath), mockInterface, mockAIVal),
	}, nil
}

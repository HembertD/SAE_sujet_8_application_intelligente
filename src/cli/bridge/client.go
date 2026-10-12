package bridge

import (
	"fmt"
	"os"
	"path/filepath"

	"sae-git-cli/models"
)

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

type BridgeClient struct {
	repoRoot           string
	appRoot            string
	forceMockInterface bool
	forceMockIA        bool
	isMock             bool
	realClient         *PythonClient
	mockClient         *MockClient
}

func NewBackendClient(repoRoot string, forceMockInterface bool, forceMockIA ...bool) BackendClient {
	appRoot := FindAppRoot(repoRoot)
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	return NewBackendClientWithAppRoot(repoRoot, appRoot, forceMockInterface, forceMockIA...)
}

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

func (b *BridgeClient) activeDelegate() BackendClient {
	if b.IsMock() {
		return b.mockClient
	}
	return b.realClient
}

func (b *BridgeClient) IsMock() bool {
	if b.forceMockInterface {
		return true
	}
	return ReadMockSettingFromEnv(b.appRoot)
}

func (b *BridgeClient) GetRepoStatus() (*models.RepoStatus, error) {
	return b.activeDelegate().GetRepoStatus()
}

func (b *BridgeClient) GetDiff() (*models.DiffResponse, error) {
	return b.activeDelegate().GetDiff()
}

func (b *BridgeClient) StageAll() (*models.ActionResult, error) {
	return b.activeDelegate().StageAll()
}

func (b *BridgeClient) GenerateCommit(feedback string) (*models.CommitProposal, error) {
	return b.activeDelegate().GenerateCommit(feedback)
}

func (b *BridgeClient) ApplyCommit(message string) (*models.ActionResult, error) {
	return b.activeDelegate().ApplyCommit(message)
}

func (b *BridgeClient) Push() (*models.ActionResult, error) {
	return b.activeDelegate().Push()
}

func (b *BridgeClient) GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error) {
	return b.activeDelegate().GetReleaseNotes(fromTag, toTag)
}

func (b *BridgeClient) PingOllama() (*models.PingResponse, error) {
	return b.activeDelegate().PingOllama()
}

func (b *BridgeClient) SaveMarkdownFile(filename, content string) error {
	return b.activeDelegate().SaveMarkdownFile(filename, content)
}

func (b *BridgeClient) GetConfig() (*models.ConfigResponse, error) {
	cfg := LoadConfigFromEnv(b.appRoot, b.IsMock())
	if b.forceMockIA {
		cfg.MockAI = true
	}
	return cfg, nil
}

func (b *BridgeClient) SaveConfig(baseURL, model string, timeout float64, lang string, mockInterface bool, mockAI ...bool) (*models.ActionResult, error) {
	envPath := FindAppEnvPath(b.appRoot)
	if err := WriteEnvFile(envPath, baseURL, model, timeout, lang, mockInterface, mockAI...); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env applicatif : %w", err)
	}

	if !b.forceMockInterface {
		b.isMock = mockInterface
	}

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

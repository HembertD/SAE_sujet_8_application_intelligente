package bridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sae-git-cli/models"
)

// MockClient fournit une implémentation de BackendClient pour les tests et le mode hors-ligne.
type MockClient struct {
	repoRoot      string
	appRoot       string
	SimulateDelay bool

	// Champs de surcharge pour les tests unitaires
	CustomRepoStatus     *models.RepoStatus
	CustomDiff           *models.DiffResponse
	CustomCommitProposal *models.CommitProposal
	CustomReleaseNotes   *models.ReleaseNotesResponse
	CustomPing           *models.PingResponse
	CustomError          error
}

// NewMockClient instancie un client de test avec les données types du projet.
func NewMockClient(repoRoot string) *MockClient {
	return NewMockClientWithAppRoot(repoRoot, repoRoot)
}

// NewMockClientWithAppRoot instancie un client de test en séparant dépôt cible et application.
func NewMockClientWithAppRoot(repoRoot, appRoot string) *MockClient {
	if appRoot == "" {
		appRoot = repoRoot
	}
	return &MockClient{
		repoRoot:      repoRoot,
		appRoot:       appRoot,
		SimulateDelay: true,
	}
}

func (m *MockClient) sleep(d time.Duration) {
	if m.SimulateDelay {
		time.Sleep(d)
	}
}

// IsMock indique si le mode démo / test est actif.
func (m *MockClient) IsMock() bool {
	return true
}

// GetRepoStatus renvoie l'état du dépôt Git.
func (m *MockClient) GetRepoStatus() (*models.RepoStatus, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	if m.CustomRepoStatus != nil {
		return m.CustomRepoStatus, nil
	}

	return &models.RepoStatus{
		IsGitRepo:     true,
		RepoPath:      m.repoRoot,
		Branch:        "interface-cli",
		StagedCount:   3,
		UnstagedCount: 1,
	}, nil
}

// GetDiff renvoie les fichiers indexés et les indicateurs de filtrage.
func (m *MockClient) GetDiff() (*models.DiffResponse, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	if m.CustomDiff != nil {
		return m.CustomDiff, nil
	}

	m.sleep(200 * time.Millisecond)
	return &models.DiffResponse{
		Success:     true,
		StagedCount: 3,
		Files: []models.DiffFile{
			{
				Path:             "src/auth/login.py",
				Status:           "M",
				Added:            24,
				Removed:          3,
				Binary:           false,
				HasSecretsMasked: false,
			},
			{
				Path:             "src/auth/token_vault.py",
				Status:           "A",
				Added:            58,
				Removed:          0,
				Binary:           false,
				HasSecretsMasked: true, // Données sensibles masquées
			},
			{
				Path:             "assets/banner.png",
				Status:           "A",
				Added:            0,
				Removed:          0,
				Binary:           true, // Fichier binaire exclu
				HasSecretsMasked: false,
			},
		},
	}, nil
}

// StageAll indexe l'ensemble des modifications locales.
func (m *MockClient) StageAll() (*models.ActionResult, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	m.sleep(300 * time.Millisecond)
	return &models.ActionResult{
		Success: true,
		Message: "Tous les fichiers ont été indexés (git add -A)",
	}, nil
}

// GenerateCommit renvoie une proposition de commit au format Conventional Commits.
func (m *MockClient) GenerateCommit(feedback string) (*models.CommitProposal, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	if m.CustomCommitProposal != nil {
		return m.CustomCommitProposal, nil
	}

	m.sleep(1400 * time.Millisecond)

	if strings.TrimSpace(feedback) != "" {
		return &models.CommitProposal{
			Success: true,
			Type:    "refactor",
			Scope:   "auth",
			Subject: "simplification du flux d'authentification et masquage des secrets",
			Body: fmt.Sprintf("Ajusté suite à la consigne : \"%s\".\nRevue des fonctions de token et séparation des couches.", feedback),
			RawFormatted: "refactor(auth): simplification du flux d'authentification et masquage des secrets",
		}, nil
	}

	return &models.CommitProposal{
		Success: true,
		Type:    "feat",
		Scope:   "auth",
		Subject: "ajout de l'authentification OAuth2 sécurisée",
		Body:    "Intègre la gestion des jetons d'accès et filtre automatiquement les variables d'environnement sensibles.",
		RawFormatted: "feat(auth): ajout de l'authentification OAuth2 sécurisée\n\nIntègre la gestion des jetons d'accès et filtre automatiquement les variables d'environnement sensibles.",
	}, nil
}

// ApplyCommit applique le commit sur la branche courante.
func (m *MockClient) ApplyCommit(message string) (*models.ActionResult, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	m.sleep(400 * time.Millisecond)
	return &models.ActionResult{
		Success: true,
		Sha:     "7a9e2f4",
		Message: "Commit appliqué avec succès sur la branche courante.",
	}, nil
}

// Push envoie les modifications vers la branche distante.
func (m *MockClient) Push() (*models.ActionResult, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	m.sleep(800 * time.Millisecond)
	return &models.ActionResult{
		Success: true,
		Message: "Modifications poussées avec succès vers origin/interface-cli.",
	}, nil
}

// GetReleaseNotes synthétise les notes de version Markdown entre deux tags.
func (m *MockClient) GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	if m.CustomReleaseNotes != nil {
		return m.CustomReleaseNotes, nil
	}

	m.sleep(1600 * time.Millisecond)
	md := fmt.Sprintf(`# Release Notes (%s ➔ %s)
Date : %s

### 🚀 Nouveautés (Features)
- **feat(cli)**: interface utilisateur interactive en Go avec moteur ANSI TrueColor.
- **feat(auth)**: support du renouvellement automatique des sessions OAuth2.

### 🐛 Corrections de bogues (Fixes)
- **fix(git)**: filtrage et masquage strict des clés API et fichiers .env sensibles.
- **fix(parser)**: exclusion propre des fichiers binaires du calcul de diff.

### 📚 Documentation & Maintenance
- **docs**: mise à jour des guides d'architecture et du Trello d'équipe.
`, fromTag, toTag, time.Now().Format("2006-01-02"))

	return &models.ReleaseNotesResponse{
		Success:     true,
		FromTag:     fromTag,
		ToTag:       toTag,
		Markdown:    md,
		CommitCount: 14,
	}, nil
}

// GetConfig lit la configuration actuelle depuis le fichier .env applicatif.
func (m *MockClient) GetConfig() (*models.ConfigResponse, error) {
	return LoadConfigFromEnv(m.appRoot, true), nil
}

// SaveConfig met à jour le fichier .env applicatif.
func (m *MockClient) SaveConfig(baseURL, model string, timeout float64, lang string, mockInterface bool) (*models.ActionResult, error) {
	envPath := FindAppEnvPath(m.appRoot)
	if err := WriteEnvFile(envPath, baseURL, model, timeout, lang, mockInterface); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env applicatif : %w", err)
	}

	m.sleep(200 * time.Millisecond)
	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Configuration sauvegardée dans %s (MOCK_INTERFACE=%t).", filepath.Base(envPath), mockInterface),
	}, nil
}

// PingOllama vérifie la joignabilité du serveur Ollama.
func (m *MockClient) PingOllama() (*models.PingResponse, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	if m.CustomPing != nil {
		return m.CustomPing, nil
	}

	m.sleep(500 * time.Millisecond)
	return &models.PingResponse{
		Success:   true,
		Reachable: true,
		LatencyMs: 42,
		InstalledModels: []string{
			"gemma4:12b",
			"gemma4:26b",
			"mistral:latest",
			"codellama:7b",
		},
	}, nil
}

// SaveMarkdownFile enregistre localement le fichier Markdown exporté.
func (m *MockClient) SaveMarkdownFile(filename, content string) error {
	path := filepath.Join(m.repoRoot, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

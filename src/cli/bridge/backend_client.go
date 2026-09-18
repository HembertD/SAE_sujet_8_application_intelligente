package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sae-git-cli/models"
)

// BackendClient orchestre les appels système vers le back-end Python.
// Il ne contient AUCUNE logique métier et se contente de désérialiser le JSON renvoyé par Python.
type BackendClient struct {
	pythonBin   string
	moduleName  string
	repoRoot    string
	UseMockMode bool
}

// NewBackendClient crée un nouveau client de pont vers Python.
func NewBackendClient(repoRoot string, forceDemo bool) *BackendClient {
	pyBin := "python3"
	if customPy := os.Getenv("PYTHON_BIN"); customPy != "" {
		pyBin = customPy
	}

	client := &BackendClient{
		pythonBin:   pyBin,
		moduleName:  "git_commit_release_notes_generator.service.core",
		repoRoot:    repoRoot,
		UseMockMode: forceDemo,
	}

	// Si le backend Python n'est pas encore présent ou non fonctionnel, on bascule en mode simulation
	if !forceDemo && !client.isBackendAvailable() {
		client.UseMockMode = true
	}

	return client
}

// isBackendAvailable teste rapidement si le point d'entrée Python répond
func (b *BackendClient) isBackendAvailable() bool {
	cmd := exec.Command(b.pythonBin, "-m", b.moduleName, "--action", "ping-backend")
	cmd.Dir = b.repoRoot
	err := cmd.Run()
	return err == nil
}

// runPythonCommand exécute la commande Python et récupère stdout
func (b *BackendClient) runPythonCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(b.pythonBin, append([]string{"-m", b.moduleName}, args...)...)
	cmd.Dir = b.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%v (détail stderr : %s)", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// GetRepoStatus interroge le backend pour connaître l'état du dépôt Git
func (b *BackendClient) GetRepoStatus() (*models.RepoStatus, error) {
	if b.UseMockMode {
		return &models.RepoStatus{
			IsGitRepo:     true,
			RepoPath:      b.repoRoot,
			Branch:        "interface-cli",
			StagedCount:   3,
			UnstagedCount: 1,
		}, nil
	}

	out, err := b.runPythonCommand("--action", "status")
	if err != nil {
		return nil, err
	}

	var status models.RepoStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("impossible de décoder le statut renvoyé par Python : %w", err)
	}
	return &status, nil
}

// GetDiff demande à Python l'analyse et la sanitisation des fichiers staged
func (b *BackendClient) GetDiff() (*models.DiffResponse, error) {
	if b.UseMockMode {
		// Simulation basée sur les situations de test de l'équipe (commit_sensitive & feat)
		time.Sleep(200 * time.Millisecond)
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
					HasSecretsMasked: true, // Démonstration du filtrage de sécurité de Dorian
				},
				{
					Path:             "assets/banner.png",
					Status:           "A",
					Added:            0,
					Removed:          0,
					Binary:           true, // Démonstration de l'exclusion binaire
					HasSecretsMasked: false,
				},
			},
		}, nil
	}

	out, err := b.runPythonCommand("--action", "diff")
	if err != nil {
		return nil, err
	}

	var diffResp models.DiffResponse
	if err := json.Unmarshal(out, &diffResp); err != nil {
		return nil, fmt.Errorf("erreur de parsing du diff JSON : %w", err)
	}
	return &diffResp, nil
}

// StageAll demande à Python d'exécuter l'indexation de tous les fichiers
func (b *BackendClient) StageAll() (*models.ActionResult, error) {
	if b.UseMockMode {
		time.Sleep(300 * time.Millisecond)
		return &models.ActionResult{
			Success: true,
			Message: "Tous les fichiers ont été indexés (git add -A)",
		}, nil
	}

	out, err := b.runPythonCommand("--action", "stage-all")
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GenerateCommit demande au backend Python d'interroger le LLM avec le diff
func (b *BackendClient) GenerateCommit(feedback string) (*models.CommitProposal, error) {
	if b.UseMockMode {
		// Simule le temps de réponse LLM sur le serveur IUT
		time.Sleep(1400 * time.Millisecond)

		if strings.TrimSpace(feedback) != "" {
			// Si un retour utilisateur a été fourni
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

	args := []string{"--action", "generate-commit"}
	if feedback != "" {
		args = append(args, "--feedback", feedback)
	}

	out, err := b.runPythonCommand(args...)
	if err != nil {
		return nil, err
	}

	var proposal models.CommitProposal
	if err := json.Unmarshal(out, &proposal); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la proposition LLM : %w", err)
	}
	return &proposal, nil
}

// ApplyCommit demande à Python d'effectuer le commit Git
func (b *BackendClient) ApplyCommit(message string) (*models.ActionResult, error) {
	if b.UseMockMode {
		time.Sleep(400 * time.Millisecond)
		return &models.ActionResult{
			Success: true,
			Sha:     "7a9e2f4",
			Message: "Commit appliqué avec succès sur la branche courante.",
		}, nil
	}

	out, err := b.runPythonCommand("--action", "apply-commit", "--message", message)
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Push demande à Python d'exécuter le git push
func (b *BackendClient) Push() (*models.ActionResult, error) {
	if b.UseMockMode {
		time.Sleep(800 * time.Millisecond)
		return &models.ActionResult{
			Success: true,
			Message: "Modifications poussées avec succès vers origin/interface-cli.",
		}, nil
	}

	out, err := b.runPythonCommand("--action", "push")
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetReleaseNotes demande à Python de générer les notes de version entre deux tags
func (b *BackendClient) GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error) {
	if b.UseMockMode {
		time.Sleep(1600 * time.Millisecond)
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

	out, err := b.runPythonCommand("--action", "release-notes", "--from", fromTag, "--to", toTag)
	if err != nil {
		return nil, err
	}

	var res models.ReleaseNotesResponse
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetConfig demande à Python de lire la configuration active (.env)
func (b *BackendClient) GetConfig() (*models.ConfigResponse, error) {
	if b.UseMockMode {
		return &models.ConfigResponse{
			Success:       true,
			OllamaBaseURL: "http://10.22.28.190:11434",
			OllamaModel:   "gemma4:12b",
			TimeoutS:      60.0,
			Language:      "fr",
		}, nil
	}

	out, err := b.runPythonCommand("--action", "get-config")
	if err != nil {
		return nil, err
	}

	var cfg models.ConfigResponse
	if err := json.Unmarshal(out, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig demande à Python d'écrire les nouveaux paramètres dans le .env
func (b *BackendClient) SaveConfig(baseURL, model string, timeout float64, lang string) (*models.ActionResult, error) {
	if b.UseMockMode {
		time.Sleep(300 * time.Millisecond)
		return &models.ActionResult{
			Success: true,
			Message: "Configuration sauvegardée avec succès dans le fichier .env.",
		}, nil
	}

	out, err := b.runPythonCommand(
		"--action", "save-config",
		"--base-url", baseURL,
		"--model", model,
		"--timeout", fmt.Sprintf("%f", timeout),
		"--language", lang,
	)
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// PingOllama demande à Python de tester la connexion avec le serveur Ollama de l'IUT
func (b *BackendClient) PingOllama() (*models.PingResponse, error) {
	if b.UseMockMode {
		time.Sleep(500 * time.Millisecond)
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

	out, err := b.runPythonCommand("--action", "ping-ollama")
	if err != nil {
		return nil, err
	}

	var ping models.PingResponse
	if err := json.Unmarshal(out, &ping); err != nil {
		return nil, err
	}
	return &ping, nil
}

// SaveMarkdownFile écrit localement le changelog exporté
func (b *BackendClient) SaveMarkdownFile(filename, content string) error {
	path := filepath.Join(b.repoRoot, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

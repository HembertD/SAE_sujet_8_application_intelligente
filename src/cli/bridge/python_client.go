package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sae-git-cli/models"
)

type PythonClient struct {
	pythonBin  string
	moduleName string
	repoRoot   string
	appRoot    string
}

// Instancie le client Python en résolvant la racine applicative.
func NewPythonClient(repoRoot string) *PythonClient {
	appRoot := FindAppRoot(repoRoot)
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	return NewPythonClientWithAppRoot(repoRoot, appRoot)
}

// Instancie le client Python avec racines explicites et interpréteur résolu.
func NewPythonClientWithAppRoot(repoRoot, appRoot string) *PythonClient {
	if appRoot == "" {
		appRoot = FindAppRoot(repoRoot)
	}
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	pyBin := "python3"
	if customPy := os.Getenv("PYTHON_BIN"); customPy != "" {
		pyBin = customPy
	} else if home, err := os.UserHomeDir(); err == nil {
		monenvPy := filepath.Join(home, "monenv", "bin", "python")
		if _, statErr := os.Stat(monenvPy); statErr == nil {
			pyBin = monenvPy
		}
	}

	return &PythonClient{
		pythonBin:  pyBin,
		moduleName: "git_commit_release_notes_generator.service.core",
		repoRoot:   repoRoot,
		appRoot:    appRoot,
	}
}

// Exécute le module Python et capture sa sortie standard.
func (p *PythonClient) runPythonCommand(args ...string) ([]byte, error) {
	fullArgs := []string{"-m", p.moduleName}
	hasRepoArg := false
	for _, a := range args {
		if a == "--repo" {
			hasRepoArg = true
			break
		}
	}
	fullArgs = append(fullArgs, args...)
	if !hasRepoArg && p.repoRoot != "" {
		fullArgs = append(fullArgs, "--repo", p.repoRoot)
	}

	cmd := exec.Command(p.pythonBin, fullArgs...)
	cmd.Dir = p.repoRoot

	env := os.Environ()

	appSrc := filepath.Join(p.appRoot, "src")
	if stat, err := os.Stat(appSrc); err == nil && stat.IsDir() {
		existing := os.Getenv("PYTHONPATH")
		if existing != "" {
			env = append(env, fmt.Sprintf("PYTHONPATH=%s:%s", appSrc, existing))
		} else {
			env = append(env, fmt.Sprintf("PYTHONPATH=%s", appSrc))
		}
	} else {
		candidates := []string{
			filepath.Join(p.repoRoot, "src"),
			filepath.Join(p.repoRoot, "SAE_sujet_8_application_intelligente", "src"),
			filepath.Join(".", "src"),
			filepath.Join("..", "src"),
		}
		for _, cand := range candidates {
			if abs, err := filepath.Abs(cand); err == nil {
				if stat, err := os.Stat(abs); err == nil && stat.IsDir() {
					existing := os.Getenv("PYTHONPATH")
					if existing != "" {
						env = append(env, fmt.Sprintf("PYTHONPATH=%s:%s", abs, existing))
					} else {
						env = append(env, fmt.Sprintf("PYTHONPATH=%s", abs))
					}
					break
				}
			}
		}
	}

	cfg := LoadConfigFromEnv(p.appRoot, false)
	mockAIStr := "false"
	if cfg.MockAI {
		mockAIStr = "true"
	}
	env = append(env,
		fmt.Sprintf("OLLAMA_BASE_URL=%s", cfg.OllamaBaseURL),
		fmt.Sprintf("OLLAMA_MODEL=%s", cfg.OllamaModel),
		fmt.Sprintf("OLLAMA_TIMEOUT_S=%.0f", cfg.TimeoutS),
		fmt.Sprintf("APP_LANGUAGE=%s", cfg.Language),
		fmt.Sprintf("MOCK_AI=%s", mockAIStr),
		fmt.Sprintf("SMART_COMMIT_APP_ROOT=%s", p.appRoot),
	)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%v (détail stderr : %s)", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// Indique que le client réel n'est jamais en simulation.
func (p *PythonClient) IsMock() bool {
	return false
}

// Interroge le backend Python pour le statut du dépôt.
func (p *PythonClient) GetRepoStatus() (*models.RepoStatus, error) {
	out, err := p.runPythonCommand("--action", "status")
	if err != nil {
		return nil, err
	}

	var status models.RepoStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("impossible de décoder le statut renvoyé par Python : %w", err)
	}
	return &status, nil
}

// Interroge le backend Python pour le diff.
func (p *PythonClient) GetDiff() (*models.DiffResponse, error) {
	out, err := p.runPythonCommand("--action", "diff")
	if err != nil {
		return nil, err
	}

	var diffResp models.DiffResponse
	if err := json.Unmarshal(out, &diffResp); err != nil {
		return nil, fmt.Errorf("erreur de parsing du diff JSON : %w", err)
	}
	if !diffResp.Success && diffResp.Error != "" {
		return nil, fmt.Errorf("%s", diffResp.Error)
	}
	return &diffResp, nil
}

// Demande au backend Python d'indexer tous les fichiers.
func (p *PythonClient) StageAll() (*models.ActionResult, error) {
	out, err := p.runPythonCommand("--action", "stage-all")
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Demande au backend Python une proposition de commit.
func (p *PythonClient) GenerateCommit(feedback string) (*models.CommitProposal, error) {
	args := []string{"--action", "generate-commit"}
	if strings.TrimSpace(feedback) != "" {
		args = append(args, "--feedback", feedback)
	}

	out, err := p.runPythonCommand(args...)
	if err != nil {
		return nil, err
	}

	var proposal models.CommitProposal
	if err := json.Unmarshal(out, &proposal); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la proposition LLM : %w", err)
	}
	if !proposal.Success {
		errMsg := proposal.Error
		if errMsg == "" {
			errMsg = "erreur : aucune réponse du serveur LLM (timeout)"
		}
		return &proposal, fmt.Errorf("%s", errMsg)
	}
	return &proposal, nil
}

// Demande au backend Python d'appliquer le commit.
func (p *PythonClient) ApplyCommit(message string) (*models.ActionResult, error) {
	out, err := p.runPythonCommand("--action", "apply-commit", "--message", message)
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Demande au backend Python de pousser vers le distant.
func (p *PythonClient) Push() (*models.ActionResult, error) {
	out, err := p.runPythonCommand("--action", "push")
	if err != nil {
		return nil, err
	}

	var res models.ActionResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Demande au backend Python les notes de version.
func (p *PythonClient) GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error) {
	out, err := p.runPythonCommand("--action", "release-notes", "--from", fromTag, "--to", toTag)
	if err != nil {
		return nil, err
	}

	var res models.ReleaseNotesResponse
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, err
	}
	if !res.Success && res.Error != "" {
		return &res, fmt.Errorf("%s", res.Error)
	}
	return &res, nil
}

// Charge la configuration réelle depuis l'environnement.
func (p *PythonClient) GetConfig() (*models.ConfigResponse, error) {
	return LoadConfigFromEnv(p.appRoot, false), nil
}

// Persiste la configuration puis notifie le backend Python.
func (p *PythonClient) SaveConfig(baseURL, model string, timeout float64, lang string, mockInterface bool, mockAI ...bool) (*models.ActionResult, error) {
	envPath := FindAppEnvPath(p.appRoot)
	if err := WriteEnvFile(envPath, baseURL, model, timeout, lang, mockInterface, mockAI...); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env applicatif : %w", err)
	}

	_, _ = p.runPythonCommand(
		"--action", "save-config",
		"--base-url", baseURL,
		"--model", model,
		"--timeout", fmt.Sprintf("%f", timeout),
		"--language", lang,
	)

	mockAIVal := ReadMockAISettingFromEnv(p.appRoot)
	if len(mockAI) > 0 {
		mockAIVal = mockAI[0]
	}

	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Configuration sauvegardée dans %s (MOCK_INTERFACE=%t, MOCK_AI=%t).", filepath.Base(envPath), mockInterface, mockAIVal),
	}, nil
}

// Interroge le backend Python sur l'état d'Ollama.
func (p *PythonClient) PingOllama() (*models.PingResponse, error) {
	out, err := p.runPythonCommand("--action", "ping-ollama")
	if err != nil {
		return nil, err
	}

	var ping models.PingResponse
	if err := json.Unmarshal(out, &ping); err != nil {
		return nil, err
	}
	return &ping, nil
}

// Sauvegarde le contenu Markdown dans le dépôt cible.
func (p *PythonClient) SaveMarkdownFile(filename, content string) error {
	path := filepath.Join(p.repoRoot, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

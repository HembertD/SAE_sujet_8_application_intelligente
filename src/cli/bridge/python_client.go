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

// PythonClient implémente BackendClient en communiquant avec le sous-processus Python.
// Il ne contient AUCUNE logique métier et délègue l'intégralité du traitement à Python.
type PythonClient struct {
	pythonBin  string
	moduleName string
	repoRoot   string
}

// NewPythonClient instancie un client communiquant avec le backend Python réel.
func NewPythonClient(repoRoot string) *PythonClient {
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
	}
}

// runPythonCommand exécute la commande Python et récupère stdout
func (p *PythonClient) runPythonCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(p.pythonBin, append([]string{"-m", p.moduleName}, args...)...)
	cmd.Dir = p.repoRoot

	env := os.Environ()
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
	cfg := LoadConfigFromEnv(p.repoRoot, false)
	env = append(env,
		fmt.Sprintf("OLLAMA_BASE_URL=%s", cfg.OllamaBaseURL),
		fmt.Sprintf("OLLAMA_MODEL=%s", cfg.OllamaModel),
		fmt.Sprintf("OLLAMA_TIMEOUT_S=%.0f", cfg.TimeoutS),
		fmt.Sprintf("APP_LANGUAGE=%s", cfg.Language),
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

// IsMock indique si le client est en mode test.
func (p *PythonClient) IsMock() bool {
	return false
}

// GetRepoStatus interroge le backend pour connaître l'état du dépôt Git
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

// GetDiff demande à Python l'analyse et la sanitisation des fichiers staged
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

// StageAll demande à Python d'exécuter l'indexation de tous les fichiers
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

// GenerateCommit demande au backend Python d'interroger le LLM avec le diff
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

// ApplyCommit demande à Python d'effectuer le commit Git
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

// Push demande à Python d'exécuter le git push
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

// GetReleaseNotes demande à Python de générer les notes de version entre deux tags
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

// GetConfig lit la configuration depuis le fichier .env
func (p *PythonClient) GetConfig() (*models.ConfigResponse, error) {
	return LoadConfigFromEnv(p.repoRoot, false), nil
}

// SaveConfig enregistre la configuration dans le fichier .env et notifie Python
func (p *PythonClient) SaveConfig(baseURL, model string, timeout float64, lang string, mockInterface bool) (*models.ActionResult, error) {
	envPath := FindEnvPath(p.repoRoot)
	if err := WriteEnvFile(envPath, baseURL, model, timeout, lang, mockInterface); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env : %w", err)
	}

	_, _ = p.runPythonCommand(
		"--action", "save-config",
		"--base-url", baseURL,
		"--model", model,
		"--timeout", fmt.Sprintf("%f", timeout),
		"--language", lang,
	)

	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Configuration sauvegardée dans %s (MOCK_INTERFACE=%t).", filepath.Base(envPath), mockInterface),
	}, nil
}

// PingOllama teste la connexion avec le serveur Ollama de l'IUT
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

// SaveMarkdownFile écrit localement le changelog exporté
func (p *PythonClient) SaveMarkdownFile(filename, content string) error {
	path := filepath.Join(p.repoRoot, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

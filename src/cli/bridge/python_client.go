package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sae-git-cli/models"
)

// PythonClient implémente BackendClient en communiquant avec le sous-processus Python.
// Il ne contient AUCUNE logique métier et délègue l'intégralité du traitement à Python.
type PythonClient struct {
	pythonBin    string
	pythonBinErr error
	moduleName   string
	repoRoot     string
	appRoot      string
}

// NewPythonClient instancie un client communiquant avec le backend Python réel.
func NewPythonClient(repoRoot string) *PythonClient {
	appRoot := FindAppRoot(repoRoot)
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	return NewPythonClientWithAppRoot(repoRoot, appRoot)
}

// NewPythonClientWithAppRoot instancie un client en précisant la racine applicative.
//
// Si aucun interpréteur Python valide n'est résolu (cf. resolvePythonBin), la
// construction n'échoue pas immédiatement : l'erreur est conservée et renvoyée
// au premier appel réel (runPythonCommand). C'est CheckPythonEnvironment,
// appelé explicitement au démarrage du CLI (main.go), qui doit stopper le
// programme avant d'en arriver là quand le mode mock-interface n'est pas actif.
func NewPythonClientWithAppRoot(repoRoot, appRoot string) *PythonClient {
	if appRoot == "" {
		appRoot = FindAppRoot(repoRoot)
	}
	if isTestDir(repoRoot) {
		appRoot = repoRoot
	}
	pyBin, err := resolvePythonBin(appRoot)

	return &PythonClient{
		pythonBin:    pyBin,
		pythonBinErr: err,
		moduleName:   "git_commit_release_notes_generator.service.core",
		repoRoot:     repoRoot,
		appRoot:      appRoot,
	}
}

// CheckPythonEnvironment vérifie qu'un interpréteur Python valide est résolu
// pour appRoot, sans construire de client. Destiné à un contrôle explicite au
// démarrage du CLI (main.go), pour un échec immédiat et clair plutôt qu'un
// plantage différé (ModuleNotFoundError) lors de la première action réelle.
func CheckPythonEnvironment(appRoot string) error {
	_, err := resolvePythonBin(appRoot)
	return err
}

// resolvePythonBin détermine quel interpréteur Python utiliser, par ordre de priorité :
//  1. PYTHON_BIN (variable d'environnement système, ou définie dans le .env applicatif) ;
//  2. l'environnement virtuel autonome créé dans les fichiers de l'application
//     (<appRoot>/.venv, via scripts/setup_python_env.py).
//
// Aucun repli sur l'interpréteur système : un repli silencieux ("python3" sous
// Unix, "python" sous Windows) ferait démarrer l'application même sans
// dépendances installées, pour planter plus tard avec un ModuleNotFoundError
// incompréhensible. On préfère échouer tout de suite, explicitement.
func resolvePythonBin(appRoot string) (string, error) {
	if customPy := ReadPythonBinFromEnv(appRoot); customPy != "" {
		return customPy, nil
	}

	venvPy := filepath.Join(appRoot, ".venv", "Scripts", "python.exe")
	if runtime.GOOS != "windows" {
		venvPy = filepath.Join(appRoot, ".venv", "bin", "python")
	}
	if _, err := os.Stat(venvPy); err == nil {
		return venvPy, nil
	}

	return "", fmt.Errorf(
		"Aucun environnement Python valide n'a été trouvé.\n" +
			"Pour initialiser l'environnement de l'application, exécutez :\n" +
			"    python scripts/setup_python_env.py\n\n" +
			"Pour utiliser un interpréteur existant, renseignez PYTHON_BIN dans votre .env :\n" +
			"    PYTHON_BIN=/chemin/vers/votre/python",
	)
}

// runPythonCommand exécute la commande Python et récupère stdout
func (p *PythonClient) runPythonCommand(args ...string) ([]byte, error) {
	if p.pythonBinErr != nil {
		return nil, p.pythonBinErr
	}

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

	// 1. PYTHONPATH pointe sur le dossier src de l'application.
	// os.PathListSeparator vaut ';' sous Windows et ':' ailleurs : un ':' en
	// dur cassait PYTHONPATH sous Windows dès qu'une valeur préexistait.
	pathSep := string(os.PathListSeparator)
	appSrc := filepath.Join(p.appRoot, "src")
	if stat, err := os.Stat(appSrc); err == nil && stat.IsDir() {
		existing := os.Getenv("PYTHONPATH")
		if existing != "" {
			env = append(env, fmt.Sprintf("PYTHONPATH=%s%s%s", appSrc, pathSep, existing))
		} else {
			env = append(env, fmt.Sprintf("PYTHONPATH=%s", appSrc))
		}
	} else {
		// Candidats de secours si appSrc non résolu
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
						env = append(env, fmt.Sprintf("PYTHONPATH=%s%s%s", abs, pathSep, existing))
					} else {
						env = append(env, fmt.Sprintf("PYTHONPATH=%s", abs))
					}
					break
				}
			}
		}
	}

	// 2. Configuration chargée exclusivement depuis l'application (p.appRoot, JAMAIS p.repoRoot)
	cfg := LoadConfigFromEnv(p.appRoot, false)
	mockAIStr := "false"
	if cfg.MockAI {
		mockAIStr = "true"
	}
	env = append(env,
		fmt.Sprintf("OLLAMA_BASE_URL=%s", cfg.OllamaBaseURL),
		fmt.Sprintf("OLLAMA_MODEL=%s", cfg.OllamaModel),
		fmt.Sprintf("OLLAMA_TIMEOUT_S=%.0f", cfg.TimeoutS),
		fmt.Sprintf("MOCK_AI=%s", mockAIStr),
		fmt.Sprintf("GITELLIGENCE_APP_ROOT=%s", p.appRoot),
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

// GetConfig lit la configuration depuis le fichier .env de l'application
func (p *PythonClient) GetConfig() (*models.ConfigResponse, error) {
	return LoadConfigFromEnv(p.appRoot, false), nil
}

// SaveConfig enregistre la configuration dans le fichier .env applicatif et notifie Python
func (p *PythonClient) SaveConfig(baseURL, model string, timeout float64, mockInterface bool, mockAI ...bool) (*models.ActionResult, error) {
	envPath := FindAppEnvPath(p.appRoot)
	if err := WriteEnvFile(envPath, baseURL, model, timeout, mockInterface, mockAI...); err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans le fichier .env applicatif : %w", err)
	}

	_, _ = p.runPythonCommand(
		"--action", "save-config",
		"--base-url", baseURL,
		"--model", model,
		"--timeout", fmt.Sprintf("%f", timeout),
	)

	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Configuration sauvegardée dans %s.", filepath.Base(envPath)),
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

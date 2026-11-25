package main

import (
	"context"
	"fmt"
	stdRuntime "runtime"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/ui"
	"sae-git-cli/views"
)

// App structure centrale orchestrant l'application Wails et le pontage vers le backend Python.
type App struct {
	ctx                context.Context
	repoRoot           string
	appRoot            string
	devMode            bool
	forceMockInterface bool
	forceMockIA        bool
	forceDockerLLM     bool
	client             bridge.BackendClient
	mu                 sync.RWMutex
}

// NewApp instancie une nouvelle structure d'application Wails avec les paramètres spécifiés.
func NewApp(repoRoot, appRoot string, devMode, forceMockInterface, forceMockIA, forceDockerLLM bool) *App {
	app := &App{
		repoRoot:           repoRoot,
		appRoot:            appRoot,
		devMode:            devMode,
		forceMockInterface: forceMockInterface,
		forceMockIA:        forceMockIA,
		forceDockerLLM:     forceDockerLLM,
	}
	app.initClient()
	return app
}

// startup est appelé par Wails au démarrage du cycle de vie de l'application.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	ui.SetNotificationsEnabled(bridge.ReadEnableNotificationsFromEnv(a.appRoot))
}

// initClient initialise ou réinitialise le BackendClient en fonction du dépôt actuel.
func (a *App) initClient() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.appRoot == "" {
		a.appRoot = bridge.FindAppRoot()
	}
	a.client = bridge.NewBackendClientWithAppRoot(
		a.repoRoot,
		a.appRoot,
		a.forceMockInterface,
		a.forceMockIA,
		a.forceDockerLLM,
	)
}

// ChooseDirectory ouvre un sélecteur natif de répertoire du système d'exploitation.
func (a *App) ChooseDirectory() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("contexte wails non initialisé")
	}
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Sélectionner un dépôt Git",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// SetTargetRepo modifie le dépôt cible analysé et réinitialise le client de pontage.
func (a *App) SetTargetRepo(repoPath string) (*models.RepoStatus, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("chemin de dépôt vide")
	}
	a.repoRoot = repoPath
	a.initClient()
	return a.GetRepoStatus()
}

// GetTargetRepo renvoie le chemin du dépôt Git actuellement analysé.
func (a *App) GetTargetRepo() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.repoRoot
}

// GetAppRoot renvoie le chemin de la racine de l'application Gitelligence.
func (a *App) GetAppRoot() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.appRoot
}

// IsDevMode indique si le mode développeur est actif.
func (a *App) IsDevMode() bool {
	return a.devMode
}

// GetRepoStatus interroge l'état du dépôt Git courant.
func (a *App) GetRepoStatus() (*models.RepoStatus, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.GetRepoStatus()
}

// GetDiff récupère la liste des fichiers modifiés et les patchs de diff.
func (a *App) GetDiff() (*models.DiffResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.GetDiff()
}

// StageAll indexe l'ensemble des modifications dans le staging area Git (git add -A).
func (a *App) StageAll() (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.StageAll()
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Tous les fichiers ont été indexés (git add -A)")
	}
	return res, err
}

// GenerateCommit sollicite le backend et le LLM pour générer une proposition Conventional Commits.
func (a *App) GenerateCommit(feedback string) (*models.CommitProposal, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}

	// Vérification intelligente de l'index Git (staging)
	status, err := client.GetRepoStatus()
	if err == nil && status != nil {
		if status.StagedCount == 0 && status.UnstagedCount > 0 {
			// Auto-staging si rien n'est indexé mais que des fichiers modifiés sont présents
			_, _ = client.StageAll()
		} else if status.StagedCount == 0 && status.UnstagedCount == 0 {
			errEmpty := fmt.Errorf("aucune modification détectée dans le dépôt (ni indexée, ni non-indexée)")
			ui.NotifyCommitGenerated(false, errEmpty.Error())
			return nil, errEmpty
		}
	}

	proposal, err := client.GenerateCommit(feedback)
	if err != nil {
		ui.NotifyCommitGenerated(false, err.Error())
		return nil, err
	}
	if proposal != nil && !proposal.Success {
		ui.NotifyCommitGenerated(false, proposal.Error)
	} else {
		ui.NotifyCommitGenerated(true)
	}
	return proposal, nil
}

// ImposeCommitType régénère une proposition de commit en imposant un type Conventional Commit strict (comme l'option [T] du CLI).
func (a *App) ImposeCommitType(chosenType string, comment string) (*models.CommitProposal, error) {
	feedback := views.BuildForcedTypeFeedback(chosenType, comment)
	return a.GenerateCommit(feedback)
}

// ApplyCommit applique et enregistre le commit Git avec le message final validé.
func (a *App) ApplyCommit(message string) (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.ApplyCommit(message)
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Commit créé et enregistré avec succès !")
	}
	return res, err
}

// Push pousse les commits locaux vers le dépôt distant (git push).
func (a *App) Push() (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.Push()
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Branches poussées vers le dépôt distant avec succès !")
	}
	return res, err
}

// GetReleaseNotes génère les release notes Markdown pour la plage de tags ou commits spécifiée.
func (a *App) GetReleaseNotes(fromTag, toTag string) (*models.ReleaseNotesResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	notes, err := client.GetReleaseNotes(fromTag, toTag)
	if err != nil {
		ui.NotifyReleaseNotesGenerated(false, err.Error())
		return nil, err
	}
	if notes != nil && !notes.Success {
		ui.NotifyReleaseNotesGenerated(false, notes.Error)
	} else {
		ui.NotifyReleaseNotesGenerated(true)
	}
	return notes, nil
}

// GetRecentCommits renvoie l'historique des commits récents du dépôt.
func (a *App) GetRecentCommits(limit int) (*models.CommitsResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.GetRecentCommits(limit)
}

// GetConfig charge les options de configuration applicative depuis le .env.
func (a *App) GetConfig() (*models.ConfigResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.GetConfig()
}

// SaveConfig enregistre les modifications de configuration dans le .env et réinitialise le client.
func (a *App) SaveConfig(baseURL, model string, timeout float64, mockInterface bool, dockerLLM bool, mockAI bool) (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.SaveConfigWithDockerLLM(baseURL, model, timeout, mockInterface, dockerLLM, mockAI)
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Configuration mise à jour dans le .env")
		a.initClient()
	}
	return res, err
}

// SaveNotifications enregistre le paramètre ENABLE_NOTIFICATIONS dans le .env et met à jour l'état runtime.
func (a *App) SaveNotifications(enabled bool) (*models.ActionResult, error) {
	a.mu.RLock()
	appRoot := a.appRoot
	a.mu.RUnlock()
	ui.SetNotificationsEnabled(enabled)
	if err := bridge.WriteNotificationsSetting(appRoot, enabled); err != nil {
		return &models.ActionResult{Success: false, Error: err.Error()}, err
	}
	return &models.ActionResult{Success: true, Message: "Paramètre de notifications enregistré."}, nil
}

// VerifyPythonEnvironment teste en profondeur l'environnement Python résolu pour l'application (comme l'option [P] du CLI).
func (a *App) VerifyPythonEnvironment() (*bridge.PythonEnvReport, error) {
	a.mu.RLock()
	appRoot := a.appRoot
	a.mu.RUnlock()
	return bridge.VerifyPythonEnvironment(appRoot)
}

// VerifyPythonCandidate teste un interpréteur Python candidat avant acceptation (comme l'option [M] du CLI).
func (a *App) VerifyPythonCandidate(candidatePyBin string, isAuto bool) (*bridge.PythonEnvReport, error) {
	a.mu.RLock()
	appRoot := a.appRoot
	a.mu.RUnlock()
	return bridge.VerifyPythonCandidate(appRoot, candidatePyBin, isAuto)
}

// SavePythonBin met à jour l'interpréteur Python dans le .env applicatif et réinitialise le client.
func (a *App) SavePythonBin(candidatePyBin string, isAuto bool) (*models.ActionResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	targetBin := candidatePyBin
	if isAuto {
		targetBin = ""
	}

	report, err := bridge.VerifyPythonCandidate(a.appRoot, targetBin, isAuto)
	if err != nil || (report != nil && !report.Success) {
		errMsg := "environnement Python invalide"
		if err != nil {
			errMsg = err.Error()
		} else if report != nil && len(report.MissingDeps) > 0 {
			errMsg = fmt.Sprintf("dépendances manquantes : %v", report.MissingDeps)
		} else if report != nil && report.AppLoadError != "" {
			errMsg = report.AppLoadError
		}
		return &models.ActionResult{
			Success: false,
			Error:   errMsg,
		}, fmt.Errorf("%s", errMsg)
	}

	if errWrite := bridge.WritePythonBinSetting(a.appRoot, targetBin); errWrite != nil {
		return &models.ActionResult{
			Success: false,
			Error:   errWrite.Error(),
		}, errWrite
	}

	// Réinitialise le client pour activer le nouvel interpréteur
	if a.appRoot == "" {
		a.appRoot = bridge.FindAppRoot()
	}
	a.client = bridge.NewBackendClientWithAppRoot(
		a.repoRoot,
		a.appRoot,
		a.forceMockInterface,
		a.forceMockIA,
		a.forceDockerLLM,
	)

	return &models.ActionResult{
		Success: true,
		Message: fmt.Sprintf("Interpréteur Python validé et configuré avec succès (%s, Python %s) !", report.Mode, report.Version),
	}, nil
}

// PingOllama teste la joignabilité et la latence du serveur Ollama.
func (a *App) PingOllama() (*models.PingResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.PingOllama()
}

// SaveReleaseNotesMarkdown enregistre le contenu des release notes dans un fichier Markdown sur le disque.
func (a *App) SaveReleaseNotesMarkdown(filename, content string) error {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return fmt.Errorf("client non initialisé")
	}
	err := client.SaveMarkdownFile(filename, content)
	if err == nil {
		ui.NotifySuccess(fmt.Sprintf("Fichier %s enregistré avec succès !", filename))
	}
	return err
}

// DockerOllamaStatus inspecte l'état du conteneur Ollama Docker Compose.
func (a *App) DockerOllamaStatus() (*models.DockerStatusResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.DockerOllamaStatus()
}

// StartDockerOllama démarre le conteneur Docker Ollama.
func (a *App) StartDockerOllama() (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.StartDockerOllama()
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Conteneur Ollama démarré avec succès !")
	}
	return res, err
}

// StopDockerOllama arrête le conteneur Docker Ollama.
func (a *App) StopDockerOllama() (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.StopDockerOllama()
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Conteneur Ollama arrêté !")
	}
	return res, err
}

// RestartDockerOllama redémarre le conteneur Docker Ollama.
func (a *App) RestartDockerOllama() (*models.ActionResult, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	res, err := client.RestartDockerOllama()
	if err == nil && res != nil && res.Success {
		ui.NotifySuccess("Conteneur Ollama redémarré !")
	}
	return res, err
}

// GetDockerLogs récupère les dernières lignes de log du conteneur Ollama.
func (a *App) GetDockerLogs(tailLines int) (*models.DockerLogsResponse, error) {
	a.mu.RLock()
	client := a.client
	a.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("client non initialisé")
	}
	return client.GetDockerLogs(tailLines)
}

// GetSystemSpecs renvoie des métriques matérielles et d'exécution du système.
func (a *App) GetSystemSpecs() map[string]interface{} {
	var mem stdRuntime.MemStats
	stdRuntime.ReadMemStats(&mem)

	return map[string]interface{}{
		"os":          stdRuntime.GOOS,
		"arch":        stdRuntime.GOARCH,
		"go_version":  stdRuntime.Version(),
		"cpus":        stdRuntime.NumCPU(),
		"goroutines":  stdRuntime.NumGoroutine(),
		"alloc_mb":    fmt.Sprintf("%.2f MB", float64(mem.Alloc)/1024/1024),
		"sys_mb":      fmt.Sprintf("%.2f MB", float64(mem.Sys)/1024/1024),
		"server_time": time.Now().Format("15:04:05"),
		"webview":     "WebKitGTK 4.1 (Wails v2)",
		"app_version": "1.0.0",
	}
}

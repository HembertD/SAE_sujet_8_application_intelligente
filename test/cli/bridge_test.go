package cli_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"sae-git-cli/bridge"
	"sae-git-cli/models"
	"sae-git-cli/views"
)

// Vérifications de conformité d'interface à la compilation
var (
	_ bridge.BackendClient = (*bridge.BridgeClient)(nil)
	_ bridge.BackendClient = (*bridge.PythonClient)(nil)
	_ bridge.BackendClient = (*bridge.MockClient)(nil)
)

func TestMockClientBasics(t *testing.T) {
	mock := bridge.NewMockClient("/tmp/fake-repo")
	mock.SimulateDelay = false // Désactive les délais pour des tests instantanés

	if !mock.IsMock() {
		t.Errorf("attendu IsMock() == true, obtenu false")
	}

	// 1. GetRepoStatus
	status, err := mock.GetRepoStatus()
	if err != nil {
		t.Fatalf("erreur inattendue sur GetRepoStatus : %v", err)
	}
	if !status.IsGitRepo || status.Branch != "interface-cli" || status.StagedCount != 3 {
		t.Errorf("statut inattendu : %+v", status)
	}

	// 2. GetDiff
	diff, err := mock.GetDiff()
	if err != nil {
		t.Fatalf("erreur inattendue sur GetDiff : %v", err)
	}
	if len(diff.Files) != 3 {
		t.Errorf("attendu 3 fichiers dans le diff, obtenu %d", len(diff.Files))
	}
	// Vérifier la détection des secrets masqués
	foundMasked := false
	foundBinary := false
	for _, f := range diff.Files {
		if f.HasSecretsMasked {
			foundMasked = true
		}
		if f.Binary {
			foundBinary = true
		}
	}
	if !foundMasked {
		t.Errorf("le mock doit contenir au moins un fichier avec HasSecretsMasked=true")
	}
	if !foundBinary {
		t.Errorf("le mock doit contenir au moins un fichier binaire")
	}

	// 3. StageAll
	resStage, err := mock.StageAll()
	if err != nil || !resStage.Success {
		t.Errorf("erreur sur StageAll : %v", err)
	}

	// 4. GenerateCommit sans feedback
	proposal, err := mock.GenerateCommit("")
	if err != nil {
		t.Fatalf("erreur GenerateCommit : %v", err)
	}
	if proposal.Type != "feat" || proposal.Scope != "auth" {
		t.Errorf("proposition de commit inattendue : %+v", proposal)
	}

	// 5. GenerateCommit avec consigne utilisateur
	feedback := "Sois plus concis et passe en refactor"
	propRefactor, err := mock.GenerateCommit(feedback)
	if err != nil {
		t.Fatalf("erreur GenerateCommit avec feedback : %v", err)
	}
	if propRefactor.Type != "refactor" {
		t.Errorf("attendu refactor, obtenu %s", propRefactor.Type)
	}
	if !strings.Contains(propRefactor.Body, feedback) {
		t.Errorf("le body doit contenir le feedback '%s'", feedback)
	}

	// 5.b GenerateCommit avec type forcé
	propFix, err := mock.GenerateCommit("type: fix")
	if err != nil || propFix.Type != "fix" {
		t.Errorf("attendu type 'fix', obtenu %s (err: %v)", propFix.Type, err)
	}

	propChore, err := mock.GenerateCommit("c'est un chore pour ranger les fichiers")
	if err != nil || propChore.Type != "chore" {
		t.Errorf("attendu type 'chore', obtenu %s (err: %v)", propChore.Type, err)
	}

	// Piège classique : 'test' sans mot d'intention ne doit pas être imposé comme type
	propNoOverride, err := mock.GenerateCommit("Consigne de test pour l'IA")
	if err != nil || propNoOverride.Type != "refactor" {
		t.Errorf("aucun override attendu sur 'Consigne de test', obtenu %s", propNoOverride.Type)
	}

	// Types contradictoires : aucun override
	propContradiction, err := mock.GenerateCommit("c'est un fix ou c'est un feat")
	if err != nil || propContradiction.Type != "refactor" {
		t.Errorf("aucun override attendu en cas de contradiction, obtenu %s", propContradiction.Type)
	}

	// 6. ApplyCommit
	resCommit, err := mock.ApplyCommit("feat(auth): commit test")
	if err != nil || !resCommit.Success || resCommit.Sha == "" {
		t.Errorf("erreur ApplyCommit : %v, res: %+v", err, resCommit)
	}

	// 7. Push
	resPush, err := mock.Push()
	if err != nil || !resPush.Success {
		t.Errorf("erreur Push : %v", err)
	}

	// 8. GetReleaseNotes
	notes, err := mock.GetReleaseNotes("v0.1.0", "v1.0.0")
	if err != nil || !notes.Success {
		t.Fatalf("erreur GetReleaseNotes : %v", err)
	}
	if !strings.Contains(notes.Markdown, "Release Notes") || !strings.Contains(notes.Markdown, "v0.1.0") {
		t.Errorf("contenu Release Notes inattendu : %s", notes.Markdown)
	}

	// 9. PingOllama
	ping, err := mock.PingOllama()
	if err != nil || !ping.Reachable || len(ping.InstalledModels) == 0 {
		t.Errorf("erreur PingOllama : %v, ping: %+v", err, ping)
	}

	// 10. GetRecentCommits
	commits, err := mock.GetRecentCommits(0)
	if err != nil || !commits.Success || len(commits.Commits) == 0 {
		t.Fatalf("erreur GetRecentCommits : %v, commits: %+v", err, commits)
	}
	if commits.Commits[0].Sha != "02b90a2" {
		t.Errorf("premier commit inattendu : %+v", commits.Commits[0])
	}
}

func TestMockClientOverrides(t *testing.T) {
	mock := bridge.NewMockClient("/tmp/fake-repo")
	mock.SimulateDelay = false

	// Test injection d'erreur
	expectedErr := errors.New("échec de simulation réseau")
	mock.CustomError = expectedErr

	if _, err := mock.GetRepoStatus(); !errors.Is(err, expectedErr) {
		t.Errorf("attendu erreur personnalisée, obtenu : %v", err)
	}
	if _, err := mock.GetDiff(); !errors.Is(err, expectedErr) {
		t.Errorf("attendu erreur personnalisée, obtenu : %v", err)
	}
	if _, err := mock.GenerateCommit(""); !errors.Is(err, expectedErr) {
		t.Errorf("attendu erreur personnalisée, obtenu : %v", err)
	}

	// Réinitialise l'erreur
	mock.CustomError = nil

	// Test surcharge de réponse Diff
	customDiff := &models.DiffResponse{
		Success:     true,
		StagedCount: 1,
		Files: []models.DiffFile{
			{Path: "custom/file.go", Status: "M"},
		},
	}
	mock.CustomDiff = customDiff

	d, err := mock.GetDiff()
	if err != nil || len(d.Files) != 1 || d.Files[0].Path != "custom/file.go" {
		t.Errorf("surcharge CustomDiff non prise en compte : %+v", d)
	}
}

func TestParseBoolFromString(t *testing.T) {
	truthy := []string{"true", "True", "TRUE", "1", "yes", "YES", "oui", "OUI"}
	for _, v := range truthy {
		if !bridge.ParseBoolFromString(v) {
			t.Errorf("attendu true pour '%s'", v)
		}
	}

	falsy := []string{"false", "False", "0", "no", "non", "", "random"}
	for _, v := range falsy {
		if bridge.ParseBoolFromString(v) {
			t.Errorf("attendu false pour '%s'", v)
		}
	}
}

func TestEnvFileReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	// Écriture du fichier .env
	err := bridge.WriteEnvFile(envPath, "http://localhost:11434", "mistral:latest", 45.0, true)
	if err != nil {
		t.Fatalf("WriteEnvFile échoué : %v", err)
	}

	// Lecture de la configuration
	cfg := bridge.LoadConfigFromEnv(tempDir, false)
	if cfg.OllamaBaseURL != "http://localhost:11434" {
		t.Errorf("URL attendue 'http://localhost:11434', obtenue '%s'", cfg.OllamaBaseURL)
	}
	if cfg.OllamaModel != "mistral:latest" {
		t.Errorf("Modèle attendu 'mistral:latest', obtenu '%s'", cfg.OllamaModel)
	}
	if cfg.TimeoutS != 45.0 {
		t.Errorf("Timeout attendu 45, obtenu %.0f", cfg.TimeoutS)
	}
	if !cfg.MockInterface {
		t.Errorf("attendu MockInterface == true")
	}
}

func TestBridgeClientSwitching(t *testing.T) {
	tempDir := t.TempDir()

	// Initialisation en forçant le mode démo
	bridgeDemo := bridge.NewBackendClient(tempDir, true, false)
	if !bridgeDemo.IsMock() {
		t.Errorf("attendu bridgeDemo.IsMock() == true avec forceDemo=true")
	}

	// Initialisation normale
	bridgeNormal := bridge.NewBackendClient(tempDir, false, false)
	// Sans configuration, le mode démo est inactif
	if bridgeNormal.IsMock() {
		t.Errorf("attendu bridgeNormal.IsMock() == false par défaut sans variable d'env")
	}

	// Activation du mode démo
	res, err := bridgeNormal.SaveConfig("http://127.0.0.1:11434", "gemma4:26b", 60, true)
	if err != nil || !res.Success {
		t.Fatalf("SaveConfig échoué : %v", err)
	}

	// Vérifie la prise en compte de la configuration
	if !bridgeNormal.IsMock() {
		t.Errorf("attendu bridgeNormal.IsMock() == true après SaveConfig(mockInterface=true)")
	}

	// Bascule inverse
	_, err = bridgeNormal.SaveConfig("http://127.0.0.1:11434", "gemma4:26b", 60, false)
	if err != nil {
		t.Fatalf("SaveConfig échoué : %v", err)
	}
	if bridgeNormal.IsMock() {
		t.Errorf("attendu bridgeNormal.IsMock() == false après SaveConfig(mockInterface=false)")
	}

	// Test de sauvegarde conjointe de MockInterface et MockAI
	_, err = bridgeNormal.SaveConfig("http://127.0.0.1:11434", "gemma4:26b", 60, false, true)
	if err != nil {
		t.Fatalf("SaveConfig échoué avec MockAI : %v", err)
	}
	cfgAfter, errCfg := bridgeNormal.GetConfig()
	if errCfg != nil {
		t.Fatalf("GetConfig a échoué : %v", errCfg)
	}
	if cfgAfter.MockInterface != false || cfgAfter.MockAI != true {
		t.Errorf("attendu MockInterface=false et MockAI=true, obtenu %+v", cfgAfter)
	}
}

func TestPythonClientRealIntegration(t *testing.T) {
	repoRoot := bridge.FindAppRoot()

	client := bridge.NewPythonClient(repoRoot)
	if client.IsMock() {
		t.Errorf("attendu IsMock() == false pour PythonClient")
	}

	// GetRepoStatus
	status, err := client.GetRepoStatus()
	if err != nil {
		t.Fatalf("GetRepoStatus a échoué : %v", err)
	}
	if !status.IsGitRepo {
		t.Errorf("attendu is_git_repo == true, obtenu false (repo: %s)", repoRoot)
	}
	if status.Branch == "" {
		t.Errorf("attendu nom de branche non vide")
	}

	// GetDiff
	diff, err := client.GetDiff()
	if err != nil {
		t.Fatalf("GetDiff a échoué : %v", err)
	}
	if !diff.Success {
		t.Errorf("attendu diff.Success == true")
	}

	// PingOllama (répond propre même si serveur inaccessible)
	ping, err := client.PingOllama()
	if err != nil {
		t.Fatalf("PingOllama a échoué : %v", err)
	}
	if !ping.Success {
		t.Errorf("attendu ping.Success == true")
	}

	// GetRecentCommits
	commitsResp, errCommits := client.GetRecentCommits(5)
	if errCommits != nil {
		t.Fatalf("GetRecentCommits a échoué : %v", errCommits)
	}
	if !commitsResp.Success || len(commitsResp.Commits) == 0 {
		t.Errorf("attendu liste de commits non vide, obtenu %+v", commitsResp)
	}
}

func TestAppEnvIsolationFromTargetRepo(t *testing.T) {
	targetRepo := t.TempDir() // Répertoire du projet cible analysé
	appRoot := t.TempDir()    // Répertoire où est installée l'application

	client := bridge.NewBackendClientWithAppRoot(targetRepo, appRoot, true, false)

	// Sauvegarde d'une configuration
	_, err := client.SaveConfig("http://10.22.28.190:11434", "qwen3.8:27b", 250.0, true)
	if err != nil {
		t.Fatalf("SaveConfig a échoué : %v", err)
	}

	// 1. Le .env DOIT être créé dans appRoot
	appEnv := filepath.Join(appRoot, ".env")
	if _, errStat := os.Stat(appEnv); errStat != nil {
		t.Errorf("Le fichier .env aurait dû être créé dans appRoot (%s)", appEnv)
	}

	// 2. Le .env NE DOIT PAS être créé dans targetRepo (aucun parasitage du projet analysé)
	targetEnv := filepath.Join(targetRepo, ".env")
	if _, errStat := os.Stat(targetEnv); errStat == nil {
		t.Errorf("ERREUR CRITIQUE : le fichier .env a été créé dans le dépôt cible analysé (%s) au lieu de l'application !", targetEnv)
	}

	// 3. GetConfig doit bien relire la configuration applicative
	cfg, errCfg := client.GetConfig()
	if errCfg != nil {
		t.Fatalf("GetConfig a échoué : %v", errCfg)
	}
	if cfg.OllamaModel != "qwen3.8:27b" {
		t.Errorf("modèle attendu 'qwen3.8:27b', obtenu '%s'", cfg.OllamaModel)
	}
	if cfg.TimeoutS != 250.0 {
		t.Errorf("timeout attendu 250.0, obtenu %.0f", cfg.TimeoutS)
	}
}

func TestAppEnvNeverInSrc(t *testing.T) {
	// Vérifie que FindAppRoot résout bien la racine du projet et JAMAIS src ou cli
	root := bridge.FindAppRoot()
	if filepath.Base(root) == "src" || filepath.Base(root) == "cli" {
		t.Fatalf("FindAppRoot ne doit jamais renvoyer src ou cli, obtenu: %s", root)
	}

	envPath := bridge.FindAppEnvPath(root)
	if filepath.Base(filepath.Dir(envPath)) == "src" {
		t.Fatalf("FindAppEnvPath a renvoyé un chemin dans src: %s", envPath)
	}

	// Tenter d'écrire en forçant un chemin contenant /src/.env
	// Sauvegarde/restauration du .env réel : ce test écrit volontairement à la
	// racine de l'application, il ne doit pas écraser la configuration locale.
	realRootEnv := filepath.Join(root, ".env")
	originalEnv, errOriginal := os.ReadFile(realRootEnv)
	t.Cleanup(func() {
		if errOriginal == nil {
			_ = os.WriteFile(realRootEnv, originalEnv, 0644)
		} else {
			_ = os.Remove(realRootEnv)
		}
	})

	forcedSrcPath := filepath.Join(root, "src", ".env")
	err := bridge.WriteEnvFile(forcedSrcPath, "http://localhost:11434", "gemma4:26b", 30.0, true)
	if err != nil {
		t.Fatalf("WriteEnvFile a échoué: %v", err)
	}

	// S'assurer formellement que src/.env n'existe pas
	if _, errStat := os.Stat(forcedSrcPath); errStat == nil {
		t.Fatalf("CRITIQUE : src/.env a été créé alors qu'il est formellement interdit !")
	}

	// S'assurer que le fichier a bien été écrit à la racine
	if _, errStat := os.Stat(realRootEnv); errStat != nil {
		t.Fatalf("Le fichier .env devrait exister à la racine (%s)", realRootEnv)
	}
}

func TestResolveForcedType(t *testing.T) {
	// 1. Tests par numéro
	numTests := map[string]string{
		"1": "feat",
		"2": "fix",
		"3": "chore",
		"4": "docs",
		"5": "refactor",
		"6": "perf",
		"7": "test",
		"8": "build",
		"9": "ci",
	}
	for in, expected := range numTests {
		res, ok := views.ResolveForcedType(in)
		if !ok || res != expected {
			t.Errorf("ResolveForcedType(%s) = (%s, %v), attendu (%s, true)", in, res, ok, expected)
		}
	}

	// 2. Tests par nom direct (insensible à la casse et aux espaces)
	nameTests := map[string]string{
		"feat":     "feat",
		"FIX":      "fix",
		"  Docs  ": "docs",
		"Chore":    "chore",
		"refactor": "refactor",
		"Perf":     "perf",
		"test":     "test",
		"build":    "build",
		"CI":       "ci",
	}
	for in, expected := range nameTests {
		res, ok := views.ResolveForcedType(in)
		if !ok || res != expected {
			t.Errorf("ResolveForcedType(%s) = (%s, %v), attendu (%s, true)", in, res, ok, expected)
		}
	}

	// 3. Tests d'entrées invalides
	invalidInputs := []string{"", "0", "10", "unknown", "invalid", "feat2", "style"}
	for _, in := range invalidInputs {
		res, ok := views.ResolveForcedType(in)
		if ok {
			t.Errorf("ResolveForcedType(%s) = (%s, true), attendu false", in, res)
		}
	}
}

func TestBuildForcedTypeFeedback(t *testing.T) {
	// Sans consigne additionnelle
	fb1 := views.BuildForcedTypeFeedback("fix", "")
	if fb1 != "type: fix" {
		t.Errorf("attendu 'type: fix', obtenu '%s'", fb1)
	}

	// Avec espaces seuls
	fb2 := views.BuildForcedTypeFeedback("feat", "   ")
	if fb2 != "type: feat" {
		t.Errorf("attendu 'type: feat', obtenu '%s'", fb2)
	}

	// Avec consigne additionnelle
	fb3 := views.BuildForcedTypeFeedback("docs", "ajoute des exemples d'utilisation")
	if fb3 != "type: docs. ajoute des exemples d'utilisation" {
		t.Errorf("attendu 'type: docs. ajoute des exemples d'utilisation', obtenu '%s'", fb3)
	}
}

func TestBridgeClientMockNeExecutePasDocker(t *testing.T) {
	tempDir := t.TempDir()
	client := bridge.NewBackendClient(tempDir, true, false)

	resStart, err := client.StartDockerOllama()
	if err != nil {
		t.Fatalf("StartDockerOllama en mode mock ne doit pas échouer : %v", err)
	}
	if resStart == nil || !resStart.Success || !strings.Contains(strings.ToLower(resStart.Message), "mock") {
		t.Errorf("attendu un démarrage simulé en mode mock, obtenu : %+v", resStart)
	}

	resStop, err := client.StopDockerOllama()
	if err != nil {
		t.Fatalf("StopDockerOllama en mode mock ne doit pas échouer : %v", err)
	}
	if resStop == nil || !resStop.Success || !strings.Contains(strings.ToLower(resStop.Message), "mock") {
		t.Errorf("attendu un arrêt simulé en mode mock, obtenu : %+v", resStop)
	}
}

func TestEffectiveOllamaBaseURL(t *testing.T) {
	const remote = "http://10.22.28.190:11434"

	if got := bridge.EffectiveOllamaBaseURL(&models.ConfigResponse{OllamaBaseURL: remote}); got != remote {
		t.Errorf("hors Docker, URL attendue %q, obtenu %q", remote, got)
	}
	if got := bridge.EffectiveOllamaBaseURL(&models.ConfigResponse{OllamaBaseURL: remote, DockerLLM: true}); got != bridge.DockerOllamaBaseURL {
		t.Errorf("en Docker, URL attendue %q, obtenu %q", bridge.DockerOllamaBaseURL, got)
	}
	if got := bridge.EffectiveOllamaBaseURL(nil); got != "" {
		t.Errorf("config nil : URL attendue vide, obtenu %q", got)
	}
}

func TestReadDockerLLMSettingFromEnvPriorite(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("USE_DOCKER_LLM=true\n"), 0644); err != nil {
		t.Fatalf("écriture .env : %v", err)
	}

	// Fichier seul
	t.Setenv("USE_DOCKER_LLM", "")
	if !bridge.ReadDockerLLMSettingFromEnv(dir) {
		t.Errorf("attendu true depuis le .env")
	}

	// Variable d'environnement prioritaire sur le fichier
	t.Setenv("USE_DOCKER_LLM", "false")
	if bridge.ReadDockerLLMSettingFromEnv(dir) {
		t.Errorf("attendu false depuis la variable d'environnement")
	}
	t.Setenv("USE_DOCKER_LLM", "true")
	if err := os.WriteFile(envPath, []byte("USE_DOCKER_LLM=false\n"), 0644); err != nil {
		t.Fatalf("écriture .env : %v", err)
	}
	if !bridge.ReadDockerLLMSettingFromEnv(dir) {
		t.Errorf("attendu true depuis la variable d'environnement prioritaire")
	}

	// PythonClient doit refléter la configuration
	t.Setenv("USE_DOCKER_LLM", "")
	if err := os.WriteFile(envPath, []byte("USE_DOCKER_LLM=true\n"), 0644); err != nil {
		t.Fatalf("écriture .env : %v", err)
	}
	pyClient := bridge.NewPythonClientWithAppRoot(dir, dir)
	if !pyClient.IsDockerLLM() {
		t.Errorf("PythonClient.IsDockerLLM() doit lire USE_DOCKER_LLM")
	}
}

func TestWriteEnvFilePreserveDockerLLM(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	if err := bridge.WriteEnvFileWithDockerLLM(envPath, "http://remote:11434", "gemma4:26b", 300, false, true); err != nil {
		t.Fatalf("WriteEnvFileWithDockerLLM(true) : %v", err)
	}
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("lecture .env : %v", err)
	}
	if !strings.Contains(string(content), "USE_DOCKER_LLM=true") {
		t.Errorf("attendu USE_DOCKER_LLM=true, contenu :\n%s", content)
	}

	// L'ancienne signature WriteEnvFile ne doit pas écraser le réglage existant
	if err := bridge.WriteEnvFile(envPath, "http://remote:11434", "gemma4:26b", 300, false); err != nil {
		t.Fatalf("WriteEnvFile : %v", err)
	}
	content, err = os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("lecture .env : %v", err)
	}
	if !strings.Contains(string(content), "USE_DOCKER_LLM=true") {
		t.Errorf("WriteEnvFile a écrasé USE_DOCKER_LLM, contenu :\n%s", content)
	}

	// La variante explicite doit pouvoir le désactiver
	if err := bridge.WriteEnvFileWithDockerLLM(envPath, "http://remote:11434", "gemma4:26b", 300, false, false); err != nil {
		t.Fatalf("WriteEnvFileWithDockerLLM(false) : %v", err)
	}
	content, err = os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("lecture .env : %v", err)
	}
	if !strings.Contains(string(content), "USE_DOCKER_LLM=false") {
		t.Errorf("attendu USE_DOCKER_LLM=false, contenu :\n%s", content)
	}
}

func TestSaveConfigDockerLLMCanBeDisabledAtRuntime(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("USE_DOCKER_LLM=true\nOLLAMA_MODEL=gemma4:26b\n"), 0644); err != nil {
		t.Fatalf("écriture .env : %v", err)
	}
	t.Setenv("USE_DOCKER_LLM", "")

	// Mock interface pour éviter de lancer le backend Python pendant le test.
	client := bridge.NewBackendClientWithAppRoot(dir, dir, true, false)

	if !client.IsDockerLLM() {
		t.Fatalf("état initial attendu : USE_DOCKER_LLM=true")
	}

	res, err := client.SaveConfigWithDockerLLM("http://10.22.28.190:11434", "gemma4:26b", 300, true, false)
	if err != nil {
		t.Fatalf("SaveConfigWithDockerLLM(false) : %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("sauvegarde attendue réussie, obtenu : %+v", res)
	}

	// Le .env et l'état exposé par le client doivent immédiatement refléter false.
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("lecture .env : %v", err)
	}
	if !strings.Contains(string(content), "USE_DOCKER_LLM=false") {
		t.Errorf("le .env doit contenir USE_DOCKER_LLM=false, contenu :\n%s", content)
	}
	if client.IsDockerLLM() {
		t.Errorf("IsDockerLLM() doit passer à false après sauvegarde")
	}
	cfg, err := client.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig : %v", err)
	}
	if cfg.DockerLLM {
		t.Errorf("GetConfig().DockerLLM doit refléter false après sauvegarde")
	}
}

func TestReadWritePythonBinSetting(t *testing.T) {
	tempDir := t.TempDir()

	// Nettoyage préventif de toute variable d'environnement système
	oldEnv := os.Getenv("PYTHON_BIN")
	_ = os.Unsetenv("PYTHON_BIN")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("PYTHON_BIN", oldEnv)
		} else {
			_ = os.Unsetenv("PYTHON_BIN")
		}
	}()

	// 1. Initialement vide
	initial := bridge.ReadPythonBinFromEnv(tempDir)
	if initial != "" {
		t.Errorf("attendu PythonBin initial vide, obtenu '%s'", initial)
	}

	// 2. Écriture d'un chemin d'interpréteur
	targetBin := "/opt/custom/python3"
	err := bridge.WritePythonBinSetting(tempDir, targetBin)
	if err != nil {
		t.Fatalf("WritePythonBinSetting a échoué : %v", err)
	}

	readVal := bridge.ReadPythonBinFromEnv(tempDir)
	if readVal != targetBin {
		t.Errorf("ReadPythonBinFromEnv attendu '%s', obtenu '%s'", targetBin, readVal)
	}

	cfg := bridge.LoadConfigFromEnv(tempDir, false)
	if cfg.PythonBin != targetBin {
		t.Errorf("LoadConfigFromEnv.PythonBin attendu '%s', obtenu '%s'", targetBin, cfg.PythonBin)
	}

	// 3. Réinitialisation vers vide (détection auto)
	err = bridge.WritePythonBinSetting(tempDir, "")
	if err != nil {
		t.Fatalf("WritePythonBinSetting(vide) a échoué : %v", err)
	}

	readAfterReset := bridge.ReadPythonBinFromEnv(tempDir)
	if readAfterReset != "" {
		t.Errorf("attendu PythonBin vide après réinitialisation, obtenu '%s'", readAfterReset)
	}
}

func TestResolvePythonBinPriority(t *testing.T) {
	tempDir := t.TempDir()

	// Nettoyage préventif de toute variable d'environnement système
	oldEnv := os.Getenv("PYTHON_BIN")
	_ = os.Unsetenv("PYTHON_BIN")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("PYTHON_BIN", oldEnv)
		} else {
			_ = os.Unsetenv("PYTHON_BIN")
		}
	}()

	// 1. Sans .venv ni PYTHON_BIN : échec explicite
	_, err := bridge.ResolvePythonBin(tempDir)
	if err == nil {
		t.Errorf("attendu erreur lorsque aucun environnement Python n'est présent")
	}

	// 2. Avec un faux .venv créé
	venvDir := filepath.Join(tempDir, ".venv", "bin")
	if runtime.GOOS == "windows" {
		venvDir = filepath.Join(tempDir, ".venv", "Scripts")
	}
	if errMkdir := os.MkdirAll(venvDir, 0755); errMkdir != nil {
		t.Fatalf("MkdirAll a échoué : %v", errMkdir)
	}
	fakePy := filepath.Join(venvDir, "python")
	if runtime.GOOS == "windows" {
		fakePy = filepath.Join(venvDir, "python.exe")
	}
	if errWrite := os.WriteFile(fakePy, []byte("#!/bin/sh\n"), 0755); errWrite != nil {
		t.Fatalf("WriteFile faux python a échoué : %v", errWrite)
	}

	resolved, errResolved := bridge.ResolvePythonBin(tempDir)
	if errResolved != nil {
		t.Fatalf("ResolvePythonBin avec .venv a échoué : %v", errResolved)
	}
	if resolved != fakePy {
		t.Errorf("ResolvePythonBin attendu '%s', obtenu '%s'", fakePy, resolved)
	}

	// 3. Avec PYTHON_BIN explicite qui surcharge le .venv
	customBin := "/usr/local/bin/python3-custom"
	if errWriteEnv := bridge.WritePythonBinSetting(tempDir, customBin); errWriteEnv != nil {
		t.Fatalf("WritePythonBinSetting a échoué : %v", errWriteEnv)
	}

	resolvedCustom, errCustom := bridge.ResolvePythonBin(tempDir)
	if errCustom != nil {
		t.Fatalf("ResolvePythonBin avec customBin a échoué : %v", errCustom)
	}
	if resolvedCustom != customBin {
		t.Errorf("attendu customBin prioritaire '%s', obtenu '%s'", customBin, resolvedCustom)
	}
}

func TestVerifyPythonEnvironmentReal(t *testing.T) {
	appRoot := bridge.FindAppRoot()

	report, err := bridge.VerifyPythonEnvironment(appRoot)
	if err != nil {
		t.Fatalf("VerifyPythonEnvironment a échoué sur appRoot : %v", err)
	}

	if !report.Success {
		t.Errorf("attendu report.Success == true, obtenu false (missing: %v, appLoadError: %s)", report.MissingDeps, report.AppLoadError)
	}
	if report.Version == "" {
		t.Errorf("attendu version Python non vide")
	}
	if report.Executable == "" {
		t.Errorf("attendu executable Python non vide")
	}

	// CheckPythonEnvironment doit réussir sans erreur
	if errCheck := bridge.CheckPythonEnvironment(appRoot); errCheck != nil {
		t.Errorf("CheckPythonEnvironment a échoué inopinément : %v", errCheck)
	}
}

func TestVerifyPythonEnvironmentInvalid(t *testing.T) {
	tempDir := t.TempDir()

	// Nettoyage préventif de toute variable d'environnement système
	oldEnv := os.Getenv("PYTHON_BIN")
	_ = os.Unsetenv("PYTHON_BIN")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("PYTHON_BIN", oldEnv)
		} else {
			_ = os.Unsetenv("PYTHON_BIN")
		}
	}()

	// 1. Sans aucun environnement Python
	_, err := bridge.VerifyPythonEnvironment(tempDir)
	if err == nil {
		t.Errorf("attendu erreur pour VerifyPythonEnvironment sans environnement")
	}

	// 2. Avec un faux exécutable qui échoue
	venvDir := filepath.Join(tempDir, ".venv", "bin")
	if runtime.GOOS == "windows" {
		venvDir = filepath.Join(tempDir, ".venv", "Scripts")
	}
	_ = os.MkdirAll(venvDir, 0755)
	fakePy := filepath.Join(venvDir, "python")
	if runtime.GOOS == "windows" {
		fakePy = filepath.Join(venvDir, "python.exe")
	}
	_ = os.WriteFile(fakePy, []byte("#!/bin/sh\necho 'crash'\nexit 1\n"), 0755)

	_, errInvalid := bridge.VerifyPythonEnvironment(tempDir)
	if errInvalid == nil {
		t.Errorf("attendu erreur pour VerifyPythonEnvironment avec exécutable défaillant")
	}
}

func TestVerifyPythonEnvironmentMissingDep(t *testing.T) {
	tempDir := t.TempDir()

	// Nettoyage préventif de toute variable d'environnement système
	oldEnv := os.Getenv("PYTHON_BIN")
	_ = os.Unsetenv("PYTHON_BIN")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("PYTHON_BIN", oldEnv)
		} else {
			_ = os.Unsetenv("PYTHON_BIN")
		}
	}()

	// Simule un script python qui renvoie missing_deps: ["git"]
	venvDir := filepath.Join(tempDir, ".venv", "bin")
	if runtime.GOOS == "windows" {
		venvDir = filepath.Join(tempDir, ".venv", "Scripts")
	}
	_ = os.MkdirAll(venvDir, 0755)
	fakePy := filepath.Join(venvDir, "python")
	if runtime.GOOS == "windows" {
		fakePy = filepath.Join(venvDir, "python.exe")
	}
	mockOutput := `{"executable": "/bin/mock", "version": "3.12.0", "missing_deps": ["git"], "app_load_error": "", "success": false}`
	_ = os.WriteFile(fakePy, []byte(fmt.Sprintf("#!/bin/sh\necho '%s'\n", mockOutput)), 0755)

	report, err := bridge.VerifyPythonEnvironment(tempDir)
	if err != nil {
		t.Fatalf("VerifyPythonEnvironment inattendu err: %v", err)
	}
	if report.Success {
		t.Errorf("attendu report.Success == false quand missing_deps est non vide")
	}
	if len(report.MissingDeps) != 1 || report.MissingDeps[0] != "git" {
		t.Errorf("attendu missing_deps=[git], obtenu %v", report.MissingDeps)
	}

	// CheckPythonEnvironment doit échouer avec l'erreur explicite
	errCheck := bridge.CheckPythonEnvironment(tempDir)
	if errCheck == nil {
		t.Errorf("attendu erreur pour CheckPythonEnvironment quand dépendances manquantes")
	}
}

func TestVerifyPythonCandidate(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Candidat introuvable
	_, errNotFound := bridge.VerifyPythonCandidate(tempDir, "/chemin/inexistant/vers/python", false)
	if errNotFound == nil {
		t.Errorf("attendu erreur pour binaire inexistant")
	}

	// 2. Mode auto sans .venv
	_, errAutoNoVenv := bridge.VerifyPythonCandidate(tempDir, "", true)
	if errAutoNoVenv == nil {
		t.Errorf("attendu erreur pour mode auto sans .venv")
	}

	// 3. Candidat valide simulé
	fakeValidPy := filepath.Join(tempDir, "fake_valid_python")
	if runtime.GOOS == "windows" {
		fakeValidPy = filepath.Join(tempDir, "fake_valid_python.exe")
	}
	validOutput := `{"executable": "/bin/python3", "version": "3.12.0", "missing_deps": [], "app_load_error": "", "success": true}`
	_ = os.WriteFile(fakeValidPy, []byte(fmt.Sprintf("#!/bin/sh\necho '%s'\n", validOutput)), 0755)

	reportValid, errValid := bridge.VerifyPythonCandidate(tempDir, fakeValidPy, false)
	if errValid != nil {
		t.Fatalf("VerifyPythonCandidate valide inattendu err : %v", errValid)
	}
	if !reportValid.Success {
		t.Errorf("attendu Success == true pour candidat valide")
	}

	// 4. Candidat avec dépendances manquantes
	fakeInvalidPy := filepath.Join(tempDir, "fake_invalid_python")
	if runtime.GOOS == "windows" {
		fakeInvalidPy = filepath.Join(tempDir, "fake_invalid_python.exe")
	}
	invalidOutput := `{"executable": "/bin/python3", "version": "3.12.0", "missing_deps": ["yaml", "git"], "app_load_error": "", "success": false}`
	_ = os.WriteFile(fakeInvalidPy, []byte(fmt.Sprintf("#!/bin/sh\necho '%s'\n", invalidOutput)), 0755)

	reportInvalid, errInvalid := bridge.VerifyPythonCandidate(tempDir, fakeInvalidPy, false)
	if errInvalid != nil {
		t.Fatalf("VerifyPythonCandidate invalide inattendu err : %v", errInvalid)
	}
	if reportInvalid.Success {
		t.Errorf("attendu Success == false quand dépendances manquantes")
	}
	if len(reportInvalid.MissingDeps) != 2 {
		t.Errorf("attendu 2 dépendances manquantes, obtenu %v", reportInvalid.MissingDeps)
	}
}

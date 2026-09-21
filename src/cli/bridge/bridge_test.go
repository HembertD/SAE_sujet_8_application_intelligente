package bridge

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"sae-git-cli/models"
)

// Vérifications de conformité d'interface à la compilation
var (
	_ BackendClient = (*BridgeClient)(nil)
	_ BackendClient = (*PythonClient)(nil)
	_ BackendClient = (*MockClient)(nil)
)

func TestMockClientBasics(t *testing.T) {
	mock := NewMockClient("/tmp/fake-repo")
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
}

func TestMockClientOverrides(t *testing.T) {
	mock := NewMockClient("/tmp/fake-repo")
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
		if !ParseBoolFromString(v) {
			t.Errorf("attendu true pour '%s'", v)
		}
	}

	falsy := []string{"false", "False", "0", "no", "non", "", "random"}
	for _, v := range falsy {
		if ParseBoolFromString(v) {
			t.Errorf("attendu false pour '%s'", v)
		}
	}
}

func TestEnvFileReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	// Écriture du fichier .env
	err := WriteEnvFile(envPath, "http://localhost:11434", "mistral:latest", 45.0, "fr", true)
	if err != nil {
		t.Fatalf("WriteEnvFile échoué : %v", err)
	}

	// Lecture de la configuration
	cfg := LoadConfigFromEnv(tempDir, false)
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
	bridgeDemo := NewBackendClient(tempDir, true)
	if !bridgeDemo.IsMock() {
		t.Errorf("attendu bridgeDemo.IsMock() == true avec forceDemo=true")
	}

	// Initialisation normale
	bridgeNormal := NewBackendClient(tempDir, false)
	// Sans configuration, le mode démo est inactif
	if bridgeNormal.IsMock() {
		t.Errorf("attendu bridgeNormal.IsMock() == false par défaut sans variable d'env")
	}

	// Activation du mode démo
	res, err := bridgeNormal.SaveConfig("http://127.0.0.1:11434", "gemma4:12b", 60, "fr", true)
	if err != nil || !res.Success {
		t.Fatalf("SaveConfig échoué : %v", err)
	}

	// Vérifie la prise en compte de la configuration
	if !bridgeNormal.IsMock() {
		t.Errorf("attendu bridgeNormal.IsMock() == true après SaveConfig(mockInterface=true)")
	}

	// Bascule inverse
	_, err = bridgeNormal.SaveConfig("http://127.0.0.1:11434", "gemma4:12b", 60, "fr", false)
	if err != nil {
		t.Fatalf("SaveConfig échoué : %v", err)
	}
	if bridgeNormal.IsMock() {
		t.Errorf("attendu bridgeNormal.IsMock() == false après SaveConfig(mockInterface=false)")
	}
}

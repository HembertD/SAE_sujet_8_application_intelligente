package cli

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"sae-git-cli/bridge"
	"sae-git-cli/ui"
)

func TestNotificationEnabledDisabled(t *testing.T) {
	// Sauvegarde de l'état initial
	initialState := ui.IsNotificationsEnabled()
	defer ui.SetNotificationsEnabled(initialState)

	ui.SetNotificationsEnabled(false)
	if ui.IsNotificationsEnabled() {
		t.Errorf("attendu IsNotificationsEnabled() == false")
	}

	called := false
	ui.SetNotifySender(func(title, message string, appIcon any) error {
		called = true
		return nil
	})
	defer ui.ResetNotifySender()

	err := ui.NotifySync("Test", "Message")
	if err != nil {
		t.Errorf("erreur inattendue sur NotifySync désactivé : %v", err)
	}
	if called {
		t.Errorf("le sender ne devrait pas être invoqué quand les notifications sont désactivées")
	}

	ui.SetNotificationsEnabled(true)
	if !ui.IsNotificationsEnabled() {
		t.Errorf("attendu IsNotificationsEnabled() == true")
	}

	err = ui.NotifySync("Test", "Message")
	if err != nil {
		t.Errorf("erreur inattendue sur NotifySync activé : %v", err)
	}
	if !called {
		t.Errorf("le sender aurait dû être invoqué quand les notifications sont activées")
	}
}

func TestNotificationMessagesFormatting(t *testing.T) {
	ui.SetNotificationsEnabled(true)
	defer ui.ResetNotifySender()

	var lastTitle, lastMessage string
	ui.SetNotifySender(func(title, message string, appIcon any) error {
		lastTitle = title
		lastMessage = message
		return nil
	})

	// 1. Commit réussi
	ui.NotifyCommitGenerated(true)
	time.Sleep(20 * time.Millisecond) // Goroutine asynchrone
	if !strings.Contains(lastTitle, "Gitelligence") || !strings.Contains(lastMessage, "succès") {
		t.Errorf("format commit succès invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	// 2. Commit échoué
	ui.NotifyCommitGenerated(false, "modèle Ollama injoignable")
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "⚠️") || !strings.Contains(lastMessage, "modèle Ollama injoignable") {
		t.Errorf("format commit erreur invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	// 3. Release notes réussies
	ui.NotifyReleaseNotesGenerated(true)
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "Gitelligence") || !strings.Contains(lastMessage, "Release notes") {
		t.Errorf("format release notes succès invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	// 4. Release notes échouées
	ui.NotifyReleaseNotesGenerated(false, "timeout dépassé")
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "⚠️") || !strings.Contains(lastMessage, "timeout dépassé") {
		t.Errorf("format release notes erreur invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	// 5. Primitives standards NotifySuccess / NotifyError / NotifyWarning / NotifyInfo
	ui.NotifySuccess("Action terminée avec succès")
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "⚡") || lastMessage != "Action terminée avec succès" {
		t.Errorf("NotifySuccess invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	ui.NotifyError("Une erreur est survenue")
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "⚠️") || lastMessage != "Une erreur est survenue" {
		t.Errorf("NotifyError invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	ui.NotifyWarning("Attention au diff volumineux")
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "⚠️") || lastMessage != "Attention au diff volumineux" {
		t.Errorf("NotifyWarning invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	ui.NotifyInfo("Mode hors-ligne activé")
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(lastTitle, "ℹ") || lastMessage != "Mode hors-ligne activé" {
		t.Errorf("NotifyInfo invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	// 6. Primitives synchrones NotifySuccessSync / NotifyErrorSync
	_ = ui.NotifySuccessSync("Succès Synchrone", "Titre Personnalisé")
	if lastTitle != "Titre Personnalisé" || lastMessage != "Succès Synchrone" {
		t.Errorf("NotifySuccessSync invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}

	_ = ui.NotifyErrorSync("Erreur Synchrone", "Titre Erreur")
	if lastTitle != "Titre Erreur" || lastMessage != "Erreur Synchrone" {
		t.Errorf("NotifyErrorSync invalide : title='%s', msg='%s'", lastTitle, lastMessage)
	}
}

func TestNotificationAsyncNonBlocking(t *testing.T) {
	ui.SetNotificationsEnabled(true)
	defer ui.ResetNotifySender()

	var wg sync.WaitGroup
	wg.Add(1)

	// Simule un sender lent pour vérifier le caractère asynchrone
	ui.SetNotifySender(func(title, message string, appIcon any) error {
		time.Sleep(50 * time.Millisecond)
		wg.Done()
		return nil
	})

	start := time.Now()
	ui.Notify("Test Asynchrone", "Contenu")
	elapsed := time.Since(start)

	if elapsed > 30*time.Millisecond {
		t.Errorf("Notify bloquant détecté : durée d'appel %v (attendu < 30ms)", elapsed)
	}

	// Attente de la fin de la goroutine en arrière-plan
	wg.Wait()
}

func TestReadWriteNotificationsSetting(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Valeur par défaut si aucun .env
	if !bridge.ReadEnableNotificationsFromEnv(tempDir) {
		t.Errorf("attendu ReadEnableNotificationsFromEnv == true par défaut")
	}

	// 2. Écriture de la valeur false
	err := bridge.WriteNotificationsSetting(tempDir, false)
	if err != nil {
		t.Fatalf("WriteNotificationsSetting échoué : %v", err)
	}

	if bridge.ReadEnableNotificationsFromEnv(tempDir) {
		t.Errorf("attendu ReadEnableNotificationsFromEnv == false après écriture")
	}

	// Vérification de la prise en compte dans LoadConfigFromEnv
	cfg := bridge.LoadConfigFromEnv(tempDir, false)
	if cfg.EnableNotifications {
		t.Errorf("attendu cfg.EnableNotifications == false dans LoadConfigFromEnv")
	}

	// 3. Réactivation avec true
	err = bridge.WriteNotificationsSetting(tempDir, true)
	if err != nil {
		t.Fatalf("WriteNotificationsSetting réactivation échouée : %v", err)
	}

	if !bridge.ReadEnableNotificationsFromEnv(tempDir) {
		t.Errorf("attendu ReadEnableNotificationsFromEnv == true après réactivation")
	}

	// 4. Préservation lors de WriteEnvFile
	envPath := filepath.Join(tempDir, ".env")
	err = bridge.WriteEnvFile(envPath, "http://localhost:11434", "gemma4:26b", 120.0, false)
	if err != nil {
		t.Fatalf("WriteEnvFile échoué : %v", err)
	}

	if !bridge.ReadEnableNotificationsFromEnv(tempDir) {
		t.Errorf("ENABLE_NOTIFICATIONS aurait dû être préservé par WriteEnvFile")
	}
}

package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gen2brain/beeep"
)

var (
	notificationsEnabled = true
	notifMutex           sync.RWMutex
	notifySender         = beeep.Notify
)

// SetNotificationsEnabled active ou désactive l'envoi de notifications de bureau.
func SetNotificationsEnabled(enabled bool) {
	notifMutex.Lock()
	defer notifMutex.Unlock()
	notificationsEnabled = enabled
}

// IsNotificationsEnabled indique si les notifications sont actuellement activées.
func IsNotificationsEnabled() bool {
	notifMutex.RLock()
	defer notifMutex.RUnlock()
	return notificationsEnabled
}

// SetNotifySender remplace le délégué d'envoi de notification (utilisé notamment par les tests unitaires).
func SetNotifySender(sender func(title, message string, appIcon any) error) {
	notifMutex.Lock()
	defer notifMutex.Unlock()
	if sender == nil {
		notifySender = beeep.Notify
	} else {
		notifySender = sender
	}
}

// ResetNotifySender rétablit le comportement par défaut de notification système via beeep.
func ResetNotifySender() {
	SetNotifySender(nil)
}

// Notify envoie une notification de bureau de façon asynchrone (non-bloquante).
// En cas d'environnement sans serveur d'affichage ou d'erreur système, l'échec est absorbé silencieusement.
func Notify(title, message string) {
	if !IsNotificationsEnabled() {
		return
	}
	notifMutex.RLock()
	sender := notifySender
	notifMutex.RUnlock()

	go func() {
		_ = sender(title, message, "")
	}()
}

// NotifySync envoie une notification de manière synchrone (utile pour les assertions des tests unitaires).
func NotifySync(title, message string) error {
	if !IsNotificationsEnabled() {
		return nil
	}
	notifMutex.RLock()
	sender := notifySender
	notifMutex.RUnlock()

	return sender(title, message, "")
}

// NotifySuccess diffuse une notification de succès de façon asynchrone ("afficher un succès").
// Suit la même convention que PrintSuccess dans l'interface utilisateur.
func NotifySuccess(message string, title ...string) {
	t := "Gitelligence ⚡"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = title[0]
	}
	Notify(t, message)
}

// NotifyError diffuse une notification d'erreur de façon asynchrone ("afficher une erreur").
// Suit la même convention que PrintError / PrintErrorCard dans l'interface utilisateur.
func NotifyError(message string, title ...string) {
	t := "Gitelligence ⚠️"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = title[0]
	}
	Notify(t, message)
}

// NotifyWarning diffuse une notification d'avertissement de façon asynchrone ("afficher un avertissement").
// Suit la même convention que PrintWarning dans l'interface utilisateur.
func NotifyWarning(message string, title ...string) {
	t := "Gitelligence ⚠️"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = title[0]
	}
	Notify(t, message)
}

// NotifyInfo diffuse une notification d'information de façon asynchrone ("afficher une information").
// Suit la même convention que PrintInfo dans l'interface utilisateur.
func NotifyInfo(message string, title ...string) {
	t := "Gitelligence ℹ"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = title[0]
	}
	Notify(t, message)
}

// NotifySuccessSync diffuse une notification de succès de façon synchrone (pour tests et assertions).
func NotifySuccessSync(message string, title ...string) error {
	t := "Gitelligence ⚡"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = title[0]
	}
	return NotifySync(t, message)
}

// NotifyErrorSync diffuse une notification d'erreur de façon synchrone (pour tests et assertions).
func NotifyErrorSync(message string, title ...string) error {
	t := "Gitelligence ⚠️"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = title[0]
	}
	return NotifySync(t, message)
}

// NotifyCommitGenerated notifie l'utilisateur dès que le LLM a terminé la génération du commit.
func NotifyCommitGenerated(success bool, errMsg ...string) {
	if success {
		NotifySuccess("Message de commit généré avec succès !")
	} else {
		detail := "délai dépassé ou indisponibilité de l'IA."
		if len(errMsg) > 0 && strings.TrimSpace(errMsg[0]) != "" {
			detail = errMsg[0]
		}
		NotifyError(fmt.Sprintf("Échec génération commit : %s", detail))
	}
}

// NotifyReleaseNotesGenerated notifie l'utilisateur dès que le LLM a terminé la rédaction des release notes.
func NotifyReleaseNotesGenerated(success bool, errMsg ...string) {
	if success {
		NotifySuccess("Release notes générées avec succès !")
	} else {
		detail := "délai dépassé ou indisponibilité de l'IA."
		if len(errMsg) > 0 && strings.TrimSpace(errMsg[0]) != "" {
			detail = errMsg[0]
		}
		NotifyError(fmt.Sprintf("Échec génération release notes : %s", detail))
	}
}

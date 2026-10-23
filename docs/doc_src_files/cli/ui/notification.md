# Module de Notifications Système UI

Fichier source associé : src/cli/ui/notification.go

## Description

Ce module implémente le système de notifications de bureau multi-plateforme (Linux, macOS, Windows) de Gitelligence via la bibliothèque `beeep`. Il avertit l'utilisateur dès que le LLM a terminé la génération d'un message de commit ou la rédaction de release notes.

## Responsabilités

- `Notify` : diffuse une notification de bureau de façon asynchrone (non-bloquante via une goroutine) ;
- `NotifySync` : diffuse une notification de bureau de façon synchrone (utilisé principalement pour les assertions de tests unitaires) ;
- `NotifySuccess` / `NotifySuccessSync` : diffuse une notification de bureau de succès (« afficher un succès ») alignée sur le standard `PrintSuccess` ;
- `NotifyError` / `NotifyErrorSync` : diffuse une notification de bureau d'erreur (« afficher une erreur ») alignée sur le standard `PrintError` / `PrintErrorCard` ;
- `NotifyWarning` : diffuse une notification de bureau d'avertissement alignée sur `PrintWarning` ;
- `NotifyInfo` : diffuse une notification de bureau d'information alignée sur `PrintInfo` ;
- `NotifyCommitGenerated` : notifie la réussite ou l'échec de la génération du message de commit avec messages contextualisés et émojis distinctifs (`⚡` pour le succès, `⚠️` pour l'erreur) ;
- `NotifyReleaseNotesGenerated` : notifie la réussite ou l'échec de la rédaction des release notes synthétisées ;
- `SetNotificationsEnabled` / `IsNotificationsEnabled` : active ou désactive l'envoi de notifications selon les préférences de configuration (`ENABLE_NOTIFICATIONS` dans le `.env`) ;
- `SetNotifySender` / `ResetNotifySender` : permet d'injecter un délégué mocké pour tester l'ensemble du cycle de notification sans dépendre du serveur graphique ou du daemon de notification de l'OS hôte.

## Points clés

- **Multi-plateforme** :
  - **Linux** : Notifications D-Bus natives en pur Go ou via `notify-send` ;
  - **macOS** : Notifications système via AppleScript (`osascript`) ;
  - **Windows** : Notifications Toast natives (Windows 10/11) ou ballons.
- **Tolérance aux pannes (Fail-safe)** : en l'absence de serveur d'affichage ou en environnement d'intégration continue (CI), les erreurs d'envoi sont absorbées silencieusement sans perturber le flux utilisateur ni bloquer le terminal ;
- **Non-bloquant** : l'envoi s'exécute dans un thread de fond (`goroutine`), garantissant que la navigation du TUI et les saisies clavier demeurent parfaitement fluides ;
- **Débrayable** : configurable via la variable d'environnement `ENABLE_NOTIFICATIONS` et ajustable directement dans le menu de configuration `[3] Configuration`.

## Commentaires de fonction (dupliqués du code source)

- `SetNotificationsEnabled()` : Active ou désactive l'envoi de notifications de bureau.
- `IsNotificationsEnabled()` : Indique si les notifications sont actuellement activées.
- `SetNotifySender()` : Remplace le délégué d'envoi de notification (utilisé notamment par les tests unitaires).
- `ResetNotifySender()` : Rétablit le comportement par défaut de notification système via beeep.
- `Notify()` : Envoie une notification de bureau de façon asynchrone (non-bloquante).
- `NotifySync()` : Envoie une notification de manière synchrone.
- `NotifySuccess()` : Diffuse une notification de bureau de succès ("afficher un succès").
- `NotifyError()` : Diffuse une notification de bureau d'erreur ("afficher une erreur").
- `NotifyWarning()` : Diffuse une notification de bureau d'avertissement.
- `NotifyInfo()` : Diffuse une notification de bureau d'information.
- `NotifySuccessSync()` : Diffuse une notification de succès de façon synchrone.
- `NotifyErrorSync()` : Diffuse une notification d'erreur de façon synchrone.
- `NotifyCommitGenerated()` : Notifie l'utilisateur dès que le LLM a terminé la génération du commit.
- `NotifyReleaseNotesGenerated()` : Notifie l'utilisateur dès que le LLM a terminé la rédaction des release notes.

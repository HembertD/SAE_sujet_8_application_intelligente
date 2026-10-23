# Tests Unitaires des Notifications de Bureau

Fichier source associé : test/cli/notification_test.go

## Description

Ce fichier contient la suite de tests unitaires automatisés validant le sous-système de notifications multi-plateforme du CLI Go, le formatage des messages, la non-régression asynchrone et la persistance de configuration dans le `.env`.

## Tests couverts

1. **`TestNotificationEnabledDisabled`** : valide l'activation et la désactivation dynamique (`SetNotificationsEnabled`), vérifiant que le sender de notification n'est jamais sollicité lorsque les notifications sont coupées.
2. **`TestNotificationMessagesFormatting`** : vérifie le formatage des messages et titres pour les scénarios métier (génération de commit, release notes) et pour les primitives génériques `NotifySuccess`, `NotifyError`, `NotifyWarning`, `NotifyInfo` (« afficher un succès / une erreur ») ainsi que leurs variantes synchrones `NotifySuccessSync` / `NotifyErrorSync`.
3. **`TestNotificationAsyncNonBlocking`** : garantit que l'appel à `Notify` délègue l'envoi dans une goroutine indépendante et retourne immédiatement (< 30ms) même si le système de notification sous-jacent accuse une latence.
4. **`TestReadWriteNotificationsSetting`** : valide la lecture et l'écriture de `ENABLE_NOTIFICATIONS` dans le `.env` temporaire, la valeur par défaut (`true`), la répercussion dans `LoadConfigFromEnv` et la bonne préservation du paramètre lors de la réécriture du `.env` via `WriteEnvFile`.

## Exécution des tests

```bash
cd test/cli
go test -v -run TestNotification ./...
```

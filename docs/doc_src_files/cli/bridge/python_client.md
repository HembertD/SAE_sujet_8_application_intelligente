# Implémentation Réelle PythonClient

Fichier source associé : src/cli/bridge/python_client.go

## Description

Ce module implémente l'interface `BackendClient` en exécutant le backend Python (`git_commit_release_notes_generator.service.core`) sous la forme de sous-processus système via `os/exec`.

## Responsabilités

- Résoudre dynamiquement l'interpréteur Python (`python3` ou environnement virtuel local `/home/maxence/monenv/bin/python`).
- Configurer les variables d'environnement du sous-processus : `PYTHONPATH` pointant sur `<appRoot>/src`, et transmission de `SMART_COMMIT_APP_ROOT`.
- Exécuter les commandes métier (`--action status`, `diff`, `stage-all`, `generate-commit`, `apply-commit`, `push`, `release-notes`, `ping-ollama`).
- Isoler les flux : capture du JSON strict sur `stdout` et redirection des logs/traces sur `stderr`.
- Désérialiser les réponses JSON dans des structures Go fortement typées (`models`).

## Points clés

- **Résilience** : En cas d'erreur du sous-processus, le message `stderr` est extrait et formaté pour un affichage propre sans panic.
- **Isolation du dépôt cible** : Le dépôt cible est transmis via `--repo`, tandis que le code et la configuration proviennent exclusivement de l'application.

## Détails déplacés du code source

Précisions issues des commentaires retirés de `python_client.go` :

- `PythonClient` implémente `BackendClient` en communiquant avec le sous-processus Python et ne contient AUCUNE logique métier : tout le traitement est délégué à Python.
- `runPythonCommand` exécute la commande Python et récupère `stdout` ; `PYTHONPATH` pointe sur le dossier `src` de l'application, avec candidats de secours si `appSrc` n'est pas résolu ; la configuration est chargée exclusivement depuis l'application (`p.appRoot`, JAMAIS `p.repoRoot`).
- `IsMock` retourne toujours `false` ; chaque méthode (`GetRepoStatus`, `GetDiff`, `StageAll`, `GenerateCommit`, `ApplyCommit`, `Push`, `GetReleaseNotes`, `GetConfig`, `SaveConfig`, `PingOllama`, `SaveMarkdownFile`) interroge le backend Python pour l'action correspondante, `SaveConfig` notifiant Python après écriture du `.env` applicatif et `SaveMarkdownFile` écrivant localement le changelog exporté.

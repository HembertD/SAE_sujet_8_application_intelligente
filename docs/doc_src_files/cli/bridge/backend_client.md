# Client de Pontage Go ↔ Python (Bridge)

Fichiers sources associés :
- `src/cli/bridge/client.go` : Interface publique `BackendClient` et coordinateur `BridgeClient`
- `src/cli/bridge/python_client.go` : Implémentation réelle exécutant le sous-processus Python
- `src/cli/bridge/mock_client.go` : Implémentation de test et mode démo hors-ligne
- `src/cli/bridge/env.go` : Gestionnaire isolé du fichier `.env` (lecture, écriture, parsing)
- `src/cli/bridge/bridge_test.go` : Tests unitaires automatisés du module bridge

## Description

Ce module implémente le pont de communication entre l'interface TUI Go et le backend Python. Il garantit la règle d'architecture **Zéro logique en Go** en s'appuyant sur le principe d'inversion des dépendances (*DIP*) : les vues front-end dépendent exclusivement de l'interface `BackendClient`.

## Architecture & Responsabilités

1. **`BackendClient` (Interface)** :
   - Définit le contrat haut niveau et fortement typé (`GetRepoStatus`, `GetDiff`, `StageAll`, `GenerateCommit`, `ApplyCommit`, `Push`, `GetReleaseNotes`, `GetConfig`, `SaveConfig`, `PingOllama`, `SaveMarkdownFile`, `IsMock`).

2. **`PythonClient` (Mode Standard)** :
   - Exécute les commandes en sous-processus via `exec.Command` (`python3 -m git_commit_release_notes_generator.service.core`) ;
   - Capture `stdout` pour la désérialisation JSON stricte via `encoding/json` ;
   - Isole les traces et logs sur `stderr` pour ne pas corrompre le flux de données.

3. **`MockClient` (Mode Démo & Tests)** :
   - Fournit les réponses conformes aux jeux d'essais du projet (secrets masqués dans `auth/token_vault.py`, exclusion des fichiers binaires, simulation de délai) ;
   - Permet l'injection d'erreurs ou de réponses sur mesure (`CustomError`, `CustomDiff`, `SimulateDelay=false`) pour les tests unitaires automatisés.

4. **`BridgeClient` (Coordinateur Adaptateur)** :
   - Instancié par `NewBackendClient(repoRoot, forceDemo)` ;
   - Achemine dynamiquement les appels vers `MockClient` ou `PythonClient` selon la variable `MOCK_INTERFACE` du `.env` ou le drapeau `--demo` sans recompilation.

5. **`env.go` (Gestionnaire .env)** :
   - Découplé du transport : localisation automatique (`FindEnvPath`), lecture des booléens (`ParseBoolFromString`) et réécriture atomique sécurisée (`WriteEnvFile`).

## Points clés

- **Découplage strict** : Le front-end dépend d'une abstraction (`BackendClient`), facilitant les tests unitaires isolés.
- **Résilience et isolation** : Séparation claire entre les appels système et l'environnement de test autonome.
- **Bascule à chaud** : Modification directe de `MOCK_INTERFACE` depuis l'écran de configuration du TUI répercutée immédiatement sans recompilation.

## Rôle dans l'architecture

Passerelle unique d'accès au domaine métier. Toute vue front-end interagit avec le système exclusivement par l'intermédiaire de l'interface `BackendClient`.

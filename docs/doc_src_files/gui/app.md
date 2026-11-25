# Orchestrateur Applicatif Wails (App)

Fichier source associé : `src/gui/app.go`

## Description

Le fichier `app.go` implémente la structure principale `App` qui centralise et expose l'ensemble des méthodes métier du backend vers l'interface utilisateur Wails (HTML/CSS/JS). Il réutilise directement l'interface `BackendClient` et assure le pontage asynchrone avec le backend Python et le serveur Ollama.

## Responsabilités

- **Gestion du cycle de vie Wails** :
  - `NewApp(repoRoot, appRoot, devMode, forceMockInterface, forceMockIA, forceDockerLLM)` : instanciation sécurisée de l'application avec configuration du mode développeur ;
  - `startup(ctx)` : capture du contexte Wails lors de l'initialisation et activation des notifications système selon `.env` ;
  - `initClient()` : instanciation dynamique du `BackendClient` (`PythonClient` ou `MockClient`) en isolant la racine de l'application et le dépôt Git analysé.
- **Sélection et gestion du dépôt cible** :
  - `ChooseDirectory()` : ouverture de la boîte de dialogue native du système d'exploitation via `wailsRuntime.OpenDirectoryDialog` pour sélectionner un dossier ;
  - `SetTargetRepo(repoPath)` : basculement dynamique du dépôt analysé sans redémarrer l'application ;
  - `GetTargetRepo()` et `GetAppRoot()` : accesseurs sur les chemins configurés.
- **Opérations Git & Analyse différentielle** :
  - `GetRepoStatus()` : renvoie l'état Git courant (branche, compteurs staged / unstaged) ;
  - `GetDiff()` : récupère les fichiers modifiés, compteurs de lignes et patchs textuels assainis ;
  - `StageAll()` : indexe l'ensemble des modifications (`git add -A`) avec notification de succès ;
  - `Push()` : transmet les commits locaux au serveur distant (`git push`).
- **Génération Intelligente Conventional Commits** :
  - `GenerateCommit(feedback)` : sollicite le LLM local ou simulé pour proposer un message normé avec gestion du feedback interactif utilisateur et auto-staging si besoin ;
  - `ImposeCommitType(chosenType, comment)` : force un type Conventional Commit strict (`feat`, `fix`, `docs`, etc.) comme l'option `[T]` du CLI ;
  - `ApplyCommit(message)` : valide et enregistre le commit sur le dépôt avec le message final.
- **Release Notes & Changelogs** :
  - `GetReleaseNotes(fromTag, toTag)` : synthétise l'historique Git et rédige les release notes au format Markdown ;
  - `GetRecentCommits(limit)` : extrait l'historique des derniers commits avec leurs tags éventuels ;
  - `SaveReleaseNotesMarkdown(filename, content)` : exporte le document Markdown sur le disque.
- **Configuration, Python & Diagnostics (Copie Conforme CLI)** :
  - `GetConfig()` et `SaveConfig(...)` : lecture et écriture des paramètres dans le fichier `.env` ;
  - `SaveNotifications(enabled)` : persistance du réglage des notifications de bureau dans le `.env` et activation/désactivation au runtime ;
  - `VerifyPythonEnvironment()` : diagnostic approfondi de l'environnement Python et des dépendances (option `[P]` du CLI) ;
  - `VerifyPythonCandidate(candidate, isAuto)` : vérification d'un interpréteur candidat avant acceptation ;
  - `SavePythonBin(candidate, isAuto)` : persistance du binaire Python dans le `.env` et réinitialisation du client de pontage ;
  - `PingOllama()` : vérification de la latence, disponibilité et liste des modèles installés sur le serveur LLM Ollama ;
  - `DockerOllamaStatus()`, `StartDockerOllama()`, `StopDockerOllama()`, `RestartDockerOllama()`, `GetDockerLogs(tailLines)` : inspection détaillée (santé, téléchargement, modèle), pilotage et streaming des logs du conteneur Ollama sous Docker Compose ;
  - `GetSystemSpecs()` : fournit les métriques d'exécution Go en direct (mémoire vive allouée, cœurs CPU, nombre de goroutines).

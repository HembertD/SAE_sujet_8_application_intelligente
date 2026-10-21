# Interface BackendClient et Coordinateur BridgeClient

Fichier source associé : src/cli/bridge/client.go

## Description

Ce fichier définit l'interface maîtresse `BackendClient` et le coordinateur adaptateur `BridgeClient` qui fait le lien entre les vues de l'interface en ligne de commande et les implémentations réelles (`PythonClient`) ou simulées (`MockClient`).

## Responsabilités

- Définir le contrat d'interface haut niveau `BackendClient` regroupant les actions du domaine Git, Ollama et configuration.
- Instancier le coordinateur via `NewBackendClientWithAppRoot(repoRoot, appRoot, forceMockInterface, forceMockIA, forceDockerLLM...)`.
- Gérer le cycle de vie du conteneur Ollama Docker (`StartDockerOllama`, `StopDockerOllama`).
- Maintenir la référence vers le dépôt cible (`repoRoot`) et la racine de l'application (`appRoot`).
- Assurer le routage dynamique vers le délégué actif en fonction du drapeau `--mockInterface` ou du paramètre `MOCK_INTERFACE` du `.env`.
- Fournir les méthodes de consultation et modification de la configuration (`GetConfig`, `SaveConfig`).

## Points clés

- **Inversion des dépendances (DIP)** : Le front-end dépend exclusivement de l'abstraction `BackendClient`.
- **Zéro logique métier en Go** : Les traitements d'analyse Git, LLM et release notes sont délégués au backend Python.

## Détails déplacés du code source

Précisions issues des commentaires retirés de `client.go` :

- `BridgeClient` coordonne l'accès au backend en déléguant à `PythonClient` ou `MockClient` ; `activeDelegate` renvoie l'implémentation active selon `IsMock`.
- `IsMock` indique si le mode simulation est actif : le drapeau forcé `forceMockInterface` est prioritaire, sinon lecture via `ReadMockSettingFromEnv`.
- `GetConfig` lit la configuration depuis le `.env` de l'application et force `MockAI` à `true` si `forceMockIA` est actif ; expose également l'état de `DockerLLM`.
- `SaveConfig` persiste la nouvelle configuration dans le `.env` applicatif puis actualise l'état du client : mise à jour dynamique de `isMock` sauf si forcé par `--mockInterface`, et notification du backend Python via `save-config` uniquement en mode réel.
- `SaveConfigWithDockerLLM` persiste la configuration avec la valeur explicite de `USE_DOCKER_LLM`.
- `StartDockerOllama` lance le conteneur Ollama via `docker compose up -d` depuis la racine de l'application.
- `StopDockerOllama` arrête le conteneur Ollama via `docker compose down`.
- `IsDockerLLM` indique si le mode LLM via Docker est activé.

## Commentaires de fonction (dupliqués du code source)

- `NewBackendClient()` : Instancie le client backend en résolvant automatiquement la racine applicative.
- `NewBackendClientWithAppRoot()` : Instancie le coordinateur avec dépôt cible et racine applicative explicites.
- `BridgeClient.activeDelegate()` : Sélectionne le délégué actif selon le mode simulation.
- `BridgeClient.IsMock()` : Indique si le mode simulation est actif.
- `BridgeClient.GetRepoStatus()` : Délègue la récupération du statut du dépôt.
- `BridgeClient.GetDiff()` : Délègue la récupération du diff.
- `BridgeClient.StageAll()` : Délègue l'indexation de tous les fichiers.
- `BridgeClient.GenerateCommit()` : Délègue la génération de la proposition de commit.
- `BridgeClient.ApplyCommit()` : Délègue l'application du message de commit.
- `BridgeClient.Push()` : Délègue la poussée vers le dépôt distant.
- `BridgeClient.GetReleaseNotes()` : Délègue la génération des notes de version.
- `BridgeClient.PingOllama()` : Délègue la vérification de joignabilité d'Ollama.
- `BridgeClient.SaveMarkdownFile()` : Délègue la sauvegarde du fichier Markdown.
- `BridgeClient.GetConfig()` : Charge la configuration depuis l'environnement applicatif.
- `BridgeClient.SaveConfig()` : Persiste la configuration (USE_DOCKER_LLM conservé) et notifie le backend réel si nécessaire.
- `BridgeClient.SaveConfigWithDockerLLM()` : Persiste la configuration avec USE_DOCKER_LLM explicite.
- `BridgeClient.IsDockerLLM()` : Indique si le mode LLM via Docker est actif.
- `BridgeClient.StartDockerOllama()` : Lance le conteneur Ollama Docker via `docker compose up -d`.
- `BridgeClient.StopDockerOllama()` : Arrête le conteneur Ollama Docker via `docker compose down`.

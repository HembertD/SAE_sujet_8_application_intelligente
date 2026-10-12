# Interface BackendClient et Coordinateur BridgeClient

Fichier source associé : src/cli/bridge/client.go

## Description

Ce fichier définit l'interface maîtresse `BackendClient` et le coordinateur adaptateur `BridgeClient` qui fait le lien entre les vues de l'interface en ligne de commande et les implémentations réelles (`PythonClient`) ou simulées (`MockClient`).

## Responsabilités

- Définir le contrat d'interface haut niveau `BackendClient` regroupant les actions du domaine Git, Ollama et configuration.
- Instancier le coordinateur via `NewBackendClientWithAppRoot(repoRoot, appRoot, forceMockInterface, forceMockIA)`.
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
- `GetConfig` lit la configuration depuis le `.env` de l'application et force `MockAI` à `true` si `forceMockIA` est actif.
- `SaveConfig` persiste la nouvelle configuration dans le `.env` applicatif puis actualise l'état du client : mise à jour dynamique de `isMock` sauf si forcé par `--mockInterface`, et notification du backend Python via `save-config` uniquement en mode réel.

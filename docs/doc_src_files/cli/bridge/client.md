# Interface BackendClient et Coordinateur BridgeClient

Fichier source associé : src/cli/bridge/client.go

## Description

Ce fichier définit l'interface maîtresse `BackendClient` et le coordinateur adaptateur `BridgeClient` qui fait le lien entre les vues de l'interface en ligne de commande et les implémentations réelles (`PythonClient`) ou simulées (`MockClient`).

## Responsabilités

- Définir le contrat d'interface haut niveau `BackendClient` regroupant les actions du domaine Git, Ollama et configuration.
- Instancier le coordinateur via `NewBackendClientWithAppRoot(repoRoot, appRoot, forceDemo)`.
- Maintenir la référence vers le dépôt cible (`repoRoot`) et la racine de l'application (`appRoot`).
- Assurer le routage dynamique vers le délégué actif en fonction du drapeau `--demo` ou du paramètre `MOCK_INTERFACE` du `.env`.
- Fournir les méthodes de consultation et modification de la configuration (`GetConfig`, `SaveConfig`).

## Points clés

- **Inversion des dépendances (DIP)** : Le front-end dépend exclusivement de l'abstraction `BackendClient`.
- **Zéro logique métier en Go** : Les traitements d'analyse Git, LLM et release notes sont délégués au backend Python.

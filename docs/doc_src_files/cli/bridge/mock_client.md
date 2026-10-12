# Implémentation Simulée MockClient

Fichier source associé : src/cli/bridge/mock_client.go

## Description

Ce module implémente l'interface `BackendClient` en fournissant des réponses synthétiques immédiates pour le mode démo hors-ligne et les tests automatisés du front-end.

## Responsabilités

- Simuler un dépôt Git actif avec des modifications réalistes (détection de secrets, fichiers binaires, ajouts/suppressions).
- Générer des messages de commit simulés respectant la convention Conventional Commits.
- Produire des Release Notes formatées en Markdown sans appel réseau.
- Simuler la connectivité et la latence avec un serveur Ollama.
- Permettre la surcharge (`CustomError`, `CustomDiff`, `SimulateDelay = false`) pour les tests unitaires Go.

## Points clés

- **Autonomie totale** : Permet de présenter l'application ou d'exécuter des tests d'interface sans nécessiter de serveur Ollama actif ni de backend Python installé.
- **Conformité aux règles de sécurité** : Démontre la détection des clés API sensibles et l'exclusion des fichiers binaires.

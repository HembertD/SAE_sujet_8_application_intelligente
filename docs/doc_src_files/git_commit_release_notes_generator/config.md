# Configuration

Fichier source associé : src/git_commit_release_notes_generator/config.py

## Description

Ce module contient les paramètres globaux utilisés par le client Ollama et le reste de l’application.

## Variables principales

- OLLAMA_BASE_URL : URL du serveur Ollama.
- OLLAMA_MODEL : modèle utilisé pour la génération des messages de commit.
- OLLAMA_TIMEOUT_S : délai de requête avant timeout.

## Points de vigilance

- La configuration est lue depuis les variables d’environnement, ce qui permet de modifier le comportement sans toucher au code.
- Les valeurs par défaut doivent rester compatibles avec l’environnement local et la CI.
- Les changements de modèle ou de serveur doivent être documentés lors d’un sprint ou d’un réglage de déploiement.

## Exemple de usage

Le client LLM lit cette configuration pour construire les appels HTTP vers Ollama.

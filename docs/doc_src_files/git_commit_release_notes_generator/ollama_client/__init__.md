# Package ollama_client

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/__init__.py

## Rôle

Ce package encapsule la communication avec Ollama pour la génération de messages de commit.

## Sous-modules

- [base.md](base.md) : primitives de base et point d’entrée technique.
- [llm.md](llm.md) : logique de génération de message et validation du résultat.
- [embedding.md](embedding.md) : gestion des embeddings, si cette fonctionnalité est ajoutée plus tard.
- [exceptions.md](exceptions.md) : exceptions métier du client.
- [prompts](prompts/commit_message_system.md) : prompt système utilisé pour imposer la structure du message.

## Objectif

Le package isole les dépendances externes LLM du reste du projet afin de garder la logique métier testable et plus simple à maintenir.

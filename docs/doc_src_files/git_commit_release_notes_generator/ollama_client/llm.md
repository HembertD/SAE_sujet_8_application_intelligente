# Client LLM

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/llm.py

## Description

Ce module gère la génération du message de commit via Ollama. Il construit le résumé du diff, envoie une requête structurée à l’API de chat, valide la réponse JSON et la transforme en objet métier.

## Flux de travail

1. Lire les fichiers différences et les convertir en résumé exploitable.
2. Construire le prompt système et le message utilisateur.
3. Appeler Ollama via une requête JSON structurée.
4. Vérifier que la réponse est bien un JSON conforme.
5. En cas d’erreur, relancer une seconde fois avec le message invalidé en contexte.
6. Convertir la réponse validée en objet CommitMessage.

## Points clés

- format : le modèle répond avec un schéma JSON strict ;
- validation : le code refuse les messages trop longs, absents ou non conformes ;
- robustesse : un retry est tenté pour corriger une réponse invalide.

## Dépendances

- [config.md](../config.md)
- [models.md](../models.md)
- [exceptions.md](exceptions.md)
- [commit_message_system.md](prompts/commit_message_system.md)

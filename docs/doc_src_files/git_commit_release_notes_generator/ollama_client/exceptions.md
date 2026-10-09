# Exceptions Ollama

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/exceptions.py

## Description

Ce module définit les erreurs spécifiques à la couche LLM et aux échanges avec Ollama.

## Exceptions prévues

- OllamaCallError : échec de l’appel au serveur.
- CommitMessageValidationError : validation du JSON de réponse impossible ou incohérente.

## Pourquoi les centraliser ?

Cela permet de distinguer clairement :

- les erreurs réseau ;
- les erreurs de format de réponse ;
- les erreurs métier liées au contenu généré.

Cette séparation simplifie le diagnostic et le test des cas limites.

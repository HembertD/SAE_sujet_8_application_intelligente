# Exceptions Ollama

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/exceptions.py

## Description

Ce module définit les erreurs spécifiques à la couche LLM et aux échanges avec Ollama.

## Exceptions prévues

- OllamaCallError : échec de l’appel au serveur.
- MalformedResponseError (sous-classe d'OllamaCallError) : le contenu renvoyé n'est pas du JSON valide malgré le `format` contraint. Porte le contenu brut (`.content`) pour permettre à `generate_commit_message()` de le réinjecter dans la conversation et de retenter une fois, au lieu de lever une erreur fatale immédiate.
- CommitMessageValidationError : validation du JSON de réponse impossible ou incohérente, y compris après un retry déclenché par une réponse non-JSON.

## Pourquoi les centraliser ?

Cela permet de distinguer clairement :

- les erreurs réseau ;
- les erreurs de format de réponse ;
- les erreurs métier liées au contenu généré.

Cette séparation simplifie le diagnostic et le test des cas limites.

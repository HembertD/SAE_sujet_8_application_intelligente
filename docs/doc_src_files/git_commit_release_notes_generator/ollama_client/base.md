# Base Ollama client

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/base.py

## Description

Ce module sert de base commune au client Ollama. Il définit les contrats et les éléments de fondation pour les intégrations avec le modèle.

## Rôle attendu

- centraliser les interfaces de communication ;
- préparer les appels HTTP ;
- exposer une structure cohérente pour les autres sous-modules.

## Utilisation

Les modules spécialisés reposent sur cette base pour éviter de dupliquer la logique de transport et de gestion des erreurs.

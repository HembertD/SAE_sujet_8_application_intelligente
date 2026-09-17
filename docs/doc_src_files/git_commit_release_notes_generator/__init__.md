# Module git_commit_release_notes_generator

Fichier source associé : src/git_commit_release_notes_generator/__init__.py

## Rôle

Ce package regroupe le cœur de la génération des messages de commit et des notes de release. Il centralise les données de configuration, les structures de données partagées et les clients externes qui interagissent avec Ollama.

## Sous-modules

- [config.md](config.md) : paramètres de configuration de l’application.
- [models.md](models.md) : modèles de données partagés.
- [ollama_client](ollama_client/__init__.md) : couche de communication avec Ollama.
- [service](service/__init__.md) : logique métier et traitements applicatifs.

## Objectif fonctionnel

Le package sert de point d’entrée logique pour :

- analyser un diff Git ;
- filtrer les données sensibles ;
- transmettre un contexte structuré au modèle LLM ;
- générer un message de commit conforme à Conventional Commits ;
- produire éventuellement des éléments exploitables pour des notes de version.

# Prompt système de génération de commit

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/prompts/commit_message_system.txt

## Description

Ce fichier contient le prompt système qui impose à l’IA de renvoyer un message de commit structuré et cohérent avec la convention Conventional Commits.

## Objectif

Le prompt sert à guider le modèle vers une sortie stable avec :

- un type de commit valide ;
- un sujet concise ;
- une structure de message compréhensible ;
- un format exploitable par le code sans ambiguïté.

## Rôle dans le flux

Il est utilisé par le module [llm.md](../llm.md) comme contexte système avant l’envoi du diff Git au modèle.

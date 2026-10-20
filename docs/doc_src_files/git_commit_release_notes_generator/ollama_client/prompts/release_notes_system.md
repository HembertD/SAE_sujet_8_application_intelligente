# Prompt système de génération de release notes

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/prompts/release_notes_system.txt

## Description

Ce fichier contient le prompt système qui guide l'IA pour rédiger des release notes en Markdown, à partir d'une liste de commits entre deux tags.

## Objectif

Le prompt sert à guider le modèle vers une sortie stable avec :

- un contenu Markdown pur, sans balises englobantes ;
- une organisation fixe en sections (Nouveautés, Corrections de bogues, Documentation & Maintenance) ;
- l'omission des sections sans commit correspondant plutôt que de les laisser vides ;
- aucun texte en dehors de l'objet JSON attendu (`{"markdown": ...}`).

## Rôle dans le flux

Il est utilisé par le module [release_notes.md](../release_notes.md) comme contexte système avant l'envoi du résumé des commits au modèle.

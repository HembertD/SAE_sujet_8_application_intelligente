# Prompt système de résolution de conflit

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/prompts/conflict_resolution_system.txt

## Description

Ce fichier contient le prompt système qui guide l'IA pour proposer une résolution à une section de conflit de merge Git, à partir des versions "ours" et "theirs" (et "base" si disponible).

## Objectif

Le prompt sert à guider le modèle vers une sortie stable avec :

- un code final qui préserve l'intention des deux côtés quand ils sont compatibles (combiner plutôt qu'écraser) ;
- un signalement explicite quand les deux côtés semblent réellement incompatibles ;
- aucun marqueur de conflit (`<<<<<<<`, `=======`, `>>>>>>>`) dans le code proposé ;
- une explication en français destinée à l'utilisateur, qui reste seul décisionnaire : la proposition n'est jamais appliquée automatiquement.

## Rôle dans le flux

Il est utilisé par le module [conflict_resolver.md](../conflict_resolver.md) comme contexte système avant l'envoi des deux côtés du conflit au modèle.

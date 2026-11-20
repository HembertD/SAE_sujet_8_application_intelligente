# Prompt système de description d'image

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/prompts/image_description_system.txt

## Description

Ce fichier contient le prompt système qui guide le modèle de vision pour décrire le contenu d'une image ajoutée/supprimée, ou le changement visuel entre deux versions d'une image modifiée.

## Objectif

Le prompt sert à guider le modèle vers une sortie stable avec :

- une interprétation correcte du nombre d'images reçues (une = contenu à décrire, deux = avant/après à comparer) ;
- une description concise (1 à 2 phrases), en français, centrée sur ce qui est pertinent pour un message de commit ;
- aucun texte en dehors de l'objet JSON attendu (`{"description": ...}`).

## Rôle dans le flux

Il est utilisé par le module [vision.md](../vision.md) comme contexte système avant l'envoi des images au modèle.

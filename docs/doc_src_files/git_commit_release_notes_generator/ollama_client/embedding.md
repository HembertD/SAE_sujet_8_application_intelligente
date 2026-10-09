# Embedding Ollama

Fichier source associé : src/git_commit_release_notes_generator/ollama_client/embedding.py

## Description

Ce module est dédié aux embeddings, c’est-à-dire à la transformation de texte en vecteurs numériques pour les traitements de similarité ou d’indexation.

## Statut actuel

Le fichier est prévu pour évoluer avec les besoins du projet et peut servir à des cas d’usage futurs, notamment la recherche ou la comparaison de diff/commit.

## Rôle fonctionnel attendu

- convertir des chaînes textuelles en représentations vectorielles ;
- collaborer avec d’autres modules de recherche ou de classification ;
- rester isolé de la logique métier principale.

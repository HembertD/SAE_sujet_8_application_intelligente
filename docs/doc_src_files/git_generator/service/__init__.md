# Service Git

Fichier source associé : src/git_generator/service/__init__.py

## Rôle

Ce package contient les outils d’interaction avec Git et le parsing des différences de dépôt.

## Sous-module principal

- [git_wrapper.md](git_wrapper.md) : wrapper chargé de lire les diff, filtrer les secrets et produire des données structurées.

## Objectif

Cette couche permet d’isoler les opérations Git du reste du code afin de rendre les traitements plus robustes et testables.

# Service du générateur

Fichier source associé : src/git_commit_release_notes_generator/service/__init__.py

## Rôle

Ce package rassemble la logique métier pure de l’application. Il sert à orchestrer les traitements et à exposer les fonctions qui manipulent les diff et les données de génération.

## Sous-modules

- [core.md](core.md) : noyau métier du système.
- [utils.md](utils.md) : utilitaires de traitement et aide fonctionnelle.

## Objectif

Cette couche permet de garder le code de génération séparé des détails de communication réseau ou de parsing Git.

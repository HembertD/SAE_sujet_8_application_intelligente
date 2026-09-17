# Neutralisation des situations de test

Fichier source associé : test/situation_test/deinit_test.py

## Description

Ce script masque les éléments Git et sensibles du projet en ajoutant le suffixe `_test` aux fichiers et dossiers concernés. Il s’agit du mécanisme inverse du script d’activation.

## Rôle fonctionnel

Il protège l’environnement de travail en neutralisant automatiquement :

- `.git` → `.git_test`
- `.gitignore` → `.gitignore_test`
- `.github` → `.github_test`
- `.gitattributes` → `.gitattributes_test`
- `.gitmodules` → `.gitmodules_test`
- `.env` → `.env_test`

## Objectif

Cette neutralisation empêche les outils externes, la plateforme de dépôt ou les scanners de sécurité de considérer les dossiers de test comme des ressources réelles du projet.

## Comportement

Le script parcourt les sous-dossiers et renomme les éléments ciblés. Il traite les chemins les plus profonds en premier pour limiter les conflits de renommage lorsque plusieurs éléments sont imbriqués.

## Utilité du projet

C’est une étape de sécurité et de reproducibilité essentielle pour les situations de test interactives, notamment dans des scénarios où le dépôt ou des variables d’environnement ne doivent pas être réellement exploités.

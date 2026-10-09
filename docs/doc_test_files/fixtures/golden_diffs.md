# Fixtures Golden Diff

Fichier source associé : test/fixtures/golden_diffs.py

## Description

Ce module fournit des scénarios de diff déjà préparés, appelés `GoldenDiff`, pour simuler des changements de développement classiques et tester le bon classement des modifications par type de commit.

## Structure de données

Chaque fixture contient :

- un `name` de scénario ;
- un `expected_type` correspondant au type attendu (`feat`, `fix`, `docs`, `refactor`, `test`, etc.) ;
- une liste de `DiffFile` représentant les fichiers modifiés.

## Rôle fonctionnel

Ces fixtures servent de références pour valider la capacité du système à reconnaître la nature d’un changement à partir de son diff, sans dépendre d’un contexte réel de développement.

## Exemple de scénarios

Parmi les cas couverts, on trouve :

- une fonctionnalité ajoutée (feature) ;
- une correction de bug ;
- une documentation mise à jour ;
- un refactoring ;
- l’ajout de tests ;
- une modification de dépendances ou de build.

## Points de vigilance

- Les scénarios doivent rester représentatifs des modifications réelles du projet.
- Les types attendus servent de repères de validation, sans remplacer la logique métier du modèle.
- Les fixtures doivent rester stables pour assurer des tests reproductibles.

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

Trois scénarios volontairement ambigus ont été ajoutés pour sonder les limites réelles du prompt (cf. carte Trello #34, « l'IA confond parfois fix/feat ») :

- `fix_disguised_as_bounds_check` : une validation ajoutée qui corrige en réalité un bug (prix négatif possible), pour vérifier que le modèle privilégie bien "fix" même quand le diff a l'air d'ajouter une capacité ;
- `chore_add_makefile_lint_target` : une modification de Makefile, pour vérifier que le modèle ne la confond pas avec "build" (réservé au packaging/dépendances selon le prompt) ;
- `refactor_hides_rounding_fix` : une extraction de fonction qui corrige au passage un bug d'arrondi, pour vérifier que le modèle priorise "fix" sur "refactor" quand les deux sont mélangés.

Mesuré via `scripts/run_metrics.py` sur le vrai serveur Ollama : 100% de conformité et de type exact sur les 9 scénarios (les 6 historiques + ces 3 nouveaux), un seul retry observé (subject 1 caractère trop long, corrigé automatiquement).

## Points de vigilance

- Les scénarios doivent rester représentatifs des modifications réelles du projet.
- Les types attendus servent de repères de validation, sans remplacer la logique métier du modèle.
- Les fixtures doivent rester stables pour assurer des tests reproductibles.

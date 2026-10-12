# Test réseau réel sur les golden files

Fichier source associé : test/test_golden_diffs_slow.py

## Description

Ce test paramétré fait tourner `generate_commit_message()` contre le vrai serveur Ollama, une fois par golden file (`test/fixtures/golden_diffs.py`). Il vérifie que le client reste correct sur plusieurs types de changement représentatifs (feat, fix, refactor, docs, test, build), pas seulement sur un cas isolé.

## Ce qui est testé

- que chaque scénario produit un `CommitMessage` valide (type et subject non vides) ;
- affichage du type attendu vs obtenu et de la latence, pour repérer les divergences de jugement du modèle sans faire échouer le test pour autant (le type exact n'est pas une assertion stricte, car la génération n'est pas déterministe).

## Stratégie de test

Aucun mock, marqué `slow`, exclu par défaut. Nécessite un accès réseau au serveur Ollama partagé. Compte environ 3 minutes pour les 6 scénarios.

## Cas importants

### comparaison type attendu / type obtenu

Chaque scénario affiche une ligne `[nom_scenario] attendu=X obtenu=Y en Zs`, utile pour suivre à l'œil les divergences (ex. `chore` vs `build`) sans avoir besoin du script de métriques.

## Points de vigilance

- Une divergence de type n'est pas forcément un bug : certaines catégories Conventional Commits (`chore`/`build`) sont ambiguës par nature.
- Ce test ne fait qu'observer ; c'est `scripts/run_metrics.py` qui calcule un taux agrégé sur l'ensemble des scénarios.

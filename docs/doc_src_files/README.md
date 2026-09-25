# Présentation du code source

Cette section regroupe la documentation technique du code applicatif du projet.

## Objet

Elle permet de suivre la structure du dossier `src/` et de comprendre le rôle de chaque module, en gardant une correspondance claire entre le code et sa documentation.

## Organisation

Le dossier `src/` contient le cœur applicatif du projet, notamment :

- la génération des messages de commit
- l’intégration avec Ollama
- la gestion Git et des diff
- les composants CLI
- les services de synthèse et d’analyse

## Structure principale

- `git_commit_release_notes_generator/` : logique métier du projet
- `ollama_client/` : intégration avec les modèles LLM
- `service/` : services de traitement, Git et utilitaires
- `cli/` : interface en ligne de commande et vues utilisateur

## Rôle de cette documentation

Chaque fichier de code a un équivalent Markdown dans ce dossier, avec la même organisation arborescente. Cela permet de :

- retrouver rapidement la description d’un module
- comprendre les responsabilités de chaque composant
- maintenir la documentation synchronisée avec l’évolution du code

## À retenir

La documentation du code source doit rester alignée sur l’implémentation réelle. Quand un module est ajouté, renommé ou modifié, la documentation associée doit être mise à jour dans ce dossier.

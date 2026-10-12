# Modèles de données

Fichier source associé : src/git_commit_release_notes_generator/models.py

## Description

Ce module définit les structures de données transverses utilisées entre les composants du projet.

## Structures principales

### DiffFile

Représente un fichier modifié dans un diff Git.

Champs importants :

- path : chemin du fichier ;
- status : état du fichier (A, M, D, R, etc.) ;
- added / removed : nombre de lignes ajoutées et supprimées ;
- is_binary : indique si le fichier est binaire ;
- patch : contenu brut du diff.

### CommitMessage

Représente le message final de commit produit par le modèle.

Champs importants :

- type : type Conventional Commits (feat, fix, docs, chore, etc.) ;
- scope : portée optionnelle ;
- subject : sujet principal ;
- body : description détaillée optionnelle.

## Rôle dans le flux

Les objets de ce module servent de contrat entre :

- le parsing Git ;
- la génération de texte via Ollama ;
- l’assemblage final du message de commit.

Ils évitent les chaînes non structurées et rendent le code plus robuste et testable.

## Détails déplacés du code source

- Contrat d'équipe : ne pas modifier ces structures sans en discuter avec l'équipe. `DiffFile` est produit par le module de parsing Git (module B) et `CommitMessage` est consommé par la couche d'affichage Go.
- `DiffFile.status` : codes de statut Git (`A`, `M`, `D`, `R`, etc.). Le constructeur accepte `is_binary` ou l'alias `binary` (mot-clé uniquement) ; si les deux sont fournis avec des valeurs divergentes, il lève `TypeError`, sinon `is_binary` prime, défaut `False`.
- Propriété `DiffFile.binary` : alias rétrocompatible avec l'ancien contrat historique, renvoie `is_binary`.
- `CommitMessage.type` : types Conventional Commits (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`).

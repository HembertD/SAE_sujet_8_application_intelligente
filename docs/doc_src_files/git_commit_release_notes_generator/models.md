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

### ConflictSection

Représente une section de conflit de merge extraite d'un fichier.

Champs importants :

- file_path : chemin du fichier concerné ;
- ours : version de notre branche ;
- theirs : version de la branche entrante ;
- base : version de l'ancêtre commun, si disponible (stratégie de merge diff3), sinon `None` ;
- section_index : index de la section dans le fichier ;
- start_line / end_line : bornes de la section dans le fichier de travail (1-based) ;
- context_before / context_after : lignes de contexte autour de la section ;
- conflict_type : type de conflit Git (`both_modified`, `both_added`, etc.) ;
- resolution_supported : indique si le résolveur peut proposer une résolution ;
- source : origine de la section (`markers` pour les marqueurs du fichier, `index` pour un repli sur les stages Git).

Les nouveaux champs ont des valeurs par défaut : les appels existants qui
construisent une `ConflictSection` avec seulement `file_path`, `ours`,
`theirs` et `base` restent compatibles.

### ConflictFile

Représente un fichier en conflit, avec ses sections et ses versions complètes.

Champs importants :

- path : chemin du fichier ;
- conflict_type : type de conflit déduit des stages Git ;
- resolution_supported : indique si une résolution textuelle est possible ;
- is_binary : vrai pour un conflit binaire ;
- sections : liste des `ConflictSection` extraites ;
- base_content / ours_content / theirs_content : contenus complets des stages
  Git (`:1:`, `:2:`, `:3:`) quand ils existent ;
- reason : explication lorsqu'une résolution n'est pas supportée ;
- truncated : indique qu'un contenu a été tronqué pour le payload.

### GitState

Décrit l'état Git utile pour adapter le payload envoyé au LLM.

Champs importants :

- branch / head : branche active et SHA court du commit courant ;
- operation : `none`, `merge`, `rebase`, `cherry-pick` ou `revert` ;
- in_progress : opération Git en cours ;
- has_conflicts : présence de chemins non fusionnés ;
- merge_head / merge_heads : SHA court(s) de la ou des branches entrantes ;
- merge_base : ancêtre commun avec `MERGE_HEAD` ;
- head_parents : parents du commit courant ;
- is_merge_commit : vrai si le commit courant est un commit de merge.

### ConflictResolution

Représente une proposition de résolution de conflit produite par le modèle.

Champs importants :

- resolved_code : le code final proposé pour la section en conflit ;
- explanation : l'explication du choix de résolution, à destination de l'utilisateur.

## Rôle dans le flux

Les objets de ce module servent de contrat entre :

- le parsing Git ;
- la génération de texte via Ollama ;
- l’assemblage final du message de commit (ou de la résolution de conflit).

Ils évitent les chaînes non structurées et rendent le code plus robuste et testable.

## Détails déplacés du code source

- Contrat d'équipe : ne pas modifier ces structures sans en discuter avec l'équipe. `DiffFile`, `ConflictSection`, `ConflictFile` et `GitState` sont produits par le module de parsing Git (module B, `GitWrapper`) ; `CommitMessage` et `ConflictResolution` sont consommés par la couche d'affichage Go.
- `DiffFile.status` : codes de statut Git (`A`, `M`, `D`, `R`, etc.). Le constructeur accepte `is_binary` ou l'alias `binary` (mot-clé uniquement) ; si les deux sont fournis avec des valeurs divergentes, il lève `TypeError`, sinon `is_binary` prime, défaut `False`.
- Propriété `DiffFile.binary` : alias rétrocompatible avec l'ancien contrat historique, renvoie `is_binary`.
- `CommitMessage.type` : types Conventional Commits (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`).
- `ConflictResolution` n'est jamais appliquée automatiquement au fichier : la couche d'affichage Go doit toujours la soumettre à validation humaine avant toute écriture sur disque, comme pour `CommitMessage`.

## Commentaires de fonction (dupliqués du code source)

- `DiffFile.__init__()` : Initialise un fichier de diff en réconciliant les indicateurs binaires.
- `DiffFile.binary()` : Expose l'indicateur binaire via l'alias historique.
- `DiffFile.to_dict()` : Sérialise le fichier de diff en dictionnaire.
- `ConflictSection.to_dict()` : Sérialise une section de conflit, contexte et métadonnées inclus.
- `ConflictFile.to_dict()` : Sérialise un fichier en conflit avec ses versions complètes et ses sections.
- `GitState.to_dict()` : Sérialise l'état Git pour le payload LLM.

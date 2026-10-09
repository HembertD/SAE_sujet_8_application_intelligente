# Composants de Boîtes et Tableaux UI

Fichier source associé : src/cli/ui/box.go

## Description

Ce module assemble les primitives de style pour construire des composants visuels d'interface riches : bannières contextuelles, cadres arrondis Unicode, tableaux de modifications Git et cartes de messages Conventional Commits.

## Responsabilités

- `PrintBanner` : affiche la bannière d'en-tête dynamique du projet avec le statut Git fourni par Python (dépôt, branche, nombre de fichiers staged/non-staged) ;
- `PrintCard` : trace des panneaux encadrés rectilignes avec titre centré et couleur de bordure personnalisée ;
- `PrintDiffTable` : affiche le tableau récapitulatif des fichiers modifiés avec coloration syntaxique du statut (`[M]` jaune/cyan, `[A]` vert, `[D]` rouge), métriques de lignes `+`/`-`, et badge d'avertissement `🔒 [Secret Masqué]` ;
- `PrintCommitCard` : présente la proposition de commit mise en valeur (Type en vert gras, Scope en jaune, Sujet en blanc gras, Body séparé par un filet discret) ;
- `PrintSuccess`, `PrintError`, `PrintWarning`, `PrintInfo` : affiche des alertes typées et colorées standardisées.

## Points clés

- Rendu professionnel basé sur les caractères de traçage de boîte Unicode (`╭`, `╮`, `╯`, `╰`, `─`, `│`) ;
- Adaptation automatique de la largeur des cadres (`width`) ;
- Visibilité immédiate des mécanismes de sécurité (secrets masqués).

## Rôle dans l’architecture

Bibliothèque de composants graphiques d'ordre supérieur consommée par les vues interactives (`views`).

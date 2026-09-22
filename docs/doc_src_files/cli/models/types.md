# Modèles de Données Typées CLI

Fichier source associé : src/cli/models/types.go

## Description

Ce module regroupe l'ensemble des structures de données Go représentant les échanges JSON entre le front-end Go et le back-end Python. Il assure un typage strict et évite les manipulations de chaînes de caractères brutes.

## Structures principales

### RepoStatus
Représente l'état contextuel du dépôt Git transmis par Python :
- `IsGitRepo` (`bool`) : indique si le répertoire courant est un dépôt Git valide ;
- `RepoPath` (`string`) : chemin absolu de la racine du dépôt ;
- `Branch` (`string`) : nom de la branche courante ;
- `StagedCount` (`int`) : nombre de fichiers indexés (`staged`) prêts pour le commit ;
- `UnstagedCount` (`int`) : nombre de fichiers modifiés non indexés.

### DiffFile
Détail d'un fichier extrait du diff Git par le backend :
- `Path` (`string`) : chemin relatif du fichier modifié ;
- `Status` (`string`) : type d'opération (`M` pour modifié, `A` pour ajouté, `D` pour supprimé) ;
- `Added` / `Removed` (`int`) : volume de lignes ajoutées et supprimées ;
- `Binary` (`bool`) : indicateur de fichier binaire (exclu de l'analyse LLM) ;
- `HasSecretsMasked` (`bool`) : signal visuel indiquant si un secret/token a été assaini par sécurité.

### CommitProposal
Proposition de commit générée par le modèle LLM au format Conventional Commits :
- `Type` (`string`) : catégorie conventionnelle (`feat`, `fix`, `docs`, `refactor`, etc.) ;
- `Scope` (`string`, optionnel) : sous-système concerné ;
- `Subject` (`string`) : message descriptif concis ;
- `Body` (`string`, optionnel) : explications détaillées de la modification ;
- `RawFormatted` (`string`) : message assemblé prêt pour `git commit`.

### ConfigResponse & PingResponse
Paramètres de configuration lus dans le `.env` et diagnostic réseau vers le serveur Ollama de l'IUT (latence et liste des modèles disponibles).

## Rôle dans l’architecture

Contrat d'interface formel entre la sortie standard JSON du backend Python et les vues d'affichage en Go.

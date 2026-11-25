# Tests Unitaires de l'Orchestrateur GUI

Fichier source associé : `src/gui/app_test.go`

## Description

Ce fichier regroupe l'ensemble des tests automatisés de l'orchestrateur GUI (`App`) sous Wails v2. Il s'assure du bon fonctionnement des accesseurs, de l'isolation des dépôts, des opérations de délégation vers le pont Go/Python, de la gestion du mode mock et du cycle de vie des options de configuration.

## Scénarios de Test

- `TestNewAppAndGetters` : valide l'initialisation de la structure `App`, la conservation des chemins (`repoRoot`, `appRoot`), l'activation du mode développeur (`devMode`) et l'exactitude des métriques système retournées par `GetSystemSpecs` ;
- `TestAppGetRepoStatusMock` : vérifie la récupération de l'état du dépôt Git simulé ;
- `TestAppGetDiffMock` : vérifie la génération et le formatage des structures de diffs en mémoire ;
- `TestAppStageAllMock` : s'assure que l'action d'indexation complète renvoie un résultat d'action positif ;
- `TestAppGenerateCommitMock` : teste la génération de proposition Conventional Commits sans et avec feedback textuel interactif ;
- `TestAppApplyCommitAndPushMock` : valide l'application du commit et l'opération de push vers le serveur distant ;
- `TestAppReleaseNotesAndRecentCommitsMock` : s'assure de la génération du Markdown des release notes et de la récupération de la liste des commits récents ;
- `TestAppConfigAndPingMock` : teste la lecture des paramètres de configuration, le diagnostic de ping du serveur Ollama, et l'enregistrement de fichiers Markdown sur le disque ;
- `TestAppSetTargetRepo` : valide le changement dynamique de répertoire cible et le rejet explicite d'un chemin vide ;
- `TestAppDockerOperationsMock` : vérifie la cohérence des opérations de démarrage et d'arrêt de conteneur simulées.

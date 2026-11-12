# Implémentation Réelle PythonClient

Fichier source associé : src/cli/bridge/python_client.go

## Description

Ce module implémente l'interface `BackendClient` en exécutant le backend Python (`git_commit_release_notes_generator.service.core`) sous la forme de sous-processus système via `os/exec`.

## Responsabilités

- Résoudre dynamiquement l'interpréteur Python (`resolvePythonBin`), par ordre de priorité : `PYTHON_BIN` (variable système ou `.env` applicatif) → environnement virtuel autonome `<appRoot>/.venv` (créé par `scripts/setup_python_env.py`). Aucun repli sur l'interpréteur système (ni `python3` ni `python`) : si aucun des deux n'est trouvé, l'erreur est conservée sur le client et renvoyée au premier appel réel plutôt que de démarrer avec un interpréteur non maîtrisé (voir `CheckPythonEnvironment`, utilisé par `main.go` pour un échec immédiat au lancement).
- Configurer les variables d'environnement du sous-processus : `PYTHONPATH` pointant sur `<appRoot>/src` (séparateur `os.PathListSeparator`, portable Windows/Unix), et transmission de `GITELLIGENCE_APP_ROOT` (avec maintien de `SMART_COMMIT_APP_ROOT` pour rétrocompatibilité).
- Choisir l'URL Ollama effective via `EffectiveOllamaBaseURL` : `http://localhost:11434` quand `USE_DOCKER_LLM=true`, sinon l'URL distante du `.env`. La variable `USE_DOCKER_LLM` est également transmise au sous-processus Python.
- Piloter le conteneur Ollama local (`DockerOllamaStatus`, `StartDockerOllama` / `StopDockerOllama`) via `docker compose`, avec `cmd.Dir` fixé à la racine applicative et des messages d'erreur explicites (binaire Docker absent, sortie vide). Le statut est lu avec `docker compose ps -q ollama` (sortie non vide = conteneur en cours). Le démarrage utilise `--build` pour reconstruire l'image si le Dockerfile ou `entrypoint.sh` a changé.
- Exécuter les commandes métier (`--action status`, `diff`, `stage-all`, `generate-commit`, `apply-commit`, `push`, `release-notes`, `ping-ollama`).
- Isoler les flux : capture du JSON strict sur `stdout` et redirection des logs/traces sur `stderr`.
- Désérialiser les réponses JSON dans des structures Go fortement typées (`models`).

## Points clés

- **Résilience** : En cas d'erreur du sous-processus, le message `stderr` est extrait et formaté pour un affichage propre sans panic.
- **Isolation du dépôt cible** : Le dépôt cible est transmis via `--repo`, tandis que le code et la configuration proviennent exclusivement de l'application.

## Détails déplacés du code source

Précisions issues des commentaires retirés de `python_client.go` :

- `PythonClient` implémente `BackendClient` en communiquant avec le sous-processus Python et ne contient AUCUNE logique métier : tout le traitement est délégué à Python.
- `ResolvePythonBin(appRoot)` (et alias interne `resolvePythonBin`) : fonction exportée résolvant l'interpréteur par ordre de priorité : `PYTHON_BIN` (système puis `.env`), puis `<appRoot>/.venv/Scripts/python.exe` (Windows) ou `<appRoot>/.venv/bin/python` (Unix) si présent. Si un chemin vers le script d'activation (`activate`) est renseigné au lieu du binaire `python`, résout automatiquement le binaire voisin dans le même dossier `bin/`. Si ni l'un ni l'autre n'existe, retourne une erreur explicite guidant vers `scripts/setup_python_env.py` ou `PYTHON_BIN` (*fail fast*, plutôt qu'un repli silencieux suivi d'un `ModuleNotFoundError` différé). Si `appRoot` est vide, résout automatiquement la racine via `FindAppRoot()`.
- `VerifyPythonEnvironment(appRoot)` : exécute une sonde en sous-processus pour vérifier en profondeur que l'interpréteur résolu fonctionne, que les dépendances requises (`GitPython`, `PyYAML`) sont installées et que le module applicatif backend peut être chargé avec succès sans erreur d'import.
- `VerifyPythonCandidate(appRoot, candidatePyBin, isAuto)` : vérifie un interpréteur candidat avant acceptation et enregistrement dans le `.env` (en mode `auto` pour tester le `.venv`, ou sur un chemin personnalisé avec tolérance `activate` -> `python`).
- `CheckPythonEnvironment(appRoot)` : exploite `VerifyPythonEnvironment` pour garantir au démarrage du CLI (`main.go`) que l'environnement Python est complet et opérationnel, arrêtant immédiatement le programme en cas de dépendance manquante ou d'environnement défaillant (sauté en mode mock-interface).
- `runPythonCommand` exécute la commande Python et récupère `stdout`, après avoir vérifié qu'aucune erreur de résolution d'interpréteur n'est en attente (`pythonBinErr`) ; `PYTHONPATH` pointe sur le dossier `src` de l'application, avec candidats de secours si `appSrc` n'est pas résolu ; la configuration est chargée exclusivement depuis l'application (`p.appRoot`, JAMAIS `p.repoRoot`).
- `IsMock` retourne toujours `false` ; chaque méthode (`GetRepoStatus`, `GetDiff`, `StageAll`, `GenerateCommit`, `ApplyCommit`, `Push`, `GetReleaseNotes`, `GetConfig`, `SaveConfig`, `PingOllama`, `SaveMarkdownFile`) interroge le backend Python pour l'action correspondante, `SaveConfig` notifiant Python après écriture du `.env` applicatif et `SaveMarkdownFile` écrivant localement le changelog exporté.
- `IsDockerLLM` lit `USE_DOCKER_LLM` via `ReadDockerLLMSettingFromEnv(p.appRoot)` (variable d'environnement prioritaire, puis `.env`).
- `dockerComposeOutput` factorise l'exécution d'une commande `docker compose` : vérification de la présence de `docker-compose.yml`, `cmd.Dir` fixé à `p.appRoot`, et retour de la sortie brute.
- `DockerOllamaStatus` exécute `docker compose ps -q ollama` : un identifiant non vide indique que le conteneur tourne. Une indisponibilité de Docker n'est pas une erreur fatale ; elle est exposée dans `DockerStatusResponse.Error` pour l'affichage TUI.
- `runDockerCompose` factorise `docker compose up -d --build` / `down` via `dockerComposeOutput`, avec repli sur `err.Error()` si `CombinedOutput` ne renvoie aucune sortie (ex: binaire `docker` introuvable). Le message de démarrage précise qu'il s'agit de l'exécution de la commande, pas d'une garantie que le conteneur restera actif : le téléchargement du modèle peut échouer ensuite et l'état réel doit être vérifié via `DockerOllamaStatus`.

## Commentaires de fonction (dupliqués du code source)

- `NewPythonClient()` : Instancie le client Python en résolvant la racine applicative.
- `NewPythonClientWithAppRoot()` : Instancie le client Python avec racines explicites et interpréteur résolu.
- `PythonClient.runPythonCommand()` : Exécute le module Python et capture sa sortie standard.
- `PythonClient.IsMock()` : Indique que le client réel n'est jamais en simulation.
- `PythonClient.GetRepoStatus()` : Interroge le backend Python pour le statut du dépôt.
- `PythonClient.GetDiff()` : Interroge le backend Python pour le diff.
- `PythonClient.StageAll()` : Demande au backend Python d'indexer tous les fichiers.
- `PythonClient.GenerateCommit()` : Demande au backend Python une proposition de commit.
- `PythonClient.ApplyCommit()` : Demande au backend Python d'appliquer le commit.
- `PythonClient.Push()` : Demande au backend Python de pousser vers le distant.
- `PythonClient.GetReleaseNotes()` : Demande au backend Python les notes de version.
- `PythonClient.GetConfig()` : Charge la configuration réelle depuis l'environnement.
- `PythonClient.SaveConfig()` : Persiste la configuration puis notifie le backend Python.
- `PythonClient.PingOllama()` : Interroge le backend Python sur l'état d'Ollama.
- `PythonClient.SaveMarkdownFile()` : Sauvegarde le contenu Markdown dans le dépôt cible.
- `PythonClient.IsDockerLLM()` : Lit l'activation du mode Docker depuis l'environnement ou le `.env`.
- `PythonClient.DockerOllamaStatus()` : Vérifie si le conteneur Ollama est en cours via `docker compose ps -q ollama`.
- `PythonClient.StartDockerOllama()` : Lance le conteneur Ollama via `docker compose up -d --build`.
- `PythonClient.StopDockerOllama()` : Arrête le conteneur Ollama via `docker compose down`.

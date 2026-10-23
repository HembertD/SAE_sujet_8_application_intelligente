# Gestionnaire d'Environnement (.env)

Fichier source associé : src/cli/bridge/env.go

## Description

Ce module isole l'ensemble de la logique de localisation, de lecture et d'écriture du fichier de configuration `.env` de l'application Gitelligence.

## Responsabilités

- **Localisation stricte de la racine applicative (`FindAppRoot`)** : exploration hiérarchique ascendante combinant `runtime.Caller`, `os.Executable` et `os.Getwd` pour déterminer l'emplacement physique du dépôt applicatif, indépendamment du dépôt Git cible analysé (`--repo`).
- **Garantie d'emplacement (`FindAppEnvPath`)** : renvoie l'emplacement unique `<racine_projet>/.env`. Élimine tout sous-chemin `src/` ou `cli/` pour empêcher la création ou lecture d'un `.env` dans les sous-dossiers.
- **Lecture typée (`LoadConfigFromEnv`)** : parse les variables `OLLAMA_BASE_URL`, `OLLAMA_MODEL`, `OLLAMA_TIMEOUT_S`, `MOCK_INTERFACE`, `MOCK_AI`, `USE_DOCKER_LLM` et `ENABLE_NOTIFICATIONS` avec valeurs par défaut de repli sécurisées.
- **Résolution de l'interpréteur Python (`ReadPythonBinFromEnv`)** : lit `PYTHON_BIN` (variable système prioritaire, puis `.env` applicatif), consommée par `python_client.go` pour choisir l'interpréteur sans dépendre d'un chemin codé en dur.
- **Persistance atomique (`WriteEnvFile` / `WriteEnvFileWithDockerLLM` / `WriteNotificationsSetting`)** : sauvegarde à chaud les réglages dans le `.env` à la racine de l'application (y compris `MOCK_INTERFACE`, `MOCK_AI`, `USE_DOCKER_LLM` et `ENABLE_NOTIFICATIONS`), purge préventivement tout fichier `.env` ou `.env.exemple` résiduel dans `src/`, et **préserve** les paramètres non gérés par cet écran (ex: `OLLAMA_NUM_CTX`, `PYTHON_BIN`) au lieu de les écraser silencieusement.
- **Parsing booléen robuste (`ParseBoolFromString`)** : interprète les représentations usuelles (`true`, `1`, `yes`, `oui`, `false`, `0`, etc.).

## Points clés

- **Isolation stricte** : Aucun fichier `.env` n'est créé ou lu dans `src/`, ni dans le projet externe analysé par l'outil.
- **Prise d'effet immédiate** : Les modifications saisies dans le TUI sont écrites sur disque et rechargées dynamiquement sans aucune recompilation du binaire Go.

## Détails déplacés du code source

Précisions issues des commentaires retirés de `env.go` :

- `isAppRoot` vérifie si un répertoire est la racine Smart Commit (présence de `src/git_commit_release_notes_generator`, exclusion des dossiers de base `src` / `cli`) ; `isTestDir` détecte les répertoires temporaires de test (préfixe `os.TempDir()` ou sous-chaîne `Test`, ex. `t.TempDir()`).
- `FindAppRoot` suit un ordre de recherche : hints valides d'abord, puis variable `SMART_COMMIT_APP_ROOT`, puis remontée depuis le code source Go (`runtime.Caller`), depuis l'exécutable Go, depuis le répertoire de travail, avec repli sur le premier hint non vide ou `"."` .
- `FindAppEnvPath` renvoie l'emplacement STRICTEMENT à la racine et ne cherche JAMAIS dans `src/` ni dans le dépôt cible ; `FindEnvPath` est un alias conservé pour compatibilité avec le code existant et les tests.
- `ParseBoolFromString` convertit `"true"`, `"1"`, `"yes"`, `"oui"` en booléen ; `ReadMockSettingFromEnv` / `ReadMockAISettingFromEnv` lisent respectivement `MOCK_INTERFACE` (avec variante historique `MOCKAGE_INTERFACE`) et `MOCK_AI`, avec priorité aux variables d'environnement système directes puis lecture du `.env` applicatif.
- `LoadConfigFromEnv` purge par sécurité tout `.env` résiduel dans `src/` ; `WriteEnvFile` (et sa variante `WriteEnvFileWithDockerLLM`) réécrit le `.env` STRICTEMENT à la racine, supprime préventivement tout ancien `src/.env`, et recopie telles quelles les lignes `clé=valeur` non gérées par cet écran (lues dans le `.env` existant avant réécriture) pour ne jamais les perdre.
- `ReadPythonBinFromEnv` suit le même ordre de priorité que les autres lecteurs `.env` du module : variable d'environnement système directe, puis ligne `PYTHON_BIN=` du `.env` applicatif ; renvoie une chaîne vide si rien n'est configuré explicitement (détection automatique à faire par l'appelant).
- `ReadDockerLLMSettingFromEnv` lit la variable `USE_DOCKER_LLM` avec priorité aux variables d'environnement système directes puis lecture du `.env` applicatif ; détermine si l'application doit s'appuyer sur un serveur Ollama local lancé via Docker.
- `ReadEnableNotificationsFromEnv` lit la variable `ENABLE_NOTIFICATIONS` avec priorité aux variables d'environnement système directes puis lecture du `.env` applicatif ; `WriteNotificationsSetting` persiste ce réglage sans écraser les autres variables du fichier.

## Commentaires de fonction (dupliqués du code source)

- `isAppRoot()` : Vérifie si un répertoire est la racine applicative.
- `isTestDir()` : Détecte si un chemin correspond à un répertoire temporaire de test.
- `FindAppRoot()` : Localise la racine applicative à partir des indices et de l'environnement.
- `FindAppEnvPath()` : Renvoie le chemin du fichier .env à la racine applicative.
- `FindEnvPath()` : Fournit un alias de compatibilité pour le chemin du .env.
- `ParseBoolFromString()` : Interprète une chaîne comme booléen.
- `ReadMockSettingFromEnv()` : Lit l'activation du mode simulation depuis l'environnement.
- `ReadMockAISettingFromEnv()` : Lit l'activation de la simulation IA depuis l'environnement.
- `LoadConfigFromEnv()` : Charge la configuration complète depuis le .env applicatif.
- `ReadPythonBinFromEnv()` : Lit l'interpréteur Python explicitement configuré, sinon chaîne vide.
- `WriteEnvFile()` : Écrit la configuration dans le fichier .env applicatif (USE_DOCKER_LLM désactivé par défaut).
- `WriteEnvFileWithDockerLLM()` : Écrit la configuration avec la valeur explicite de USE_DOCKER_LLM.
- `ReadDockerLLMSettingFromEnv()` : Lit l'activation du serveur Ollama via Docker depuis l'environnement.
- `ReadEnableNotificationsFromEnv()` : Lit l'activation des notifications de bureau depuis l'environnement.
- `WriteNotificationsSetting()` : Persiste l'activation des notifications dans le fichier .env applicatif.

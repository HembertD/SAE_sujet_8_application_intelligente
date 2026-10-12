# Gestionnaire d'Environnement (.env)

Fichier source associé : src/cli/bridge/env.go

## Description

Ce module isole l'ensemble de la logique de localisation, de lecture et d'écriture du fichier de configuration `.env` de l'application Smart Commit.

## Responsabilités

- **Localisation stricte de la racine applicative (`FindAppRoot`)** : exploration hiérarchique ascendante combinant `runtime.Caller`, `os.Executable` et `os.Getwd` pour déterminer l'emplacement physique du dépôt applicatif, indépendamment du dépôt Git cible analysé (`--repo`).
- **Garantie d'emplacement (`FindAppEnvPath`)** : renvoie l'emplacement unique `<racine_projet>/.env`. Élimine tout sous-chemin `src/` ou `cli/` pour empêcher la création ou lecture d'un `.env` dans les sous-dossiers.
- **Lecture typée (`LoadConfigFromEnv`)** : parse les variables `OLLAMA_BASE_URL`, `OLLAMA_MODEL`, `OLLAMA_TIMEOUT_S`, `APP_LANGUAGE`, `MOCK_INTERFACE` et `MOCK_AI` avec valeurs par défaut de repli sécurisées.
- **Persistance atomique (`WriteEnvFile`)** : sauvegarde à chaud les réglages dans le `.env` à la racine de l'application (y compris `MOCK_INTERFACE` et `MOCK_AI`) et purge préventivement tout fichier `.env` ou `.env.exemple` résiduel dans `src/`.
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
- `LoadConfigFromEnv` purge par sécurité tout `.env` résiduel dans `src/` ; `WriteEnvFile` réécrit le `.env` STRICTEMENT à la racine et supprime préventivement tout ancien `src/.env` pour ne garder que la racine du projet.

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
- `WriteEnvFile()` : Écrit la configuration dans le fichier .env applicatif.

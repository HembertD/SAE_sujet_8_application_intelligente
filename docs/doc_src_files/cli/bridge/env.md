# Gestionnaire d'Environnement (.env)

Fichier source associé : src/cli/bridge/env.go

## Description

Ce module isole l'ensemble de la logique de localisation, de lecture et d'écriture du fichier de configuration `.env` de l'application Smart Commit.

## Responsabilités

- **Localisation stricte de la racine applicative (`FindAppRoot`)** : exploration hiérarchique ascendante combinant `runtime.Caller`, `os.Executable` et `os.Getwd` pour déterminer l'emplacement physique du dépôt applicatif, indépendamment du dépôt Git cible analysé (`--repo`).
- **Garantie d'emplacement (`FindAppEnvPath`)** : renvoie l'emplacement unique `<racine_projet>/.env`. Élimine tout sous-chemin `src/` ou `cli/` pour empêcher la création ou lecture d'un `.env` dans les sous-dossiers.
- **Lecture typée (`LoadConfigFromEnv`)** : parse les variables `OLLAMA_BASE_URL`, `OLLAMA_MODEL`, `OLLAMA_TIMEOUT_S`, `APP_LANGUAGE` et `MOCK_INTERFACE` avec valeurs par défaut de repli sécurisées.
- **Persistance atomique (`WriteEnvFile`)** : sauvegarde à chaud les réglages dans le `.env` à la racine de l'application et purge préventivement tout fichier `.env` ou `.env.exemple` résiduel dans `src/`.
- **Parsing booléen robuste (`ParseBoolFromString`)** : interprète les représentations usuelles (`true`, `1`, `yes`, `oui`, `false`, `0`, etc.).

## Points clés

- **Isolation stricte** : Aucun fichier `.env` n'est créé ou lu dans `src/`, ni dans le projet externe analysé par l'outil.
- **Prise d'effet immédiate** : Les modifications saisies dans le TUI sont écrites sur disque et rechargées dynamiquement sans aucune recompilation du binaire Go.

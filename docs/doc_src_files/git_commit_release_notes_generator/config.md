# Configuration

Fichier source associé : src/git_commit_release_notes_generator/config.py

## Description

Ce module contient les paramètres globaux utilisés par le client Ollama et le reste de l’application.

## Variables principales

- `OLLAMA_BASE_URL` : URL du serveur Ollama (ex: `http://10.22.28.190:11434` ou local).
- `OLLAMA_MODEL` : modèle utilisé pour la génération des messages de commit (ex: `gemma4:26b`, `gemma4:12b`).
- `OLLAMA_TIMEOUT_S` : délai de requête avant timeout en secondes.
- `APP_LANGUAGE` : langue de l'application et des messages générés (`fr` ou `en`).
- `MOCK_INTERFACE` : activation du mode simulation/démo hors-ligne (`true` ou `false`).

## Emplacement et chargement du fichier `.env`

- **Emplacement strict** : Le fichier `.env` réside **exclusivement à la racine du dépôt du projet** (`SAE_sujet_8_application_intelligente/.env`), ou à l'emplacement indiqué par la variable d'environnement système `SMART_COMMIT_APP_ROOT`.
- **Aucune pollution** : Le module ne lit ni n'écrit jamais de `.env` dans `src/` ou dans le répertoire de travail courant `Path.cwd()`, évitant toute confusion ou parasitage d'un dépôt Git tiers analysé par l'outil.
- **Surcharge** : Les variables définies directement dans l'environnement système ont priorité sur le contenu du fichier `.env`.

## Points de vigilance

- La configuration est lue dynamiquement sans nécessiter de recompilation ni de modification de code.
- Les valeurs par défaut restent compatibles avec l'environnement de l'IUT et la CI.
- Le fichier modèle `.env.exemple` est versionné à la racine pour guider le déploiement initial.

## Exemple d'usage

Le client LLM et le service `core.py` importent cette configuration pour construire les requêtes HTTP vers Ollama et configurer la langue de traitement.


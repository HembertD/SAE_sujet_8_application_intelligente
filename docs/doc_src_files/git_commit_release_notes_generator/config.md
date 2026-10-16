# Configuration

Fichier source associé : src/git_commit_release_notes_generator/config.py

## Description

Ce module contient les paramètres globaux utilisés par le client Ollama et le reste de l’application.

## Variables principales

- `OLLAMA_BASE_URL` : URL du serveur Ollama (ex: `http://10.22.28.190:11434` ou local).
- `OLLAMA_MODEL` : modèle utilisé pour la génération des messages de commit (ex: `gemma4:26b`, `gemma4:12b`).
- `OLLAMA_TIMEOUT_S` : délai de requête avant timeout en secondes.
- `MOCK_INTERFACE` : activation du mode simulation complète de l'interface en Go (`true` ou `false`).
- `MOCK_AI` : activation de la simulation des réponses de l'IA Ollama hors-ligne (`true` ou `false`).

## Emplacement et chargement du fichier `.env`

- **Emplacement strict** : Le fichier `.env` réside **exclusivement à la racine du dépôt du projet** (`SAE_sujet_8_application_intelligente/.env`), ou à l'emplacement indiqué par la variable d'environnement système `GITELLIGENCE_APP_ROOT` (avec repli sur `SMART_COMMIT_APP_ROOT`).
- **Aucune pollution** : Le module ne lit ni n'écrit jamais de `.env` dans `src/` ou dans le répertoire de travail courant `Path.cwd()`, évitant toute confusion ou parasitage d'un dépôt Git tiers analysé par l'outil.
- **Surcharge** : Les variables définies directement dans l'environnement système ont priorité sur le contenu du fichier `.env`.

## Points de vigilance

- La configuration est lue dynamiquement sans nécessiter de recompilation ni de modification de code.
- Les valeurs par défaut restent compatibles avec l'environnement de l'IUT et la CI.
- Le fichier modèle `.env.exemple` est versionné à la racine pour guider le déploiement initial.

## Exemple d'usage

Le client LLM et le service `core.py` importent cette configuration pour construire les requêtes HTTP vers Ollama et initialiser les paramètres d'exécution.

## Détails déplacés du code source

- `parse_bool(val)` : accepte un `bool` (retourné tel quel), ou une chaîne insensible à la casse et aux espaces. Valeurs vraies : `"true"`, `"1"`, `"yes"`, `"oui"`. Toute valeur absente, vide ou `None` vaut `False`.
- `load_env()` : construit une liste de chemins candidats (`.env` sous `SMART_COMMIT_APP_ROOT` si défini, puis `.env` à la racine du dépôt déduite de `Path(__file__).resolve().parent.parent`). Chaque candidat est dédupliqué via `resolve()`, seul le premier fichier existant est lu ; chaque ligne est nettoyée (`strip()`, guillemets retirés) et les lignes vides, commençant par `#` ou sans `=` sont ignorées. Les variables système `OLLAMA_BASE_URL`, `OLLAMA_MODEL`, `OLLAMA_TIMEOUT_S`, `OLLAMA_NUM_CTX`, `APP_LANGUAGE`, `MOCK_INTERFACE`, `MOCK_AI` surchargent ensuite le fichier si elles sont non vides.
- `OLLAMA_NUM_CTX` vaut `"32768"` par défaut : sans ce paramètre explicite, Ollama utilise sa fenêtre de contexte réduite par défaut (souvent 2048-4096 tokens selon le `Modelfile`) même si le modèle supporte bien plus, et tronque silencieusement le début du contexte sur les diffs volumineux.

## Commentaires de fonction (dupliqués du code source)

- `parse_bool()` : Convertit une valeur texte ou booléenne en booléen.
- `load_env()` : Charge les variables depuis le fichier .env et l'environnement système.


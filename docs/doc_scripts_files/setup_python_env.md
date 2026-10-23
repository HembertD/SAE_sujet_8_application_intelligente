# Script de création de l'environnement Python autonome

Fichier source associé : scripts/setup_python_env.py

## Description

Outil de développement/installation (pas le module lui-même) qui crée un environnement virtuel Python dédié dans les fichiers de l'application (`.venv` à la racine) et y installe les dépendances de `requirements.txt`. Résout le problème d'un interpréteur système deviné (`python3` sous Windows, souvent absent) ou d'un chemin d'environnement virtuel personnel codé en dur côté Go.

## Flux de travail

1. Crée `<racine>/.venv` via le module standard `venv` s'il n'existe pas déjà.
2. Met à jour `pip` puis installe `requirements.txt` dans cet environnement.
3. Affiche le chemin de l'interpréteur créé.

## Points clés

- écrit volontairement sur stdout : ce n'est pas le module de production ;
- `.venv/` est ignoré par git (`.gitignore`) : chaque poste crée le sien ;
- le pont Go (`python_client.go`, `resolvePythonBin`) détecte automatiquement cet environnement s'il existe, sans configuration supplémentaire ;
- pour forcer un autre interpréteur (ex: un environnement déjà existant), renseigner `PYTHON_BIN` dans le `.env` plutôt que d'exécuter ce script.

## Dépendances

- [python_client.md](../doc_src_files/cli/bridge/python_client.md)
- [env.md](../doc_src_files/cli/bridge/env.md)

## Usage

```bash
python scripts/setup_python_env.py
```

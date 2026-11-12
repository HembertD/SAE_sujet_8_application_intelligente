# entrypoint.sh - Script d'initialisation du conteneur Ollama

Fichier source associé : `entrypoint.sh` (à la racine du projet)

## Description

Ce script shell est le point d'entrée (`ENTRYPOINT`) du conteneur Docker Ollama. Il est copié dans l'image via le `Dockerfile` et rendu exécutable.

## Comportement

1. Lance `ollama serve` en arrière-plan et mémorise son PID
2. Installe un `trap` `SIGTERM`/`SIGINT` qui transmet proprement l'arrêt au serveur Ollama (évite le `SIGKILL` de Docker après timeout et protège le stockage des modèles)
3. Attend que l'API Ollama soit joignable pendant 30 secondes maximum (1 tentative par seconde) via `ollama list` — `curl` n'est pas présent dans l'image `ollama/ollama:latest`
4. Lit le modèle cible dans `OLLAMA_MODEL` (valeur du `.env` propagée par `docker compose`), avec `gemma4:26b` par défaut (modèle de référence IUT ; `gemma4:12b` possible en local)
5. Si le modèle n'est pas déjà installé, exécute `ollama pull "$OLLAMA_MODEL"`
6. Appelle `wait` sur le PID d'Ollama pour maintenir le conteneur en vie

## Sortie console exemple

```
Ollama server is ready.
Pulling model gemma4:26b (this may take a while)...
...
```

## Usage

Ce script n'est pas appelé directement par un utilisateur ; il est invoqué automatiquement par Docker au lancement du conteneur (`docker compose up`).

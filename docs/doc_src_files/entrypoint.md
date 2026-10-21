# entrypoint.sh - Script d'initialisation du conteneur Ollama

Fichier source associé : `entrypoint.sh` (à la racine du projet)

## Description

Ce script shell est le point d'entrée (`ENTRYPOINT`) du conteneur Docker Ollama. Il est copié dans l'image via le `Dockerfile` et rendu exécutable.

## Comportement

1. Lance `ollama serve` en arrière-plan pour démarrer le serveur Ollama
2. Tente de joindre `http://localhost:11434/api/version` pendant 30 secondes maximum (1 tentative par seconde)
3. Une fois le serveur prêt, vérifie si le modèle `gemma4:26b` est déjà installé (`ollama list`)
4. Si absent, exécute `ollama pull gemma4:26b` pour télécharger le modèle
5. Appelle `wait` pour maintenir le conteneur en vie tant que le serveur Ollama tourne

## Sortie console exemple

```
Ollama server is ready.
Pulling model gemma4:26b (this may take a while)...
...
```

## Usage

Ce script n'est pas appelé directement par un utilisateur ; il est invoqué automatiquement par Docker au lancement du conteneur (`docker compose up`).

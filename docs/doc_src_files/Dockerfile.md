# Dockerfile - Serveur Ollama avec modèle gemma4:26b

Fichier source associé : `Dockerfile` (à la racine du projet)

## Description

Ce fichier définit l'image Docker du serveur Ollama préconfiguré avec le modèle de langage `gemma4:26b`. Il est utilisé par l'application Gitelligence quand l'option `USE_DOCKER_LLM` est activée : le front-end Go pilote le lancement et l'arrêt de ce conteneur via `docker compose`.

## Contenu

- **Image de base** : `ollama/ollama:latest` (serveur Ollama officiel)
- **Modèle pré-téléchargé** : `gemma4:26b` (téléchargé automatiquement au premier démarrage via `entrypoint.sh`)
- **Port exposé** : `11434` (API Ollama standard)
- **Volume de persistance** : `/root/.ollama` (les modèles survivent aux redémarrages du conteneur)

## Script d'initialisation (`entrypoint.sh`)

L'image exécute `/entrypoint.sh` au lancement :

1. Démarre le serveur Ollama en arrière-plan (`ollama serve`)
2. Attend que l'API `http://localhost:11434/api/version` réponde (timeout 30s)
3. Vérifie si `gemma4:26b` est déjà présent ; sinon le télécharge (`ollama pull gemma4:26b`)
4. Maintient le conteneur en vie (`wait`)

## Utilisation

```bash
# Construire l'image
docker build -t gitelligence-ollama .

# Lancer via docker compose (recommandé)
docker compose up -d
```

## Sécurité

Le conteneur n'expose que le port 11434. Aucune autre ouverture réseau n'est configurée.

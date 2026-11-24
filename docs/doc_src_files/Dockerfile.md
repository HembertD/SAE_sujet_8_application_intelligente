# Dockerfile - Serveur Ollama pour Gitelligence

Fichier source associé : `Dockerfile` (à la racine du projet)

## Description

Ce fichier définit l'image Docker du serveur Ollama utilisé par Gitelligence quand l'option `USE_DOCKER_LLM` est activée : le front-end Go pilote le lancement et l'arrêt de ce conteneur via `docker compose`.

## Contenu

- **Image de base** : `ollama/ollama:latest` (serveur Ollama officiel)
- **Modèle par défaut** : `gemma4:26b` (modèle de référence IUT) via `ENV OLLAMA_MODEL`, qui peut être remplacé par la valeur du `.env` applicatif (docker compose). Pour un mode Docker local plus léger, utiliser par exemple `gemma4:12b` dans le `.env`.
- **Port exposé** : `11434` (API Ollama standard)
- **Volume de persistance** : `/root/.ollama` (les modèles survivent aux redémarrages du conteneur)

## Script d'initialisation (`entrypoint.sh`)

L'image exécute `/entrypoint.sh` au lancement :

1. Démarre le serveur Ollama en arrière-plan (`ollama serve`)
2. Installe un `trap` `SIGTERM`/`SIGINT` pour un arrêt propre du serveur
3. Attend que l'API Ollama réponde via `ollama list` (timeout 30s), sans dépendre de `curl`
4. Vérifie si le modèle `OLLAMA_MODEL` est déjà présent ; sinon le télécharge (`ollama pull`)
5. Maintient le conteneur en vie (`wait`)

## Utilisation

```bash
# Construire l'image
docker build -t gitelligence-ollama .

# Lancer via docker compose (recommandé)
docker compose up -d --build
```

## Sécurité

Le conteneur n'expose que le port 11434. Aucune autre ouverture réseau n'est configurée.

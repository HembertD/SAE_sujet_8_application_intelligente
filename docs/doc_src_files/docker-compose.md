# docker-compose.yml - Orchestration du serveur Ollama local

Fichier source associé : `docker-compose.yml` (à la racine du projet)

## Description

Ce fichier définit le service `ollama` qui fournit un serveur LLM local pour Gitelligence. Il est orchestré par Docker Compose et piloté par le front-end Go quand l'option `USE_DOCKER_LLM` est activée.

## Service : `ollama`

| Clé | Valeur |
|-----|--------|
| Build | `.` (Dockerfile à la racine) |
| Nom du conteneur | `gitelligence-ollama` |
| Port hôte | `11434` → `11434` (conteneur) |
| Volume | `ollama_data` → `/root/.ollama` |
| Variable d'environnement | `OLLAMA_HOST=0.0.0.0:11434` |
| Healthcheck | `curl -f http://localhost:11434/api/version` (toutes les 10s) |

## Volumes

- `ollama_data` : persiste les modèles téléchargés entre les arrêts/démarrages du conteneur.

## Utilisation

```bash
# Lancer le serveur Ollama en arrière-plan
docker compose up -d

# Voir les logs
docker compose logs -f ollama

# Arrêter et supprimer le conteneur
docker compose down

# Reconstruire l'image (après modification du Dockerfile)
docker compose build
docker compose up -d
```

## Intégration avec l'application

Quand l'option **LLM via Docker** est activée dans le TUI (menu Configuration > [M] Modifier > `LLM via Docker (true/false)`) :

1. L'utilisateur valide la sauvegarde dans le `.env`
2. L'application propose de démarrer le conteneur Ollama (`docker compose up -d`) avec une barre de progression
3. Le backend Python communique alors avec `http://localhost:11434` au lieu du serveur distant de l'IUT

Quand l'option est désactivée :

1. L'application propose d'arrêter le conteneur (`docker compose down`)
2. Le backend revient au serveur Ollama de l'IUT configuré par `OLLAMA_BASE_URL`

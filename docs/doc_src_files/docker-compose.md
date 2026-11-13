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
| Variables d'environnement | `OLLAMA_HOST=0.0.0.0:11434`, `OLLAMA_MODEL=${OLLAMA_MODEL:-gemma4:26b}` |
| Healthcheck | `ollama list` (toutes les 10s) — `curl` n'existe pas dans l'image officielle |

Le modèle utilisé est celui défini dans le `.env` applicatif (`OLLAMA_MODEL`). La valeur de référence du projet et du serveur IUT est `gemma4:26b` ; pour un mode Docker local sur une machine limitée en RAM, utiliser par exemple `gemma4:12b` (~8 Go) dans le `.env`.

## Volumes

- `ollama_data` : persiste les modèles téléchargés entre les arrêts/démarrages du conteneur.

## Utilisation

```bash
# Lancer le serveur Ollama en arrière-plan en reconstruisant l'image si besoin
docker compose up -d --build

# Voir les logs (utile lors du premier téléchargement du modèle)
docker compose logs -f ollama

# Arrêter et supprimer le conteneur
docker compose down

# Reconstruire explicitement l'image
docker compose build

# Activer le GPU NVIDIA (override optionnel)
docker compose -f docker-compose.yml -f docker-compose.gpu.yml up -d --build
```

## GPU NVIDIA (optionnel)

Le fichier `docker-compose.gpu.yml` ajoute une réservation de périphérique GPU
au service `ollama`. Il est séparé du fichier principal pour que les machines
sans GPU NVIDIA puissent continuer à démarrer le conteneur en CPU sans erreur.
Prérequis : pilote NVIDIA et `nvidia-container-toolkit` installés sur l'hôte.

## Intégration avec l'application

Quand l'option **LLM via Docker** est activée dans le TUI (menu Configuration > [M] Modifier > `LLM via Docker (true/false)`) :

1. L'utilisateur valide la sauvegarde dans le `.env`
2. L'application propose de démarrer le conteneur Ollama (`docker compose up -d --build`) avec une barre de progression
3. Le backend Python communique alors avec `http://localhost:11434` au lieu du serveur distant de l'IUT ; `config.py` applique cette bascule même lorsque le module Python est lancé seul

Quand l'option est désactivée :

1. L'application propose d'arrêter le conteneur (`docker compose down`)
2. Le backend revient au serveur Ollama de l'IUT configuré par `OLLAMA_BASE_URL`

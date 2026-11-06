#!/bin/bash
# entrypoint.sh - Démarre Ollama et télécharge le modèle configuré (OLLAMA_MODEL)

set -e

# Démarre le serveur Ollama en arrière-plan
ollama serve &
OLLAMA_PID=$!

# Transmet proprement SIGTERM/SIGINT au serveur : sans cela, Docker attend son
# timeout puis tue brutalement le conteneur (SIGKILL), ce qui peut corrompre
# le stockage des modèles.
trap 'kill -TERM "$OLLAMA_PID" 2>/dev/null || true; wait "$OLLAMA_PID" 2>/dev/null || true; exit 0' SIGTERM SIGINT

# Attend que l'API Ollama soit joignable (timeout 30s). On utilise la CLI
# `ollama` déjà présente dans l'image : `curl` n'est pas installé dans
# ollama/ollama:latest.
for i in $(seq 1 30); do
    if ollama list > /dev/null 2>&1; then
        echo "Ollama server is ready."
        break
    fi
    echo "Waiting for Ollama server... ($i/30)"
    sleep 1
done

# Modèle cible : celui du .env applicatif (propagé par docker compose),
# sinon gemma4:12b par défaut.
TARGET_MODEL="${OLLAMA_MODEL:-gemma4:12b}"

# Télécharge le modèle configuré s'il n'est pas déjà présent
if ! ollama list | awk 'NR > 1 {print $1}' | grep -qx "$TARGET_MODEL"; then
    echo "Pulling model $TARGET_MODEL (this may take a while)..."
    ollama pull "$TARGET_MODEL"
else
    echo "Model $TARGET_MODEL already present."
fi

# Garde le conteneur en vie ; le trap gère l'arrêt propre.
wait "$OLLAMA_PID"

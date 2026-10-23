#!/bin/bash
# entrypoint.sh - Démarre Ollama et télécharge le modèle gemma4:26b au lancement du conteneur

set -e

# Démarre le serveur Ollama en arrière-plan
ollama serve &

# Attend que l'API Ollama soit joignable (timeout 30s)
for i in $(seq 1 30); do
    if curl -s http://localhost:11434/api/version > /dev/null 2>&1; then
        echo "Ollama server is ready."
        break
    fi
    echo "Waiting for Ollama server... ($i/30)"
    sleep 1
done

# Télécharge le modèle gemma4:26b s'il n'est pas déjà présent
if ! ollama list | grep -q "gemma4:26b"; then
    echo "Pulling model gemma4:26b (this may take a while)..."
    ollama pull gemma4:26b
else
    echo "Model gemma4:26b already present."
fi

# Garde le conteneur en vie
wait

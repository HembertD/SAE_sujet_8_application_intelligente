# Dockerfile - Serveur Ollama pour Gitelligence
# Ce conteneur fournit un serveur Ollama local pour le générateur de commits
# et de release notes quand l'option "LLM via Docker" est activée.

FROM ollama/ollama:latest

# Modèle par défaut si OLLAMA_MODEL n'est pas fourni (docker compose le
# surcharge avec la valeur du .env applicatif).
ENV OLLAMA_MODEL=gemma4:12b

# Copie le script d'initialisation qui démarre Ollama et télécharge le modèle
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Port exposé par Ollama
EXPOSE 11434

# Volume pour persister les modèles entre les redémarrages
VOLUME ["/root/.ollama"]

ENTRYPOINT ["/entrypoint.sh"]

"""Configuration du module C (client Ollama), lue depuis l'environnement."""
import os

OLLAMA_BASE_URL = os.environ.get("OLLAMA_BASE_URL", "http://10.22.28.190:11434")
OLLAMA_MODEL = os.environ.get("OLLAMA_MODEL", "gemma4:26b")

# Généreux : un ~26B peut prendre 15-30s en génération, +10-15s si le modèle
# doit être chargé en mémoire côté serveur (load_duration ~11.4s observé à froid).
OLLAMA_TIMEOUT_S = float(os.environ.get("OLLAMA_TIMEOUT_S", "60"))

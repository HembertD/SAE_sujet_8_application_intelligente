"""Module ollama_client pour l'interaction avec le serveur Ollama."""
from git_commit_release_notes_generator.ollama_client.llm import (
    PingResult,
    check_ollama_reachable,
    generate_commit_message,
    ping_ollama,
)

__all__ = [
    "PingResult",
    "check_ollama_reachable",
    "ping_ollama",
    "generate_commit_message",
]

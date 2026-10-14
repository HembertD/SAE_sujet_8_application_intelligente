"""Module de simulation (Mock) de l'IA Ollama.

Permet de simuler les appels et réponses de l'IA hors-ligne afin de
travailler sur le reste de l'application sans serveur Ollama.
"""
from __future__ import annotations

from typing import Any

from git_commit_release_notes_generator.models import CommitMessage, DiffFile


def mock_generate_commit_message(
    diff_files: list[DiffFile],
    feedback: str = "",
) -> CommitMessage:
    """Génère un message de commit mocké simple et prévisible."""
    if not diff_files:
        raise ValueError("generate_commit_message() nécessite au moins un DiffFile.")

    body = "Commit généré en mode mock pour permettre le travail hors-ligne."
    if feedback and feedback.strip():
        body += f"\nConsigne utilisateur : {feedback.strip()}"

    return CommitMessage(
        type="feat",
        scope="mock",
        subject="commit mocker pour test hors ligne",
        body=body,
    )


def mock_call_chat(messages: list[dict]) -> dict:
    """Simule la charge utile JSON d'un appel chat Ollama (/api/chat)."""
    return {
        "type": "feat",
        "scope": "mock",
        "subject": "commit mocker pour test hors ligne",
        "body": "Réponse brute simulée par le mock IA.",
    }


def mock_release_notes(commits: list[Any] | None = None, from_tag: str = "", to_tag: str = "") -> str:
    """Génère une réponse mockée statique pour les release notes."""
    from_tag = from_tag or "v0.1.0"
    to_tag = to_tag or "HEAD"
    return (
        f"# Release Notes ({from_tag} ➔ {to_tag})\n\n"
        "Release notes mockées pour test hors ligne."
    )


def mock_ping() -> dict:
    """Simule un test de connectivité réussi avec le serveur Ollama."""
    return {
        "success": True,
        "reachable": True,
        "latency_ms": 10,
        "installed_models": [
            "gemma4:26b",
            "gemma4:12b",
            "mistral:latest",
            "codellama:7b",
            "mock-ia:latest",
        ],
        "mock": True,
    }

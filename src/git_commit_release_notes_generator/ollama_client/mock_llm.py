from __future__ import annotations

from typing import Any

from git_commit_release_notes_generator.models import CommitMessage, DiffFile

def mock_generate_commit_message(
    diff_files: list[DiffFile],
    feedback: str = "",
) -> CommitMessage:
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
    return {
        "type": "feat",
        "scope": "mock",
        "subject": "commit mocker pour test hors ligne",
        "body": "Réponse brute simulée par le mock IA.",
    }

def mock_release_notes(commits: list[Any] | None = None, from_tag: str = "", to_tag: str = "") -> str:
    from_tag = from_tag or "v0.1.0"
    to_tag = to_tag or "HEAD"
    return (
        f"# Release Notes ({from_tag} ➔ {to_tag})\n\n"
        "Release notes mockées pour test hors ligne."
    )

def mock_ping() -> dict:
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

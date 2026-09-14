"""Client Ollama pour la génération de messages de commit (module C).

Contrat : DiffFile (entrée) -> CommitMessage (sortie) validé Conventional
Commits. Le modèle renvoie du JSON structuré (contraint côté serveur via le
paramètre `format`), jamais une chaîne déjà formatée : c'est ce module qui
assemble la ligne finale, pour garantir le format par le code plutôt que par
la bonne volonté du LLM.
"""
from __future__ import annotations

import json
import logging
import urllib.error
import urllib.request
from pathlib import Path

from git_commit_release_notes_generator.config import (
    OLLAMA_BASE_URL,
    OLLAMA_MODEL,
    OLLAMA_TIMEOUT_S,
)
from git_commit_release_notes_generator.models import CommitMessage, DiffFile
from git_commit_release_notes_generator.ollama_client.exceptions import (
    CommitMessageValidationError,
    OllamaCallError,
)

logger = logging.getLogger(__name__)

_PROMPT_PATH = Path(__file__).parent / "prompts" / "commit_message_system.txt"

_ALLOWED_TYPES = {
    "feat", "fix", "chore", "docs", "refactor", "test", "perf", "build", "ci",
}
_MAX_SUBJECT_LENGTH = 72

_JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "type": {"type": "string", "enum": sorted(_ALLOWED_TYPES)},
        "scope": {"type": ["string", "null"]},
        "subject": {"type": "string"},
        "body": {"type": ["string", "null"]},
    },
    "required": ["type", "subject"],
    "additionalProperties": False,
}


def generate_commit_message(diff_files: list[DiffFile]) -> CommitMessage:
    """Génère un CommitMessage validé à partir d'une liste de DiffFile.

    Un seul retry est tenté en cas d'échec de validation, en réinjectant
    l'erreur constatée dans la conversation. Si le second essai échoue aussi,
    lève CommitMessageValidationError plutôt que de renvoyer un message
    bancal en silence.
    """
    system_prompt = _PROMPT_PATH.read_text(encoding="utf-8")
    messages: list[dict] = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": _build_diff_summary(diff_files)},
    ]

    raw = _call_chat(messages)
    error = _validate(raw)
    if error is None:
        return _to_commit_message(raw)

    logger.warning("Génération invalide (%s), tentative de retry.", error)
    messages.append({"role": "assistant", "content": json.dumps(raw, ensure_ascii=False)})
    messages.append({
        "role": "user",
        "content": (
            f"Ta réponse précédente est invalide : {error}. "
            "Corrige-la et renvoie uniquement un JSON conforme au schéma."
        ),
    })

    raw = _call_chat(messages)
    error = _validate(raw)
    if error is not None:
        raise CommitMessageValidationError(
            f"Échec de validation après retry : {error}. Réponse brute : {raw!r}"
        )
    return _to_commit_message(raw)


def _build_diff_summary(diff_files: list[DiffFile]) -> str:
    parts = [
        f"Fichier: {f.path} (statut={f.status}, +{f.added}/-{f.removed})\n{f.patch}"
        for f in diff_files
    ]
    return "\n\n".join(parts)


def _call_chat(messages: list[dict]) -> dict:
    body = {
        "model": OLLAMA_MODEL,
        "messages": messages,
        "stream": False,
        "format": _JSON_SCHEMA,
    }
    request = urllib.request.Request(
        url=f"{OLLAMA_BASE_URL}/api/chat",
        data=json.dumps(body).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )

    try:
        with urllib.request.urlopen(request, timeout=OLLAMA_TIMEOUT_S) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except urllib.error.URLError as e:
        raise OllamaCallError(f"Impossible de joindre Ollama à {OLLAMA_BASE_URL} : {e}") from e
    except json.JSONDecodeError as e:
        raise OllamaCallError(f"Réponse non-JSON depuis Ollama : {e}") from e

    if not payload.get("done") or payload.get("done_reason") != "stop":
        raise OllamaCallError(f"Génération incomplète : done_reason={payload.get('done_reason')!r}")

    content = payload.get("message", {}).get("content")
    if not isinstance(content, str):
        raise OllamaCallError(f"Réponse Ollama sans contenu exploitable : {payload!r}")

    try:
        return json.loads(content)
    except json.JSONDecodeError as e:
        raise OllamaCallError(f"Contenu non-JSON malgré le format contraint : {content!r}") from e


def _validate(raw: dict) -> str | None:
    """Retourne None si `raw` est valide, sinon un message d'erreur explicite."""
    commit_type = raw.get("type")
    if commit_type not in _ALLOWED_TYPES:
        return f"type {commit_type!r} hors de la liste autorisée {sorted(_ALLOWED_TYPES)}"

    subject = raw.get("subject")
    if not isinstance(subject, str) or not subject.strip():
        return "subject manquant ou vide"
    if len(subject) > _MAX_SUBJECT_LENGTH:
        return f"subject trop long ({len(subject)} > {_MAX_SUBJECT_LENGTH} caractères)"
    if subject.rstrip().endswith("."):
        return "subject ne doit pas se terminer par un point"

    return None


def _to_commit_message(raw: dict) -> CommitMessage:
    return CommitMessage(
        type=raw["type"],
        scope=raw.get("scope"),
        subject=raw["subject"],
        body=raw.get("body"),
    )

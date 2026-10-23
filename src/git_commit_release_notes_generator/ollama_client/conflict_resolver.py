"""Suggestion de résolution de conflits de merge via Ollama (module C).

Contrat : ConflictSection (entrée, une section de conflit d'un fichier) ->
ConflictResolution (sortie, proposée par le LLM) validée syntaxiquement avant
d'être renvoyée. La proposition n'est JAMAIS appliquée automatiquement au
fichier : c'est toujours à l'utilisateur de la relire et de valider, comme
pour le reste de l'application (aucune sortie LLM n'est écrite sur disque
sans validation humaine).

Périmètre : ce module ne lit ni n'écrit aucun fichier, ne détecte pas les
conflits sur le disque (rôle du GitWrapper, module B) et n'écrit rien sur
stdout. Il prend en entrée le texte déjà extrait des deux côtés du conflit.

ConflictSection (entrée) et ConflictResolution (sortie) sont définies dans
models.py, au même titre que DiffFile/CommitMessage : ce sont des contrats
partagés avec le reste de l'équipe (GitWrapper les produit, la couche
d'affichage Go les consomme).
"""
from __future__ import annotations

import json
import logging
from pathlib import Path

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.models import ConflictResolution, ConflictSection
from git_commit_release_notes_generator.ollama_client.exceptions import (
    ConflictResolutionError,
    MalformedResponseError,
)
from git_commit_release_notes_generator.ollama_client import llm
from git_commit_release_notes_generator.ollama_client.mock_llm import mock_conflict_resolution

logger = logging.getLogger(__name__)

_PROMPT_PATH = Path(__file__).parent / "prompts" / "conflict_resolution_system.txt"

_CONFLICT_MARKERS = ("<<<<<<<", "=======", ">>>>>>>")

_JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "resolved_code": {"type": "string"},
        "explanation": {"type": "string"},
    },
    "required": ["resolved_code", "explanation"],
    "additionalProperties": False,
}


def suggest_conflict_resolution(conflict: ConflictSection) -> ConflictResolution:
    """Propose une résolution pour une section de conflit de merge.

    Un seul retry est tenté si la réponse contient encore des marqueurs de
    conflit non résolus ou un contenu non-JSON, suivant le même principe que
    generate_commit_message(). Si le second essai échoue aussi, lève
    ConflictResolutionError plutôt que de renvoyer une résolution invalide.
    """
    if not conflict.ours.strip() and not conflict.theirs.strip():
        raise ValueError(
            "suggest_conflict_resolution() nécessite au moins un côté du conflit non vide."
        )

    if config.MOCK_AI:
        return mock_conflict_resolution(conflict)

    system_prompt = _PROMPT_PATH.read_text(encoding="utf-8")
    messages: list[dict] = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": _build_conflict_prompt(conflict)},
    ]

    last_error: str | None = None
    for attempt in range(2):
        if attempt == 1:
            messages.append({
                "role": "user",
                "content": (
                    f"Ta réponse précédente est invalide : {last_error}. "
                    "Corrige-la et renvoie uniquement un JSON conforme au schéma, "
                    "sans aucun marqueur de conflit restant."
                ),
            })

        try:
            raw = llm.call_chat_json(messages, _JSON_SCHEMA)
        except MalformedResponseError as e:
            last_error = f"réponse non-JSON malgré le format contraint ({e.content!r})"
            logger.warning("Résolution de conflit invalide (%s), tentative de retry.", last_error)
            messages.append({"role": "assistant", "content": e.content})
            continue

        error = _validate(raw)
        if error is None:
            return ConflictResolution(
                resolved_code=raw["resolved_code"], explanation=raw["explanation"]
            )

        last_error = error
        logger.warning("Résolution de conflit invalide (%s), tentative de retry.", error)
        messages.append({"role": "assistant", "content": json.dumps(raw, ensure_ascii=False)})

    raise ConflictResolutionError(
        f"Échec de la proposition de résolution après retry : {last_error}"
    )


def _build_conflict_prompt(conflict: ConflictSection) -> str:
    parts = [f"Fichier : {conflict.file_path}"]
    if conflict.base:
        parts.append(f"Version de base (ancêtre commun) :\n{conflict.base}")
    parts.append(f'Version "ours" (notre branche) :\n{conflict.ours}')
    parts.append(f'Version "theirs" (branche entrante) :\n{conflict.theirs}')
    return "\n\n".join(parts)


def _validate(raw: dict) -> str | None:
    """Retourne None si `raw` est valide, sinon un message d'erreur explicite."""
    resolved = raw.get("resolved_code")
    if not isinstance(resolved, str) or not resolved.strip():
        return "resolved_code manquant ou vide"
    if any(marker in resolved for marker in _CONFLICT_MARKERS):
        return "resolved_code contient encore des marqueurs de conflit non résolus"

    explanation = raw.get("explanation")
    if not isinstance(explanation, str) or not explanation.strip():
        return "explanation manquante ou vide"

    return None

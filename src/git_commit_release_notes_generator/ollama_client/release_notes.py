"""Génération de release notes via Ollama (module C).

Contrat : une liste de commits (objets exposant `.sha` et `.message`, produits
par le GitWrapper, module B) + deux tags -> un Markdown validé. Même rigueur
que generate_commit_message() et suggest_conflict_resolution() : sortie JSON
contrainte côté serveur, retry unique sur JSON malformé ou contenu vide.

Périmètre : ce module ne lit ni n'écrit aucun fichier, ne résout pas les tags
Git (rôle du GitWrapper) et n'écrit rien sur stdout. Le mode hors-ligne
(MOCK_AI) est géré en interne. L'appelant (service/core.py) ne doit PAS
remplacer silencieusement un échec de cette fonction par un contenu généré
autrement : une exception ici doit remonter comme un échec explicite.
"""
from __future__ import annotations

import json
import logging
from pathlib import Path
from typing import Any

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.ollama_client.exceptions import (
    MalformedResponseError,
    ReleaseNotesError,
)
from git_commit_release_notes_generator.ollama_client import llm
from git_commit_release_notes_generator.ollama_client.mock_llm import mock_release_notes

logger = logging.getLogger(__name__)

_PROMPT_PATH = Path(__file__).parent / "prompts" / "release_notes_system.txt"

_JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "markdown": {"type": "string"},
    },
    "required": ["markdown"],
    "additionalProperties": False,
}


def generate_release_notes(commits: list[Any], from_tag: str, to_tag: str) -> str:
    """Génère le Markdown des release notes pour une liste de commits.

    Un seul retry est tenté si la réponse est non-JSON ou vide, suivant le
    même principe que generate_commit_message(). Si le second essai échoue
    aussi, lève ReleaseNotesError plutôt que de renvoyer un Markdown vide.
    """
    if not commits:
        raise ValueError("generate_release_notes() nécessite au moins un commit.")

    if config.MOCK_AI:
        return mock_release_notes(commits, from_tag, to_tag)

    system_prompt = _PROMPT_PATH.read_text(encoding="utf-8")
    messages: list[dict] = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": _build_commits_prompt(commits, from_tag, to_tag)},
    ]

    last_error: str | None = None
    for attempt in range(2):
        if attempt == 1:
            messages.append({
                "role": "user",
                "content": (
                    f"Ta réponse précédente est invalide : {last_error}. "
                    "Corrige-la et renvoie uniquement un JSON conforme au schéma."
                ),
            })

        try:
            raw = llm.call_chat_json(messages, _JSON_SCHEMA)
        except MalformedResponseError as e:
            last_error = f"réponse non-JSON malgré le format contraint ({e.content!r})"
            logger.warning("Release notes invalides (%s), tentative de retry.", last_error)
            messages.append({"role": "assistant", "content": e.content})
            continue

        error = _validate(raw)
        if error is None:
            return raw["markdown"].strip()

        last_error = error
        logger.warning("Release notes invalides (%s), tentative de retry.", error)
        messages.append({"role": "assistant", "content": json.dumps(raw, ensure_ascii=False)})

    raise ReleaseNotesError(f"Échec de la génération des release notes après retry : {last_error}")


def _build_commits_prompt(commits: list[Any], from_tag: str, to_tag: str) -> str:
    commits_summary = "\n".join(f"- [{c.sha}] {c.message.splitlines()[0]}" for c in commits)
    return (
        f"Version {from_tag} -> {to_tag}.\n"
        f"Voici la liste des commits :\n{commits_summary}"
    )


def _validate(raw: dict) -> str | None:
    """Retourne None si `raw` est valide, sinon un message d'erreur explicite."""
    markdown = raw.get("markdown")
    if not isinstance(markdown, str) or not markdown.strip():
        return "markdown manquant ou vide"
    return None

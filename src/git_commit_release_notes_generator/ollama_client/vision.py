"""Description de contenu/changement d'image via un modèle de vision Ollama
(module C).

Contrat : contenu(s) d'image bruts (bytes) -> description textuelle du
contenu (ajout) ou du changement perçu (modification), pour que les fichiers
image puissent alimenter generate_commit_message() avec un texte exploitable
plutôt que d'être simplement ignorés comme n'importe quel autre binaire.

Périmètre : ce module ne lit ni n'écrit aucun fichier (les octets d'image
sont fournis par l'appelant, GitWrapper côté module B) et n'écrit rien sur
stdout. Il ne décide pas non plus quand l'appeler : action_generate_commit()
(service/core.py) choisit d'invoquer ce module uniquement au moment de
générer un message de commit, pas lors d'une simple extraction de diff
(qui doit rester rapide et sans réseau).
"""
from __future__ import annotations

import base64
import json
import logging
from pathlib import Path

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.ollama_client.exceptions import (
    ImageDescriptionError,
    MalformedResponseError,
)
from git_commit_release_notes_generator.ollama_client import llm
from git_commit_release_notes_generator.ollama_client.mock_llm import mock_describe_image

logger = logging.getLogger(__name__)

_PROMPT_PATH = Path(__file__).parent / "prompts" / "image_description_system.txt"

_JSON_SCHEMA = {
    "type": "object",
    "properties": {
        "description": {"type": "string"},
    },
    "required": ["description"],
    "additionalProperties": False,
}


def describe_image_change(
    file_path: str, status: str, new_image: bytes | None, old_image: bytes | None = None
) -> str:
    """Décrit le contenu d'une image ajoutée/supprimée, ou le changement entre deux versions.

    - `status="A"` (ajout) : seule `new_image` doit être fournie, décrite comme
      un nouveau contenu.
    - `status="D"` (suppression) : seule `old_image` doit être fournie, décrite
      comme le contenu de l'image retirée.
    - sinon (modification/renommage) : les deux images sont envoyées au modèle
      en une seule requête pour qu'il décrive la différence perçue (avant/après).

    Un seul retry est tenté en cas de réponse non-JSON ou de description vide,
    suivant le même principe que generate_commit_message(). Si le second essai
    échoue aussi, lève ImageDescriptionError plutôt que de renvoyer une
    description bancale en silence.
    """
    if not new_image and not old_image:
        raise ValueError("describe_image_change() nécessite au moins une image (new_image ou old_image).")

    if config.MOCK_AI:
        return mock_describe_image(file_path, has_old=bool(old_image) and bool(new_image))

    system_prompt = _PROMPT_PATH.read_text(encoding="utf-8")

    if status == "D":
        images_b64 = [base64.b64encode(old_image).decode("ascii")]
        intent = "Cette image a été supprimée : décris son contenu pour référence."
    elif status == "A" or not old_image:
        images_b64 = [base64.b64encode(new_image).decode("ascii")]
        intent = "Nouvelle image ajoutée : décris son contenu."
    else:
        images_b64 = [
            base64.b64encode(old_image).decode("ascii"),
            base64.b64encode(new_image).decode("ascii"),
        ]
        intent = "Voici la version avant (1ère image) et après (2ème image) : décris ce qui a changé."

    messages: list[dict] = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": f"Fichier : {file_path}. {intent}", "images": images_b64},
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
            logger.warning("Description d'image invalide (%s), tentative de retry.", last_error)
            messages.append({"role": "assistant", "content": e.content})
            continue

        error = _validate(raw)
        if error is None:
            return raw["description"].strip()

        last_error = error
        logger.warning("Description d'image invalide (%s), tentative de retry.", error)
        messages.append({"role": "assistant", "content": json.dumps(raw, ensure_ascii=False)})

    raise ImageDescriptionError(f"Échec de la description d'image après retry : {last_error}")


def _validate(raw: dict) -> str | None:
    """Retourne None si `raw` est valide, sinon un message d'erreur explicite."""
    description = raw.get("description")
    if not isinstance(description, str) or not description.strip():
        return "description manquante ou vide"
    return None

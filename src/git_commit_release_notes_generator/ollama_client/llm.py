"""Client Ollama pour la génération de messages de commit (module C).

Contrat : DiffFile (entrée) -> CommitMessage (sortie) validé Conventional
Commits. Le modèle renvoie du JSON structuré (contraint côté serveur via le
paramètre `format`), jamais une chaîne déjà formatée : c'est ce module qui
assemble la ligne finale, pour garantir le format par le code plutôt que par
la bonne volonté du LLM.
"""
from __future__ import annotations

from dataclasses import dataclass, field
import json
import logging
import time
import urllib.error
import urllib.request
from pathlib import Path

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.config import (
    OLLAMA_BASE_URL,
    OLLAMA_MODEL,
    OLLAMA_NUM_CTX,
    OLLAMA_TIMEOUT_S,
)
from git_commit_release_notes_generator.models import CommitMessage, DiffFile
from git_commit_release_notes_generator.ollama_client.exceptions import (
    CommitMessageValidationError,
    OllamaCallError,
)
from git_commit_release_notes_generator.ollama_client.mock_llm import (
    mock_call_chat,
    mock_generate_commit_message,
    mock_ping,
)

logger = logging.getLogger(__name__)

_PROMPT_PATH = Path(__file__).parent / "prompts" / "commit_message_system.txt"

_ALLOWED_TYPES = {
    "feat", "fix", "chore", "docs", "refactor", "test", "perf", "build", "ci",
}
_MAX_SUBJECT_LENGTH = 72
_PING_TIMEOUT_S = 3.0

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


@dataclass
class PingResult:
    """Résultat de la vérification de joignabilité d'Ollama.

    Permet à la fois l'accès par attributs (.reachable, .latency_ms, .error, .installed_models)
    et le déballage direct en tuple (reachable, error) pour une compatibilité totale.
    """
    reachable: bool
    error: str | None = None
    latency_ms: int = 0
    installed_models: list[str] = field(default_factory=list)

    def __iter__(self):
        return iter((self.reachable, self.error))

    def __getitem__(self, index):
        return (self.reachable, self.error)[index]

    def __eq__(self, other):
        if isinstance(other, tuple) and len(other) == 2:
            return (self.reachable, self.error) == other
        return super().__eq__(other)

    def to_dict(self) -> dict:
        """Formate le résultat en dictionnaire pour le CLI Go."""
        data = {
            "success": True,
            "reachable": self.reachable,
            "latency_ms": self.latency_ms,
            "installed_models": self.installed_models,
        }
        if self.error:
            data["error"] = self.error
        return data


def check_ollama_reachable(timeout: float = _PING_TIMEOUT_S, fetch_models: bool = False) -> PingResult:
    """Vérifie la joignabilité du serveur Ollama et optionnellement les modèles installés.

    Mesure la latence en millisecondes et retourne un PingResult.
    Peut être déballé comme un tuple `reachable, error = check_ollama_reachable()`
    ou manipulé avec ses attributs (`res.reachable`, `res.latency_ms`, `res.error`, `res.installed_models`).
    Ne lève jamais d'exception.
    """
    if config.MOCK_AI:
        mock_data = mock_ping()
        models = mock_data.get("installed_models", []) if fetch_models else []
        return PingResult(reachable=True, latency_ms=10, installed_models=models)

    start = time.perf_counter()
    try:
        req = urllib.request.Request(
            f"{OLLAMA_BASE_URL.rstrip('/')}/api/version",
            headers={"User-Agent": "SmartCommit/1.0"},
        )
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            _ = resp.read()
        latency_ms = int((time.perf_counter() - start) * 1000)

        installed_models: list[str] = []
        if fetch_models:
            try:
                tags_url = f"{OLLAMA_BASE_URL.rstrip('/')}/api/tags"
                req_t = urllib.request.Request(tags_url, headers={"User-Agent": "SmartCommit/1.0"})
                with urllib.request.urlopen(req_t, timeout=timeout) as resp_t:
                    data = json.loads(resp_t.read().decode("utf-8"))
                    raw_models = data.get("models", [])
                    for m in raw_models:
                        if isinstance(m, dict) and "name" in m:
                            installed_models.append(str(m["name"]))
            except Exception as e:
                logger.debug(f"Impossible de récupérer les modèles Ollama : {e}")

        return PingResult(reachable=True, latency_ms=latency_ms, installed_models=installed_models)
    except urllib.error.URLError as e:
        return PingResult(
            reachable=False,
            error=f"Serveur Ollama non joignable à {OLLAMA_BASE_URL} : {e}",
            latency_ms=0,
            installed_models=[],
        )
    except Exception as e:
        return PingResult(
            reachable=False,
            error=f"Serveur Ollama non joignable à {OLLAMA_BASE_URL} : {e}",
            latency_ms=0,
            installed_models=[],
        )


# Alias sémantique
ping_ollama = check_ollama_reachable


def generate_commit_message(diff_files: list[DiffFile], feedback: str = "") -> CommitMessage:
    """Génère un CommitMessage validé à partir d'une liste de DiffFile.

    Un seul retry est tenté en cas d'échec de validation, en réinjectant
    l'erreur constatée dans la conversation. Si le second essai échoue aussi,
    lève CommitMessageValidationError plutôt que de renvoyer un message
    bancal en silence.
    """
    if not diff_files:
        raise ValueError("generate_commit_message() nécessite au moins un DiffFile.")

    if config.MOCK_AI:
        return mock_generate_commit_message(diff_files, feedback=feedback)

    system_prompt = _PROMPT_PATH.read_text(encoding="utf-8")
    messages: list[dict] = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": _build_diff_summary(diff_files)},
    ]
    if feedback and feedback.strip():
        messages.append({
            "role": "user",
            "content": f"Consigne additionnelle de révision : {feedback.strip()}",
        })

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
    if config.MOCK_AI:
        return mock_call_chat(messages)

    body = {
        "model": OLLAMA_MODEL,
        "messages": messages,
        "stream": False,
        "format": _JSON_SCHEMA,
        # Sans ça, Ollama utilise sa fenêtre de contexte par défaut (souvent
        # bien plus petite que les 262K annoncés pour gemma4:26b) et tronque
        # silencieusement le début du diff sur les cas volumineux.
        "options": {"num_ctx": OLLAMA_NUM_CTX},
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

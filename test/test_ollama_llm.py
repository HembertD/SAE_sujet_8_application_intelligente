"""Tests du module C (client Ollama) : parsing, validation, retry.

Aucun appel réseau réel : `urllib.request.urlopen` est monkeypatché pour
simuler la réponse d'Ollama, comme demandé par les conventions du projet.
Les tests contre le vrai serveur sont marqués `@pytest.mark.slow` ailleurs.
"""
import json
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from git_commit_release_notes_generator.models import DiffFile
from git_commit_release_notes_generator.ollama_client import llm
from git_commit_release_notes_generator.ollama_client.exceptions import CommitMessageValidationError


class _FakeResponse:
    def __init__(self, payload: dict):
        self._body = json.dumps(payload).encode("utf-8")

    def read(self):
        return self._body

    def __enter__(self):
        return self

    def __exit__(self, *exc):
        return False


def _ollama_payload(content: dict, done_reason: str = "stop") -> dict:
    return {
        "message": {"role": "assistant", "content": json.dumps(content), "thinking": "..."},
        "done": True,
        "done_reason": done_reason,
    }


SAMPLE_DIFF = [DiffFile(path="auth.py", status="M", added=10, removed=2, is_binary=False, patch="+ def login(): ...")]


def test_generate_commit_message_valid_first_try(monkeypatch):
    valid = {"type": "feat", "scope": "auth", "subject": "ajout du bouton de connexion", "body": None}
    monkeypatch.setattr(llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(valid)))

    result = llm.generate_commit_message(SAMPLE_DIFF)

    assert result.type == "feat"
    assert result.scope == "auth"
    assert result.subject == "ajout du bouton de connexion"


def test_generate_commit_message_retries_then_succeeds(monkeypatch):
    invalid = {"type": "banana", "scope": None, "subject": "test.", "body": None}
    valid = {"type": "fix", "scope": None, "subject": "corrige le crash au démarrage", "body": None}
    responses = [_ollama_payload(invalid), _ollama_payload(valid)]

    monkeypatch.setattr(llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = llm.generate_commit_message(SAMPLE_DIFF)

    assert result.type == "fix"
    assert not responses  # les deux appels ont bien été consommés


def test_generate_commit_message_fails_after_second_invalid(monkeypatch):
    invalid = {"type": "banana", "scope": None, "subject": "test.", "body": None}
    monkeypatch.setattr(llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(invalid)))

    with pytest.raises(CommitMessageValidationError):
        llm.generate_commit_message(SAMPLE_DIFF)


def test_validate_rejects_subject_ending_with_dot():
    error = llm._validate({"type": "feat", "subject": "ajout du bouton."})
    assert error is not None
    assert "point" in error


def test_validate_rejects_unknown_type():
    error = llm._validate({"type": "banana", "subject": "ok"})
    assert error is not None
    assert "banana" in error


def test_validate_accepts_well_formed_message():
    error = llm._validate({"type": "docs", "subject": "met à jour le README"})
    assert error is None

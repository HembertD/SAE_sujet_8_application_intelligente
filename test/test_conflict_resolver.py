"""Tests du module de résolution de conflits de merge (module C).

Aucun appel réseau réel : `urllib.request.urlopen` est monkeypatché pour
simuler la réponse d'Ollama, comme pour les autres tests du module.
"""
import json
import sys
from pathlib import Path
from unittest.mock import patch

import pytest

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.models import ConflictSection
from git_commit_release_notes_generator.ollama_client import conflict_resolver as cr
from git_commit_release_notes_generator.ollama_client.exceptions import ConflictResolutionError


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


def _ollama_raw_content_payload(raw_content: str, done_reason: str = "stop") -> dict:
    return {
        "message": {"role": "assistant", "content": raw_content, "thinking": "..."},
        "done": True,
        "done_reason": done_reason,
    }


SAMPLE_CONFLICT = ConflictSection(
    file_path="src/calc.py",
    ours="def add(a, b):\n    return a + b\n",
    theirs="def add(a, b):\n    return a + b  # fix overflow\n",
)


def test_suggest_conflict_resolution_valid_first_try(monkeypatch):
    valid = {
        "resolved_code": "def add(a, b):\n    return a + b  # fix overflow\n",
        "explanation": "Les deux cotes ajoutent la meme fonction, seule theirs ajoute un commentaire utile.",
    }
    monkeypatch.setattr(cr.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(valid)))

    result = cr.suggest_conflict_resolution(SAMPLE_CONFLICT)

    assert "fix overflow" in result.resolved_code
    assert result.explanation


def test_suggest_conflict_resolution_rejects_leftover_markers_then_succeeds(monkeypatch):
    invalid = {"resolved_code": "<<<<<<< HEAD\nfoo\n=======\nbar\n>>>>>>> branch", "explanation": "pas resolu"}
    valid = {"resolved_code": "foo_et_bar()", "explanation": "fusion des deux appels"}
    responses = [_ollama_payload(invalid), _ollama_payload(valid)]

    monkeypatch.setattr(cr.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = cr.suggest_conflict_resolution(SAMPLE_CONFLICT)

    assert result.resolved_code == "foo_et_bar()"
    assert not responses


def test_suggest_conflict_resolution_fails_after_second_invalid(monkeypatch):
    invalid = {"resolved_code": "<<<<<<< toujours pas resolu", "explanation": "x"}
    monkeypatch.setattr(cr.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(invalid)))

    with pytest.raises(ConflictResolutionError):
        cr.suggest_conflict_resolution(SAMPLE_CONFLICT)


def test_suggest_conflict_resolution_retries_on_malformed_json_then_succeeds(monkeypatch):
    valid = {"resolved_code": "resultat_final()", "explanation": "ok"}
    responses = [
        _ollama_raw_content_payload("ceci n'est pas du JSON"),
        _ollama_payload(valid),
    ]
    monkeypatch.setattr(cr.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = cr.suggest_conflict_resolution(SAMPLE_CONFLICT)

    assert result.resolved_code == "resultat_final()"
    assert not responses


def test_suggest_conflict_resolution_rejects_empty_sides():
    empty_conflict = ConflictSection(file_path="x.py", ours="   ", theirs="")
    with pytest.raises(ValueError):
        cr.suggest_conflict_resolution(empty_conflict)


def test_validate_rejects_leftover_conflict_markers():
    error = cr._validate({"resolved_code": "=======", "explanation": "x"})
    assert error is not None
    assert "marqueur" in error


def test_validate_accepts_well_formed_resolution():
    error = cr._validate({"resolved_code": "def f(): pass", "explanation": "explication"})
    assert error is None


def test_mock_ai_delegates_without_network():
    with patch.object(config, "MOCK_AI", True):
        result = cr.suggest_conflict_resolution(SAMPLE_CONFLICT)
        assert result.resolved_code == SAMPLE_CONFLICT.ours
        assert "mockée" in result.explanation

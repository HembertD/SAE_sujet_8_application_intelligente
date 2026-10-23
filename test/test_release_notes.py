"""Tests du module de génération de release notes (module C).

Aucun appel réseau réel : `urllib.request.urlopen` est monkeypatché pour
simuler la réponse d'Ollama, comme pour les autres tests du module.
"""
import json
import sys
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch

import pytest

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.ollama_client import release_notes as rn
from git_commit_release_notes_generator.ollama_client.exceptions import ReleaseNotesError


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


SAMPLE_COMMITS = [
    SimpleNamespace(sha="1a2b3c4", message="feat(auth): ajout du login OAuth2"),
    SimpleNamespace(sha="5d6e7f8", message="fix: correction du bug de division par zéro"),
]


def test_generate_release_notes_valid_first_try(monkeypatch):
    valid = {"markdown": "# Release Notes (v0.1.0 ➔ v1.0.0)\n\n### 🚀 Nouveautés (Features)\n- login OAuth2\n"}
    monkeypatch.setattr(rn.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(valid)))

    result = rn.generate_release_notes(SAMPLE_COMMITS, "v0.1.0", "v1.0.0")

    assert "Nouveautés" in result
    assert "login OAuth2" in result


def test_generate_release_notes_retries_on_empty_then_succeeds(monkeypatch):
    invalid = {"markdown": "   "}
    valid = {"markdown": "# Release Notes (v0.1.0 ➔ v1.0.0)\n\nOk."}
    responses = [_ollama_payload(invalid), _ollama_payload(valid)]

    monkeypatch.setattr(rn.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = rn.generate_release_notes(SAMPLE_COMMITS, "v0.1.0", "v1.0.0")

    assert result == "# Release Notes (v0.1.0 ➔ v1.0.0)\n\nOk."
    assert not responses


def test_generate_release_notes_fails_after_second_invalid(monkeypatch):
    invalid = {"markdown": ""}
    monkeypatch.setattr(rn.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(invalid)))

    with pytest.raises(ReleaseNotesError):
        rn.generate_release_notes(SAMPLE_COMMITS, "v0.1.0", "v1.0.0")


def test_generate_release_notes_retries_on_malformed_json_then_succeeds(monkeypatch):
    valid = {"markdown": "# Release Notes\n\nOk."}
    responses = [
        _ollama_raw_content_payload("ceci n'est pas du JSON"),
        _ollama_payload(valid),
    ]
    monkeypatch.setattr(rn.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = rn.generate_release_notes(SAMPLE_COMMITS, "v0.1.0", "v1.0.0")

    assert result == "# Release Notes\n\nOk."
    assert not responses


def test_generate_release_notes_rejects_empty_commits():
    with pytest.raises(ValueError):
        rn.generate_release_notes([], "v0.1.0", "v1.0.0")


def test_validate_rejects_missing_or_empty_markdown():
    assert rn._validate({"markdown": ""}) is not None
    assert rn._validate({}) is not None


def test_validate_accepts_well_formed_markdown():
    assert rn._validate({"markdown": "# Release Notes\n\nOk."}) is None


def test_mock_ai_delegates_without_network():
    with patch.object(config, "MOCK_AI", True):
        result = rn.generate_release_notes(SAMPLE_COMMITS, "v0.1.0", "v1.0.0")
        assert "mockées" in result

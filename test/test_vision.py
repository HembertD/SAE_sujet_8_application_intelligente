"""Tests du module de description d'image via un modèle de vision (module C).

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
from git_commit_release_notes_generator.ollama_client import vision
from git_commit_release_notes_generator.ollama_client.exceptions import ImageDescriptionError


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


FAKE_IMAGE_BYTES = b"\x89PNG\r\n\x1a\n" + b"\x00" * 16


def test_describe_image_change_addition_sends_single_image(monkeypatch):
    sent = {}

    def _fake_urlopen(request, timeout=None):
        body = json.loads(request.data.decode("utf-8"))
        sent["messages"] = body["messages"]
        return _FakeResponse(_ollama_payload({"description": "Un éléphant dans la savane."}))

    monkeypatch.setattr(vision.llm.urllib.request, "urlopen", _fake_urlopen)

    result = vision.describe_image_change("photo.png", "A", FAKE_IMAGE_BYTES, old_image=None)

    assert result == "Un éléphant dans la savane."
    user_message = sent["messages"][1]
    assert len(user_message["images"]) == 1
    assert "ajoutée" in user_message["content"]


def test_describe_image_change_deletion_sends_old_image_only(monkeypatch):
    sent = {}

    def _fake_urlopen(request, timeout=None):
        body = json.loads(request.data.decode("utf-8"))
        sent["messages"] = body["messages"]
        return _FakeResponse(_ollama_payload({"description": "Logo bleu supprimé."}))

    monkeypatch.setattr(vision.llm.urllib.request, "urlopen", _fake_urlopen)

    result = vision.describe_image_change("logo.png", "D", None, old_image=FAKE_IMAGE_BYTES)

    assert result == "Logo bleu supprimé."
    user_message = sent["messages"][1]
    assert len(user_message["images"]) == 1
    assert "supprimée" in user_message["content"]


def test_describe_image_change_modification_sends_both_images(monkeypatch):
    sent = {}

    def _fake_urlopen(request, timeout=None):
        body = json.loads(request.data.decode("utf-8"))
        sent["messages"] = body["messages"]
        return _FakeResponse(_ollama_payload({"description": "Le logo passe du bleu au vert."}))

    monkeypatch.setattr(vision.llm.urllib.request, "urlopen", _fake_urlopen)

    result = vision.describe_image_change("logo.png", "M", FAKE_IMAGE_BYTES, old_image=FAKE_IMAGE_BYTES)

    assert result == "Le logo passe du bleu au vert."
    user_message = sent["messages"][1]
    assert len(user_message["images"]) == 2
    assert "avant" in user_message["content"] and "après" in user_message["content"]


def test_describe_image_change_retries_on_empty_then_succeeds(monkeypatch):
    invalid = {"description": "   "}
    valid = {"description": "Une photo de montagne."}
    responses = [_ollama_payload(invalid), _ollama_payload(valid)]

    monkeypatch.setattr(vision.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = vision.describe_image_change("photo.png", "A", FAKE_IMAGE_BYTES)

    assert result == "Une photo de montagne."
    assert not responses


def test_describe_image_change_fails_after_second_invalid(monkeypatch):
    invalid = {"description": ""}
    monkeypatch.setattr(vision.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(_ollama_payload(invalid)))

    with pytest.raises(ImageDescriptionError):
        vision.describe_image_change("photo.png", "A", FAKE_IMAGE_BYTES)


def test_describe_image_change_retries_on_malformed_json_then_succeeds(monkeypatch):
    valid = {"description": "Ok."}
    responses = [
        _ollama_raw_content_payload("ceci n'est pas du JSON"),
        _ollama_payload(valid),
    ]
    monkeypatch.setattr(vision.llm.urllib.request, "urlopen", lambda *a, **k: _FakeResponse(responses.pop(0)))

    result = vision.describe_image_change("photo.png", "A", FAKE_IMAGE_BYTES)

    assert result == "Ok."
    assert not responses


def test_describe_image_change_rejects_no_images():
    with pytest.raises(ValueError):
        vision.describe_image_change("photo.png", "A", None, old_image=None)


def test_validate_rejects_missing_or_empty_description():
    assert vision._validate({"description": ""}) is not None
    assert vision._validate({}) is not None


def test_validate_accepts_well_formed_description():
    assert vision._validate({"description": "Un contenu valide."}) is None


def test_mock_ai_delegates_without_network():
    with patch.object(config, "MOCK_AI", True):
        result = vision.describe_image_change("photo.png", "A", FAKE_IMAGE_BYTES)
        assert "mock" in result.lower()

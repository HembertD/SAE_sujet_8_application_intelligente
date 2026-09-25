"""Tests unitaires pour le point d'entrée service/core.py.

Vérifie que les 9 actions requises par le contrat Go / Python fonctionnent
correctement et retournent du JSON strict.
"""
from __future__ import annotations

import json
from pathlib import Path
from unittest.mock import patch

import pytest

ROOT = Path(__file__).resolve().parent.parent

from git import Repo
from git_commit_release_notes_generator.models import CommitMessage, DiffFile
from git_commit_release_notes_generator.service.git_wrapper import CommitInfo
import git_commit_release_notes_generator.service.core as core
from git_commit_release_notes_generator.service.core import (
    action_apply_commit,
    action_diff,
    action_generate_commit,
    action_ping_ollama,
    action_push,
    action_release_notes,
    action_save_config,
    action_stage_all,
    action_status,
)


class _FakeHTTPResponse:
    def __init__(self, data: dict | bytes):
        if isinstance(data, dict):
            self._body = json.dumps(data).encode("utf-8")
        else:
            self._body = data

    def read(self):
        return self._body

    def __enter__(self):
        return self

    def __exit__(self, *exc):
        return False


def test_action_status_on_valid_repo():
    result = core.action_status(str(ROOT))
    assert result["is_git_repo"] is True
    assert "branch" in result
    assert "repo_path" in result
    assert isinstance(result["staged_count"], int)
    assert isinstance(result["unstaged_count"], int)


def test_action_status_on_non_git_repo(tmp_path):
    result = core.action_status(str(tmp_path))
    assert result["is_git_repo"] is False
    assert result["staged_count"] == 0
    assert "error" in result


def test_action_diff_structure():
    result = core.action_diff(str(ROOT))
    assert result["success"] is True
    assert "files" in result
    assert "staged_count" in result
    for f in result["files"]:
        assert "path" in f
        assert "status" in f
        assert "binary" in f
        assert "has_secrets_masked" in f
        assert "patch" in f


def test_action_generate_commit_no_staged_files():
    # Sur le dépôt principal où rien n'est staged (ou via mock)
    with patch("git_commit_release_notes_generator.service.core.GitWrapper.get_staged_diff", return_value=[]):
        result = core.action_generate_commit(str(ROOT))
        assert result["success"] is False
        assert "error" in result


def test_action_generate_commit_success():
    fake_diff = [
        DiffFile(path="calc.py", status="M", added=5, removed=1, is_binary=False, patch="+def add(a, b): return a + b")
    ]
    fake_commit = CommitMessage(
        type="feat",
        scope="calc",
        subject="ajout de la fonction addition",
        body="Permet d'effectuer des additions simples.",
    )

    with patch("git_commit_release_notes_generator.service.core.GitWrapper.get_staged_diff", return_value=fake_diff):
        with patch("git_commit_release_notes_generator.service.core.generate_commit_message", return_value=fake_commit):
            result = core.action_generate_commit(str(ROOT))

            assert result["success"] is True
            assert result["type"] == "feat"
            assert result["scope"] == "calc"
            assert result["subject"] == "ajout de la fonction addition"
            assert "raw_formatted" in result
            assert "feat(calc): ajout de la fonction addition" in result["raw_formatted"]


def test_action_apply_commit_validation():
    # Message vide
    result = core.action_apply_commit(str(ROOT), message="")
    assert result["success"] is False
    assert "vide" in result["error"]


def test_action_apply_commit_success():
    with patch("git_commit_release_notes_generator.service.core.GitWrapper.commit", return_value="abc1234"):
        result = core.action_apply_commit(str(ROOT), message="feat: test commit")
        assert result["success"] is True
        assert result["sha"] == "abc1234"


def test_action_push_no_remote(tmp_path):
    # Dépôt temporaire sans remote
    r = Repo.init(tmp_path)
    result = core.action_push(str(tmp_path))
    assert result["success"] is False
    assert "remote" in result["error"].lower()


def test_action_release_notes_formatting():
    mock_commits = [
        CommitInfo(sha="1a2b3c4", message="feat(auth): ajout du login OAuth2", author="Maxence", date="2026-09-22"),
        CommitInfo(sha="5d6e7f8", message="fix: correction du bug de division par zéro", author="Dorian", date="2026-09-22"),
        CommitInfo(sha="9g0h1i2", message="docs: mise à jour du guide d'architecture", author="Enzo", date="2026-09-22"),
    ]

    with patch("git_commit_release_notes_generator.service.core.GitWrapper.get_commits_between_tags", return_value=mock_commits):
        # Force le fallback déterministe pour tester le formateur
        with patch("urllib.request.urlopen", side_effect=Exception("Ollama offline")):
            result = core.action_release_notes(str(ROOT), from_tag="v0.1.0", to_tag="v1.0.0")

            assert result["success"] is True
            assert result["from_tag"] == "v0.1.0"
            assert result["to_tag"] == "v1.0.0"
            assert result["commit_count"] == 3
            assert "### 🚀 Nouveautés (Features)" in result["markdown"]
            assert "feat(auth)" in result["markdown"]
            assert "### 🐛 Corrections de bogues (Fixes)" in result["markdown"]
            assert "### 📚 Documentation" in result["markdown"]


def test_action_release_notes_empty():
    with patch("git_commit_release_notes_generator.service.core.GitWrapper.get_commits_between_tags", return_value=[]):
        result = core.action_release_notes(str(ROOT), from_tag="v1.0.0", to_tag="v1.0.0")
        assert result["success"] is True
        assert result["commit_count"] == 0
        assert "Aucun commit trouvé" in result["markdown"]


def test_action_save_config():
    result = core.action_save_config("http://localhost:11434", "gemma4:12b", "60", "fr")
    assert result["success"] is True
    assert "message" in result


def test_action_ping_ollama_reachable():
    fake_tags = {
        "models": [
            {"name": "gemma4:12b"},
            {"name": "mistral:latest"},
        ]
    }

    responses = [
        _FakeHTTPResponse({"version": "0.1.30"}),
        _FakeHTTPResponse(fake_tags),
    ]

    with patch("urllib.request.urlopen", side_effect=lambda *a, **k: responses.pop(0)):
        result = core.action_ping_ollama()
        assert result["success"] is True
        assert result["reachable"] is True
        assert "gemma4:12b" in result["installed_models"]
        assert "mistral:latest" in result["installed_models"]


def test_action_ping_ollama_unreachable():
    with patch("urllib.request.urlopen", side_effect=Exception("Connection refused")):
        result = core.action_ping_ollama()
        assert result["success"] is True
        assert result["reachable"] is False
        assert result["installed_models"] == []
        assert "error" in result

"""Tests unitaires pour le Mock de l'IA (simulation hors-ligne).

Vérifie que la simulation de l'IA produit un message mocké conforme aux spécifications
Conventional Commits, supporte les feedbacks, simule le ping et les release notes,
et s'active via la variable MOCK_AI.
"""
from __future__ import annotations

import sys
from pathlib import Path
from unittest.mock import patch

import pytest

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT / "src"))

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.config import parse_bool
from git_commit_release_notes_generator.models import DiffFile
from git_commit_release_notes_generator.ollama_client import llm
from git_commit_release_notes_generator.ollama_client.mock_llm import (
    mock_call_chat,
    mock_generate_commit_message,
    mock_ping,
    mock_release_notes,
)
from git_commit_release_notes_generator.service import core
from git_commit_release_notes_generator.service.git_wrapper import CommitInfo


def test_parse_bool():
    """Valide les différentes représentations de booléens dans le .env."""
    assert parse_bool("true") is True
    assert parse_bool("TRUE") is True
    assert parse_bool("True") is True
    assert parse_bool("1") is True
    assert parse_bool("yes") is True
    assert parse_bool("oui") is True
    assert parse_bool(True) is True

    assert parse_bool("false") is False
    assert parse_bool("FALSE") is False
    assert parse_bool("False") is False
    assert parse_bool("0") is False
    assert parse_bool("no") is False
    assert parse_bool("") is False
    assert parse_bool(None) is False
    assert parse_bool(False) is False


def test_mock_generate_commit_empty_diff_raises():
    """Doit lever une exception si aucun diff n'est fourni."""
    with pytest.raises(ValueError):
        mock_generate_commit_message([])


def test_mock_generate_commit_static_message():
    """Vérifie le message statique retourné par le mock."""
    diff = [
        DiffFile(path="test/test_service.py", status="M", added=12, removed=2, patch="+ def test_new(): pass"),
    ]
    msg = mock_generate_commit_message(diff)
    assert msg.type == "feat"
    assert msg.scope == "mock"
    assert msg.subject == "commit mocker pour test hors ligne"
    assert "hors-ligne" in msg.body
    assert llm._validate({"type": msg.type, "subject": msg.subject, "scope": msg.scope, "body": msg.body}) is None


def test_mock_generate_commit_with_user_feedback():
    """Vérifie la prise en compte de la consigne utilisateur (feedback) dans le body."""
    diff = [
        DiffFile(path="src/auth/service.py", status="M", added=20, removed=5, patch="+ def authenticate(): ..."),
    ]
    feedback = "Attention aux tokens expirés"
    msg = mock_generate_commit_message(diff, feedback=feedback)

    assert msg.type == "feat"
    assert msg.scope == "mock"
    assert feedback in (msg.body or "")
    assert llm._validate({"type": msg.type, "subject": msg.subject, "scope": msg.scope, "body": msg.body}) is None


def test_mock_call_chat():
    """Vérifie que la réponse brute simulée est conforme au schéma JSON."""
    raw = mock_call_chat([{"role": "user", "content": "diff summary"}])
    assert raw["type"] == "feat"
    assert raw["scope"] == "mock"
    assert raw["subject"] == "commit mocker pour test hors ligne"
    assert llm._validate(raw) is None


def test_mock_release_notes():
    """Vérifie la génération statique des release notes simulées."""
    md = mock_release_notes([], from_tag="v1.0.0", to_tag="v1.1.0")
    assert "# Release Notes (v1.0.0 ➔ v1.1.0)" in md
    assert "Release notes mockées pour test hors ligne." in md


def test_mock_ping():
    """Vérifie que mock_ping renvoie un statut joignable avec modèles simulés."""
    res = mock_ping()
    assert res["success"] is True
    assert res["reachable"] is True
    assert res["latency_ms"] > 0
    assert len(res["installed_models"]) > 0


def test_llm_check_ollama_reachable_when_mock_ai():
    """check_ollama_reachable doit renvoyer True sans requête HTTP si MOCK_AI=True."""
    with patch.object(config, "MOCK_AI", True):
        reachable, err = llm.check_ollama_reachable()
        assert reachable is True
        assert err is None


def test_llm_generate_commit_delegates_to_mock_when_mock_ai():
    """generate_commit_message doit appeler le mock sans toucher au réseau si MOCK_AI=True."""
    diff = [DiffFile(path="src/cli/main.go", status="M", added=10, removed=2, patch="+ func newFeature() {}")]

    with patch.object(config, "MOCK_AI", True):
        msg = llm.generate_commit_message(diff, feedback="Ajoute le nouveau flag")
        assert msg.type == "feat"
        assert msg.scope == "mock"
        assert msg.subject == "commit mocker pour test hors ligne"
        assert llm._validate({"type": msg.type, "subject": msg.subject, "scope": msg.scope, "body": msg.body}) is None


def test_core_action_generate_commit_with_mock_ai():
    """action_generate_commit doit réussir hors-ligne grâce à MOCK_AI."""
    diff = [DiffFile(path="src/service.py", status="M", added=5, removed=1, patch="+ def run(): pass")]

    with patch.object(config, "MOCK_AI", True):
        with patch.object(core.GitWrapper, "get_staged_diff", return_value=diff):
            res = core.action_generate_commit(repo_path=".", feedback="Consigne de test")
            assert res["success"] is True
            assert res["type"] == "feat"
            assert res["scope"] == "mock"
            assert res["subject"] == "commit mocker pour test hors ligne"
            assert res["raw_formatted"] == "feat(mock): commit mocker pour test hors ligne\n\nCommit généré en mode mock pour permettre le travail hors-ligne.\nConsigne utilisateur : Consigne de test"


def test_core_action_release_notes_with_mock_ai():
    """action_release_notes doit renvoyer les notes de version simulées sans requête réseau."""
    commits = [
        CommitInfo(sha="abcdef1", message="feat: nouvelle fonctionnalité", author="Dev", date="2026-09-28"),
    ]

    with patch.object(config, "MOCK_AI", True):
        with patch.object(core.GitWrapper, "get_commits_between_tags", return_value=commits):
            res = core.action_release_notes(repo_path=".", from_tag="v1.0.0", to_tag="v1.1.0")
            assert res["success"] is True
            assert "Release Notes" in res["markdown"]
            assert res["commit_count"] == 1


def test_core_action_ping_ollama_with_mock_ai():
    """action_ping_ollama doit renvoyer un ping réussi avec MOCK_AI sans requête réseau."""
    with patch.object(config, "MOCK_AI", True):
        res = core.action_ping_ollama()
        assert res["success"] is True
        assert res["reachable"] is True
        assert "gemma4:26b" in res["installed_models"]

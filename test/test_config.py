"""Tests de la configuration Python, notamment la bascule Docker / IUT.

Le module `config` lit le fichier `.env` au moment de l'import. Ces tests
rechargent le module avec un `.env` temporaire pour vérifier que
`USE_DOCKER_LLM` ne casse pas la configuration de référence du serveur IUT.
"""
import importlib
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from git_commit_release_notes_generator import config


def _reload_with_env(monkeypatch, tmp_path: Path, env_content: str):
    """Recharge `config` avec un `.env` isolé et sans variables système."""
    (tmp_path / ".env").write_text(env_content, encoding="utf-8")
    monkeypatch.setenv("GITELLIGENCE_APP_ROOT", str(tmp_path))
    for key in ("OLLAMA_BASE_URL", "OLLAMA_MODEL", "USE_DOCKER_LLM"):
        monkeypatch.delenv(key, raising=False)
    return importlib.reload(config)


def _restore_config(monkeypatch):
    monkeypatch.delenv("GITELLIGENCE_APP_ROOT", raising=False)
    importlib.reload(config)


def test_use_docker_llm_forces_localhost(monkeypatch, tmp_path):
    try:
        cfg = _reload_with_env(
            monkeypatch,
            tmp_path,
            "OLLAMA_BASE_URL=http://10.22.28.190:11434\n"
            "OLLAMA_MODEL=gemma4:26b\n"
            "USE_DOCKER_LLM=true\n",
        )
        assert cfg.OLLAMA_BASE_URL == "http://localhost:11434"
        assert cfg.OLLAMA_MODEL == "gemma4:26b"
        assert cfg.USE_DOCKER_LLM is True
    finally:
        _restore_config(monkeypatch)


def test_iut_url_is_preserved_when_docker_is_disabled(monkeypatch, tmp_path):
    try:
        cfg = _reload_with_env(
            monkeypatch,
            tmp_path,
            "OLLAMA_BASE_URL=http://10.22.28.190:11434\n"
            "OLLAMA_MODEL=gemma4:26b\n"
            "USE_DOCKER_LLM=false\n",
        )
        assert cfg.OLLAMA_BASE_URL == "http://10.22.28.190:11434"
        assert cfg.OLLAMA_MODEL == "gemma4:26b"
        assert cfg.USE_DOCKER_LLM is False
    finally:
        _restore_config(monkeypatch)


def test_default_model_remains_iut_reference(monkeypatch, tmp_path):
    try:
        cfg = _reload_with_env(monkeypatch, tmp_path, "USE_DOCKER_LLM=false\n")
        assert cfg.OLLAMA_MODEL == "gemma4:26b"
        assert cfg.OLLAMA_BASE_URL == "http://10.22.28.190:11434"
    finally:
        _restore_config(monkeypatch)

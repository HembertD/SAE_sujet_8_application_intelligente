"""Configuration pytest globale pour l'isolation des tests."""
import pytest

from git_commit_release_notes_generator import config
from git_commit_release_notes_generator.ollama_client import llm
from git_commit_release_notes_generator.service import core


@pytest.fixture(autouse=True)
def _isolate_test_environment(monkeypatch):
    """Garantit que les tests unitaires existants s'exécutent avec un environnement
    neutre (MOCK_AI=False), quel que soit le contenu du fichier .env local."""
    monkeypatch.setattr(config, "MOCK_AI", False)

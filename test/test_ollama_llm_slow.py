"""Test contre le vrai serveur Ollama partagé (réseau IUT requis).

Non exécuté par défaut : `python -m pytest -m slow` pour le lancer explicitement.
"""
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from git_commit_release_notes_generator.models import DiffFile
from git_commit_release_notes_generator.ollama_client.llm import generate_commit_message

SAMPLE_DIFF = [
    DiffFile(
        path="src/auth/login.py",
        status="M",
        added=18,
        removed=3,
        is_binary=False,
        patch=(
            "+ def login(username: str, password: str) -> bool:\n"
            "+     user = find_user(username)\n"
            "+     if user is None:\n"
            "+         return False\n"
            "+     return check_password(user, password)\n"
        ),
    )
]


@pytest.mark.slow
def test_generate_commit_message_against_real_ollama():
    result = generate_commit_message(SAMPLE_DIFF)

    print(f"\ntype={result.type} scope={result.scope}")
    print(f"subject={result.subject}")
    print(f"body={result.body}")

    assert result.type
    assert result.subject

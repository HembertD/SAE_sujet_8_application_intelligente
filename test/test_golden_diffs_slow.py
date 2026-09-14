"""Vérifie chaque golden file contre le vrai serveur Ollama (réseau requis).

Non exécuté par défaut : `python -m pytest -m slow` pour le lancer.
Sert de base au futur script de métriques (taux de conformité, latence).
"""
import sys
import time
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))
sys.path.insert(0, str(Path(__file__).parent.parent))

from git_commit_release_notes_generator.ollama_client.llm import generate_commit_message
from test.fixtures.golden_diffs import GOLDEN_DIFFS


@pytest.mark.slow
@pytest.mark.parametrize("golden", GOLDEN_DIFFS, ids=[g.name for g in GOLDEN_DIFFS])
def test_golden_diff_produces_valid_commit_message(golden):
    start = time.perf_counter()
    result = generate_commit_message(golden.files)
    elapsed = time.perf_counter() - start

    print(
        f"\n[{golden.name}] attendu={golden.expected_type} "
        f"obtenu={result.type}({result.scope}) en {elapsed:.1f}s"
    )
    print(f"  subject: {result.subject}")
    if result.body:
        print(f"  body: {result.body}")

    assert result.type
    assert result.subject

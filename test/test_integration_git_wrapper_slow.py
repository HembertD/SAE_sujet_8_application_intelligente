"""Test d'intégration réel : GitWrapper (module Git) -> generate_commit_message (module C).

But : vérifier que le vrai `list[DiffFile]` produit par
`GitWrapper.get_staged_diff()` est consommé tel quel par le module C, sans
conversion manuelle.

Nécessite :
- `pip install GitPython`
- les fixtures activées : `python test/situation_test/init_test.py`
- réseau vers le serveur Ollama partagé (test marqué "slow")
"""
import subprocess
import sys
import time
from pathlib import Path

import pytest

ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(ROOT / "src"))

from git_commit_release_notes_generator.service.git_wrapper import GitWrapper
from git_commit_release_notes_generator.ollama_client.llm import generate_commit_message

FIXTURES_DIR = ROOT / "test" / "situation_test" / "commit"

# Sous-dossier -> type Conventional Commits attendu (voir README_situation_test.md)
SCENARIOS = {
    "commit_feat": "feat",
    "commit_fix": "fix",
    "commit_docs": "docs",
    "commit_refactor": "refactor",
    "commit_test": "test",
    "commit_chore": "chore",
    "commit_perf": "perf",
    "commit_build": "build",
}


def _stage_all_changes(repo_path: Path) -> None:
    """Indexe toutes les modifications locales (équivalent `git add -A`)."""
    subprocess.run(["git", "add", "-A"], cwd=repo_path, check=True)


@pytest.mark.slow
@pytest.mark.parametrize("scenario", SCENARIOS.keys())
def test_real_diff_files_are_usable_by_llm_module(scenario):
    repo_path = FIXTURES_DIR / scenario
    _stage_all_changes(repo_path)

    wrapper = GitWrapper(repo_path)
    diff_files = wrapper.get_staged_diff()

    assert diff_files, f"Aucun DiffFile produit pour {scenario} - vérifier que les fixtures sont activées"

    start = time.perf_counter()
    result = generate_commit_message(diff_files)
    elapsed = time.perf_counter() - start

    expected = SCENARIOS[scenario]
    print(f"\n[{scenario}] attendu={expected} obtenu={result.type}({result.scope}) en {elapsed:.1f}s")
    print(f"  subject: {result.subject}")

    assert result.type
    assert result.subject

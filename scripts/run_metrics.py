"""Mesure le taux de conformité et la latence du module C sur les golden files.

Outil de dev, pas le module lui-même : celui-ci peut écrire sur stdout.
Nécessite un accès réseau au serveur Ollama partagé.

Usage : python scripts/run_metrics.py
"""
from __future__ import annotations

import logging
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))
sys.path.insert(0, str(Path(__file__).parent.parent))

from git_commit_release_notes_generator.ollama_client.exceptions import CommitMessageValidationError
from git_commit_release_notes_generator.ollama_client.llm import generate_commit_message
from test.fixtures.golden_diffs import GOLDEN_DIFFS


class _RetryCounter(logging.Handler):
    """Compte les warnings de retry émis par llm.py, sans toucher à son API publique."""

    def __init__(self) -> None:
        super().__init__()
        self.count = 0

    def emit(self, record: logging.LogRecord) -> None:
        if "invalide" in record.getMessage():
            self.count += 1


def main() -> None:
    retry_counter = _RetryCounter()
    llm_logger = logging.getLogger("git_commit_release_notes_generator.ollama_client.llm")
    llm_logger.addHandler(retry_counter)
    llm_logger.setLevel(logging.WARNING)

    results: list[dict] = []
    for golden in GOLDEN_DIFFS:
        retries_before = retry_counter.count
        start = time.perf_counter()
        try:
            message = generate_commit_message(golden.files)
            elapsed = time.perf_counter() - start
            results.append({
                "name": golden.name,
                "ok": True,
                "elapsed": elapsed,
                "needed_retry": retry_counter.count > retries_before,
                "type": message.type,
                "expected": golden.expected_type,
            })
        except CommitMessageValidationError:
            elapsed = time.perf_counter() - start
            results.append({"name": golden.name, "ok": False, "elapsed": elapsed})

    _print_report(results)


def _print_report(results: list[dict]) -> None:
    n = len(results)
    successes = [r for r in results if r["ok"]]
    type_matches = sum(1 for r in successes if r["type"] == r["expected"])
    retries = sum(1 for r in successes if r["needed_retry"])
    latencies = [r["elapsed"] for r in results]

    print(f"\n{'Scénario':<32} {'Résultat':<9} {'Attendu':<10} {'Obtenu':<10} {'Retry':<6} {'Latence'}")
    print("-" * 85)
    for r in results:
        if r["ok"]:
            retry_flag = "oui" if r["needed_retry"] else "non"
            print(f"{r['name']:<32} {'OK':<9} {r['expected']:<10} {r['type']:<10} {retry_flag:<6} {r['elapsed']:.1f}s")
        else:
            print(f"{r['name']:<32} {'ECHEC':<9} {'-':<10} {'-':<10} {'-':<6} {r['elapsed']:.1f}s")

    print("-" * 85)
    print(f"Taux de conformité (JSON valide + règles respectées) : {len(successes)}/{n} ({len(successes) / n:.0%})")
    print(f"Type exact attendu vs obtenu                          : {type_matches}/{len(successes)} ({type_matches / len(successes):.0%})")
    print(f"Générations ayant nécessité un retry                  : {retries}/{len(successes)} ({retries / len(successes):.0%})")
    print(f"Latence moyenne / min / max                           : {sum(latencies) / n:.1f}s / {min(latencies):.1f}s / {max(latencies):.1f}s")


if __name__ == "__main__":
    main()

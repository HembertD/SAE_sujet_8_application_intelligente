"""Diffs réalistes figés, utilisés comme fixtures tant que le module A/B
(Dorian) ne produit pas encore de vrais DiffFile.

Un scénario = un type de changement de développement courant (feature, fix,
refactor, docs, test, chore) associé au type Conventional Commits qu'on
s'attend à voir ressortir du LLM. Le type attendu sert de repère pour les
métriques de conformité, pas d'assertion stricte : le modèle reste libre du
scope/subject exacts.
"""
from __future__ import annotations

from dataclasses import dataclass

from git_commit_release_notes_generator.models import DiffFile


@dataclass
class GoldenDiff:
    name: str
    expected_type: str
    files: list[DiffFile]


GOLDEN_DIFFS: list[GoldenDiff] = [
    GoldenDiff(
        name="feature_login_lockout",
        expected_type="feat",
        files=[
            DiffFile(
                path="src/auth/login.py",
                status="M",
                added=6,
                removed=0,
                is_binary=False,
                patch=(
                    "@@ -10,6 +10,12 @@ def login(username: str, password: str) -> bool:\n"
                    "     user = find_user(username)\n"
                    "     if user is None:\n"
                    "         return False\n"
                    "+    if is_locked(user):\n"
                    "+        return False\n"
                    "+    if failed_attempts(user) >= MAX_ATTEMPTS:\n"
                    "+        lock_account(user)\n"
                    "+        return False\n"
                    "     return check_password(user, password)\n"
                ),
            ),
            DiffFile(
                path="tests/test_login.py",
                status="M",
                added=4,
                removed=0,
                is_binary=False,
                patch=(
                    "@@ -1,3 +1,7 @@\n"
                    " def test_login_ok():\n"
                    "     ...\n"
                    "+\n"
                    "+def test_login_locked_after_max_attempts():\n"
                    "+    ...\n"
                ),
            ),
        ],
    ),
    GoldenDiff(
        name="fix_none_crash_on_export",
        expected_type="fix",
        files=[
            DiffFile(
                path="src/reports/export.py",
                status="M",
                added=3,
                removed=1,
                is_binary=False,
                patch=(
                    "@@ -42,7 +42,9 @@ def export_report(report):\n"
                    "-    return report.to_csv()\n"
                    "+    if report is None:\n"
                    "+        raise ValueError(\"report ne peut pas être None\")\n"
                    "+    return report.to_csv()\n"
                ),
            ),
        ],
    ),
    GoldenDiff(
        name="refactor_extract_price_helper",
        expected_type="refactor",
        files=[
            DiffFile(
                path="src/shop/cart.py",
                status="M",
                added=2,
                removed=6,
                is_binary=False,
                patch=(
                    "@@ -20,12 +20,8 @@ def total_price(items):\n"
                    "-    total = 0\n"
                    "-    for item in items:\n"
                    "-        total += item.price * item.quantity\n"
                    "-        if item.discount:\n"
                    "-            total -= item.price * item.discount\n"
                    "-    return total\n"
                    "+    return compute_total(items)\n"
                ),
            ),
            DiffFile(
                path="src/shop/pricing.py",
                status="A",
                added=8,
                removed=0,
                is_binary=False,
                patch=(
                    "@@ -0,0 +1,8 @@\n"
                    "+def compute_total(items):\n"
                    "+    total = 0\n"
                    "+    for item in items:\n"
                    "+        total += item.price * item.quantity\n"
                    "+        if item.discount:\n"
                    "+            total -= item.price * item.discount\n"
                    "+    return total\n"
                ),
            ),
        ],
    ),
    GoldenDiff(
        name="docs_update_readme_setup",
        expected_type="docs",
        files=[
            DiffFile(
                path="README.md",
                status="M",
                added=5,
                removed=1,
                is_binary=False,
                patch=(
                    "@@ -12,7 +12,11 @@\n"
                    "-Lancer l'application avec `python main.py`.\n"
                    "+## Installation\n"
                    "+\n"
                    "+1. `pip install -r requirements.txt`\n"
                    "+2. Copier `.env.example` en `.env`\n"
                    "+3. Lancer l'application avec `python main.py`\n"
                ),
            ),
        ],
    ),
    GoldenDiff(
        name="test_add_coverage_pricing",
        expected_type="test",
        files=[
            DiffFile(
                path="tests/test_pricing.py",
                status="A",
                added=10,
                removed=0,
                is_binary=False,
                patch=(
                    "@@ -0,0 +1,10 @@\n"
                    "+def test_compute_total_without_discount():\n"
                    "+    ...\n"
                    "+\n"
                    "+def test_compute_total_with_discount():\n"
                    "+    ...\n"
                    "+\n"
                    "+def test_compute_total_empty_cart():\n"
                    "+    ...\n"
                ),
            ),
        ],
    ),
    GoldenDiff(
        name="build_bump_httpx_dependency",
        expected_type="build",
        files=[
            DiffFile(
                path="requirements.txt",
                status="M",
                added=1,
                removed=1,
                is_binary=False,
                patch=(
                    "@@ -3,4 +3,4 @@\n"
                    "-httpx==0.24.1\n"
                    "+httpx==0.27.0\n"
                ),
            ),
        ],
    ),
]

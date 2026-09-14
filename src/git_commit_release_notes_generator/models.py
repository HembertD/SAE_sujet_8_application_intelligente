"""Structures de données partagées entre les modules (contrat d'équipe).

Ne pas modifier sans en discuter avec l'équipe : DiffFile est produit par le
module de parsing Git (module B), CommitMessage est consommé par la couche
d'affichage Go.
"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass
class DiffFile:
    path: str
    status: str  # A, M, D, R...
    added: int
    removed: int
    is_binary: bool
    patch: str


@dataclass
class CommitMessage:
    type: str  # feat, fix, chore, docs, refactor, test, perf, build, ci
    scope: str | None
    subject: str
    body: str | None

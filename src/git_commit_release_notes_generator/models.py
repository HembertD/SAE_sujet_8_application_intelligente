"""Structures de données partagées entre les modules (contrat d'équipe).

Ne pas modifier sans en discuter avec l'équipe : DiffFile est produit par le
module de parsing Git (module B), CommitMessage est consommé par la couche
d'affichage Go.
"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(init=False)
class DiffFile:
    path: str
    status: str  # A, M, D, R...
    added: int
    removed: int
    is_binary: bool
    patch: str

    def __init__(
        self,
        path: str,
        status: str,
        added: int,
        removed: int,
        is_binary: bool | None = None,
        patch: str = "",
        *,
        binary: bool | None = None,
    ) -> None:
        if is_binary is not None and binary is not None and is_binary != binary:
            raise TypeError("Les champs 'is_binary' et 'binary' ne peuvent pas diverger.")

        resolved_binary = is_binary if is_binary is not None else binary
        if resolved_binary is None:
            resolved_binary = False

        self.path = path
        self.status = status
        self.added = added
        self.removed = removed
        self.is_binary = bool(resolved_binary)
        self.patch = patch

    @property
    def binary(self) -> bool:
        """Alias rétrocompatible avec l'ancien contrat historique."""
        return self.is_binary

    def to_dict(self) -> dict[str, str | int | bool]:
        return {
            "path": self.path,
            "status": self.status,
            "added": self.added,
            "removed": self.removed,
            "patch": self.patch,
            "is_binary": self.is_binary,
        }


@dataclass
class CommitMessage:
    type: str  # feat, fix, chore, docs, refactor, test, perf, build, ci
    scope: str | None
    subject: str
    body: str | None

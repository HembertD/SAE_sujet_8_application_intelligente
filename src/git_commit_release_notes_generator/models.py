"""Structures de données partagées entre les modules (contrat d'équipe).

Ne pas modifier sans en discuter avec l'équipe : DiffFile est produit par le
module de parsing Git (module B), CommitMessage est consommé par la couche
d'affichage Go.
"""
from __future__ import annotations

from dataclasses import dataclass, field


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


@dataclass
class ConflictSection:
    """Une section de conflit de merge extraite d'un fichier.

    Produite par le module de parsing Git (module B, GitWrapper) lors de la
    détection d'un conflit sur le disque ; consommée par le module C
    (ollama_client.conflict_resolver) pour proposer une résolution.

    Les champs historiques (`file_path`, `ours`, `theirs`, `base`) restent
    compatibles avec les appels existants. Les champs complémentaires sont
    optionnels pour permettre au wrapper d'exposer le contexte utile au LLM
    sans casser les tests ou les consommateurs actuels.
    """
    file_path: str
    ours: str
    theirs: str
    base: str | None = None
    section_index: int = 0
    start_line: int = 0
    end_line: int = 0
    context_before: str = ""
    context_after: str = ""
    conflict_type: str = "text"
    resolution_supported: bool = True
    source: str = "markers"

    def to_dict(self) -> dict[str, object]:
        return {
            "file_path": self.file_path,
            "section_index": self.section_index,
            "start_line": self.start_line,
            "end_line": self.end_line,
            "base": self.base,
            "ours": self.ours,
            "theirs": self.theirs,
            "context_before": self.context_before,
            "context_after": self.context_after,
            "conflict_type": self.conflict_type,
            "resolution_supported": self.resolution_supported,
            "source": self.source,
        }


@dataclass
class ConflictFile:
    """Conflit Git sur un fichier, avec ses sections et ses versions complètes.

    `base_content`, `ours_content` et `theirs_content` proviennent des stages
    Git (`:1:`, `:2:`, `:3:`) quand ils existent. Ils servent de repli lorsque
    le fichier de travail ne contient pas de marqueurs exploitables.
    """
    path: str
    conflict_type: str
    resolution_supported: bool
    is_binary: bool = False
    sections: list[ConflictSection] = field(default_factory=list)
    base_content: str | None = None
    ours_content: str | None = None
    theirs_content: str | None = None
    reason: str = ""
    truncated: bool = False

    def to_dict(self) -> dict[str, object]:
        return {
            "path": self.path,
            "conflict_type": self.conflict_type,
            "resolution_supported": self.resolution_supported,
            "is_binary": self.is_binary,
            "reason": self.reason,
            "truncated": self.truncated,
            "base_content": self.base_content,
            "ours_content": self.ours_content,
            "theirs_content": self.theirs_content,
            "sections": [section.to_dict() for section in self.sections],
        }


@dataclass
class GitState:
    """État Git utile pour adapter le payload envoyé au LLM.

    `operation` vaut `none`, `merge`, `rebase`, `cherry-pick` ou `revert`.
    Les SHA sont volontairement tronqués à 7 caractères, comme les commits
    exposés par `CommitInfo`.
    """
    branch: str = ""
    head: str = ""
    operation: str = "none"
    in_progress: bool = False
    has_conflicts: bool = False
    merge_head: str = ""
    merge_heads: list[str] = field(default_factory=list)
    merge_base: str = ""
    head_parents: list[str] = field(default_factory=list)
    is_merge_commit: bool = False

    def to_dict(self) -> dict[str, object]:
        return {
            "branch": self.branch,
            "head": self.head,
            "operation": self.operation,
            "in_progress": self.in_progress,
            "has_conflicts": self.has_conflicts,
            "merge_head": self.merge_head,
            "merge_heads": self.merge_heads,
            "merge_base": self.merge_base,
            "head_parents": self.head_parents,
            "is_merge_commit": self.is_merge_commit,
        }


@dataclass
class ConflictResolution:
    """Proposition de résolution d'un conflit, produite par le module C.

    Jamais appliquée automatiquement au fichier : consommée par la couche
    d'affichage Go, qui doit toujours la soumettre à validation humaine
    avant toute écriture sur disque.
    """
    resolved_code: str
    explanation: str

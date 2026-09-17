from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path

from git import InvalidGitRepositoryError, Repo

from git_commit_release_notes_generator.models import DiffFile


class GitWrapperError(Exception):
    """Exception métier levée lors d'un échec de commande Git."""


@dataclass(frozen=True)
class CommitInfo:
    sha: str
    message: str
    author: str
    date: str


class GitWrapper:
    """Wrapper pour extraire l'état d'un dépôt Git et assainir les données avant inférence."""

    SECRET_PATTERNS = [
        re.compile(r"(?i)(api[_-]?key|secret|token|password|auth)\s*[:=]\s*['\"]?([a-zA-Z0-9_\-\.]{8,})['\"]?"),
        re.compile(r"ghp_[a-zA-Z0-9]{36}"),
        re.compile(r"ey[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*"),
    ]

    def __init__(self, repo_path: str | Path = ".") -> None:
        try:
            self.repo = Repo(repo_path, search_parent_directories=True)
        except InvalidGitRepositoryError as exc:
            raise GitWrapperError(f"Le dossier '{repo_path}' n'est pas un dépôt Git valide.") from exc

    @classmethod
    def _sanitize_diff_text(cls, raw_diff: str) -> str:
        sanitized = raw_diff
        for pattern in cls.SECRET_PATTERNS:
            sanitized = pattern.sub(
                lambda match: f"{match.group(1)}: [REDACTED_SECRET]" if match.lastindex and match.group(1) else "[REDACTED_SECRET]",
                sanitized,
            )
        return sanitized

    def sanitize_diff(self, raw_diff: str) -> str:
        """Supprime ou masque les secrets et tokens sensibles détectés dans le diff."""
        return self._sanitize_diff_text(raw_diff)

    @staticmethod
    def _parse_diff(raw_diff: str) -> list[DiffFile]:
        """Découpe un diff brut Git en objets DiffFile, un par fichier modifié."""
        if raw_diff is None or not raw_diff.strip():
            return []

        sanitized = GitWrapper._sanitize_diff_text(raw_diff)
        normalized = re.sub(r"(?m)^\s*diff --git ", "diff --git ", sanitized)
        chunks = re.split(r"(?m)^diff --git ", normalized)
        diff_files: list[DiffFile] = []

        for chunk in chunks:
            if not chunk.strip():
                continue
            section = "diff --git " + chunk.strip()
            header_match = re.search(r"^diff --git (?P<old>\S+) (?P<new>\S+)$", section, flags=re.MULTILINE)
            if header_match is None:
                continue

            old_path = header_match.group("old")
            new_path = header_match.group("new")

            if old_path.startswith("a/"):
                old_file = old_path[2:]
            else:
                old_file = old_path

            if new_path.startswith("b/"):
                new_file = new_path[2:]
            else:
                new_file = new_path

            file_path = new_file if new_file != "/dev/null" else old_file
            if not file_path:
                file_path = old_file

            if old_path == "/dev/null":
                status = "A"
            elif new_path == "/dev/null":
                status = "D"
            elif old_file != new_file:
                status = "R"
            else:
                status = "M"

            if "Binary files" in section or "GIT binary patch" in section:
                diff_files.append(
                    DiffFile(
                        path=file_path,
                        status=status,
                        added=0,
                        removed=0,
                        is_binary=True,
                        patch=section,
                    )
                )
                continue

            added = 0
            removed = 0
            for line in section.splitlines():
                if line.startswith("+") and not line.startswith("+++"):
                    added += 1
                elif line.startswith("-") and not line.startswith("---"):
                    removed += 1

            diff_files.append(
                DiffFile(
                    path=file_path,
                    status=status,
                    added=added,
                    removed=removed,
                    is_binary=False,
                    patch=section,
                )
            )

        return diff_files

    def get_staged_diff(self, max_chars: int = 8000) -> list[DiffFile]:
        """Extrait le diff indexé en objets DiffFile structurés, sans fichiers binaires."""
        try:
            diff_text = self.repo.git.diff("--cached", "--no-color", "--diff-filter=ACMRTUXB")
            if not diff_text.strip():
                return []

            parsed = self._parse_diff(diff_text)
            sanitized_files: list[DiffFile] = []
            for diff_file in parsed:
                if diff_file.is_binary:
                    continue
                patch = diff_file.patch
                if len(patch) > max_chars:
                    patch = patch[:max_chars] + "\n\n... [DIFF TRUNCATED: TROP VOLUMINEUX] ..."
                sanitized_files.append(
                    DiffFile(
                        path=diff_file.path,
                        status=diff_file.status,
                        added=diff_file.added,
                        removed=diff_file.removed,
                        is_binary=False,
                        patch=patch,
                    )
                )
            return sanitized_files
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la récupération du diff indexé : {exc}") from exc

    def get_staged_diff_as_json(self, max_chars: int = 8000) -> list[dict[str, str | int | bool]]:
        """Retourne le diff indexé sous forme de liste JSON-ready, par fichier modifié."""
        return [diff_file.to_dict() for diff_file in self.get_staged_diff(max_chars=max_chars)]

    def get_staged_diff_payload(self, max_chars: int = 8000) -> dict[str, object]:
        """Retourne le payload JSON final sans contexte de prompt, prêt pour le LLM."""
        files = self.get_staged_diff(max_chars=max_chars)
        return {
            "summary": {
                "file_count": len(files),
                "binary_file_count": sum(1 for diff in files if diff.is_binary),
                "truncated": any(len(diff.patch) >= max_chars for diff in files),
            },
            "files": [diff.to_dict() for diff in files],
        }

    def get_llm_payload(self, max_chars: int = 8000) -> dict[str, object]:
        """Retourne le payload directement exploitable par le modèle, sans bloc prompt interne."""
        return self.get_staged_diff_payload(max_chars=max_chars)

    def get_commits_between_tags(self, tag_a: str, tag_b: str) -> list[CommitInfo]:
        """Récupère l'historique des commits structurés entre deux tags ou révisions."""
        try:
            rev_range = f"{tag_a}..{tag_b}"
            commits = list(self.repo.iter_commits(rev_range))
            return [
                CommitInfo(
                    sha=c.hexsha[:7],
                    message=c.message.strip(),
                    author=c.author.name,
                    date=c.committed_datetime.isoformat(),
                )
                for c in commits
            ]
        except Exception as exc:
            raise GitWrapperError(f"Erreur d'extraction entre {tag_a} et {tag_b} : {exc}") from exc

    def commit(self, message: str) -> str:
        """Applique le commit sur les modifications actuellement indexées."""
        if not self.repo.is_dirty(index=True):
            raise GitWrapperError("Aucune modification indexée (staged) à commiter.")
        try:
            commit_obj = self.repo.index.commit(message)
            return commit_obj.hexsha[:7]
        except Exception as exc:
            raise GitWrapperError(f"Échec de l'application du commit : {exc}") from exc

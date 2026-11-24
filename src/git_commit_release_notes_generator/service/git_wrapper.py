from __future__ import annotations

import difflib
import io
import json
import re
import subprocess
from dataclasses import dataclass
from pathlib import Path

from git import InvalidGitRepositoryError, Repo

from git_commit_release_notes_generator.models import (
    ConflictFile,
    ConflictSection,
    DiffFile,
    GitState,
)


class GitWrapperError(Exception):
    """Exception métier levée lors d'un échec de commande Git."""


def _extract_docx_text(data: bytes) -> str:
    """Extrait le texte brut d'un document Word (.docx), paragraphes et tableaux inclus.

    `document.paragraphs` ignore le contenu situé à l'intérieur des tableaux
    (`python-docx`) : on parcourt donc aussi `document.tables` pour ne pas
    perdre le texte d'un cahier des charges, d'un chiffrage, etc.
    """
    from docx import Document  # import différé : dépendance optionnelle, coûteuse à charger

    document = Document(io.BytesIO(data))
    lines = [p.text for p in document.paragraphs if p.text.strip()]
    for table in document.tables:
        for row in table.rows:
            cells = [cell.text.strip() for cell in row.cells if cell.text.strip()]
            if cells:
                lines.append(" | ".join(cells))
    return "\n".join(lines)


def _extract_pdf_text(data: bytes) -> str:
    """Extrait le texte brut d'un PDF, page par page."""
    from pypdf import PdfReader  # import différé : dépendance optionnelle, coûteuse à charger

    reader = PdfReader(io.BytesIO(data))
    return "\n".join(page.extract_text() or "" for page in reader.pages)


# Formats de document "riches" pour lesquels on tente une extraction de texte
# et un diff ligne à ligne, plutôt que de les traiter comme des binaires opaques
# (images, archives, ...) qui restent exclus du payload LLM.
_DOCUMENT_TEXT_EXTRACTORS = {
    ".docx": _extract_docx_text,
    ".pdf": _extract_pdf_text,
}

# Formats image pour lesquels get_staged_image_changes() expose les octets
# bruts, à destination d'une description par un modèle de vision (cf.
# ollama_client.vision). Contrairement aux documents riches ci-dessus, on ne
# tente aucune extraction de texte ici : seul l'appelant décide d'invoquer
# le modèle de vision, et seulement quand il en a réellement besoin.
_IMAGE_EXTENSIONS = {".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp"}


@dataclass(frozen=True)
class ImageChange:
    """Un fichier image ajouté, modifié ou supprimé, avec son contenu avant/après."""
    path: str
    status: str
    old_bytes: bytes | None
    new_bytes: bytes | None


@dataclass(frozen=True)
class CommitInfo:
    sha: str
    message: str
    author: str
    date: str
    tag: str = ""

    def to_dict(self) -> dict[str, str]:
        return {
            "sha": self.sha,
            "message": self.message,
            "author": self.author,
            "date": self.date,
            "tag": self.tag,
        }


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

            # Les marqueurs "new/deleted file mode" sont des en-têtes Git en début
            # de ligne : on ne doit pas matcher le contenu ajouté/supprimé
            # (lignes préfixées par "+" / "-") qui citerait ces mêmes mots.
            section_lines = section.splitlines()
            if any(line.startswith("new file mode") for line in section_lines):
                status = "A"
            elif any(line.startswith("deleted file mode") for line in section_lines):
                status = "D"
            elif old_path == "/dev/null":
                status = "A"
            elif new_path == "/dev/null":
                status = "D"
            elif old_file != new_file:
                status = "R"
            else:
                status = "M"

            section_lines = section.splitlines()
            if any(line.startswith("Binary files") for line in section_lines) or any(
                line.startswith("GIT binary patch") for line in section_lines
            ):
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

    def get_staged_file_names(self) -> list[str]:
        """Renvoie la liste des noms de fichiers indexés (staged), même si le diff est vide.

        Permet de distinguer "aucun fichier staged" de "fichier(s) staged mais
        sans diff exploitable" (fichier vide, par exemple).
        """
        try:
            # U (unmerged) est volontairement exclu : les conflits sont exposés
            # séparément par get_conflict_files()/get_merge_payload().
            names = self.repo.git.diff("--cached", "--name-only", "--diff-filter=ACMRTBD")
            return [n for n in names.splitlines() if n.strip()]
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la récupération des fichiers indexés : {exc}") from exc

    def _read_staged_blob_bytes(self, file_path: str) -> bytes | None:
        """Lit le contenu brut (octets) tel qu'indexé (staged), sans dépendre du
        contenu actuel sur disque (qui peut avoir divergé depuis le `git add`)."""
        entry = self.repo.index.entries.get((file_path, 0))
        if entry is None:
            return None
        return self.repo.odb.stream(entry.binsha).read()

    def _read_head_blob_bytes(self, file_path: str) -> bytes | None:
        """Lit le contenu brut (octets) de la version HEAD d'un fichier, ou None s'il est nouveau."""
        try:
            if not self.repo.head.is_valid():
                return None
            blob = self.repo.head.commit.tree / file_path
            return blob.data_stream.read()
        except Exception:
            return None

    def _get_staged_rename_sources(self) -> dict[str, str]:
        """Associe, pour les fichiers renommés (status "R") indexés, le nouveau
        chemin à son chemin d'origine dans HEAD.

        Nécessite une détection explicite des renommages (`-M`), car
        `get_staged_diff()` ne l'active pas par défaut (coût de calcul inutile
        pour l'essentiel des diffs). Best-effort : renvoie un dict vide en cas
        d'échec plutôt que de faire planter l'appelant, cette information ne
        servant qu'à affiner un diff déjà fonctionnel sans elle (cf.
        `_try_build_document_diff`).
        """
        try:
            output = self.repo.git.diff("--cached", "--no-color", "-M", "--diff-filter=R", "--name-status")
        except Exception:
            return {}

        sources: dict[str, str] = {}
        for line in output.splitlines():
            parts = line.split("\t")
            if len(parts) == 3 and parts[0].startswith("R"):
                sources[parts[2]] = parts[1]
        return sources

    def _try_build_document_diff(self, diff_file: DiffFile, old_path: str | None = None) -> DiffFile | None:
        """Tente de transformer un DiffFile binaire en diff textuel exploitable
        par le LLM, pour les formats de document "riches" supportés (.docx, .pdf).

        `old_path` permet de lire la version HEAD depuis le chemin d'origine
        d'un fichier renommé (`diff_file.path` porte alors déjà le nouveau nom,
        introuvable dans HEAD) ; par défaut, identique à `diff_file.path`.

        Renvoie None si l'extension n'est pas supportée ou si l'extraction
        échoue (document corrompu, format inattendu, dépendance manquante) :
        l'appelant doit alors conserver le comportement binaire existant
        (fichier exclu du payload LLM) plutôt que de faire planter le diff.
        """
        extractor = _DOCUMENT_TEXT_EXTRACTORS.get(Path(diff_file.path).suffix.lower())
        if extractor is None:
            return None

        try:
            new_bytes = None if diff_file.status == "D" else self._read_staged_blob_bytes(diff_file.path)
            old_bytes = None if diff_file.status == "A" else self._read_head_blob_bytes(old_path or diff_file.path)

            new_text = extractor(new_bytes) if new_bytes else ""
            old_text = extractor(old_bytes) if old_bytes else ""
        except Exception:
            return None

        if not new_text and not old_text:
            return None

        diff_lines = list(
            difflib.unified_diff(
                old_text.splitlines(),
                new_text.splitlines(),
                fromfile=f"a/{diff_file.path}",
                tofile=f"b/{diff_file.path}",
                lineterm="",
            )
        )
        added = sum(1 for line in diff_lines if line.startswith("+") and not line.startswith("+++"))
        removed = sum(1 for line in diff_lines if line.startswith("-") and not line.startswith("---"))

        header = f"diff --git a/{diff_file.path} b/{diff_file.path}"
        patch = header + "\n" + "\n".join(diff_lines) if diff_lines else header + "\n(contenu textuel identique)"

        return DiffFile(
            path=diff_file.path,
            status=diff_file.status,
            added=added,
            removed=removed,
            is_binary=False,
            patch=self._sanitize_diff_text(patch),
        )

    def get_staged_diff(self, max_chars: int = 8000) -> list[DiffFile]:
        """Extrait le diff indexé en objets DiffFile structurés.

        Les formats de document "riches" supportés (.docx, .pdf) sont convertis
        en diff textuel par extraction de contenu (cf. `_try_build_document_diff`).
        Les autres binaires (images, archives, ...) sont conservés avec
        `is_binary=True` et un patch vide : présents dans la liste, sans contenu.
        """
        try:
            diff_text = self.repo.git.diff("--cached", "--no-color", "--diff-filter=ACMRTBD")
            if not diff_text.strip():
                staged_files = self.get_staged_file_names()
                if staged_files:
                    # Fichiers indexés mais sans diff exploitable (ex: fichier vide).
                    # Retourne un DiffFile vide pour que l'appelant sache qu'un fichier est staged.
                    return [
                        DiffFile(
                            path=staged_files[0],
                            status="M",
                            added=0,
                            removed=0,
                            is_binary=False,
                            patch="",
                        )
                    ]
                return []

            parsed = self._parse_diff(diff_text)
            rename_sources: dict[str, str] | None = None
            sanitized_files: list[DiffFile] = []
            for diff_file in parsed:
                # Un PDF n'a pas forcément d'octet nul dans ses premiers kilo-octets :
                # Git peut le classer comme "texte" (is_binary=False) et produire un
                # diff ligne à ligne du bytecode brut. On tente donc l'extraction pour
                # toute extension supportée, que Git l'ait étiqueté binaire ou non —
                # sinon ce bytecode (avec ses octets non-UTF8) finit dans le patch et
                # fait planter la sérialisation JSON en aval (surrogates invalides).
                is_supported_doc = Path(diff_file.path).suffix.lower() in _DOCUMENT_TEXT_EXTRACTORS
                if diff_file.is_binary or is_supported_doc:
                    old_path = None
                    if is_supported_doc and diff_file.status == "R":
                        # Calculé une seule fois, à la demande : un document
                        # renommé (diff_file.path = le nouveau nom) a sa version
                        # HEAD sous l'ancien nom, pas sous le nouveau.
                        if rename_sources is None:
                            rename_sources = self._get_staged_rename_sources()
                        old_path = rename_sources.get(diff_file.path)
                    document_diff = self._try_build_document_diff(diff_file, old_path=old_path)
                    if document_diff is not None:
                        diff_file = document_diff
                    else:
                        # Binaire opaque (image, archive, document illisible) : on le garde
                        # pour que le TUI l'affiche ("[Binaire]") et que le LLM sache qu'il
                        # a changé, mais sans patch, qui contiendrait des octets bruts non-UTF8.
                        diff_file = DiffFile(
                            path=diff_file.path,
                            status=diff_file.status,
                            added=0,
                            removed=0,
                            is_binary=True,
                            patch="",
                        )

                patch = diff_file.patch
                if len(patch) > max_chars:
                    patch = patch[:max_chars] + "\n\n... [DIFF TRUNCATED: TROP VOLUMINEUX] ..."
                sanitized_files.append(
                    DiffFile(
                        path=diff_file.path,
                        status=diff_file.status,
                        added=diff_file.added,
                        removed=diff_file.removed,
                        is_binary=diff_file.is_binary,
                        patch=patch,
                    )
                )
            return sanitized_files
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la récupération du diff indexé : {exc}") from exc

    def get_staged_image_changes(self) -> list[ImageChange]:
        """Liste les fichiers image indexés (ajoutés/modifiés/supprimés), avec
        leurs octets avant/après.

        N'effectue aucun appel réseau : c'est à l'appelant (service/core.py,
        uniquement au moment de générer un message de commit) de décider s'il
        invoque un modèle de vision sur ces octets. Reste donc indépendant de
        get_staged_diff(), qui doit rester rapide et sans réseau.
        """
        try:
            diff_text = self.repo.git.diff("--cached", "--no-color", "--diff-filter=ACMRTUXBD")
            if not diff_text.strip():
                return []

            changes: list[ImageChange] = []
            for diff_file in self._parse_diff(diff_text):
                if not diff_file.is_binary:
                    continue
                if Path(diff_file.path).suffix.lower() not in _IMAGE_EXTENSIONS:
                    continue

                new_bytes = None if diff_file.status == "D" else self._read_staged_blob_bytes(diff_file.path)
                old_bytes = None if diff_file.status == "A" else self._read_head_blob_bytes(diff_file.path)
                if not new_bytes and not old_bytes:
                    continue

                changes.append(
                    ImageChange(path=diff_file.path, status=diff_file.status, old_bytes=old_bytes, new_bytes=new_bytes)
                )
            return changes
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la détection des images modifiées : {exc}") from exc

    @staticmethod
    def _short_sha(sha: str) -> str:
        """Tronque un SHA Git à 7 caractères, comme les commits exposés au LLM."""
        value = sha.strip().splitlines()[0] if sha.strip() else ""
        return value[:7]

    def _read_git_file(self, name: str) -> str:
        """Lit un fichier interne du répertoire .git sans lever si absent."""
        try:
            return (Path(self.repo.git_dir) / name).read_text(encoding="utf-8", errors="replace").strip()
        except OSError:
            return ""

    def _get_unmerged_stages(self) -> dict[str, dict[int, str]]:
        """Retourne les stages Git (1=base, 2=ours, 3=theirs) des fichiers en conflit."""
        try:
            raw = self.repo.git.ls_files("-u", "-z")
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la récupération des conflits Git : {exc}") from exc

        if isinstance(raw, bytes):
            raw = raw.decode("utf-8", errors="replace")

        result: dict[str, dict[int, str]] = {}
        for entry in raw.split("\0"):
            if not entry or "\t" not in entry:
                continue
            meta, path = entry.split("\t", 1)
            parts = meta.split()
            if len(parts) < 3:
                continue
            try:
                stage = int(parts[2])
            except ValueError:
                continue
            result.setdefault(path, {})[stage] = parts[1]
        return result

    def get_unmerged_paths(self) -> list[str]:
        """Retourne les chemins actuellement en conflit (index non fusionné)."""
        return sorted(self._get_unmerged_stages())

    @staticmethod
    def _conflict_type_from_stages(stages: set[int]) -> str:
        """Déduit le type de conflit Git à partir des stages présents dans l'index."""
        if stages == {1, 2, 3}:
            return "both_modified"
        if stages == {2, 3}:
            return "both_added"
        if stages == {1, 2}:
            return "deleted_by_them"
        if stages == {1, 3}:
            return "deleted_by_us"
        if stages == {1}:
            return "both_deleted"
        if stages == {2}:
            return "added_by_us"
        if stages == {3}:
            return "added_by_them"
        return "unknown"

    def get_git_state(self) -> GitState:
        """Expose l'état Git utile au LLM : branche, opération et parents de merge."""
        try:
            branch = ""
            try:
                branch = self.repo.active_branch.name
            except (TypeError, ValueError):
                branch = "HEAD" if self.repo.head.is_valid() else ""

            head = self._short_sha(self.repo.head.commit.hexsha) if self.repo.head.is_valid() else ""
            head_parents = (
                [self._short_sha(parent.hexsha) for parent in self.repo.head.commit.parents]
                if self.repo.head.is_valid()
                else []
            )

            git_dir = Path(self.repo.git_dir)
            operation = "none"
            if (git_dir / "MERGE_HEAD").exists():
                operation = "merge"
            elif (git_dir / "rebase-merge").exists() or (git_dir / "rebase-apply").exists() or (git_dir / "REBASE_HEAD").exists():
                operation = "rebase"
            elif (git_dir / "CHERRY_PICK_HEAD").exists():
                operation = "cherry-pick"
            elif (git_dir / "REVERT_HEAD").exists():
                operation = "revert"

            merge_heads_full: list[str] = []
            if operation == "merge":
                merge_heads_full = [line.strip() for line in self._read_git_file("MERGE_HEAD").splitlines() if line.strip()]

            merge_base = ""
            if merge_heads_full and head:
                try:
                    bases = self.repo.merge_base("HEAD", merge_heads_full[0])
                    if bases:
                        merge_base = self._short_sha(bases[0].hexsha)
                except Exception:
                    merge_base = ""

            has_conflicts = bool(self._get_unmerged_stages())
            return GitState(
                branch=branch,
                head=head,
                operation=operation,
                in_progress=operation != "none",
                has_conflicts=has_conflicts,
                merge_head=self._short_sha(merge_heads_full[0]) if merge_heads_full else "",
                merge_heads=[self._short_sha(sha) for sha in merge_heads_full],
                merge_base=merge_base,
                head_parents=head_parents,
                is_merge_commit=len(head_parents) > 1,
            )
        except GitWrapperError:
            raise
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la lecture de l'état Git : {exc}") from exc

    def _show_stage(self, path: str, stage: int) -> tuple[str | None, bool]:
        """Retourne le contenu texte d'un stage Git et indique s'il est binaire."""
        cwd = self.repo.working_tree_dir or self.repo.git_dir
        try:
            proc = subprocess.run(
                ["git", "show", f":{stage}:{path}"],
                cwd=cwd,
                capture_output=True,
                check=False,
            )
        except OSError as exc:
            raise GitWrapperError(f"Impossible d'exécuter git show pour '{path}' : {exc}") from exc

        if proc.returncode != 0:
            return None, False
        data = proc.stdout
        if b"\x00" in data:
            return None, True
        try:
            return data.decode("utf-8"), False
        except UnicodeDecodeError:
            return None, True

    def _read_worktree_text(self, path: str) -> tuple[str | None, bool]:
        """Lit le fichier de travail courant et détecte un contenu binaire."""
        root = self.repo.working_tree_dir
        if not root:
            return None, False
        try:
            data = (Path(root) / path).read_bytes()
        except OSError:
            return None, False
        if b"\x00" in data:
            return None, True
        try:
            return data.decode("utf-8"), False
        except UnicodeDecodeError:
            return None, True

    @staticmethod
    def _parse_conflict_sections(
        file_path: str,
        text: str,
        context_lines: int = 3,
        conflict_type: str = "text",
    ) -> list[ConflictSection]:
        """Découpe un fichier contenant des marqueurs de conflit en sections."""
        lines = text.splitlines()
        sections: list[ConflictSection] = []
        section_index = 0
        index = 0

        while index < len(lines):
            if not lines[index].startswith("<<<<<<<"):
                index += 1
                continue

            start = index
            index += 1
            ours_lines: list[str] = []
            base_lines: list[str] = []
            theirs_lines: list[str] = []
            in_base = False

            while index < len(lines) and not lines[index].startswith("=======") and not lines[index].startswith(">>>>>>>"):
                if lines[index].startswith("|||||||"):
                    in_base = True
                elif in_base:
                    base_lines.append(lines[index])
                else:
                    ours_lines.append(lines[index])
                index += 1

            if index < len(lines) and lines[index].startswith("======="):
                index += 1
                while index < len(lines) and not lines[index].startswith(">>>>>>>"):
                    theirs_lines.append(lines[index])
                    index += 1

            end = index if index < len(lines) and lines[index].startswith(">>>>>>>") else index - 1
            if index < len(lines) and lines[index].startswith(">>>>>>>"):
                index += 1

            context_before = "\n".join(lines[max(0, start - context_lines):start])
            context_after = "\n".join(lines[end + 1:end + 1 + context_lines]) if end + 1 <= len(lines) else ""

            sections.append(
                ConflictSection(
                    file_path=file_path,
                    ours="\n".join(ours_lines),
                    theirs="\n".join(theirs_lines),
                    base="\n".join(base_lines) if in_base else None,
                    section_index=section_index,
                    start_line=start + 1,
                    end_line=end + 1,
                    context_before=context_before,
                    context_after=context_after,
                    conflict_type=conflict_type,
                    resolution_supported=True,
                    source="markers",
                )
            )
            section_index += 1

        return sections

    @staticmethod
    def _truncate(text: str | None, limit: int) -> tuple[str | None, bool]:
        """Tronque un texte à une limite et signale si une troncature a eu lieu."""
        if text is None:
            return None, False
        if limit < 0:
            limit = 0
        if len(text) <= limit:
            return text, False
        return text[:limit] + "\n... [TRUNCATED] ...", True

    def get_conflict_files(self, max_chars: int = 8000, context_lines: int = 3) -> list[ConflictFile]:
        """Retourne les fichiers en conflit avec leurs sections et versions Git."""
        try:
            stages_map = self._get_unmerged_stages()
            conflicts: list[ConflictFile] = []

            for path in sorted(stages_map):
                stages = stages_map[path]
                conflict_type = self._conflict_type_from_stages(set(stages))

                base_content, base_binary = self._show_stage(path, 1)
                ours_content, ours_binary = self._show_stage(path, 2)
                theirs_content, theirs_binary = self._show_stage(path, 3)
                worktree_text, worktree_binary = self._read_worktree_text(path)
                is_binary = base_binary or ours_binary or theirs_binary or worktree_binary

                sections: list[ConflictSection] = []
                if not is_binary and worktree_text is not None:
                    sections = self._parse_conflict_sections(path, worktree_text, context_lines, conflict_type)

                if not is_binary and not sections:
                    side_text = ours_content if ours_content is not None else theirs_content
                    if side_text is not None:
                        sections = [
                            ConflictSection(
                                file_path=path,
                                ours=ours_content or "",
                                theirs=theirs_content or "",
                                base=base_content,
                                section_index=0,
                                start_line=1,
                                end_line=max(1, len(side_text.splitlines())),
                                conflict_type=conflict_type,
                                resolution_supported=True,
                                source="index",
                            )
                        ]

                resolution_supported = not is_binary and conflict_type not in {"both_deleted", "unknown"}
                reason = ""
                if is_binary:
                    reason = "fichier binaire"
                elif conflict_type == "both_deleted":
                    reason = "fichier supprimé des deux côtés"
                elif conflict_type == "unknown":
                    reason = "type de conflit non pris en charge"

                truncated = False
                base_content, truncated_base = self._truncate(base_content, max_chars)
                ours_content, truncated_ours = self._truncate(ours_content, max_chars)
                theirs_content, truncated_theirs = self._truncate(theirs_content, max_chars)
                truncated = truncated or truncated_base or truncated_ours or truncated_theirs

                for section in sections:
                    section.ours, truncated_section_ours = self._truncate(section.ours, max_chars)
                    section.theirs, truncated_section_theirs = self._truncate(section.theirs, max_chars)
                    section.base, truncated_section_base = self._truncate(section.base, max_chars)
                    section.context_before, truncated_before = self._truncate(section.context_before, max_chars)
                    section.context_after, truncated_after = self._truncate(section.context_after, max_chars)
                    truncated = (
                        truncated
                        or truncated_section_ours
                        or truncated_section_theirs
                        or truncated_section_base
                        or truncated_before
                        or truncated_after
                    )

                conflicts.append(
                    ConflictFile(
                        path=path,
                        conflict_type=conflict_type,
                        resolution_supported=resolution_supported,
                        is_binary=is_binary,
                        sections=sections,
                        base_content=base_content,
                        ours_content=ours_content,
                        theirs_content=theirs_content,
                        reason=reason,
                        truncated=truncated,
                    )
                )

            return conflicts
        except GitWrapperError:
            raise
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de l'extraction des conflits : {exc}") from exc

    def get_conflict_sections(self, max_chars: int = 8000, context_lines: int = 3) -> list[ConflictSection]:
        """Retourne toutes les sections de conflit, tous fichiers confondus."""
        sections: list[ConflictSection] = []
        for conflict_file in self.get_conflict_files(max_chars=max_chars, context_lines=context_lines):
            sections.extend(conflict_file.sections)
        return sections

    @staticmethod
    def _contains_truncation(value: object) -> bool:
        if isinstance(value, str):
            return "TRUNCATED" in value
        if isinstance(value, list):
            return any(GitWrapper._contains_truncation(item) for item in value)
        if isinstance(value, dict):
            return any(GitWrapper._contains_truncation(item) for item in value.values())
        return False

    @staticmethod
    def _apply_total_budget(
        payload: dict[str, object],
        total_max_chars: int,
        truncated_files: set[str],
    ) -> dict[str, object]:
        """Réduit progressivement le payload pour respecter un budget global."""
        if total_max_chars <= 0:
            return payload

        for _ in range(12):
            if len(json.dumps(payload, ensure_ascii=False)) <= total_max_chars:
                break

            # Les conflits sont prioritaires : on réduit d'abord les fichiers
            # indexés classiques, puis seulement les sections de conflit.
            changed_files = False
            files = payload.get("files", [])
            if isinstance(files, list):
                for file_payload in files:
                    if not isinstance(file_payload, dict):
                        continue
                    patch = file_payload.get("patch")
                    if isinstance(patch, str) and len(patch) > 200:
                        file_payload["patch"] = patch[: max(200, len(patch) // 2)] + "\n... [TRUNCATED] ..."
                        truncated_files.add(str(file_payload.get("path", "")))
                        changed_files = True
            if changed_files:
                continue

            changed_conflicts = False
            conflicts = payload.get("conflicts", [])
            if isinstance(conflicts, list):
                for conflict in conflicts:
                    if not isinstance(conflict, dict):
                        continue
                    for key in ("base_content", "ours_content", "theirs_content"):
                        content = conflict.get(key)
                        if isinstance(content, str) and len(content) > 200:
                            conflict[key] = content[: max(200, len(content) // 2)] + "\n... [TRUNCATED] ..."
                            truncated_files.add(str(conflict.get("path", "")))
                            changed_conflicts = True
                    sections = conflict.get("sections", [])
                    if isinstance(sections, list):
                        for section in sections:
                            if not isinstance(section, dict):
                                continue
                            for key in ("base", "ours", "theirs", "context_before", "context_after"):
                                content = section.get(key)
                                if isinstance(content, str) and len(content) > 200:
                                    section[key] = content[: max(200, len(content) // 2)] + "\n... [TRUNCATED] ..."
                                    truncated_files.add(str(conflict.get("path", "")))
                                    changed_conflicts = True

            if not changed_conflicts:
                break

        return payload

    def get_merge_payload(
        self,
        max_chars: int = 8000,
        total_max_chars: int = 60000,
        context_lines: int = 3,
    ) -> dict[str, object]:
        """Construit le payload complet diff + état Git + conflits pour le LLM."""
        state = self.get_git_state()
        files = self.get_staged_diff(max_chars=max_chars)
        conflicts = self.get_conflict_files(max_chars=max_chars, context_lines=context_lines)

        files_payload = [diff.to_dict() for diff in files]
        conflicts_payload = [conflict.to_dict() for conflict in conflicts]
        truncated_files: set[str] = set()

        for diff_payload in files_payload:
            if self._contains_truncation(diff_payload.get("patch", "")):
                truncated_files.add(str(diff_payload.get("path", "")))
        for conflict_payload in conflicts_payload:
            if conflict_payload.get("truncated") or self._contains_truncation(conflict_payload):
                truncated_files.add(str(conflict_payload.get("path", "")))

        payload: dict[str, object] = {
            "repository": state.to_dict(),
            "summary": {
                "file_count": len(files_payload),
                "conflict_count": len(conflicts_payload),
                "binary_file_count": sum(1 for conflict in conflicts if conflict.is_binary),
                "binary_conflict_count": sum(1 for conflict in conflicts if conflict.is_binary),
                "truncated": bool(truncated_files),
                "truncated_files": sorted(truncated_files),
            },
            "files": files_payload,
            "conflicts": conflicts_payload,
        }

        payload = self._apply_total_budget(payload, total_max_chars, truncated_files)
        summary = payload.get("summary")
        if isinstance(summary, dict):
            summary["truncated"] = bool(truncated_files)
            summary["truncated_files"] = sorted(truncated_files)
        return payload

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

    def get_llm_payload(
        self,
        max_chars: int = 8000,
        total_max_chars: int = 60000,
        context_lines: int = 3,
    ) -> dict[str, object]:
        """Retourne le payload directement exploitable par le modèle.

        Le payload contient le diff indexé, l'état Git et les conflits non
        résolus. Il reste compatible avec les clés historiques `summary` et
        `files`, tout en ajoutant `repository` et `conflicts`.
        """
        return self.get_merge_payload(
            max_chars=max_chars,
            total_max_chars=total_max_chars,
            context_lines=context_lines,
        )

    def get_recent_commits(self, limit: int = 0) -> list[CommitInfo]:
        """Récupère l'historique complet ou récent des commits avec leurs tags associés."""
        try:
            if not self.repo.head.is_valid():
                return []
            tags_by_sha: dict[str, str] = {}
            for tag in self.repo.tags:
                try:
                    target_sha = tag.commit.hexsha
                    tags_by_sha[target_sha] = tag.name
                except Exception:
                    pass

            kwargs: dict[str, object] = {}
            if limit > 0:
                kwargs["max_count"] = limit

            commits = list(self.repo.iter_commits("HEAD", **kwargs))
            return [
                CommitInfo(
                    sha=c.hexsha[:7],
                    message=c.message.splitlines()[0].strip() if c.message else "",
                    author=c.author.name,
                    date=c.committed_datetime.strftime("%Y-%m-%d"),
                    tag=tags_by_sha.get(c.hexsha, ""),
                )
                for c in commits
            ]
        except Exception as exc:
            raise GitWrapperError(f"Erreur lors de la récupération des commits : {exc}") from exc

    def get_commits_between_tags(self, tag_a: str, tag_b: str) -> list[CommitInfo]:
        """Récupère l'historique des commits structurés entre deux tags ou révisions."""
        try:
            rev_a = tag_a
            rev_b = tag_b

            # Détection et inversion si l'utilisateur a inversé l'ordre chronologique
            try:
                commit_a = self.repo.commit(tag_a)
                commit_b = self.repo.commit(tag_b)
                if self.repo.is_ancestor(commit_b, commit_a):
                    rev_a, rev_b = tag_b, tag_a
            except Exception:
                pass

            # Si rev_a est un hash de commit (et pas un tag Git explicite), on cherche à inclure le commit de départ
            is_tag = any(t.name == rev_a for t in self.repo.tags)
            if not is_tag:
                try:
                    c_start = self.repo.commit(rev_a)
                    if c_start.parents:
                        rev_range = f"{c_start.hexsha}^..{rev_b}"
                    else:
                        rev_range = rev_b
                except Exception:
                    rev_range = f"{rev_a}..{rev_b}"
            else:
                rev_range = f"{rev_a}..{rev_b}"

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

from __future__ import annotations

import difflib
import io
import re
from dataclasses import dataclass
from pathlib import Path

from git import InvalidGitRepositoryError, Repo

from git_commit_release_notes_generator.models import DiffFile


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
            names = self.repo.git.diff("--cached", "--name-only", "--diff-filter=ACMRTUXBD")
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

        Les fichiers binaires "opaques" (images, archives, ...) restent exclus.
        Les formats de document "riches" supportés (.docx, .pdf) sont convertis
        en diff textuel par extraction de contenu (cf. `_try_build_document_diff`)
        plutôt que d'être traités comme des binaires silencieusement ignorés.
        """
        try:
            diff_text = self.repo.git.diff("--cached", "--no-color", "--diff-filter=ACMRTUXBD")
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
                    if document_diff is None:
                        continue
                    diff_file = document_diff

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

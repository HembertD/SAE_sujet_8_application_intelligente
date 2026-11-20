import io
import subprocess
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from git_commit_release_notes_generator.service.git_wrapper import (
    GitWrapper,
    _extract_docx_text,
    _extract_pdf_text,
)


class GitWrapperParsingTests(unittest.TestCase):
    def test_parse_diff_into_structured_files(self):
        raw_diff = '''diff --git a/app.py b/app.py
index 1111111..2222222 100644
--- a/app.py
+++ b/app.py
@@ -1,2 +1,3 @@
 old
+new
 diff --git a/config.yml b/config.yml
index 3333333..4444444 100644
--- a/config.yml
+++ b/config.yml
@@ -1 +1 @@
-secret: abcdefghijklmnopqrst
+secret: [REDACTED_SECRET]
'''

        diffs = GitWrapper._parse_diff(raw_diff)

        self.assertEqual(len(diffs), 2)
        self.assertEqual(diffs[0].path, "app.py")
        self.assertIn("old", diffs[0].patch)
        self.assertIn("+new", diffs[0].patch)
        self.assertEqual(diffs[1].path, "config.yml")
        self.assertIn("[REDACTED_SECRET]", diffs[1].patch)

    def test_sanitize_removes_sensitive_tokens(self):
        wrapper = GitWrapper.__new__(GitWrapper)
        sanitized = wrapper.sanitize_diff("api_key=abc1234567890")
        self.assertIn("[REDACTED_SECRET]", sanitized)

    def test_json_payload_is_structured_per_file(self):
        raw_diff = '''diff --git a/app.py b/app.py
index 1111111..2222222 100644
--- a/app.py
+++ b/app.py
@@ -1,2 +1,3 @@
 old
+new
 diff --git a/config.yml b/config.yml
index 3333333..4444444 100644
--- a/config.yml
+++ b/config.yml
@@ -1 +1 @@
-secret: abcdefghijklmnopqrst
+secret: [REDACTED_SECRET]
'''

        files = GitWrapper._parse_diff(raw_diff)
        payload = [item.to_dict() for item in files]

        self.assertEqual(len(payload), 2)
        self.assertIn("path", payload[0])
        self.assertIn("status", payload[0])
        self.assertIn("added", payload[0])
        self.assertIn("removed", payload[0])
        self.assertIn("patch", payload[0])
        self.assertIn("is_binary", payload[0])
        self.assertNotIn("binary", payload[0])
        self.assertEqual(payload[0]["path"], "app.py")
        self.assertIn("+new", payload[0]["patch"])

    def test_llm_payload_has_summary_and_files_without_context(self):
        payload = {
            "summary": {
                "file_count": 2,
                "binary_file_count": 0,
                "truncated": False,
            },
            "files": [
                {
                    "path": "app.py",
                    "status": "M",
                    "added": 1,
                    "removed": 0,
                    "patch": "diff --git a/app.py b/app.py\n+new",
                    "binary": False,
                },
                {
                    "path": "config.yml",
                    "status": "M",
                    "added": 0,
                    "removed": 1,
                    "patch": "diff --git a/config.yml b/config.yml\n-secret: test",
                    "binary": False,
                },
            ],
        }

        self.assertNotIn("context", payload)
        self.assertEqual(payload["summary"]["file_count"], 2)
        self.assertIn("files", payload)
        self.assertEqual(payload["files"][0]["path"], "app.py")


def _make_docx_bytes(text: str) -> bytes:
    from docx import Document

    document = Document()
    for line in text.splitlines():
        document.add_paragraph(line)
    buf = io.BytesIO()
    document.save(buf)
    return buf.getvalue()


def _make_pdf_bytes(text: str) -> bytes:
    """Construit un PDF minimal mais valide avec un unique flux de texte,
    exploitable par `pypdf.PdfReader.extract_text()` (pas de dépendance à
    reportlab : seul pypdf, déjà requis par le module, est utilisé)."""
    from pypdf import PdfWriter
    from pypdf.generic import DecodedStreamObject, DictionaryObject, NameObject

    writer = PdfWriter()
    page = writer.add_blank_page(width=300, height=300)

    content = f"BT /F1 14 Tf 10 250 Td ({text}) Tj ET".encode("latin-1")
    stream = DecodedStreamObject()
    stream.set_data(content)
    stream_ref = writer._add_object(stream)
    page[NameObject("/Contents")] = stream_ref

    font = DictionaryObject()
    font[NameObject("/Type")] = NameObject("/Font")
    font[NameObject("/Subtype")] = NameObject("/Type1")
    font[NameObject("/BaseFont")] = NameObject("/Helvetica")
    font_ref = writer._add_object(font)

    fonts_dict = DictionaryObject()
    fonts_dict[NameObject("/F1")] = font_ref
    resources = DictionaryObject()
    resources[NameObject("/Font")] = fonts_dict
    page[NameObject("/Resources")] = resources

    buf = io.BytesIO()
    writer.write(buf)
    return buf.getvalue()


def test_extract_docx_text_reads_paragraphs():
    data = _make_docx_bytes("première ligne\nseconde ligne")
    text = _extract_docx_text(data)
    assert "première ligne" in text
    assert "seconde ligne" in text


def test_extract_pdf_text_reads_content():
    data = _make_pdf_bytes("Hello PDF")
    text = _extract_pdf_text(data)
    assert "Hello PDF" in text


def _run_git(args: list[str], cwd: Path) -> None:
    subprocess.run(["git", *args], cwd=cwd, check=True, capture_output=True, text=True)


def _init_repo_with_base_commit(repo_dir: Path) -> None:
    repo_dir.mkdir()
    _run_git(["init"], repo_dir)
    _run_git(["config", "user.email", "test@example.com"], repo_dir)
    _run_git(["config", "user.name", "Test"], repo_dir)
    (repo_dir / "README.md").write_text("base\n")
    _run_git(["add", "-A"], repo_dir)
    _run_git(["commit", "-m", "base"], repo_dir)


def test_get_staged_diff_converts_modified_docx_to_text_diff(tmp_path):
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    docx_path = repo_dir / "rapport.docx"
    docx_path.write_bytes(_make_docx_bytes("version initiale"))
    _run_git(["add", "-A"], repo_dir)
    _run_git(["commit", "-m", "ajout du rapport"], repo_dir)

    docx_path.write_bytes(_make_docx_bytes("version modifiee"))
    _run_git(["add", "-A"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    files = wrapper.get_staged_diff()

    docx_entries = [f for f in files if f.path == "rapport.docx"]
    assert len(docx_entries) == 1
    entry = docx_entries[0]
    assert entry.is_binary is False
    assert entry.status == "M"
    assert "version initiale" in entry.patch
    assert "version modifiee" in entry.patch
    assert entry.added > 0
    assert entry.removed > 0


def test_get_staged_diff_converts_newly_added_pdf_to_text_diff(tmp_path):
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    pdf_path = repo_dir / "notice.pdf"
    pdf_path.write_bytes(_make_pdf_bytes("Contenu du PDF"))
    _run_git(["add", "-A"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    files = wrapper.get_staged_diff()

    pdf_entries = [f for f in files if f.path == "notice.pdf"]
    assert len(pdf_entries) == 1
    entry = pdf_entries[0]
    assert entry.is_binary is False
    assert entry.status == "A"
    assert "Contenu du PDF" in entry.patch
    assert entry.added > 0
    assert entry.removed == 0


def test_get_staged_diff_still_excludes_unsupported_binary_files(tmp_path):
    """Les binaires sans extracteur dédié (images, archives, ...) restent
    exclus du diff exploitable par le LLM, comportement inchangé."""
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    (repo_dir / "logo.png").write_bytes(b"\x89PNG\r\n\x1a\n" + b"\x00" * 50)
    _run_git(["add", "-A"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    files = wrapper.get_staged_diff()

    assert all(f.path != "logo.png" for f in files)


_PNG_BYTES_V1 = b"\x89PNG\r\n\x1a\n" + bytes(range(50))
_PNG_BYTES_V2 = b"\x89PNG\r\n\x1a\n" + bytes(range(50, 100))


def test_get_staged_image_changes_detects_addition(tmp_path):
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    (repo_dir / "photo.png").write_bytes(_PNG_BYTES_V1)
    _run_git(["add", "-A"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    changes = wrapper.get_staged_image_changes()

    assert len(changes) == 1
    assert changes[0].path == "photo.png"
    assert changes[0].status == "A"
    assert changes[0].new_bytes == _PNG_BYTES_V1
    assert changes[0].old_bytes is None


def test_get_staged_image_changes_detects_modification(tmp_path):
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    photo = repo_dir / "photo.png"
    photo.write_bytes(_PNG_BYTES_V1)
    _run_git(["add", "-A"], repo_dir)
    _run_git(["commit", "-m", "ajout photo"], repo_dir)

    photo.write_bytes(_PNG_BYTES_V2)
    _run_git(["add", "-A"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    changes = wrapper.get_staged_image_changes()

    assert len(changes) == 1
    assert changes[0].status == "M"
    assert changes[0].old_bytes == _PNG_BYTES_V1
    assert changes[0].new_bytes == _PNG_BYTES_V2


def test_get_staged_image_changes_detects_deletion(tmp_path):
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    photo = repo_dir / "photo.png"
    photo.write_bytes(_PNG_BYTES_V1)
    _run_git(["add", "-A"], repo_dir)
    _run_git(["commit", "-m", "ajout photo"], repo_dir)

    _run_git(["rm", "photo.png"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    changes = wrapper.get_staged_image_changes()

    assert len(changes) == 1
    assert changes[0].status == "D"
    assert changes[0].old_bytes == _PNG_BYTES_V1
    assert changes[0].new_bytes is None


def test_get_staged_image_changes_ignores_text_and_document_files(tmp_path):
    repo_dir = tmp_path / "repo"
    _init_repo_with_base_commit(repo_dir)

    (repo_dir / "notes.txt").write_text("hello\n")
    (repo_dir / "rapport.docx").write_bytes(_make_docx_bytes("contenu"))
    _run_git(["add", "-A"], repo_dir)

    wrapper = GitWrapper(repo_dir)
    changes = wrapper.get_staged_image_changes()

    assert changes == []


if __name__ == "__main__":
    unittest.main()

"""Tests du contexte de merge exposé par GitWrapper.

Les scénarios créent de vrais dépôts Git temporaires : ils vérifient les
stages d'index, les marqueurs de conflit et le payload envoyé au LLM, sans
réseau ni fixture externe.
"""
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from git_commit_release_notes_generator.service.git_wrapper import GitWrapper


class GitWrapperMergeTests(unittest.TestCase):
    def _git(self, repo: Path, *args: str, check: bool = True) -> subprocess.CompletedProcess:
        return subprocess.run(
            ["git", *args],
            cwd=repo,
            check=check,
            text=True,
            capture_output=True,
        )

    def _init_repo(self, repo: Path) -> None:
        self._git(repo, "init", "-b", "main")
        self._git(repo, "config", "user.email", "test@example.com")
        self._git(repo, "config", "user.name", "Test User")
        self._git(repo, "config", "commit.gpgsign", "false")
        self._git(repo, "config", "tag.gpgsign", "false")

    def _write_and_commit(self, repo: Path, path: str, content: str, message: str) -> None:
        file_path = repo / path
        file_path.parent.mkdir(parents=True, exist_ok=True)
        file_path.write_text(content, encoding="utf-8")
        self._git(repo, "add", path)
        self._git(repo, "commit", "-m", message)

    def _write_binary_and_commit(self, repo: Path, path: str, content: bytes, message: str) -> None:
        file_path = repo / path
        file_path.parent.mkdir(parents=True, exist_ok=True)
        file_path.write_bytes(content)
        self._git(repo, "add", path)
        self._git(repo, "commit", "-m", message)

    def _diverged_file(self, repo: Path, ours: str = "main\n", theirs: str = "feature\n") -> GitWrapper:
        self._write_and_commit(repo, "f.txt", "base\n", "base")
        self._git(repo, "checkout", "-b", "feature")
        self._write_and_commit(repo, "f.txt", theirs, "feature change")
        self._git(repo, "checkout", "main")
        self._write_and_commit(repo, "f.txt", ours, "main change")
        return GitWrapper(repo)

    def test_merge_conflict_payload_contains_state_and_sections(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            wrapper = self._diverged_file(repo)

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            state = wrapper.get_git_state()
            self.assertEqual(state.branch, "main")
            self.assertEqual(state.operation, "merge")
            self.assertTrue(state.in_progress)
            self.assertTrue(state.has_conflicts)
            self.assertFalse(state.is_merge_commit)
            self.assertTrue(state.merge_head)
            self.assertTrue(state.merge_base)
            self.assertEqual(len(state.merge_heads), 1)

            self.assertEqual(wrapper.get_unmerged_paths(), ["f.txt"])
            self.assertEqual(wrapper.get_staged_diff(), [])
            self.assertEqual(wrapper.get_staged_file_names(), [])

            conflicts = wrapper.get_conflict_files()
            self.assertEqual(len(conflicts), 1)
            conflict = conflicts[0]
            self.assertEqual(conflict.path, "f.txt")
            self.assertEqual(conflict.conflict_type, "both_modified")
            self.assertFalse(conflict.is_binary)
            self.assertTrue(conflict.resolution_supported)
            self.assertFalse(conflict.truncated)
            self.assertEqual(conflict.base_content, "base\n")
            self.assertEqual(conflict.ours_content, "main\n")
            self.assertEqual(conflict.theirs_content, "feature\n")
            self.assertEqual(len(conflict.sections), 1)

            section = conflict.sections[0]
            self.assertEqual(section.file_path, "f.txt")
            self.assertEqual(section.ours, "main")
            self.assertEqual(section.theirs, "feature")
            self.assertIsNone(section.base)
            self.assertEqual(section.source, "markers")
            self.assertEqual(section.conflict_type, "both_modified")

            sections = wrapper.get_conflict_sections()
            self.assertEqual(len(sections), 1)
            self.assertEqual(sections[0].file_path, "f.txt")

            payload = wrapper.get_llm_payload()
            self.assertEqual(
                set(payload),
                {"repository", "summary", "files", "conflicts"},
            )
            self.assertEqual(payload["repository"]["operation"], "merge")
            self.assertEqual(payload["summary"]["file_count"], 0)
            self.assertEqual(payload["summary"]["conflict_count"], 1)
            self.assertFalse(payload["summary"]["truncated"])
            self.assertEqual(payload["conflicts"][0]["path"], "f.txt")
            self.assertEqual(payload["conflicts"][0]["sections"][0]["ours"], "main")

    def test_diff3_markers_expose_base_section(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._git(repo, "config", "merge.conflictStyle", "diff3")
            wrapper = self._diverged_file(repo)

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            sections = wrapper.get_conflict_sections()
            self.assertEqual(len(sections), 1)
            self.assertEqual(sections[0].base, "base")
            self.assertEqual(sections[0].ours, "main")
            self.assertEqual(sections[0].theirs, "feature")

    def test_add_add_conflict_has_no_base_content(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "README.md", "base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "new.txt", "feature\n", "feature add")
            self._git(repo, "checkout", "main")
            self._write_and_commit(repo, "new.txt", "main\n", "main add")

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            conflicts = GitWrapper(repo).get_conflict_files()
            self.assertEqual(len(conflicts), 1)
            conflict = conflicts[0]
            self.assertEqual(conflict.path, "new.txt")
            self.assertEqual(conflict.conflict_type, "both_added")
            self.assertIsNone(conflict.base_content)
            self.assertEqual(conflict.ours_content, "main\n")
            self.assertEqual(conflict.theirs_content, "feature\n")
            self.assertTrue(conflict.resolution_supported)

    def test_delete_modify_conflict_exposes_missing_side(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "del.txt", "base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "del.txt", "feature\n", "feature change")
            self._git(repo, "checkout", "main")
            (repo / "del.txt").unlink()
            self._git(repo, "add", "-A")
            self._git(repo, "commit", "-m", "main delete")

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            conflicts = GitWrapper(repo).get_conflict_files()
            self.assertEqual(len(conflicts), 1)
            conflict = conflicts[0]
            self.assertEqual(conflict.path, "del.txt")
            self.assertEqual(conflict.conflict_type, "deleted_by_us")
            self.assertIsNone(conflict.ours_content)
            self.assertEqual(conflict.theirs_content, "feature\n")
            self.assertFalse(conflict.is_binary)

    def test_binary_conflict_is_metadata_only(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_binary_and_commit(repo, "bin.dat", b"\x00base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_binary_and_commit(repo, "bin.dat", b"\x00feature\n", "feature")
            self._git(repo, "checkout", "main")
            self._write_binary_and_commit(repo, "bin.dat", b"\x00main\n", "main")

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            conflicts = GitWrapper(repo).get_conflict_files()
            self.assertEqual(len(conflicts), 1)
            conflict = conflicts[0]
            self.assertEqual(conflict.path, "bin.dat")
            self.assertTrue(conflict.is_binary)
            self.assertFalse(conflict.resolution_supported)
            self.assertEqual(conflict.sections, [])
            self.assertEqual(conflict.reason, "fichier binaire")

    def test_clean_merge_in_progress_exposes_staged_diff(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "a.txt", "base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "a.txt", "feature\n", "feature change")
            self._git(repo, "checkout", "main")
            self._write_and_commit(repo, "b.txt", "main\n", "main change")

            result = self._git(repo, "merge", "--no-ff", "--no-commit", "feature", check=False)
            self.assertEqual(result.returncode, 0)

            wrapper = GitWrapper(repo)
            state = wrapper.get_git_state()
            self.assertEqual(state.operation, "merge")
            self.assertTrue(state.in_progress)
            self.assertFalse(state.has_conflicts)

            files = wrapper.get_staged_diff()
            self.assertEqual(len(files), 1)
            self.assertEqual(files[0].path, "a.txt")
            self.assertIn("+feature", files[0].patch)

            payload = wrapper.get_merge_payload()
            self.assertEqual(payload["summary"]["file_count"], 1)
            self.assertEqual(payload["summary"]["conflict_count"], 0)
            self.assertEqual(payload["conflicts"], [])

    def test_completed_merge_commit_exposes_two_parents(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "a.txt", "base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "a.txt", "feature\n", "feature change")
            self._git(repo, "checkout", "main")
            self._write_and_commit(repo, "b.txt", "main\n", "main change")

            result = self._git(repo, "merge", "--no-ff", "feature", "-m", "merge feature", check=False)
            self.assertEqual(result.returncode, 0)

            state = GitWrapper(repo).get_git_state()
            self.assertEqual(state.operation, "none")
            self.assertFalse(state.in_progress)
            self.assertFalse(state.has_conflicts)
            self.assertTrue(state.is_merge_commit)
            self.assertEqual(len(state.head_parents), 2)

    def test_resolved_conflict_is_staged_and_no_longer_conflicting(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            wrapper = self._diverged_file(repo)

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            (repo / "f.txt").write_text("resolved\n", encoding="utf-8")
            self._git(repo, "add", "f.txt")

            state = wrapper.get_git_state()
            self.assertEqual(state.operation, "merge")
            self.assertTrue(state.in_progress)
            self.assertFalse(state.has_conflicts)
            self.assertEqual(wrapper.get_unmerged_paths(), [])

            files = wrapper.get_staged_diff()
            self.assertEqual(len(files), 1)
            self.assertEqual(files[0].path, "f.txt")
            self.assertIn("+resolved", files[0].patch)

            payload = wrapper.get_merge_payload()
            self.assertEqual(payload["summary"]["conflict_count"], 0)
            self.assertEqual(payload["summary"]["file_count"], 1)

    def test_cherry_pick_conflict_is_detected(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "f.txt", "base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "f.txt", "feature\n", "feature change")
            self._git(repo, "checkout", "main")
            self._write_and_commit(repo, "f.txt", "main\n", "main change")

            result = self._git(repo, "cherry-pick", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            wrapper = GitWrapper(repo)
            state = wrapper.get_git_state()
            self.assertEqual(state.operation, "cherry-pick")
            self.assertTrue(state.in_progress)
            self.assertTrue(state.has_conflicts)
            self.assertEqual(len(wrapper.get_conflict_files()), 1)

    def test_rebase_conflict_is_detected(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "f.txt", "base\n", "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "f.txt", "feature\n", "feature change")
            self._git(repo, "checkout", "main")
            self._write_and_commit(repo, "f.txt", "main\n", "main change")
            self._git(repo, "checkout", "feature")

            result = self._git(repo, "rebase", "main", check=False)
            self.assertNotEqual(result.returncode, 0)

            wrapper = GitWrapper(repo)
            state = wrapper.get_git_state()
            self.assertEqual(state.operation, "rebase")
            self.assertTrue(state.in_progress)
            self.assertTrue(state.has_conflicts)
            self.assertEqual(len(wrapper.get_conflict_files()), 1)

    def test_payload_without_merge_keeps_historical_keys(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "f.txt", "base\n", "base")
            (repo / "f.txt").write_text("changed\n", encoding="utf-8")
            self._git(repo, "add", "f.txt")

            wrapper = GitWrapper(repo)
            payload = wrapper.get_llm_payload()

            self.assertEqual(payload["repository"]["operation"], "none")
            self.assertFalse(payload["repository"]["in_progress"])
            self.assertFalse(payload["repository"]["has_conflicts"])
            self.assertEqual(payload["summary"]["file_count"], 1)
            self.assertEqual(payload["summary"]["conflict_count"], 0)
            self.assertEqual(payload["conflicts"], [])
            self.assertEqual(payload["files"][0]["path"], "f.txt")

    def test_large_conflict_is_truncated(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            self._init_repo(repo)
            self._write_and_commit(repo, "big.txt", "base\n" * 400, "base")

            self._git(repo, "checkout", "-b", "feature")
            self._write_and_commit(repo, "big.txt", "feature\n" * 400, "feature change")
            self._git(repo, "checkout", "main")
            self._write_and_commit(repo, "big.txt", "main\n" * 400, "main change")

            result = self._git(repo, "merge", "feature", check=False)
            self.assertNotEqual(result.returncode, 0)

            wrapper = GitWrapper(repo)
            conflicts = wrapper.get_conflict_files(max_chars=120)
            self.assertEqual(len(conflicts), 1)
            self.assertTrue(conflicts[0].truncated)
            self.assertIn("[TRUNCATED]", conflicts[0].sections[0].ours)

            payload = wrapper.get_merge_payload(max_chars=120, total_max_chars=2000)
            self.assertTrue(payload["summary"]["truncated"])
            self.assertIn("big.txt", payload["summary"]["truncated_files"])
            self.assertLessEqual(len(json.dumps(payload, ensure_ascii=False)), 2000)

    def test_parse_conflict_sections_with_multiple_blocks(self):
        text = """before
<<<<<<< HEAD
ours1
||||||| base
base1
=======
theirs1
>>>>>>> feature
middle
<<<<<<< HEAD
ours2
=======
theirs2
>>>>>>> feature
after
"""

        sections = GitWrapper._parse_conflict_sections("x.py", text, context_lines=1)

        self.assertEqual(len(sections), 2)
        self.assertEqual(sections[0].section_index, 0)
        self.assertEqual(sections[0].start_line, 2)
        self.assertEqual(sections[0].end_line, 8)
        self.assertEqual(sections[0].ours, "ours1")
        self.assertEqual(sections[0].theirs, "theirs1")
        self.assertEqual(sections[0].base, "base1")
        self.assertEqual(sections[0].context_before, "before")
        self.assertEqual(sections[0].context_after, "middle")
        self.assertEqual(sections[1].section_index, 1)
        self.assertEqual(sections[1].start_line, 10)
        self.assertEqual(sections[1].end_line, 14)
        self.assertEqual(sections[1].ours, "ours2")
        self.assertEqual(sections[1].theirs, "theirs2")
        self.assertIsNone(sections[1].base)
        self.assertEqual(sections[1].context_before, "middle")
        self.assertEqual(sections[1].context_after, "after")

    def test_conflict_type_from_stages(self):
        cases = [
            ({1, 2, 3}, "both_modified"),
            ({2, 3}, "both_added"),
            ({1, 2}, "deleted_by_them"),
            ({1, 3}, "deleted_by_us"),
            ({1}, "both_deleted"),
            ({2}, "added_by_us"),
            ({3}, "added_by_them"),
            (set(), "unknown"),
        ]
        for stages, expected in cases:
            with self.subTest(stages=stages):
                self.assertEqual(GitWrapper._conflict_type_from_stages(stages), expected)


if __name__ == "__main__":
    unittest.main()

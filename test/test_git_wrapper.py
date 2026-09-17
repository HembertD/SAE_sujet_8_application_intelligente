import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from git_commit_release_notes_generator.service.git_wrapper import GitWrapper


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


if __name__ == "__main__":
    unittest.main()

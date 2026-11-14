import hashlib
import unittest
from unittest.mock import patch

import update_dependencies as updater


def tag(name, architectures=("amd64", "arm64")):
    return {"name": name, "images": [
        {"os": "linux", "architecture": arch} for arch in architectures]}


class UpdateTests(unittest.TestCase):
    def test_selects_newest_numeric_alpine_across_pages_with_both_platforms(self):
        pages = [
            {"results": [tag("1.27.4-alpine3.9"), tag("1.27.4-alpine3.25", ("amd64",)),
                         tag("1.27.4-alpine"), tag("1.27.3-alpine3.26")], "next": "page2"},
            {"results": [tag("1.27.4-alpine3.24"), tag("1.27.4-alpine3.23")], "next": None},
        ]
        with patch.object(updater, "json_url", side_effect=pages):
            self.assertEqual(updater.latest_alpine("golang", "1.27.4", "3.23"), "3.24")

    def test_refuses_alpine_downgrade_or_missing_architecture(self):
        page = {"results": [tag("3.15.2-alpine3.23"), tag("3.15.2-alpine3.25", ("amd64",))]}
        with patch.object(updater, "json_url", return_value=page):
            with self.assertRaises(ValueError):
                updater.latest_alpine("python", "3.15.2", "3.24")

    def test_checks_downloaded_executable_against_published_checksum(self):
        artifact = b"release executable"
        digest = hashlib.sha256(artifact).hexdigest()
        release = {"tag_name": "2026.10.26", "draft": False, "prerelease": False}
        with patch.object(updater, "json_url", return_value=release):
            with patch.object(updater, "download", side_effect=[f"{digest}  yt-dlp\n".encode(), artifact]):
                self.assertEqual(updater.latest_yt_dlp(), ("2026.10.26", digest))
            with patch.object(updater, "download", side_effect=[f"{digest}  yt-dlp\n".encode(), b"corrupt"]):
                with self.assertRaisesRegex(ValueError, "checksum verification failed"):
                    updater.latest_yt_dlp()

    def test_rejects_prerelease(self):
        release = {"tag_name": "2026.10.26", "draft": False, "prerelease": True}
        with patch.object(updater, "json_url", return_value=release):
            with self.assertRaises(ValueError):
                updater.latest_yt_dlp()

    def test_rejects_missing_or_duplicate_pins(self):
        for text in ("missing", "go 1.27.4\ngo 1.27.3\n"):
            with self.assertRaises(ValueError):
                updater.replace_once(text, r"^go \S+$", "go 1.27.5")


if __name__ == "__main__":
    unittest.main()

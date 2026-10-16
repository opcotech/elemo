#!/usr/bin/env python3
"""Tests for CLA signature branch selection."""

from __future__ import annotations

import importlib.util
import unittest
from pathlib import Path


def load_module():
    path = Path(__file__).with_name("cla-sync-signatures.py")
    spec = importlib.util.spec_from_file_location("cla_sync_signatures", path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"could not load {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CommitTargetTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.mod = load_module()

    def test_signature_branch_name(self) -> None:
        self.assertEqual(self.mod.signature_branch(464), "cla/pr-464")

    def test_same_repo_uses_head_branch(self) -> None:
        branch, is_fork = self.mod.commit_target("opcotech/elemo", "opcotech/elemo", "feature", 464)
        self.assertEqual(branch, "feature")
        self.assertFalse(is_fork)

    def test_fork_uses_signature_branch(self) -> None:
        branch, is_fork = self.mod.commit_target("opcotech/elemo", "alice/elemo", "feature", 464)
        self.assertEqual(branch, "cla/pr-464")
        self.assertTrue(is_fork)

    def test_missing_head_repo_uses_signature_branch(self) -> None:
        branch, is_fork = self.mod.commit_target("opcotech/elemo", None, "feature", 464)
        self.assertEqual(branch, "cla/pr-464")
        self.assertTrue(is_fork)

    def test_read_branch_uses_existing_signature_branch(self) -> None:
        source = self.mod.read_branch(
            is_fork=True, target="cla/pr-464", base_ref="main", target_exists=True
        )
        self.assertEqual(source, "cla/pr-464")

    def test_read_branch_falls_back_to_base_when_signature_branch_is_missing(self) -> None:
        source = self.mod.read_branch(
            is_fork=True, target="cla/pr-464", base_ref="main", target_exists=False
        )
        self.assertEqual(source, "main")

    def test_read_branch_keeps_same_repo_head(self) -> None:
        source = self.mod.read_branch(
            is_fork=False, target="feature", base_ref="main", target_exists=False
        )
        self.assertEqual(source, "feature")

    def test_signature_pr_text_matches_workflow(self) -> None:
        workflow = Path(__file__).resolve().parents[1].joinpath("workflows", "cla.yml").read_text()
        title = self.mod.signature_pr_title(464)
        body = self.mod.signature_pr_body(464)
        self.assertEqual(title, "chore: record CLA signature from #464")
        self.assertIn("chore: record CLA signature from #${PR}", workflow)
        shared = "are stored on this branch because the pull request was opened from a fork."
        self.assertIn(shared, body)
        self.assertIn(shared, workflow)
        self.assertIn("Merging this pull request updates", workflow)


if __name__ == "__main__":
    unittest.main()

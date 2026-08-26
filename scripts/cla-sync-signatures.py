#!/usr/bin/env python3
"""Import CLA signature comments from a pull request into that PR's cla.json.

The CLA GitHub Action only records PR committers. Anyone else who posts the
sign-off comment is ignored. This script finds every matching comment and merges
new signers into .github/cla.json on the pull request head branch.
"""

from __future__ import annotations

import argparse
import base64
import json
import subprocess
import sys
from typing import Any

SIGN_PHRASE = "I have read the CLA Document and I hereby sign the CLA"
DEFAULT_PATH = ".github/cla.json"


def run_gh(*args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["gh", *args],
        check=check,
        capture_output=True,
        text=True,
    )


def gh_json(*args: str) -> Any:
    result = run_gh(*args)
    return json.loads(result.stdout) if result.stdout.strip() else None


def default_repo() -> str:
    data = gh_json("repo", "view", "--json", "nameWithOwner")
    if not data or "nameWithOwner" not in data:
        raise SystemExit("could not determine GitHub repo; pass --repo owner/name")
    return str(data["nameWithOwner"])


def fetch_comments(repo: str, pr: int) -> list[dict[str, Any]]:
    result = run_gh("api", "--paginate", f"repos/{repo}/issues/{pr}/comments")
    comments: list[dict[str, Any]] = []
    decoder = json.JSONDecoder()
    raw = result.stdout
    idx = 0
    while idx < len(raw):
        while idx < len(raw) and raw[idx].isspace():
            idx += 1
        if idx >= len(raw):
            break
        page, offset = decoder.raw_decode(raw, idx)
        if not isinstance(page, list):
            raise SystemExit("unexpected GitHub comments response")
        comments.extend(page)
        idx = offset
    return comments


def is_signature(body: str | None) -> bool:
    return (body or "").strip() == SIGN_PHRASE


def signatures_from_comments(
    comments: list[dict[str, Any]], repo_id: int, pr: int
) -> dict[int, dict[str, Any]]:
    found: dict[int, dict[str, Any]] = {}
    for comment in comments:
        if not is_signature(comment.get("body")):
            continue
        user = comment.get("user") or {}
        user_id = user.get("id")
        login = user.get("login")
        if user_id is None or not login or str(login).endswith("[bot]"):
            continue
        entry = {
            "name": login,
            "id": int(user_id),
            "comment_id": int(comment["id"]),
            "created_at": comment["created_at"],
            "repoId": repo_id,
            "pullRequestNo": pr,
        }
        existing = found.get(int(user_id))
        if existing is None or entry["created_at"] < existing["created_at"]:
            found[int(user_id)] = entry
    return found


def pr_head(repo: str, pr: int) -> str:
    data = gh_json("api", f"repos/{repo}/pulls/{pr}")
    head_repo = data["head"]["repo"]["full_name"]
    if head_repo != repo:
        raise SystemExit(f"PR #{pr} is from {head_repo}; signatures are committed to same-repo PR branches only")
    return str(data["head"]["ref"])


def load_cla(repo: str, branch: str, path: str) -> tuple[list[dict[str, Any]], str | None]:
    result = run_gh(
        "api",
        f"repos/{repo}/contents/{path}?ref={branch}",
        check=False,
    )
    if result.returncode != 0:
        combined = (result.stderr or "") + (result.stdout or "")
        if "Not Found" in combined:
            return [], None
        sys.stderr.write(result.stderr or result.stdout)
        raise SystemExit(result.returncode)
    payload = json.loads(result.stdout)
    raw = base64.b64decode(payload["content"])
    data = json.loads(raw.decode())
    return list(data.get("signedContributors") or []), str(payload["sha"])


def merge_signers(
    existing: list[dict[str, Any]], incoming: dict[int, dict[str, Any]]
) -> tuple[list[dict[str, Any]], list[str]]:
    by_id: dict[int, dict[str, Any]] = {}
    for signer in existing:
        by_id[int(signer["id"])] = signer
    added: list[str] = []
    for user_id, signer in incoming.items():
        if user_id not in by_id:
            by_id[user_id] = signer
            added.append(str(signer["name"]))
    merged = sorted(by_id.values(), key=lambda item: str(item["name"]).lower())
    return merged, added


def push_cla(
    repo: str,
    branch: str,
    path: str,
    sha: str | None,
    signers: list[dict[str, Any]],
    pr: int,
    added: list[str],
) -> None:
    raw = json.dumps({"signedContributors": signers}, indent=2) + "\n"
    content_b64 = base64.b64encode(raw.encode()).decode()
    message = f"chore: import CLA signatures from #{pr} ({', '.join(added)})"
    args = [
        "api",
        "--method",
        "PUT",
        f"repos/{repo}/contents/{path}",
        "-f",
        f"message={message}",
        "-f",
        f"content={content_b64}",
        "-f",
        f"branch={branch}",
    ]
    if sha:
        args.extend(["-f", f"sha={sha}"])
    run_gh(*args)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Merge CLA sign-off comments from a pull request into that PR's cla.json."
    )
    parser.add_argument("pr", type=int, help="Pull request number")
    parser.add_argument("--repo", help="owner/name (default: current gh repo)")
    parser.add_argument("--path", default=DEFAULT_PATH, help=f"Path to cla.json (default: {DEFAULT_PATH})")
    parser.add_argument(
        "--push",
        action="store_true",
        help="Commit the merged file onto the pull request branch",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo = args.repo or default_repo()
    branch = pr_head(repo, args.pr)
    repo_info = gh_json("api", f"repos/{repo}")
    repo_id = int(repo_info["id"])

    comments = fetch_comments(repo, args.pr)
    incoming = signatures_from_comments(comments, repo_id, args.pr)
    existing, sha = load_cla(repo, branch, args.path)
    merged, added = merge_signers(existing, incoming)

    print(f"PR #{args.pr} ({branch}): {len(incoming)} matching signature comment(s)", file=sys.stderr)
    print(f"{args.path}: {len(existing)} existing signer(s)", file=sys.stderr)
    if not added:
        print("no new signers to add", file=sys.stderr)
        if not args.push:
            json.dump({"signedContributors": merged}, sys.stdout, indent=2)
            sys.stdout.write("\n")
        return 0

    print("new signers: " + ", ".join(added), file=sys.stderr)
    if not args.push:
        json.dump({"signedContributors": merged}, sys.stdout, indent=2)
        sys.stdout.write("\n")
        print("re-run with --push to commit onto the PR branch", file=sys.stderr)
        return 0

    push_cla(repo, branch, args.path, sha, merged, args.pr, added)
    print(f"updated {repo}@{branch}:{args.path}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

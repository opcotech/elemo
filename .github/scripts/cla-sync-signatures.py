#!/usr/bin/env python3
"""Import CLA signature comments from a pull request into .github/cla.json.

The CLA GitHub Action only records PR committers. Anyone else who posts the
sign-off comment is ignored. This script finds every matching comment and merges
new signers into .github/cla.json.

Same-repository pull requests are updated on the head branch. Fork pull requests
are updated on ``cla/pr-<number>`` in the base repository, and ``--push`` opens a
pull request to merge that branch.
"""

from __future__ import annotations

import argparse
import base64
import json
import subprocess
import sys
from typing import Any
from urllib.parse import quote

SIGN_PHRASE = "I have read the CLA Document and I hereby sign the CLA"
DEFAULT_PATH = ".github/cla.json"
SIGNATURE_PR_BODY = (
    "Signatures for #{pr} are stored on this branch because the pull request was opened from a fork. "
    "Merging this pull request updates `.github/cla.json` on the base branch."
)


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


def signature_branch(pr: int) -> str:
    """Branch in the base repository that stores a fork pull request's signatures."""
    return f"cla/pr-{pr}"


def commit_target(
    base_repo: str, head_repo: str | None, head_ref: str, pr: int
) -> tuple[str, bool]:
    """Return the branch that should receive signature commits, and whether the PR is a fork.

    The commit always lands in the base repository. A same-repository pull request
    uses its head branch. A fork, including one whose head repository was deleted,
    uses ``cla/pr-<number>``.
    """
    if head_repo == base_repo:
        return head_ref, False
    return signature_branch(pr), True


def read_branch(*, is_fork: bool, target: str, base_ref: str, target_exists: bool) -> str:
    """Return the branch whose cla.json should be read before merging signers."""
    if not is_fork or target_exists:
        return target
    return base_ref


def signature_pr_title(pr: int) -> str:
    """Title for the pull request that merges a fork's CLA signatures."""
    return f"chore: record CLA signature from #{pr}"


def signature_pr_body(pr: int) -> str:
    """Body for the pull request that merges a fork's CLA signatures."""
    return SIGNATURE_PR_BODY.format(pr=pr)


def pull_request_refs(repo: str, pr: int) -> tuple[str | None, str, str]:
    """Return ``(head_repo, head_ref, base_ref)``. ``head_repo`` is None when the fork is gone."""
    data = gh_json("api", f"repos/{repo}/pulls/{pr}")
    if not data:
        raise SystemExit(f"could not load pull request #{pr}")
    head = data["head"]
    head_repo_obj = head.get("repo") or {}
    full_name = head_repo_obj.get("full_name")
    head_repo = str(full_name) if full_name else None
    return head_repo, str(head["ref"]), str(data["base"]["ref"])


def ref_exists(repo: str, branch: str) -> bool:
    result = run_gh("api", f"repos/{repo}/git/ref/heads/{branch}", check=False)
    if result.returncode == 0:
        return True
    combined = (result.stderr or "") + (result.stdout or "")
    if "Not Found" in combined:
        return False
    sys.stderr.write(result.stderr or result.stdout)
    raise SystemExit(result.returncode)


def ensure_signature_branch(repo: str, branch: str, base_ref: str) -> None:
    """Create ``branch`` from ``base_ref`` when the signature branch is missing."""
    if ref_exists(repo, branch):
        return
    base = gh_json("api", f"repos/{repo}/git/ref/heads/{base_ref}")
    if not base or "object" not in base:
        raise SystemExit(f"could not resolve {repo}@{base_ref}")
    run_gh(
        "api",
        "--method",
        "POST",
        f"repos/{repo}/git/refs",
        "-f",
        f"ref=refs/heads/{branch}",
        "-f",
        f"sha={base['object']['sha']}",
    )


def ahead_by(repo: str, base: str, head: str) -> int:
    base_q = quote(base, safe="")
    head_q = quote(head, safe="")
    result = run_gh("api", f"repos/{repo}/compare/{base_q}...{head_q}", "--jq", ".ahead_by")
    raw = result.stdout.strip()
    if not raw or raw == "null":
        return 0
    return int(raw)


def find_open_signature_pr(repo: str, base: str, branch: str) -> int | None:
    """Return the number of an open signature pull request, if one exists."""
    owner = repo.split("/", 1)[0]
    base_q = quote(base, safe="")
    head_q = quote(branch, safe="")
    result = run_gh(
        "api",
        f"repos/{repo}/pulls?state=open&per_page=100&base={base_q}&head={owner}:{head_q}",
    )
    pulls = json.loads(result.stdout) if result.stdout.strip() else []
    if not isinstance(pulls, list):
        raise SystemExit("unexpected GitHub pulls response")
    for pull in pulls:
        head_ref = str((pull.get("head") or {}).get("ref") or "")
        base_ref = str((pull.get("base") or {}).get("ref") or "")
        if head_ref == branch and base_ref == base:
            return int(pull["number"])
    return None


def open_signature_pr(repo: str, base: str, branch: str, pr: int) -> None:
    """Open a pull request for a fork signature branch when one is not already open."""
    if ahead_by(repo, base, branch) == 0:
        print(f"{branch} is not ahead of {base}; no pull request opened", file=sys.stderr)
        return
    existing = find_open_signature_pr(repo, base, branch)
    if existing is not None:
        print(f"signature pull request already open: #{existing}", file=sys.stderr)
        return
    result = run_gh(
        "pr",
        "create",
        "--repo",
        repo,
        "--base",
        base,
        "--head",
        branch,
        "--title",
        signature_pr_title(pr),
        "--body",
        signature_pr_body(pr),
    )
    print(result.stdout.strip(), file=sys.stderr)


def load_cla(repo: str, branch: str, path: str) -> tuple[list[dict[str, Any]], str | None]:
    ref = quote(branch, safe="")
    result = run_gh(
        "api",
        f"repos/{repo}/contents/{path}?ref={ref}",
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
        description="Merge CLA sign-off comments from a pull request into .github/cla.json."
    )
    parser.add_argument("pr", type=int, help="Pull request number")
    parser.add_argument("--repo", help="owner/name (default: current gh repo)")
    parser.add_argument("--path", default=DEFAULT_PATH, help=f"Path to cla.json (default: {DEFAULT_PATH})")
    parser.add_argument(
        "--push",
        action="store_true",
        help="Commit the merged file. Fork pull requests also open a signature pull request.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo = args.repo or default_repo()
    head_repo, head_ref, base_ref = pull_request_refs(repo, args.pr)
    branch, is_fork = commit_target(repo, head_repo, head_ref, args.pr)
    target_exists = False
    if is_fork:
        if args.push:
            ensure_signature_branch(repo, branch, base_ref)
            target_exists = True
        else:
            target_exists = ref_exists(repo, branch)
    source = read_branch(is_fork=is_fork, target=branch, base_ref=base_ref, target_exists=target_exists)
    repo_info = gh_json("api", f"repos/{repo}")
    repo_id = int(repo_info["id"])

    comments = fetch_comments(repo, args.pr)
    incoming = signatures_from_comments(comments, repo_id, args.pr)
    existing, sha = load_cla(repo, source, args.path)
    merged, added = merge_signers(existing, incoming)

    print(f"PR #{args.pr} ({source}): {len(incoming)} matching signature comment(s)", file=sys.stderr)
    print(f"{args.path}: {len(existing)} existing signer(s)", file=sys.stderr)
    if is_fork and source != branch:
        print(f"signature branch {branch} does not exist yet; showing signatures from {source}", file=sys.stderr)
    if not added:
        print("no new signers to add", file=sys.stderr)
        if not args.push:
            json.dump({"signedContributors": merged}, sys.stdout, indent=2)
            sys.stdout.write("\n")
            return 0
        if is_fork:
            open_signature_pr(repo, base_ref, branch, args.pr)
        return 0

    print("new signers: " + ", ".join(added), file=sys.stderr)
    if not args.push:
        json.dump({"signedContributors": merged}, sys.stdout, indent=2)
        sys.stdout.write("\n")
        if is_fork:
            print(f"re-run with --push to commit onto {branch} and open a pull request", file=sys.stderr)
        else:
            print("re-run with --push to commit onto the PR branch", file=sys.stderr)
        return 0

    if source != branch:
        raise SystemExit(f"refusing to commit {source} contents onto {branch}")
    push_cla(repo, branch, args.path, sha, merged, args.pr, added)
    print(f"updated {repo}@{branch}:{args.path}", file=sys.stderr)
    if is_fork:
        open_signature_pr(repo, base_ref, branch, args.pr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

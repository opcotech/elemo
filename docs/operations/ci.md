# CI

Helper configuration for CI/CD.

## Validation and deployment flow

The main [CI workflow](../../.github/workflows/ci.yml) uses
[shared path filters](../../.github/config/path-filters.yml) to run only the affected backend, web, website, and
end-to-end checks.

1. Change detection and pre-commit checks run first.
2. Applicable backend and frontend/website lint jobs must pass before the quality gate allows tests or builds to run.
3. Backend unit tests and benchmarks run in parallel; integration tests run only after both pass. Frontend unit tests
   gate the frontend production and Storybook builds.
4. The E2E image and nine Playwright browser/shard jobs run only after their applicable prerequisites pass. The matrix
   cancels remaining jobs after the first failure.
5. The validation gate accepts path-filtered skips but fails if any selected check failed or was cancelled.

On pushes to `main`, production API-documentation and website deployments, plus the reusable
[Release Please workflow](../../.github/workflows/release-please.yml), start only after the validation gate succeeds.
Deployments are selected by their own narrow path filters and share the
[`deploy-vercel`](../../.github/actions/deploy-vercel/action.yml) composite action. Manual Release Please dispatch
remains available as an explicit operator override.

## License policy

License and dependency policy runs in [`.github/workflows/license-check.yml`](../../.github/workflows/license-check.yml)
via [`scripts/ort/ort.sh`](../../scripts/ort/ort.sh). Pull requests run `./scripts/ort/ort.sh pr` (analyzer + evaluator only).
Pushes to `main`, the nightly schedule, and manual `workflow_dispatch` runs run `./scripts/ort/ort.sh run` (ScanCode, advisor, reports).

- **Image:** `ghcr.io/oss-review-toolkit/ort-minimal:92.4.0` (digest-pinned in `scripts/ort/ort.sh`; includes ScanCode)
- **Go licenses:** `go-licenses` on the host module cache that ORT reuses inside the container
- **npm licenses:** declared in `package.json`
- **PR gate:** declared/concluded dependency licenses only. Copyleft-in-source findings wait for the full pipeline
- **ScanCode:** Elemo project source only (`--package-types PROJECT`) on `main` / nightly / manual / release. Full
  package ScanCode is `ORT_SCAN_PACKAGE_TYPES=PACKAGE,PROJECT ./scripts/ort/ort.sh scan`
- **Policy:** evaluate against Apache-2.0 (the future license of FSL-1.1-ALv2). Fails on OSADL-incompatible inbound
  licenses and commercial, proprietary-free, unknown, unstated, or missing licenses. Full runs also fail on copyleft
  in project source
- **Not a required check:** advisor (OSV) findings are reported on full runs but do not fail CI
- **Artifacts:** `.ort/results/` (evaluation JSON on PRs; SPDX, CycloneDX, WebApp HTML, NOTICE on full runs) and
  `.ort/results/legal/` (`LICENSE`, `LICENSE-COMMERCIAL`, `NOTICE`, SBOMs) on full runs
- **Releases:** [`.github/workflows/release-please.yml`](../../.github/workflows/release-please.yml) is called after
  successful main-branch validation, runs ORT on a new tag, and attaches `.ort/results/legal/` to the draft GitHub
  Release. NOTICE/SBOMs are generated, not committed.
- **Cache:** Mise Go toolchain, `.ort/bin` (`go-licenses`), and `.ort/scanner` on full runs (per-SHA save)

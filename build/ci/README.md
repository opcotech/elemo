# CI

Helper configuration for CI/CD.

License and dependency policy runs in [`.github/workflows/license-check.yml`](../../.github/workflows/license-check.yml)
via [`scripts/ort/ort.sh`](../../scripts/ort/ort.sh). Pull requests run `make ort.pr` (analyzer + evaluator only).
Pushes to `main`, the nightly schedule, and manual `workflow_dispatch` runs run `make ort` (ScanCode, advisor, reports).

- **Image:** `ghcr.io/oss-review-toolkit/ort-minimal:92.4.0` (digest-pinned in `scripts/ort/ort.sh`; includes ScanCode)
- **Go licenses:** `go-licenses` on the host module cache that ORT reuses inside the container
- **npm licenses:** declared in `package.json`
- **PR gate:** declared/concluded dependency licenses only. Copyleft-in-source findings wait for the full pipeline
- **ScanCode:** Elemo project source only (`--package-types PROJECT`) on `main` / nightly / manual / release. Full
  package ScanCode is `make ort.scan.packages`
- **Policy:** evaluate against Apache-2.0 (the future license of FSL-1.1-ALv2). Fails on OSADL-incompatible inbound
  licenses and commercial, proprietary-free, unknown, unstated, or missing licenses. Full runs also fail on copyleft
  in project source
- **Not a required check:** advisor (OSV) findings are reported on full runs but do not fail CI
- **Artifacts:** `.ort/results/` (evaluation JSON on PRs; SPDX, CycloneDX, WebApp HTML, NOTICE on full runs) and
  `.ort/results/legal/` (`LICENSE`, `LICENSE-COMMERCIAL`, `NOTICE`, SBOMs) on full runs
- **Releases:** [`.github/workflows/release-please.yml`](../../.github/workflows/release-please.yml) runs ORT on the
  new tag and attaches `.ort/results/legal/` to the draft GitHub Release. NOTICE/SBOMs are generated, not committed.
- **Cache:** host Go module cache (setup-go), `.ort/bin` (`go-licenses`), and `.ort/scanner` on full runs (per-SHA save)

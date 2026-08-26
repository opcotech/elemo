# CI

Helper configuration for CI/CD.

License and dependency policy runs in [`.github/workflows/license-check.yml`](../../.github/workflows/license-check.yml)
via `make ort` ([`scripts/ort/ort.sh`](../../scripts/ort/ort.sh)). Target CI time is analyzer-scale (tens of minutes), not a
full ScanCode of every dependency.

- **Image:** `ghcr.io/oss-review-toolkit/ort:92.4.0` (pinned in `scripts/ort/ort.sh`; includes ScanCode)
- **Go licenses:** `go-licenses` on the module cache (declared licenses are missing from `go.mod`)
- **npm licenses:** declared in `package.json`
- **ScanCode:** Elemo project source only (`--package-types PROJECT`). Full package ScanCode is `make ort.scan.packages`
- **Policy:** evaluate against Apache-2.0 (the future license of FSL-1.1-ALv2). Fails on copyleft in project source,
  OSADL-incompatible inbound licenses, and commercial, proprietary-free, unknown, unstated, or missing licenses
- **Not a required check:** advisor (OSV) findings are reported but do not fail CI
- **Artifacts:** `.ort/results/` (SPDX, CycloneDX, WebApp HTML, evaluation JSON, NOTICE) and `.ort/results/legal/`
  (`LICENSE`, `LICENSE-COMMERCIAL`, `NOTICE`, SBOMs), uploaded from the workflow
- **Releases:** [`.github/workflows/release-please.yml`](../../.github/workflows/release-please.yml) runs ORT on the
  new tag and attaches `.ort/results/legal/` to the draft GitHub Release. NOTICE/SBOMs are generated, not committed.
- **Cache:** `.ort/scanner` (project ScanCode reuse)

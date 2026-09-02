#!/usr/bin/env python3
"""Map go-licenses CSV + `go list -m all` into ORT package curations YAML.

Emits versionless Go package IDs (`Go::module:`) so ORT matches analyzer versions
even when go-licenses saw a different module version. Modules missing from the CSV
are filled from LICENSE files in GOMODCACHE when possible.
"""

from __future__ import annotations

import csv
import os
import re
import sys
from collections import defaultdict
from pathlib import Path

UNKNOWN = {"unknown", "none", ""}
PROJECT_PREFIXES = (
    "github.com/opcotech/elemo",
    "github.com/opcotech/elemo-openapi-assemble",
    "github.com/opcotech/elemo-pre-mailer",
)

LICENSE_FILE_NAMES = (
    "LICENSE",
    "LICENSE.txt",
    "LICENSE.md",
    "COPYING",
    "COPYING.txt",
    "LICENCE",
    "LICENCE.txt",
)

# Order matters: more specific patterns first.
LICENSE_PATTERNS: list[tuple[re.Pattern[str], str]] = [
    (re.compile(r"Apache License\s+Version 2\.0", re.I), "Apache-2.0"),
    (re.compile(r"Mozilla Public License\s+Version 2\.0", re.I), "MPL-2.0"),
    (re.compile(r"\bGNU LESSER GENERAL PUBLIC LICENSE\b.*Version 3", re.I | re.S), "LGPL-3.0-only"),
    (re.compile(r"\bGNU GENERAL PUBLIC LICENSE\b.*Version 3", re.I | re.S), "GPL-3.0-only"),
    (re.compile(r"\bGNU GENERAL PUBLIC LICENSE\b.*Version 2", re.I | re.S), "GPL-2.0-only"),
    (re.compile(r"Permission is hereby granted, free of charge, to any person obtaining a copy", re.I), "MIT"),
    # Go Authors BSD often breaks "are met:" across lines.
    (
        re.compile(
            r"Redistribution and use in source and binary forms, with or without\s+"
            r"modification, are permitted provided that the following conditions are\s+met:",
            re.I,
        ),
        "BSD-3-Clause",
    ),
    (re.compile(r"Boost Software License\s+-\s+Version 1\.0", re.I), "BSL-1.0"),
    (re.compile(r"ISC License", re.I), "ISC"),
    (re.compile(r"Creative Commons Legal Code\s+CC0 1\.0", re.I), "CC0-1.0"),
    (re.compile(r"\bUnlicense\b", re.I), "Unlicense"),
]


def ort_version(mod_version: str) -> str:
    if re.fullmatch(r"v\d+.*", mod_version):
        return mod_version[1:]
    return mod_version


def cache_version(mod_version: str) -> str:
    """Module cache directory uses the go.mod version string (usually with v)."""
    if mod_version.startswith("v") or mod_version.startswith("0.0.0-"):
        return mod_version
    if re.fullmatch(r"\d+.*", mod_version):
        return f"v{mod_version}"
    return mod_version


def encode_module_path(module_path: str) -> str:
    """Encode a module path the way the Go module cache does (! before uppercase)."""
    return "".join(f"!{c.lower()}" if c.isupper() else c for c in module_path)


def load_modules(path: Path) -> list[tuple[str, str]]:
    """Return (module_path, go_list_version) with the original go list version string."""
    modules: list[tuple[str, str]] = []
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split()
        if len(parts) < 2:
            continue
        modules.append((parts[0], parts[1]))
    modules.sort(key=lambda item: len(item[0]), reverse=True)
    return modules


def module_for(import_path: str, modules: list[tuple[str, str]]) -> tuple[str, str] | None:
    for path, version in modules:
        if import_path == path or import_path.startswith(path + "/"):
            return path, version
    return None


def is_project(module_path: str) -> bool:
    return any(module_path == p or module_path.startswith(p + "/") for p in PROJECT_PREFIXES)


def detect_license_text(text: str) -> str | None:
    for pattern, spdx in LICENSE_PATTERNS:
        if pattern.search(text):
            return spdx
    return None


def license_from_modcache(gomodcache: Path, module_path: str, version: str) -> str | None:
    encoded = encode_module_path(module_path)
    candidates = [
        gomodcache / f"{encoded}@{cache_version(version)}",
        gomodcache / f"{encoded}@v{ort_version(version)}",
        gomodcache / f"{encoded}@{version}",
    ]
    for directory in candidates:
        if not directory.is_dir():
            continue
        for name in LICENSE_FILE_NAMES:
            license_file = directory / name
            if license_file.is_file():
                detected = detect_license_text(license_file.read_text(errors="ignore"))
                if detected:
                    return detected
        # Some modules keep the license under a docs path; prefer root only.
    return None


def main() -> int:
    if len(sys.argv) != 4:
        print("usage: go-licenses-curations.py MODULES.txt LICENSES.csv OUT.yml", file=sys.stderr)
        return 2

    modules = load_modules(Path(sys.argv[1]))
    # Key by module path only so one curation covers every ORT version.
    licenses_by_path: dict[str, set[str]] = defaultdict(set)
    sources_by_path: dict[str, str] = {}

    with Path(sys.argv[2]).open(newline="") as fh:
        reader = csv.reader(fh)
        for row in reader:
            if len(row) < 3:
                continue
            import_path, _url, license_name = row[0].strip(), row[1].strip(), row[2].strip()
            if license_name.lower() in UNKNOWN:
                continue
            found = module_for(import_path, modules)
            if found is None or is_project(found[0]):
                continue
            licenses_by_path[found[0]].add(license_name)
            sources_by_path.setdefault(found[0], "go-licenses")

    gomodcache = Path(os.environ.get("GOMODCACHE") or "")
    if gomodcache.is_dir():
        for module_path, version in modules:
            if is_project(module_path) or module_path in licenses_by_path:
                continue
            detected = license_from_modcache(gomodcache, module_path, version)
            if detected:
                licenses_by_path[module_path].add(detected)
                sources_by_path[module_path] = "module-cache LICENSE"

    lines = [
        "# Generated by scripts/ort/go-licenses-curations.py. Do not edit.",
        "# Versionless IDs match every ORT Go package version for the module path.",
        "",
    ]
    for module_path, names in sorted(licenses_by_path.items()):
        expression = " AND ".join(sorted(names))
        source = sources_by_path.get(module_path, "go-licenses")
        pkg_id = f"Go::{module_path}:"
        lines.extend(
            [
                f"- id: \"{pkg_id}\"",
                "  curations:",
                f"    comment: \"Concluded from {source}.\"",
                f"    concluded_license: \"{expression}\"",
            ]
        )

    Path(sys.argv[3]).write_text("\n".join(lines) + ("\n" if lines else ""))
    print(f"wrote {len(licenses_by_path)} Go package curations to {sys.argv[3]}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

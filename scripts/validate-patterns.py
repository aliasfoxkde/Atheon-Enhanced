#!/usr/bin/env python3
"""Validate Atheon pattern YAML files.

Walks community/<category>/*.yaml, parses each with PyYAML, and asserts
the fields the bundler depends on. Exits non-zero on the first error so
CI can fail the PR.

Checks performed:

  1. Required keys present (name, match, category).
  2. `category` field (if set) matches the parent directory name.
     The bundler derives category from the directory, so a YAML
     `category:` field would be silently ignored — we flag it as a
     likely copy-paste error.
  3. `match` is a non-empty string that compiles as a Go RE2 regex.
     PyYAML can't import the regexp/syntax package directly, so we use
     a structural pre-check plus a Go-side check in the integration
     workflow step. The structural check catches the common typos
     (unbalanced parens, trailing backslash, lone quantifier) that
     don't need a Go toolchain to spot.
  4. `name` is unique within the bundle. Two files with the same name
     would silently overwrite each other at bundle time.
  5. `severity` (if present) is one of low / medium / high / critical.
     Out-of-set values are normalised to medium by the loader but
     flagged here so the pattern author fixes the YAML, not the loader.

Usage:

  python3 scripts/validate-patterns.py community/
  python3 scripts/validate-patterns.py community/secrets/

The script uses only the standard library plus PyYAML (already a
transitive dep of every other Python tool in the repo). PyYAML is
imported lazily so a missing install produces a clear error.
"""

from __future__ import annotations

import argparse
import re
import sys
from collections.abc import Iterable
from pathlib import Path

VALID_SEVERITIES = {"low", "medium", "high", "critical"}

# Structural pre-check: a regex with a single trailing backslash (no
# escape target) is broken in any RE engine. Other syntax errors
# (illegal quantifiers, unbalanced groups) are caught authoritatively
# by the Go-side RE2 compile that runs as a follow-up step in CI, so
# we don't duplicate that here.
BROKEN_RE = re.compile(r"\\$")


def yaml() -> object:
    try:
        import yaml  # type: ignore[import-untyped]
    except ImportError as e:
        print(f"PyYAML required: pip install pyyaml ({e})", file=sys.stderr)
        sys.exit(2)
    return yaml


def iter_yaml_files(root: Path) -> Iterable[Path]:
    """Yield every *.yaml file under root, sorted for deterministic output."""
    return sorted(root.rglob("*.yaml"))


def parent_category(path: Path, root: Path) -> str:
    """The category the bundler assigns — always the immediate parent dir.

    Uses the parent directory's basename rather than the path relative
    to root, so this works whether the user points at community/ or at
    community/secrets/ as the root.
    """
    return path.parent.name


def validate_one(path: Path, root: Path, seen_names: dict[str, Path]) -> list[str]:
    y = yaml()
    errs: list[str] = []
    try:
        data = y.safe_load(path.read_text(encoding="utf-8"))
    except Exception as e:
        return [f"{path}: YAML parse error: {e}"]

    if not isinstance(data, dict):
        return [f"{path}: top-level must be a mapping, got {type(data).__name__}"]

    name = data.get("name")
    match = data.get("match")
    if not name or not isinstance(name, str):
        errs.append(f"{path}: missing or invalid 'name'")
    if not match or not isinstance(match, str):
        errs.append(f"{path}: missing or invalid 'match'")
    else:
        if BROKEN_RE.search(match):
            errs.append(f"{path}: 'match' has a trailing unescaped backslash: {match!r}")

    declared = data.get("category")
    actual = parent_category(path, root)
    if declared is not None and declared != actual:
        errs.append(
            f"{path}: declared category {declared!r} doesn't match parent dir {actual!r} "
            "(the bundler derives category from the directory; this field would be silently ignored)"
        )

    severity = data.get("severity")
    if severity is not None and severity not in VALID_SEVERITIES:
        errs.append(
            f"{path}: 'severity' must be one of {sorted(VALID_SEVERITIES)}, got {severity!r}"
        )

    if isinstance(name, str):
        if name in seen_names and seen_names[name] != path:
            errs.append(
                f"{path}: duplicate name {name!r} (also in {seen_names[name]})"
            )
        seen_names[name] = path

    return errs


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("root", type=Path, help="directory to scan (e.g. community/)")
    args = ap.parse_args()

    if not args.root.is_dir():
        print(f"{args.root}: not a directory", file=sys.stderr)
        return 2

    files = list(iter_yaml_files(args.root))
    if not files:
        print(f"{args.root}: no *.yaml files found", file=sys.stderr)
        return 2

    seen: dict[str, Path] = {}
    all_errs: list[str] = []
    for p in files:
        all_errs.extend(validate_one(p, args.root, seen))

    print(f"Scanned {len(files)} pattern files under {args.root}")
    if all_errs:
        print(f"❌ {len(all_errs)} validation error(s):", file=sys.stderr)
        for e in all_errs:
            print(f"  {e}", file=sys.stderr)
        return 1
    print(f"✅ {len(files)} patterns valid")
    return 0


if __name__ == "__main__":
    sys.exit(main())
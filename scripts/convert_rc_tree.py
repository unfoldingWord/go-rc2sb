#!/usr/bin/env python3
"""Batch-convert RC repositories into SB directories using the rc2sb binary."""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


def parse_args() -> argparse.Namespace:
    script_path = Path(__file__).resolve()
    default_bin = script_path.parent.parent / "rc2sb"

    parser = argparse.ArgumentParser(
        description=(
            "Convert RC trees laid out as <input>/<owner>/<repo> into "
            "<output>/<owner>/<repo> using rc2sb."
        )
    )
    parser.add_argument("--input", required=True, help="RC root directory")
    parser.add_argument("--output", required=True, help="SB output root directory")
    parser.add_argument(
        "--owner",
        default="",
        help="Optional owner name to process only one owner directory",
    )
    parser.add_argument(
        "--bin",
        default=str(default_bin),
        help=f"Path to rc2sb binary (default: {default_bin})",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print commands without running conversions",
    )
    parser.add_argument(
        "--fail-fast",
        action="store_true",
        help="Stop at first conversion failure",
    )
    return parser.parse_args()


def list_targets(input_root: Path, owner_filter: str) -> list[tuple[str, str]]:
    targets: list[tuple[str, str]] = []

    if owner_filter:
        owner_dir = input_root / owner_filter
        if not owner_dir.is_dir():
            raise FileNotFoundError(f"Owner directory not found: {owner_dir}")
        owners = [owner_dir]
    else:
        owners = sorted(p for p in input_root.iterdir() if p.is_dir())

    for owner_dir in owners:
        repos = sorted(p for p in owner_dir.iterdir() if p.is_dir())
        for repo_dir in repos:
            targets.append((owner_dir.name, repo_dir.name))

    return targets


def choose_usfm_dir(owner_dir: Path, repo_name: str) -> Path | None:
    for suffix in ("_tq", "_twl", "_tn"):
        if repo_name.endswith(suffix):
            lang = repo_name[: -len(suffix)]
            for usfm_suffix in ("ult", "ust", "glt", "gst"):
                candidate = owner_dir / f"{lang}_{usfm_suffix}"
                if candidate.is_dir():
                    return candidate
            return None
    return None


def choose_payload_dir(owner_dir: Path, repo_name: str) -> Path | None:
    if not repo_name.endswith("_twl"):
        return None
    lang = repo_name[: -len("_twl")]
    candidate = owner_dir / f"{lang}_tw"
    if candidate.is_dir():
        return candidate
    return None


def build_command(
    binary_path: Path, rc_repo_dir: Path, sb_repo_dir: Path, owner_dir: Path, repo_name: str
) -> list[str]:
    cmd = [str(binary_path)]

    payload_dir = choose_payload_dir(owner_dir, repo_name)
    if payload_dir is not None:
        cmd.extend(["--payload", str(payload_dir)])

    usfm_dir = choose_usfm_dir(owner_dir, repo_name)
    if usfm_dir is not None:
        cmd.extend(["--usfm", str(usfm_dir)])

    cmd.extend([str(rc_repo_dir), str(sb_repo_dir)])
    return cmd


def main() -> int:
    args = parse_args()

    input_root = Path(args.input).resolve()
    output_root = Path(args.output).resolve()
    binary_path = Path(args.bin).resolve()

    if not input_root.is_dir():
        print(f"Input directory not found: {input_root}", file=sys.stderr)
        return 1
    if not binary_path.is_file():
        print(f"rc2sb binary not found: {binary_path}", file=sys.stderr)
        return 1

    output_root.mkdir(parents=True, exist_ok=True)

    try:
        targets = list_targets(input_root, args.owner)
    except FileNotFoundError as err:
        print(str(err), file=sys.stderr)
        return 1

    total = len(targets)
    converted = 0
    failed = 0

    for idx, (owner, repo) in enumerate(targets, start=1):
        owner_dir = input_root / owner
        rc_repo_dir = owner_dir / repo
        sb_repo_dir = output_root / owner / repo
        sb_repo_dir.parent.mkdir(parents=True, exist_ok=True)

        cmd = build_command(binary_path, rc_repo_dir, sb_repo_dir, owner_dir, repo)
        print(f"[{idx}/{total}] {' '.join(cmd)}", flush=True)

        if args.dry_run:
            converted += 1
            continue

        result = subprocess.run(cmd, check=False)
        if result.returncode == 0:
            converted += 1
            continue

        failed += 1
        print(
            f"[{idx}/{total}] FAILED ({result.returncode}) owner={owner} repo={repo}",
            file=sys.stderr,
        )
        if args.fail_fast:
            break

    print(f"Done. total={total} converted={converted} failed={failed}")
    return 1 if failed > 0 else 0


if __name__ == "__main__":
    raise SystemExit(main())

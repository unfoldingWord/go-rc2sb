#!/usr/bin/env python3
"""Fetch tc-ready RC repos from Door43 catalog and extract zipballs by owner."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import sys
import tempfile
import urllib.error
import urllib.request
import zipfile
from pathlib import Path
from typing import Any


DEFAULT_API_URL = "https://git.door43.org/api/v1/catalog/search?topic=tc-ready"
USER_AGENT = "go-rc2sb-tc-ready-fetch/0.1"


def fetch_catalog_entries(api_url: str, timeout: float) -> list[dict[str, Any]]:
    req = urllib.request.Request(
        api_url,
        headers={"Accept": "application/json", "User-Agent": USER_AGENT},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        charset = resp.headers.get_content_charset() or "utf-8"
        raw = resp.read().decode(charset)

    payload = json.loads(raw)
    if not isinstance(payload, dict):
        raise ValueError("catalog response is not a JSON object")

    data = payload.get("data")
    if not isinstance(data, list):
        raise ValueError("catalog response is missing a list 'data' property")

    return [entry for entry in data if isinstance(entry, dict)]


def download_zip_to_temp(zip_url: str, timeout: float) -> Path:
    req = urllib.request.Request(zip_url, headers={"User-Agent": USER_AGENT})
    fd, tmp_path = tempfile.mkstemp(prefix="tc-ready-", suffix=".zip")
    os.close(fd)
    path = Path(tmp_path)

    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp, path.open("wb") as out:
            shutil.copyfileobj(resp, out)
        return path
    except Exception:
        path.unlink(missing_ok=True)
        raise


def extract_zip(zip_path: Path, dest_dir: Path) -> None:
    with zipfile.ZipFile(zip_path) as zf:
        zf.extractall(dest_dir)


def resolve_owner_name(entry: dict[str, Any]) -> str | None:
    owner = entry.get("owner")
    if isinstance(owner, str) and owner:
        return owner
    if isinstance(owner, dict):
        for key in ("name", "username", "login"):
            value = owner.get(key)
            if isinstance(value, str) and value:
                return value

    repo = entry.get("repo")
    if isinstance(repo, dict):
        repo_owner = repo.get("owner")
        if isinstance(repo_owner, dict):
            for key in ("username", "login", "name"):
                value = repo_owner.get(key)
                if isinstance(value, str) and value:
                    return value

    full_name = entry.get("full_name")
    if isinstance(full_name, str) and "/" in full_name:
        return full_name.split("/", 1)[0]
    return None


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Fetch tc-ready catalog items and extract each zipball under "
            "OUTPUT/<owner>/"
        )
    )
    parser.add_argument("--api-url", default=DEFAULT_API_URL, help="Catalog API URL")
    parser.add_argument(
        "--output",
        default="rc",
        help="Output root directory (default: rc)",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=0,
        help="Optional max number of entries to process (0 = all)",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=60.0,
        help="HTTP timeout in seconds",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    out_root = Path(args.output)
    out_root.mkdir(parents=True, exist_ok=True)

    try:
        entries = fetch_catalog_entries(args.api_url, args.timeout)
    except Exception as err:
        print(f"Failed to fetch catalog: {err}", file=sys.stderr)
        return 1

    if args.limit and args.limit > 0:
        entries = entries[: args.limit]

    total = len(entries)
    downloaded = 0
    skipped = 0
    failed = 0

    for idx, entry in enumerate(entries, start=1):
        zip_url = entry.get("zipball_url")
        owner_name = resolve_owner_name(entry)

        if not isinstance(zip_url, str) or not zip_url:
            skipped += 1
            print(f"[{idx}/{total}] Skipping item with missing zipball_url", file=sys.stderr)
            continue
        if not isinstance(owner_name, str) or not owner_name:
            skipped += 1
            print(f"[{idx}/{total}] Skipping item with missing owner.name", file=sys.stderr)
            continue

        owner_dir = out_root / owner_name
        owner_dir.mkdir(parents=True, exist_ok=True)
        print(f"[{idx}/{total}] Downloading {zip_url} -> {owner_dir}")

        tmp_zip: Path | None = None
        try:
            tmp_zip = download_zip_to_temp(zip_url, args.timeout)
            extract_zip(tmp_zip, owner_dir)
            downloaded += 1
        except (urllib.error.URLError, zipfile.BadZipFile, OSError) as err:
            failed += 1
            print(f"[{idx}/{total}] Failed for owner={owner_name}: {err}", file=sys.stderr)
        finally:
            if tmp_zip is not None:
                tmp_zip.unlink(missing_ok=True)

    print(
        f"Done. processed={total} downloaded={downloaded} skipped={skipped} failed={failed}"
    )
    return 1 if failed > 0 else 0


if __name__ == "__main__":
    raise SystemExit(main())

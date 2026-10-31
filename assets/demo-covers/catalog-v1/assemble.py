#!/usr/bin/env python3
"""Assemble the reviewed delivery files from the retained local originals."""

import hashlib
import json
import pathlib
import shutil
import sys


def main():
    if len(sys.argv) != 3:
        raise SystemExit("usage: assemble.py SOURCE_ROOT NEW_OUTPUT_DIRECTORY")

    source_root = pathlib.Path(sys.argv[1]).resolve(strict=True)
    output = pathlib.Path(sys.argv[2])
    manifest_path = pathlib.Path(__file__).with_name("manifest.json")
    document = json.loads(manifest_path.read_text(encoding="utf-8"))
    entries = document["covers"]
    slugs = [entry["slug"] for entry in entries]
    if len(slugs) != 109 or len(set(slugs)) != 109:
        raise SystemExit("manifest must contain 109 unique book slugs")

    files = []
    names = set()
    for entry in entries:
        for kind in ("main", "thumbnail"):
            item = entry[kind]
            name = item["file"]
            if pathlib.PurePath(name).name != name or name in names:
                raise SystemExit(f"invalid or duplicate output filename: {name}")
            names.add(name)
            source = (source_root / item["source"]).resolve(strict=True)
            if not source.is_relative_to(source_root) or not source.is_file():
                raise SystemExit(f"invalid source: {source}")
            digest = hashlib.sha256(source.read_bytes()).hexdigest()
            if digest != item["sha256"]:
                raise SystemExit(f"checksum mismatch: {source}")
            files.append((source, name))

    output.mkdir(parents=True, exist_ok=False)
    for source, name in files:
        shutil.copyfile(source, output / name)
    shutil.copyfile(manifest_path, output / "inventory.json")
    install_manifest = {
        "version": 2,
        "bundle_id": document["dataset_id"],
        "covers": [],
    }
    for entry in entries:
        install_manifest["covers"].append({
            "slug": entry["slug"],
            "reviewed": True,
            "usage_basis": (
                "Project demo illustration, not a specific edition; "
                f"reviewed by {entry['reviewed_by']} on {entry['reviewed_on']}. "
                "See retained provenance and review records."
            ),
            "main": {key: entry["main"][key] for key in ("file", "sha256")},
            "thumbnail": {key: entry["thumbnail"][key] for key in ("file", "sha256")},
        })
    manifest_bytes = (json.dumps(install_manifest, indent=2, ensure_ascii=False) + "\n").encode()
    (output / "manifest.json").write_bytes(manifest_bytes)
    print(f"Installer manifest SHA-256: {hashlib.sha256(manifest_bytes).hexdigest()}")
    print(f"Assembled {len(entries)} covers and {len(files)} images in {output}")


if __name__ == "__main__":
    main()

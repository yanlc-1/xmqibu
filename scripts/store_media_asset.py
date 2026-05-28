#!/usr/bin/env python3

import argparse
import shutil
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description="Copy a local file into the site's media directory.")
    parser.add_argument("--source", required=True, help="Local source file path")
    parser.add_argument("--media-dir", required=True, help="Media root directory")
    parser.add_argument("--slug", required=True, help="Subdirectory to store the file under")
    args = parser.parse_args()

    source = Path(args.source).resolve()
    if not source.is_file():
        raise SystemExit(f"source file not found: {source}")

    media_dir = Path(args.media_dir).resolve()
    target_dir = media_dir / args.slug
    target_dir.mkdir(parents=True, exist_ok=True)

    target = target_dir / source.name
    shutil.copy2(source, target)
    print(f"/media/{args.slug}/{source.name}")


if __name__ == "__main__":
    main()

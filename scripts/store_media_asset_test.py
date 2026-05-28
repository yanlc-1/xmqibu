import subprocess
from pathlib import Path


def test_store_media_asset_copies_file(tmp_path):
    source = tmp_path / "source.txt"
    source.write_bytes(b"demo-image")
    media_dir = tmp_path / "media"

    result = subprocess.run(
        [
            "python3",
            "scripts/store_media_asset.py",
            "--source",
            str(source),
            "--media-dir",
            str(media_dir),
            "--slug",
            "github-top-weekly",
        ],
        check=True,
        capture_output=True,
        text=True,
    )

    relative_url = result.stdout.strip()
    target = media_dir / "github-top-weekly" / "source.txt"

    assert relative_url == "/media/github-top-weekly/source.txt"
    assert target.read_bytes() == b"demo-image"


def test_env_example_mentions_media_dir():
    content = Path(".env.example").read_text(encoding="utf-8")
    assert "MEDIA_DIR=" in content

from pathlib import Path
import json
import subprocess


ROOT = Path(__file__).resolve().parents[1]
SOURCE_DIR = ROOT / "resources/images/originals"
TARGET_DIR = ROOT / "resources/images"
DEFAULT_VARIANT = "canyon-falls"
SUPPORTED_EXTENSIONS = {".jpg", ".jpeg", ".png"}


def main() -> None:
    TARGET_DIR.mkdir(parents=True, exist_ok=True)
    SOURCE_DIR.mkdir(parents=True, exist_ok=True)

    for path in sorted(TARGET_DIR.rglob("*")):
        if not path.is_file() or path.suffix.lower() not in SUPPORTED_EXTENSIONS:
            continue
        if SOURCE_DIR in path.parents:
            continue

        backup = SOURCE_DIR / path.relative_to(TARGET_DIR)
        backup.parent.mkdir(parents=True, exist_ok=True)
        if not backup.exists():
            path.replace(backup)
        else:
            path.unlink()

    variants = {}
    for path in sorted(SOURCE_DIR.rglob("*")):
        if not path.is_file() or path.suffix.lower() not in SUPPORTED_EXTENSIONS:
            continue

        variant = path.stem
        target = TARGET_DIR / f"{variant}.webp"
        subprocess.run(
            [
                "cwebp",
                "-quiet",
                "-q",
                "75",
                str(path),
                "-o",
                str(target),
            ],
            check=True,
        )
        variants[variant] = {"file": target.name}

    default = (
        DEFAULT_VARIANT if DEFAULT_VARIANT in variants else next(iter(variants), "")
    )
    manifest = {"default": default, "variants": variants}
    (TARGET_DIR / "images.json").write_text(json.dumps(manifest, indent=2) + "\n")


if __name__ == "__main__":
    main()

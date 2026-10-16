#!/usr/bin/env python3
"""Verify ticket 02's recorded source and offline evidence SHA-256 values."""
import hashlib
import json
import subprocess
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REPORT = Path(__file__).with_name("02-evidence.json")


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify_file(relative, expected):
    path = ROOT / relative
    if sha256(path) != expected:
        raise SystemExit(f"SHA-256 mismatch: {relative}")
    return path


def main():
    report = json.loads(REPORT.read_text())
    for relative, expected in report["sourceFiles"].items():
        verify_file(relative, expected)
    tracked = subprocess.check_output(["git", "ls-files", "-z"], cwd=ROOT).decode().split("\0")
    go_sources = {name: sha256(ROOT / name) for name in tracked
                  if name.endswith(".go") or name in ("go.mod", "go.sum")}
    tree_digest = hashlib.sha256(json.dumps(go_sources, sort_keys=True).encode()).hexdigest()
    if tree_digest != report["trackedGoSourceSHA256"]:
        raise SystemExit("tracked Go source inventory changed")
    for name, entry in report["runs"].items():
        manifest = verify_file(entry["manifest"], entry["manifestSHA256"])
        data = json.loads(manifest.read_text())
        artifacts = 0
        for case in data["cases"]:
            for relative, expected in case["artifacts"].items():
                path = manifest.parent / case["dir"] / relative
                if sha256(path) != expected:
                    raise SystemExit(f"artifact mismatch: {case['id']}/{relative}")
                artifacts += 1
        audit = verify_file(entry["audit"], entry["auditSHA256"])
        if json.loads(audit.read_text())["findings"]:
            raise SystemExit(f"redaction audit failed: {name}")
        statuses = dict(Counter(case["status"] for case in data["cases"]))
        if statuses != entry["statuses"] or artifacts != entry["verifiedArtifacts"]:
            raise SystemExit(f"inventory mismatch: {name}")
        print(f"{name}: {len(data['cases'])} cases, {artifacts} artifacts verified")
    print("source SHA-256 values and all redaction audits verified")


if __name__ == "__main__":
    main()

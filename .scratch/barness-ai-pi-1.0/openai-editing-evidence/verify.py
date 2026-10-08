#!/usr/bin/env python3
"""Verify the committed delivery hashes and locate the repeatable run bundle."""
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parents[3]
folder = Path(__file__).resolve().parent
manifest = json.loads((folder / "delivery-hashes.json").read_text())
for path, expected in manifest.items():
    actual = hashlib.sha256((root / path).read_bytes()).hexdigest()
    if actual != expected:
        raise SystemExit(f"hash mismatch: {path}")
report = json.loads((folder / "report.json").read_text())
if report["failed_cases"] or report["audit_findings"]:
    raise SystemExit("failed scenario or audit finding")
artifacts = json.loads((folder / "manifest.json").read_text())["artifacts"]
for name, expected in artifacts.items():
    if hashlib.sha256((folder / name).read_bytes()).hexdigest() != expected:
        raise SystemExit(f"artifact hash mismatch: {name}")
bundle = root / report["bundle"]
if bundle.exists():
    actual = hashlib.sha256((bundle / "manifest.json").read_bytes()).hexdigest()
    if actual != report["manifest_sha256"]:
        raise SystemExit("source manifest hash mismatch")
    index = json.loads((folder / "image-cases.json").read_text())
    for case in index["cases"]:
        for name, expected in case["artifacts"].items():
            path = bundle / case["dir"] / name
            if hashlib.sha256(path.read_bytes()).hexdigest() != expected:
                raise SystemExit(f"source artifact hash mismatch: {case['id']}/{name}")
print(f"verified {len(manifest)} delivery hashes; {report['passing_cases']} scenarios PASS")
print("full replay: " + report["command"])
print("bundle: " + report["bundle"])

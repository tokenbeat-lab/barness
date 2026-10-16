#!/usr/bin/env python3
"""Verify ticket-12 artifacts and write an append-only receipt, from repo root."""
import argparse
import collections
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile


def raw(path):
    path = Path(path)
    for parent in (path, *path.parents):
        if parent.is_symlink():
            raise ValueError("symlink in evidence path")
    if not path.is_file():
        raise ValueError("missing regular evidence artifact")
    return path.read_bytes()


def digest(data):
    return hashlib.sha256(data).hexdigest()


def document(path):
    return json.loads(raw(path))


def inventory(directory, files):
    for name, expected in files.items():
        rel = Path(name)
        if rel.is_absolute() or ".." in rel.parts:
            raise ValueError("unsafe artifact path")
        if digest(raw(directory / rel)) != expected:
            raise ValueError("artifact checksum mismatch")
    return len(files)


def audited(directory):
    report = document(directory / "audit.json")
    if report.get("findings") != []:
        raise ValueError("redaction audit is not clean")


def package(directory, binary, live=False):
    manifest = document(directory / "manifest.json")
    count = inventory(directory, manifest["artifacts"])
    files = {str(p.relative_to(directory)) for p in directory.rglob("*") if p.is_file()}
    if files != set(manifest["artifacts"]) | {"manifest.json"}:
        raise ValueError("missing or unlisted artifact")
    audited(directory)
    # The two cancellation-accounting tracers deliberately leave resealed,
    # corrupted packages; their own public E2E asserts verifier rejection.
    reason = None
    if directory.name.startswith("visual-cancel-accounting-"):
        reason = b"cancelled usage/model unsupported by captured response"
    if directory.name.startswith("visual-source-set-"):
        reason = b"source snapshot drift"
    reject = reason is not None
    command = [str(binary), "-verify", str(directory)]
    result = subprocess.run(command, capture_output=True, timeout=45)
    if (result.returncode == 0) == reject:
        raise ValueError("unexpected public package verification result")
    if reject and reason not in result.stderr:
        raise ValueError("corruption rejected for an unexpected reason")
    report = document(directory / "report.json")
    if live:
        if report["mode"] != "live" or report["status"] != "NOT_RUN" or report["budget"]["calls"] != 0 or report["budget"]["attempts"] != 0:
            raise ValueError("unexpected real-route outcome")
    return {"path": str(directory), "reportSHA256": digest(raw(directory / "report.json")),
            "manifestSHA256": digest(raw(directory / "manifest.json")),
            "auditSHA256": digest(raw(directory / "audit.json")), "artifacts": count,
            "verification": "expected-rejection" if reject else "accepted",
            "route": report["route"], "mode": report["mode"], "status": report["status"],
            "realQuality": report["realQuality"], "catalogHash": report["catalogHash"],
            "codeHash": report["codeHash"], "budget": report["budget"]}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run", action="append", required=True, type=Path)
    parser.add_argument("--live", action="append", required=True, type=Path)
    parser.add_argument("--out", required=True, type=Path)
    args = parser.parse_args()
    receipt = {"schema": 1, "kind": "ticket-12-controlled-and-NOT_RUN-only", "runs": [], "packages": [], "live": []}
    with tempfile.TemporaryDirectory() as temporary:
        binary = Path(temporary) / "visualeval"
        subprocess.run(["go", "build", "-o", str(binary), "./ai/examples/visualeval"], check=True, capture_output=True)
        for directory in args.run:
            manifest = document(directory / "manifest.json")
            audited(directory)
            selected = [case for case in manifest["cases"] if case["id"].startswith("VQ12-")]
            if any(case["status"] != "PASS" for case in selected):
                raise ValueError("ticket E2E did not pass")
            count = sum(inventory(directory / case["dir"], case["artifacts"]) for case in selected)
            receipt["runs"].append({"path": str(directory), "manifestSHA256": digest(raw(directory / "manifest.json")),
                                    "auditSHA256": digest(raw(directory / "audit.json")),
                                    "cases": dict(collections.Counter(case["status"] for case in manifest["cases"])),
                                    "ticket12Cases": len(selected), "ticket12Artifacts": count})
            for child in sorted(directory.glob("visual-*")):
                if child.is_dir():
                    receipt["packages"].append(package(child, binary))
        for directory in args.live:
            receipt["live"].append(package(directory, binary, live=True))
    pins = Path("ai/examples/visualeval/testdata/pins.json")
    receipt["fixturePinsSHA256"] = digest(raw(pins))
    receipt["sealerSHA256"] = digest(raw(Path(__file__)))
    for name, expected in document(pins).items():
        if digest(raw(pins.parent / name)) != expected:
            raise ValueError("frozen fixture/source pin drift")
    output = (json.dumps(receipt, ensure_ascii=False, indent=2) + "\n").encode()
    if args.out.exists() or args.out.is_symlink():
        if raw(args.out) != output:
            raise ValueError("refusing to replace an existing differing receipt")
    else:
        with args.out.open("xb") as stream:
            stream.write(output)
    print(f"sealed {len(receipt['packages'])} Host packages and {len(receipt['live'])} real NOT_RUN packages")


if __name__ == "__main__":
    main()

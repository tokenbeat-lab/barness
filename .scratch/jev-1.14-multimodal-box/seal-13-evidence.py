#!/usr/bin/env python3
"""Seal or recheck ticket-13 receipts from the repository root, without keys."""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path
import subprocess


def raw(path):
    path = Path(path)
    if any(p.is_symlink() for p in (path, *path.parents)) or not path.is_file():
        raise ValueError("nonregular evidence input")
    return path.read_bytes()


def digest(path):
    return hashlib.sha256(raw(path)).hexdigest()


def doc(path):
    return json.loads(raw(path))


def check_files(root, files):
    for name, expected in files.items():
        p = Path(name)
        if p.is_absolute() or ".." in p.parts or digest(root / p) != expected:
            raise ValueError("artifact inventory mismatch")
    return len(files)


def verify_support(path):
    env = {k: os.environ[k] for k in ("PATH", "HOME", "TMPDIR", "GOCACHE", "GOMODCACHE", "GOPATH") if k in os.environ}
    subprocess.run(["go", "run", "./ai/live/cmd/jevformal", "-verify", str(path), "-require-live"],
                   env=env, check=True, capture_output=True, timeout=180)


def case_files(root, case):
    count = check_files(root / case["dir"], case["artifacts"])
    redacted = 0
    for name, original in case["fixtures"].items():
        if name not in case["artifacts"]:
            raise ValueError("fixture missing from stored-byte artifact inventory")
        # The existing recorder fingerprints fixture input before writeFile's
        # secret redaction. artifacts always fingerprints the retained bytes.
        # New UC13 fixtures contain no secrets and must match both identities.
        if case["artifacts"][name] != original:
            if case["id"].startswith("UC13-"):
                raise ValueError("ticket-13 fixture identity differs")
            redacted += 1
    return count, redacted


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--checks", type=Path)
    parser.add_argument("--formal", type=Path)
    parser.add_argument("--out", type=Path)
    parser.add_argument("--verify", type=Path)
    args = parser.parse_args()
    if args.verify:
        receipt = doc(args.verify)
        check_files(Path("."), receipt["files"])
        for run in receipt["runs"]:
            root = Path(run["path"])
            for case in doc(root / "manifest.json")["cases"]:
                case_files(root, case)
        check_files(Path("."), doc(Path(".scratch/jev-1.14-multimodal-box/13-admission.json"))["files"])
        verify_support(receipt["formalSupportBundle"])
        print("ticket-13 receipt and sealed files match")
        return
    if not args.checks or not args.formal or not args.out:
        parser.error("provide --checks, --formal and --out, or --verify")
    receipt = {"schema": 1, "kind": "ticket-13-fixed-jev-formal-support",
               "files": {}, "runs": [], "redRuns": [], "commands": []}
    files = receipt["files"]

    def seal(path):
        files[str(path)] = digest(path)

    for base in (args.checks, args.formal):
        for path in sorted(base.glob("*-command.json")):
            command = doc(path)
            if command["exitCode"] != 0:
                raise ValueError("unsuccessful final command")
            seal(path)
            log = Path(command["log"])
            if not log.is_file():
                log = base / log
            seal(log)
            receipt["commands"].append(command)
    for path in sorted(args.checks.glob("*/*/manifest.json")):
        directory = path.parent
        manifest = doc(path)
        if not manifest["versions"].get("model_catalog_hash"):
            continue  # Selection-only helper processes do not claim acceptance.
        if doc(directory / "audit.json").get("findings") != []:
            raise ValueError("final redaction audit is not clean")
        count, redacted = 0, 0
        for case in manifest["cases"]:
            if case["status"] not in ("PASS", "NOT_RUN"):
                raise ValueError("failed final case")
            artifacts, fingerprints = case_files(directory, case)
            count += artifacts
            redacted += fingerprints
        seal(path)
        seal(directory / "audit.json")
        receipt["runs"].append({"path": str(directory), "package": manifest["package"],
                                "cases": dict(collections.Counter(c["status"] for c in manifest["cases"])),
                                "storedArtifactsVerified": count,
                                "legacyPreRedactionFixtureFingerprints": redacted,
                                "ticket13Cases": sum(c["id"].startswith("UC13-") for c in manifest["cases"])})
    red = ["20261016T222756.729186000Z", "20261016T222905.203922000Z",
           "20261016T223444.666466000Z", "20261016T223607.547597000Z",
           "20261016T223747.688305000Z", "20261016T224551.586165000Z",
           "20261016T224756.033975000Z", "20261016T225150.274777000Z"]
    red_paths = [Path(".evidence/barness-ai") / name / "manifest.json" for name in red]
    red_paths.append(Path(".evidence/ticket13-final-e633b3f/full/20261016T225305.033331000Z/manifest.json"))
    seal(Path(".evidence/ticket13-final-e633b3f/full-command.json"))
    seal(Path(".evidence/ticket13-final-e633b3f/full.log"))
    for path in red_paths:
        manifest = doc(path)
        seal(path)
        if (path.parent / "audit.json").is_file():
            seal(path.parent / "audit.json")
        receipt["redRuns"].append({"path": str(path.parent),
                                  "cases": dict(collections.Counter(c["status"] for c in manifest["cases"]))})
    summary = doc(args.formal / "summary.json")
    support = Path(summary["supportBundle"])
    verify_support(support)
    protocol = doc(support / "protocol/multimodal-report.json")
    effects = doc(support / "effects/report.json")
    if summary["protocol"] != protocol or summary["effects"] != effects or summary["catalogHash"] != protocol["catalogHash"]:
        raise ValueError("summary differs from independently verified raw reports")
    seal(support / "formal-integrity.json")
    check_files(support, doc(support / "formal-integrity.json"))
    for name in ("summary.json", "formal-catalog.json", "authorization-context.json",
                 "protocol-authorization.json", "effect-authorization.json"):
        seal(args.formal / name)
    admission = Path(".scratch/jev-1.14-multimodal-box/13-admission.json")
    check_files(Path("."), doc(admission)["files"])
    seal(admission)
    pins = Path("ai/examples/visualeval/testdata/pins.json")
    check_files(pins.parent, doc(pins))
    seal(pins)
    matrix = Path("ai/live/support-matrix.json")
    old = doc(Path("ai/e2e/testdata/typesafe/formal-matrix-before-1.14.json"))
    new = doc(matrix)
    new["rows"] = [r for r in new["rows"] if r["combo"] != "typesafe-classifier-jev-1.14.0"]
    if new != old:
        raise ValueError("historical matrix rows changed")
    seal(matrix)
    seal(Path(__file__).resolve().relative_to(Path.cwd()))
    receipt["catalogHash"] = protocol["catalogHash"]
    receipt["formalSupportBundle"] = str(support)
    receipt["protocol"] = {k: protocol[k] for k in ("outcome", "mode", "complete", "budget")}
    receipt["effects"] = {k: effects[k] for k in ("status", "realQuality", "mode", "budget", "metrics")}
    receipt["realCallsAcrossAdmissionAndFormal"] = doc(admission)["calls"] + protocol["budget"]["callsUsed"] + effects["budget"]["calls"]
    receipt["realAttemptsAcrossAdmissionAndFormal"] = doc(admission)["attempts"] + protocol["budget"]["httpAttempts"] + effects["budget"]["attempts"]
    receipt["realRetries"] = doc(admission)["retries"] + protocol["budget"]["retries"] + effects["budget"]["retries"]
    receipt["otherRoutes"] = "historical rows unchanged; Decisions real image evidence NOT_RUN"
    encoded = (json.dumps(receipt, ensure_ascii=False, indent=2) + "\n").encode()
    if args.out.exists():
        if raw(args.out) != encoded:
            raise ValueError("refusing to replace a differing receipt")
    else:
        args.out.write_bytes(encoded)
    print(f"sealed {len(receipt['runs'])} final runs and fresh real formal support")


if __name__ == "__main__":
    main()

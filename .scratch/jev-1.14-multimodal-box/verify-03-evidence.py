#!/usr/bin/env python3
"""Verify ticket 03 source, artifact inventory, audit and final image identities."""
import base64
import hashlib
import json
import subprocess
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REPORT = Path(__file__).with_name("03-evidence.json")


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify_file(relative, expected):
    path = ROOT / relative
    if sha256(path) != expected:
        raise SystemExit(f"SHA-256 mismatch: {relative}")
    return path


def verify_identities(case_dir):
    result = json.loads((case_dir / "result.json").read_text())
    if result["stopReason"] != "stop":
        if result["answers"] is not None or "imageInputs" in result:
            raise SystemExit(f"non-atomic failure: {case_dir.name}")
        return
    requests = json.loads((case_dir / "requests.json").read_text())
    final_input = requests[-1]["body"]["input"]
    expected = []
    if isinstance(final_input, list):
        for part in final_input[0]["content"][1:]:
            mime, data = part["image_url"].split(";base64,", 1)
            expected.append({"mimeType": mime.removeprefix("data:"),
                             "sha256": hashlib.sha256(base64.b64decode(data, validate=True)).hexdigest()})
    if result.get("imageInputs", []) != expected or (not expected and "imageInputs" in result):
        raise SystemExit(f"final native image identity mismatch: {case_dir.name}")
    shared = (case_dir / "observer-records.json").read_text() + json.dumps(result["metadata"])
    if any(image["sha256"] in shared for image in expected):
        raise SystemExit(f"image identity in shared diagnostics: {case_dir.name}")


def main():
    report = json.loads(REPORT.read_text())
    for relative, expected in report["sourceFiles"].items():
        verify_file(relative, expected)
    tracked = subprocess.check_output(["git", "ls-files", "-z"], cwd=ROOT).decode().split("\0")
    sources = {name: sha256(ROOT / name) for name in tracked
               if name.endswith(".go") or name in ("go.mod", "go.sum")}
    if hashlib.sha256(json.dumps(sources, sort_keys=True).encode()).hexdigest() != report["trackedGoSourceSHA256"]:
        raise SystemExit("tracked Go source inventory changed")
    for name, entry in report["runs"].items():
        manifest = verify_file(entry["manifest"], entry["manifestSHA256"])
        data = json.loads(manifest.read_text())
        artifacts, ticket_cases = 0, 0
        for case in data["cases"]:
            case_dir = manifest.parent / case["dir"]
            for relative, expected in case["artifacts"].items():
                if sha256(case_dir / relative) != expected:
                    raise SystemExit(f"artifact mismatch: {case['id']}/{relative}")
                artifacts += 1
            if case["id"].startswith("UC03-"):
                if case["status"] != "PASS":
                    raise SystemExit(f"ticket case not PASS: {case['id']}")
                verify_identities(case_dir)
                ticket_cases += 1
        audit = verify_file(entry["audit"], entry["auditSHA256"])
        if json.loads(audit.read_text())["findings"]:
            raise SystemExit(f"redaction audit failed: {name}")
        statuses = dict(Counter(case["status"] for case in data["cases"]))
        if statuses != entry["statuses"] or artifacts != entry["verifiedArtifacts"] or ticket_cases != entry["ticketCases"]:
            raise SystemExit(f"inventory mismatch: {name}")
        print(f"{name}: {len(data['cases'])} cases, {artifacts} artifacts, {ticket_cases} final-identity checks verified")
    print("source SHA-256 values, atomic results, final file identities and redaction audits verified")


if __name__ == "__main__":
    main()

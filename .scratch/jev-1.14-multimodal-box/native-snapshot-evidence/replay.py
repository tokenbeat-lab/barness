#!/usr/bin/env python3
"""Replay issue 01 offline; preserve each run and hash every delivered file."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[3]
FOCUSED = "^Test(TypeSafeNative|JevRegression|JevBaselineEvidence|Classifier|Unary|Catalog|OperationCatalog|TrustedCallback|ProtocolCallbacks|ProtocolNoEnvironmentFallback)"
PI = "^TestPiDifferential$/^typesafe-system-one$|^TestClassifierCatalogParity$"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path, data):
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n")


def inventory():
    sources = {str(p.relative_to(ROOT)): sha(p) for p in sorted((ROOT / "ai").rglob("*"))
               if p.is_file() and (p.suffix == ".go" or "testdata" in p.parts)
               and "node_modules" not in p.parts}
    stable = [p for p in (ROOT / "ai").glob("*catalog*") if p.is_file()]
    stable += list((ROOT / "ai" / "catalogdata").rglob("*"))
    stable += list((ROOT / "ai" / "e2e" / "testdata" / "typesafe").rglob("*"))
    stable += [ROOT / "ai" / "live" / "support-matrix.json"]
    stable += list((ROOT / "ai" / "internal" / "testkit" / "pioracle" / "node").glob("*"))
    stable += list((ROOT / ".scratch" / "barness-ai-openai-decisions" / "jev-baseline").glob("*"))
    return {"head": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip(),
            "sources": sources,
            "stable_inputs": {str(p.relative_to(ROOT)): sha(p) for p in sorted(set(stable)) if p.is_file()}}


def execute(out, name, argv, extra=None):
    env = dict(os.environ)
    for key in ("BARNESS_AI_LIVE", "BARNESS_AI_PRESSURE", "BARNESS_AI_PIDIFF"):
        env.pop(key, None)
    env.update(extra or {})
    env["BARNESS_AI_EVIDENCE_DIR"] = str(out / "runs" / name)
    with (out / (name + ".log")).open("w") as log:
        process = subprocess.Popen(argv, cwd=ROOT, env=env, stdout=subprocess.PIPE,
                                   stderr=subprocess.STDOUT, text=True)
        for line in process.stdout:
            log.write(line.replace(str(ROOT), "$REPO").replace(str(Path.home()), "$HOME"))
        code = process.wait()
    print(name + ": " + ("PASS" if code == 0 else "FAIL"), flush=True)
    return {"name": name, "argv": [a.replace(str(out), "$OUT") for a in argv],
            "environment": dict(extra or {}, BARNESS_AI_EVIDENCE_DIR="$OUT/runs/" + name),
            "exit_code": code}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--phase", choices=("baseline", "final"), required=True)
    parser.add_argument("--baseline", type=Path)
    args = parser.parse_args()
    if args.phase == "final" and args.baseline is None:
        parser.error("--baseline is required for final comparison")
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=False)
    before = inventory()
    write(out / "inputs.json", before)
    configuration = ["ai/e2e/main_test.go", "ai/e2e/classifier_fixture_test.go",
                     "ai/e2e/unary_fixture_test.go", "ai/e2e/typesafe_native_snapshot_test.go"]
    write(out / "configuration.json", {"transport": "controlled loopback HTTP",
          "images": "none", "sources": {name: (ROOT / name).read_text() for name in configuration},
          "source_hashes": {name: before["sources"][name] for name in configuration}})
    records = []
    commands = [
        ("focused", ["go", "test", "./ai/e2e", "-run", FOCUSED, "-count=1", "-json"], {}),
        ("jev-differential", ["go", "test", "./ai/e2e", "-run", PI, "-count=1", "-json"], {"BARNESS_AI_PIDIFF": "1"}),
        ("oracle-startup", ["node", "--test", "ai/internal/testkit/pioracle/node/runner.test.mjs"], {}),
        ("frozen-baseline", ["python3", ".scratch/barness-ai-openai-decisions/jev-baseline/verify.py", "--audit"], {}),
    ]
    if args.phase == "final":
        commands += [
            ("vet", ["go", "vet", "./..."], {}),
            ("vet-live", ["go", "vet", "-tags", "live", "./..."], {}),
            ("race", ["go", "test", "-race", "./ai/e2e", "-run", "^Test(TypeSafeNative|JevRegression|Classifier|Unary)", "-count=1", "-json"], {"BARNESS_AI_PIDIFF": "1"}),
            ("full", ["go", "test", "./...", "-count=1", "-json"], {"BARNESS_AI_PIDIFF": "1"}),
        ]
    for name, argv, extra in commands:
        records.append(execute(out, name, argv, extra))
        write(out / "commands.json", records)
    after = inventory()
    stable = before["stable_inputs"] == after["stable_inputs"]
    if args.baseline:
        stable = stable and before["stable_inputs"] == json.loads((args.baseline / "inputs.json").read_text())["stable_inputs"]
    summaries = []
    for path in sorted((out / "runs").rglob("manifest.json")):
        data = json.loads(path.read_text())
        counts = {}
        for case in data["cases"]:
            counts[case["status"]] = counts.get(case["status"], 0) + 1
        summaries.append({"manifest": str(path.relative_to(out)), "counts": counts,
                          "versions": data["versions"], "sha256": sha(path)})
    write(out / "manifest.json", {"kind": "offline-typesafe-native-snapshot", "phase": args.phase,
          "stable_inputs_unchanged": stable, "sources_unchanged_during_run": before["sources"] == after["sources"],
          "images": "not applicable: text-only refactor", "live": "NOT_RUN", "runs": summaries})
    records.append(execute(out, "audit", ["go", "run", "./ai/internal/testkit/audit/cmd/auditbundle", str(out)]))
    write(out / "commands.json", records)
    write(out / "files.json", {str(p.relative_to(out)): sha(p) for p in sorted(out.rglob("*")) if p.is_file() and p != out / "files.json"})
    good = stable and before["sources"] == after["sources"] and all(r["exit_code"] == 0 for r in records)
    return 0 if good else 1


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""Execute current Jev regressions into a new directory; never reseal history."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[3]
COMMANDS = [
    ("existing", ["go", "test", "./ai/e2e", "-run", "^Test(Classifier|Unary|Catalog|OperationCatalog|TrustedCallback|ProtocolCallbacks|ProtocolNoEnvironmentFallback)", "-count=1", "-json"], {}),
    ("jev-differential", ["go", "test", "./ai/e2e", "-run", "^TestPiDifferential$/^typesafe-system-one$|^TestClassifierCatalogParity$", "-count=1", "-json"], {"BARNESS_AI_PIDIFF": "1"}),
    ("supplement", ["go", "test", "./ai/e2e", "-run", "^TestJevRegression", "-count=1", "-json"], {}),
    ("artifact", ["go", "test", "./ai/e2e", "-run", "^TestJevBaselineEvidence$", "-count=1", "-json"], {}),
    ("oracle-startup", ["node", "--test", "ai/internal/testkit/pioracle/node/runner.test.mjs"], {}),
    ("race", ["go", "test", "-race", "./ai/e2e", "-run", "^Test(JevRegression|Classifier|Unary)", "-count=1", "-json"], {"BARNESS_AI_PIDIFF": "1"}),
    ("vet", ["go", "vet", "./..."], {}),
    ("vet-live", ["go", "vet", "-tags", "live", "./..."], {}),
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=False)
    sources = {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest()
               for p in sorted((ROOT / "ai").rglob("*.go"))}
    identity = {"head": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip(), "sources": sources}
    (out / "inputs.json").write_text(json.dumps(identity, indent=2) + "\n")
    records = []
    env = dict(os.environ)
    for key in ("BARNESS_AI_LIVE", "BARNESS_AI_PRESSURE", "BARNESS_AI_PIDIFF"):
        env.pop(key, None)
    for name, argv, extra in COMMANDS:
        started = datetime.datetime.now(datetime.timezone.utc).isoformat()
        call_env = dict(env, BARNESS_AI_EVIDENCE_DIR=str(out / name), **extra)
        with (out / (name + ".log")).open("w") as log:
            process = subprocess.Popen(argv, cwd=ROOT, env=call_env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
            for line in process.stdout:
                log.write(line.replace(str(ROOT), "$REPO").replace(str(Path.home()), "$HOME"))
            code = process.wait()
        records.append({"name": name, "argv": argv, "environment": extra, "started_at": started, "exit_code": code})
        (out / "commands.json").write_text(json.dumps(records, indent=2) + "\n")
        print(name + ": " + ("PASS" if code == 0 else "FAIL"), flush=True)
    return 1 if any(r["exit_code"] for r in records) else 0


if __name__ == "__main__":
    sys.exit(main())

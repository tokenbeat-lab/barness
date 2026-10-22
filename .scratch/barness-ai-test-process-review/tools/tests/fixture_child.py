"""A synthetic external process producing native-shape bundles, never Go/live."""
import hashlib
import json
import os
from pathlib import Path
import sys
import time

mode = sys.argv[1]
if mode == "timeout":
    time.sleep(10)
if mode == "fail":
    print("synthetic failure", file=sys.stderr)
    sys.exit(7)
if mode == "empty":
    sys.exit(0)
base = Path(os.environ["BARNESS_AI_EVIDENCE_DIR"])
for index in range(2 if mode in ("ambiguous", "auxiliary") else 1):
    bundle = base / str(index)
    case = bundle / "fixture-case"
    case.mkdir(parents=True)
    (case / "assertions.json").write_text('{"assertions":[{"passed":true}]}\n')
    manifest = {"package": "./ai/e2e", "git_commit": os.environ.get("BARNESS_RUN_COMMIT", ""),
        "cases": [{"id": "fixture-case", "test": "TestFixture", "status": "PASS", "dir": "fixture-case",
                   "artifacts": {"assertions.json": hashlib.sha256((case / "assertions.json").read_bytes()).hexdigest()}}]}
    if mode == 'live':
        manifest['package'] = './ai/live'
    if mode == "auxiliary" and index == 1:
        manifest["package"] = "./ai/examples/visualeval"
    if mode == "no-cases":
        manifest["cases"] = []
    (bundle / "manifest.json").write_text(json.dumps(manifest))
    if mode != "no-audit":
        (bundle / "audit.json").write_text('{"findings":[]}' if mode != "bad-audit" else '{"findings":[{"rule":"synthetic"}]}')
    if mode == "tampered":
        (case / "assertions.json").write_text('{}')
print(json.dumps({"Action": "run", "Package": "example.test/fixture/ai/e2e", "Test": "TestFixture"}))
print(json.dumps({"Action": "pass", "Package": "example.test/fixture/ai/e2e", "Test": "TestFixture"}))
print(json.dumps({"Action": "pass", "Package": "example.test/fixture/ai/e2e"}))
print(json.dumps({"environment": {k: os.environ[k] for k in os.environ if k.startswith("BARNESS") or k == "OPENAI_API_KEY"}}))

if mode == 'input-change':
    Path('policy.json').write_text('{"limit":99}')

if mode == "doc-change":
    Path("ai/README.md").write_text("Prose only; no execution input changed.\n")

#!/usr/bin/env python3
"""Compare captured public Client behavior, verifying each bundle's hashes."""
import argparse
import hashlib
import json
import re
from pathlib import Path


def sha(data):
    return hashlib.sha256(data).hexdigest()


def load(path):
    return json.loads(path.read_text())


def normalize(value):
    if isinstance(value, dict):
        # Wall clock and random retry jitter are asserted by the E2E cases,
        # not byte-comparable across independent runs.
        return {k: normalize(v) for k, v in value.items()
                if k not in {"timestamp", "durationMs", "retryDelay"}}
    if isinstance(value, list):
        return [normalize(v) for v in value]
    if isinstance(value, str):
        return re.sub(r"(https?://127\.0\.0\.1):[0-9]+", r"\1:$PORT", value)
    return value


def selected(name):
    return (name in {"events.json", "result.json", "error.json", "outcome.json",
                     "response-script.json", "fixture.json", "input.json", "options.json",
                     "first.json", "round-1.json", "round-2.json"}
            or name.startswith(("requests", "envelopes-", "result-")))


def read_bundle(path):
    manifest = load(path)
    cases = {c["id"]: c for c in manifest["cases"]}
    for case in cases.values():
        if case["status"] == "FAIL":
            raise ValueError(f"{case['id']}: {case['status']}")
        for name, expected in case["artifacts"].items():
            actual = sha((path.parent / case["dir"] / name).read_bytes())
            if actual != expected:
                raise ValueError(f"{case['id']}/{name}: hash mismatch")
    audit = load(path.parent / "audit.json")
    if audit.get("findings"):
        raise ValueError("redaction audit contains findings")
    return manifest, cases


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("before", type=Path)
    parser.add_argument("after", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    before, bcases = read_bundle(args.before)
    after, acases = read_bundle(args.after)
    differences, skipped, records = [], [], []
    # The after bundle may contain the wider full test suite. Every original
    # case must still exist, pass and retain every selected artifact.
    for case_id, bcase in sorted(bcases.items()):
        acase = acases.get(case_id)
        if acase is None:
            differences.append({"case": case_id, "reason": "missing after case"})
            continue
        if bcase["status"] != acase["status"]:
            differences.append({"case": case_id, "reason": "status changed"})
        names = {n for n in bcase["artifacts"] | acase["artifacts"] if selected(n)}
        for name in sorted(names):
            if "-E06-thinking-sampling-concurrent-" in case_id and name == "result.json":
                # This existing evidence filename is overwritten by whichever
                # tenant finishes last. Per-tenant captured requests remain
                # compared and the case's isolation assertions must pass.
                skipped.append({"case": case_id, "artifact": name,
                                "reason": "last-finisher result filename"})
                continue
            if name not in bcase["artifacts"] or name not in acase["artifacts"]:
                differences.append({"case": case_id, "artifact": name, "reason": "presence"})
                continue
            bp = args.before.parent / bcase["dir"] / name
            ap = args.after.parent / acase["dir"] / name
            bvalue, avalue = normalize(load(bp)), normalize(load(ap))
            if ("-E06-thinking-sampling-concurrent-" in case_id and name.startswith("requests")
                    and isinstance(bvalue, list) and isinstance(avalue, list)):
                # Concurrent requests have no specified arrival order.
                bvalue = sorted(bvalue, key=lambda v: json.dumps(v, sort_keys=True))
                avalue = sorted(avalue, key=lambda v: json.dumps(v, sort_keys=True))
            bsha = sha(json.dumps(bvalue, sort_keys=True, ensure_ascii=False).encode())
            asha = sha(json.dumps(avalue, sort_keys=True, ensure_ascii=False).encode())
            records.append({"case": case_id, "artifact": name, "before": bsha, "after": asha})
            if bsha != asha:
                differences.append({"case": case_id, "artifact": name, "reason": "value"})
    report = {
        "status": "FAIL" if differences else "PASS",
        "before_manifest_sha256": sha(args.before.read_bytes()),
        "after_manifest_sha256": sha(args.after.read_bytes()),
        "before_commit": before["git_commit"], "after_commit": after["git_commit"],
        "versions": after["versions"], "cases_compared": len(bcases),
        "not_run_cases": [c for c in bcases if bcases[c]["status"] == "NOT_RUN"],
        "artifacts_compared": len(records),
        "normalizations": ["timestamp", "durationMs", "retryDelay", "loopback port",
                           "concurrent request arrival order"],
        "excluded_artifacts": skipped, "differences": differences, "artifacts": records,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    print(f"{report['status']}: {len(bcases)} cases; {len(records)} artifacts; {len(differences)} differences")
    raise SystemExit(bool(differences))


if __name__ == "__main__":
    main()

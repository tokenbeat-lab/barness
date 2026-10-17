#!/usr/bin/env python3
"""Verify saved ticket10 bundles; no network, credentials or model calls."""
import argparse
import collections
import hashlib
import json
from pathlib import Path


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify_bundle(path):
    manifest = json.loads((path / "manifest.json").read_text())
    audit = json.loads((path / "audit.json").read_text())
    assert audit["findings"] == [], f"audit findings: {path}"
    cases = {}
    for case in manifest["cases"]:
        assert case["id"] not in cases, f"duplicate case: {case['id']}"
        for name, sha in case["artifacts"].items():
            artifact = path / case["dir"] / name
            assert digest(artifact) == sha, f"artifact changed: {artifact}"
        cases[case["id"]] = case
    return manifest, cases


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("raw_root", type=Path)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[4]
    ledger_path = root / "ai/e2e/testdata/glm-pidiff/ledger.json"
    ledger = json.loads(ledger_path.read_text())
    raw_root = args.raw_root.resolve()
    bundles, case_sets = {}, {}
    for name in ["public", "differential", "ledger"]:
        path = raw_root / "runs" / name
        manifest, cases = verify_bundle(path)
        statuses = dict(collections.Counter(c["status"] for c in cases.values()))
        assert not any(c["status"] not in ("PASS", "NOT_RUN") for c in cases.values())
        if name == "public":
            assert [c["id"] for c in cases.values() if c["status"] == "NOT_RUN"] == ["GLM10-frozen-route-identity"]
        else:
            assert all(c["status"] == "PASS" for c in cases.values())
        bundles[name] = {"path": str(path), "manifestSHA256": digest(path / "manifest.json"),
                         "auditSHA256": digest(path / "audit.json"), "statuses": statuses,
                         "gitCommit": manifest["git_commit"], "versions": manifest["versions"]}
        case_sets[name] = cases
    proof = []
    for index, decision in enumerate(ledger["decisions"]):
        assert decision["classification"] in ("fixed", "extension")
        assert decision["cases"] and decision["protocol"].startswith("zhipu/")
        assert decision["path"] not in ("request.body", "result", "events")
        for case_id in decision["cases"]:
            case = case_sets["public"][case_id]
            assert case["status"] == "PASS", case_id
            proof.append({"decision": index, "classification": decision["classification"],
                          "protocol": decision["protocol"], "path": decision["path"],
                          "case": case_id, "dir": case["dir"], "replay": case["replay"],
                          "artifacts": case["artifacts"]})
    differential_counts = collections.Counter()
    for case in case_sets["differential"].values():
        record = json.loads((raw_root / "runs/differential" / case["dir"] / "pidiff.json").read_text())
        assert record["pi_version"] == "1.1.0"
        assert record["pi_commit"] == "abe508e1b89912adde45528136c3221eb69acdd7"
        assert record["pi_model_data_sha256"] == "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8"
        assert record["pi_source_sha256"] == "39a9951108d0e7faf60886302ce4bc953d8c51d5ebad22370301f23b29aa2504"
        assert record["sdk_lock_sha256"] == "e0b4084c372bc764b73faa25f2a311a3649e5b7ffb341352f25c13a891b7de65"
        assert record["pending"] == 0 and record["pass"]
        assert not case["id"].startswith("PIDIFF-GLM")
        differential_counts[record["protocol"]] += 1
    # Exact proof is separate to keep this index compact and inspectable.
    proof_path = raw_root / "case-path-proof.json"
    proof_path.write_text(json.dumps(proof, ensure_ascii=False, indent=2) + "\n")
    result = {"scope": "offline own GLM extension acceptance plus unchanged native differential; no live/support-matrix promotion",
              "bundles": bundles, "ledger": {"path": str(ledger_path), "sha256": digest(ledger_path),
              "decisions": len(ledger["decisions"]), "casePathPairs": len(proof), "cases": len({p['case'] for p in proof}),
              "pending": 0, "proof": str(proof_path), "proofSHA256": digest(proof_path)},
              "nativeDifferentialCounts": dict(differential_counts), "glmNativePiParity": "NOT_APPLICABLE_EXTENSION_ROUTE"}
    args.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"statuses": {k: v['statuses'] for k, v in bundles.items()},
                      "decisions": len(ledger["decisions"]), "casePathPairs": len(proof),
                      "nativeDifferentialCounts": dict(differential_counts)}))


if __name__ == "__main__":
    main()

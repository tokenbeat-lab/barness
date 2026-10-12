#!/usr/bin/env python3
"""Archive already-executed issue 08 runs and audit the resulting bundle."""
import argparse
import collections
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def write_json(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")


def applicable(case):
    return (case["id"].startswith("E08-mixed-pressure-")
            or case["id"].startswith("E08-classifier-images-import-")
            or case["id"].startswith("E10-mixed-policy-"))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--executions", type=Path, required=True)
    args = parser.parse_args()
    repo = Path(__file__).resolve().parents[3]
    output = Path(__file__).resolve().parent
    execution_root = args.executions.resolve()
    validation = json.loads((output / "validation.json").read_text())
    summaries = []
    measurements = []
    with tempfile.TemporaryDirectory(prefix="barness-classifier-load-") as tmp:
        root = Path(tmp)
        for category in ["history", "full", "full-final", "race", "full-current", "race-current"]:
            manifests = sorted((execution_root / category).glob("*/manifest.json"))
            if not manifests:
                raise SystemExit(f"missing executed {category} evidence")
            for manifest in manifests:
                report = json.loads(manifest.read_text())
                destination = root / "runs" / category / manifest.parent.name
                if category == "history":
                    shutil.copytree(manifest.parent, destination)
                    retained = report["cases"]
                else:
                    destination.mkdir(parents=True)
                    for name in ["manifest.json", "audit.json"]:
                        shutil.copy2(manifest.parent / name, destination / name)
                    retained = [c for c in report["cases"] if applicable(c) or c["status"] == "FAIL"]
                    for case in retained:
                        shutil.copytree(manifest.parent / case["dir"], destination / case["dir"])
                summaries.append({"path": str(destination.relative_to(root)),
                                  "gitCommit": report["git_commit"],
                                  "statuses": dict(collections.Counter(c["status"] for c in report["cases"])),
                                  "cases": len(report["cases"]), "retainedCases": len(retained),
                                  "uniqueCaseDirectories": len({c["dir"] for c in report["cases"]}),
                                  "repeatedCaseNote": "count=5 diagnostic retained last files per same directory; independent count=1 reruns also archived" if len({c["dir"] for c in report["cases"]})!=len(report["cases"]) else None,
                                  "retention": "complete" if category == "history" else "applicable_complete_others_manifest_audit_only"})
                for case in retained:
                    path = destination / case["dir"] / "classifier-images-pressure.json"
                    if path.exists():
                        load = json.loads(path.read_text())
                        measurements.append({"run": category, "path": str(path.relative_to(root)),
                                             "case": case["id"], "status": case["status"],
                                             "budget": load["heapBudget"], "batches": load["batches"],
                                             "headroom": load["headroom"], "identity": load["identity"],
                                             "currentQuestionReading": load["headroom"]["Classifier.MaxQuestionBytes"]["design"] == 4050})
        # Baseline-to-source diff is frozen by the implementation commits; no
        # unrelated caller files (notably modified GLOSSARY) enter the archive.
        files = subprocess.check_output(["git", "diff", "--name-only", "c4cc495ddba6c7fbe625ccaede74989d4341fe88...HEAD"], cwd=repo, text=True).splitlines()
        files = [name for name in files if "/classifier-load-evidence/" not in name]
        files += [".scratch/barness-ai-openai-decisions/" + name for name in ["spec.md", "design.md", "acceptance.md", "issues/08-classifier-image-design-load.md"]]
        for name in sorted(set(files)):
            destination = root / "source" / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(repo / name, destination)
        for name in ["README.md", "REVIEW.md", "capacity.md", "validation.json", "pack.py"]:
            shutil.copy2(output / name, root / name)
        (root / "logs").mkdir()
        for path in sorted(execution_root.glob("*.log")):
            raw = path.read_text().replace(str(repo), "<workspace>").replace(str(Path.home()), "<home>")
            (root / "logs" / path.name).write_text(raw)
        write_json(root / "manifest.json", {"issue": "08", "baseline": "c4cc495", "runs": summaries,
                                            "measurements": measurements, "executions": validation,
                                            "live": "NOT_RUN", "releaseGate": "NOT_RUN"})
        subprocess.run(["go", "run", "./ai/internal/testkit/audit/cmd/auditbundle", str(root)], cwd=repo, check=True)
        members = []
        archive = output / "load.zip"
        with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as bundle:
            for path in sorted(root.rglob("*")):
                if path.is_file():
                    raw = path.read_bytes()
                    name = str(path.relative_to(root))
                    bundle.writestr(name, raw)
                    members.append({"path": name, "bytes": len(raw), "sha256": digest(raw)})
        write_json(output / "manifest.json", {"issue": "08", "archive": archive.name,
                                               "sha256": digest(archive.read_bytes()), "runs": summaries,
                                               "measurements": measurements, "members": members})
        shutil.copy2(root / "audit.json", output / "audit.json")
    names = ["load.zip", "manifest.json", "validation.json", "audit.json", "README.md", "REVIEW.md", "capacity.md", "pack.py"]
    (output / "SHA256SUMS").write_text("".join(f"{digest((output / name).read_bytes())}  {name}\n" for name in names))
    print(f"Packaged {len(members)} members; audit PASS")


if __name__ == "__main__":
    main()

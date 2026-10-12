#!/usr/bin/env python3
"""Package issue 06's existing executions; never manufacture a test verdict."""
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


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--validation", type=Path, required=True)
    parser.add_argument("--history", type=Path, required=True)
    parser.add_argument("--race", type=Path, required=True)
    parser.add_argument("--full", type=Path, required=True)
    parser.add_argument("--log", type=Path, action="append", default=[])
    args = parser.parse_args()
    repo = Path(__file__).resolve().parents[3]
    output = Path(__file__).resolve().parent
    baseline = "26d955b260c2ac3f08bdce3bba810772d5948d49"
    validation = json.loads(args.validation.read_text())
    write_json(output / "validation.json", validation)
    summaries = []
    with tempfile.TemporaryDirectory(prefix="barness-runtime-package-") as tmp:
        root = Path(tmp)
        for category, base in [("history", args.history), ("race", args.race), ("full", args.full)]:
            manifests = sorted(base.glob("*/manifest.json"))
            if not manifests:
                raise SystemExit(f"missing {category} evidence")
            for manifest in manifests:
                report = json.loads(manifest.read_text())
                relative = Path("runs", category, manifest.parent.name)
                destination = root / relative
                if category == "full":
                    destination.mkdir(parents=True)
                    for name in ["manifest.json", "audit.json"]:
                        shutil.copy2(manifest.parent / name, destination / name)
                    retained = [c["dir"] for c in report["cases"]
                                if (c["id"].startswith("P10-OD") and "-runtime-" in c["id"])
                                or c["id"].startswith("P10-OD41-usage-")
                                or c["id"] == "P10-OD42-images-validation-cancellation"]
                    for name in retained:
                        shutil.copytree(manifest.parent / name, destination / name)
                    retention = "applicable_cases_complete_other_cases_manifest_and_audit_only"
                else:
                    shutil.copytree(manifest.parent, destination)
                    retention = "complete_case_artifacts"
                    retained = [c["dir"] for c in report["cases"]]
                summaries.append({"path": str(relative), "retention": retention,
                                  "commit": report["git_commit"], "cases": len(report["cases"]),
                                  "retainedCaseCount": len(retained),
                                  "statuses": dict(collections.Counter(c["status"] for c in report["cases"]))})
        files = subprocess.check_output(["git", "diff", "--name-only", f"{baseline}...HEAD"], cwd=repo, text=True).splitlines()
        # Copy the implementation and approved spec as provenance, not unrelated
        # caller changes or credentials. Evidence packaging is not recursive.
        files = [name for name in files if not name.startswith(".scratch/barness-ai-openai-decisions/runtime-evidence/")]
        files += [".scratch/barness-ai-openai-decisions/" + name for name in ["spec.md", "design.md", "acceptance.md"]]
        for name in sorted(set(files)):
            destination = root / "source" / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(repo / name, destination)
        for name in ["README.md", "REVIEW.md", "retry-disclosures.json", "pack.py"]:
            shutil.copy2(output / name, root / name)
        write_json(root / "validation.json", validation)
        logs = root / "logs"
        logs.mkdir()
        for path in args.log:
            # Go's output includes workspace locations; preserve the actual
            # execution text after replacing only that private path prefix.
            (logs / path.name).write_text(path.read_text().replace(str(repo), "<workspace>").replace(str(Path.home()), "<home>"))

        # The existing auditor activates the Observer schema whitelist only
        # for observations*.json. Preserve historical names and add identical
        # audit views, including congested cases without per-call delivery.
        observer_views = []
        for original in sorted(root.rglob("observer-records.json")):
            raw = original.read_bytes()
            alias = original.with_name("observations-observer-records.json")
            alias.write_bytes(raw)
            observer_views.append({"original": str(original.relative_to(root)),
                                   "auditView": str(alias.relative_to(root)),
                                   "sha256": digest(raw)})
        write_json(root / "manifest.json", {"issue": "06", "baseline": baseline,
                                            "runs": summaries, "live": "NOT_RUN",
                                            "observerAuditViews": observer_views,
                                            "decisionsParity": "NOT_RUN", "pressure": "NOT_RUN"})
        subprocess.run(["go", "run", "./ai/internal/testkit/audit/cmd/auditbundle", str(root)], cwd=repo, check=True)
        members = []
        archive = output / "runtime.zip"
        with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as bundle:
            for path in sorted(root.rglob("*")):
                if path.is_file():
                    raw = path.read_bytes()
                    name = str(path.relative_to(root))
                    bundle.writestr(name, raw)
                    members.append({"path": name, "bytes": len(raw), "sha256": digest(raw)})
        write_json(output / "manifest.json", {"issue": "06", "archive": archive.name,
                                               "sha256": digest(archive.read_bytes()),
                                               "runs": summaries, "members": members})
        shutil.copy2(root / "audit.json", output / "audit.json")
    names = ["runtime.zip", "manifest.json", "validation.json", "audit.json", "README.md", "REVIEW.md", "retry-disclosures.json", "pack.py"]
    (output / "SHA256SUMS").write_text("".join(f"{digest((output / name).read_bytes())}  {name}\n" for name in names))
    print(f"Packaged {len(members)} members; audit PASS")


if __name__ == "__main__":
    main()

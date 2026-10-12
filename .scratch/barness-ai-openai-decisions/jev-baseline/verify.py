#!/usr/bin/env python3
"""Read-only verification of the fixed Jev evidence; no provider calls."""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import sys
import subprocess
import tempfile
import zipfile

# Filled once from the reviewed capture. Rewriting a manifest and all its child
# hashes cannot turn a later run into this baseline. Updating this anchor is an
# explicit change of historical evidence, never a replay operation.
MANIFEST_SHA256 = "9847ee653eff6d9534db94f590946bb487fd4ffcee4ba2bfbf88558b026086cb"
PI_COMMIT = "abe508e1b89912adde45528136c3221eb69acdd7"
MODEL_DATA_SHA256 = "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8"


def require(ok, reason):
    if not ok:
        raise ValueError(reason)


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def unique(pairs):
    obj = {}
    for key, value in pairs:
        require(key not in obj, "duplicate JSON field")
        obj[key] = value
    return obj


def decode(raw):
    return json.loads(raw, object_pairs_hook=unique)


def safe_name(name):
    path = PurePosixPath(name)
    require(name and not path.is_absolute() and ".." not in path.parts
            and "\\" not in name and str(path) == name, "unsafe artifact path")
    return name


def verify(folder):
    raw = (folder / "manifest.json").read_bytes()
    require(digest(raw) == MANIFEST_SHA256, "fixed manifest changed")
    manifest = decode(raw)
    archive = folder / "baseline.zip"
    require(archive.stat().st_size <= 32 << 20, "compressed archive exceeds budget")
    require(digest(archive.read_bytes()) == manifest["archive_sha256"], "fixed archive changed")
    with zipfile.ZipFile(archive) as bundle:
        members = bundle.infolist()
        require(len(members) <= 10000, "too many artifacts")
        require(sum(m.file_size for m in members) <= 128 << 20, "expanded archive exceeds budget")
        names = [safe_name(m.filename) for m in members]
        require(len(set(names)) == len(names), "duplicate archive entry")
        require(set(names) == set(manifest["files"]), "archive inventory changed")
        for member in members:
            require(0 <= member.file_size <= 8 << 20 and not member.is_dir()
                    and member.external_attr >> 16 == 0o100644
                    and member.compress_type == zipfile.ZIP_DEFLATED, "invalid archive member")
            require(digest(bundle.read(member)) == manifest["files"][member.filename], "artifact bytes changed")

        def read(name):
            return decode(bundle.read(name))

        # Source identity is the actual checkout plus the full byte inventory;
        # recording a HEAD never pretends the initial worktree was clean.
        inputs = read("inputs.json")
        require(inputs["head"] == manifest["tested_head"], "tested checkout changed")
        require(inputs["production_changed"] == [], "production sources changed during capture")
        require(inputs["initial_inventory_sha256"] == digest(bundle.read("initial-inventory.json")), "initial inventory changed")
        for name, sha in read("initial-inventory.json").items():
            archived = "source/" + safe_name(name)
            if archived in manifest["files"]:
                require(manifest["files"][archived] == sha, "original sources changed")

        node = "source/ai/internal/testkit/pioracle/node/"
        prov = read(node + "provenance.json")
        require(prov["pi_version"] == "1.1.0" and prov["pi_commit"] == PI_COMMIT
                and prov["npm_git_head"] == "" and prov["model_data_sha256"] == MODEL_DATA_SHA256,
                "oracle pins changed")
        package, lock = read(node + "package.json"), read(node + "package-lock.json")
        for name in ("@earendil-works/pi-ai", "@earendil-works/pi-telemetry"):
            require(package["dependencies"][name] == "1.1.0"
                    and lock["packages"]["node_modules/" + name]["version"] == "1.1.0", "oracle version is not unique")
        require(lock["packages"]["node_modules/@earendil-works/pi-ai"]["integrity"] == prov["tarball_integrity"], "release integrity changed")
        data_names = sorted(n.removeprefix("installed-model-data/") for n in names if n.startswith("installed-model-data/"))
        data_sums = "".join(digest(bundle.read("installed-model-data/" + n)) + "  " + n + "\n" for n in data_names).encode()
        require(len(data_names) == 43 and digest(data_sums) == MODEL_DATA_SHA256
                and data_sums == bundle.read(node + "model-data.sha256"), "installed model data changed")
        require(read("installed-packages.json")["pi-ai"] == "1.1.0", "installed oracle version changed")

        # The 2026-10-11 source rebuilds remain recorded proofs. Validate their
        # original inventories; this command performs no new source rebuild.
        proof = "source/.scratch/barness-ai-thinking-sampling/release-verification/"
        for name in ("evidence-b", "evidence-c"):
            directory = proof + name + "/"
            for line in bundle.read(directory + "SHA256SUMS").decode().splitlines():
                sha, file = line.split("  ", 1)
                require("/" not in safe_name(file) and digest(bundle.read(directory + file)) == sha, "historical source proof changed")

        for entry in manifest["runs"]:
            directory = entry["directory"] + "/"
            recorded = read(directory + "manifest.json")
            require(recorded["git_commit"] == manifest["tested_head"], "run HEAD changed")
            counts = {}
            for case in recorded["cases"]:
                counts[case["status"]] = counts.get(case["status"], 0) + 1
                prefix = directory + safe_name(case["dir"]) + "/"
                for file, sha in case["artifacts"].items():
                    require(digest(bundle.read(prefix + safe_name(file))) == sha, "case artifact changed")
                assertions = read(prefix + "assertions.json")
                require(assertions["status"] == case["status"] and assertions["replay"] == case["replay"], "case status or replay changed")
                if case["status"] == "PASS":
                    require(all(a["pass"] for a in assertions["assertions"]), "false assertion marked PASS")
            require(counts == entry["statuses"] and len(recorded["cases"]) == entry["cases"], "run verdicts changed")
            require(not read(directory + "audit.json")["findings"], "recorded run audit failed")

        release = read("history/release-report.json")
        require(release["pass"] is False and manifest["recorded_release"] == "FAIL", "historical release failure erased")
        require(not read("audit.json")["findings"], "whole baseline audit failed")
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bundle", type=Path, default=Path(__file__).resolve().parent)
    parser.add_argument("--audit", action="store_true", help="also rerun the existing whole-bundle redaction audit")
    args = parser.parse_args()
    try:
        m = verify(args.bundle)
        if args.audit:
            with tempfile.TemporaryDirectory(prefix="jev-baseline-audit-") as directory:
                work = Path(directory)
                with zipfile.ZipFile(args.bundle / "baseline.zip") as bundle:
                    for member in bundle.infolist():
                        target = work / member.filename
                        target.parent.mkdir(parents=True, exist_ok=True)
                        target.write_bytes(bundle.read(member))
                root = Path(__file__).resolve().parents[3]
                result = subprocess.run(["go", "run", "./ai/internal/testkit/audit/cmd/auditbundle", str(work)], cwd=root, capture_output=True, timeout=60)
                require(result.returncode == 0, "fresh whole-bundle audit failed")
    except (OSError, ValueError, KeyError, zipfile.BadZipFile) as err:
        # OS exception text can disclose private host paths. Stable reasons
        # are enough at this public artifact boundary.
        reason = str(err) if isinstance(err, ValueError) else "baseline input missing or invalid"
        print("FAIL: " + reason)
        return 1
    print(f"PASS: {sum(r['cases'] for r in m['runs'])} fixed cases verified; recorded release FAIL; no live execution")
    return 0


if __name__ == "__main__":
    sys.exit(main())

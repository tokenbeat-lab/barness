#!/usr/bin/env python3
"""Verify fixed Decisions text evidence without accessing any Provider."""
import argparse
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import subprocess
import tempfile
import zipfile

# Anchored once after capture. A replay creates a new bundle; it never rewrites
# this reviewed evidence or changes a recorded red run into a green run.
MANIFEST_SHA256 = "56eebbe07a9c339d675c48ab94bc4399c22140bf8489d28b1d517458edaeb34c"


def require(ok, reason):
    if not ok:
        raise ValueError(reason)


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def unique(pairs):
    value = {}
    for key, item in pairs:
        require(key not in value, "duplicate JSON key")
        value[key] = item
    return value


def read_json(raw):
    return json.loads(raw, object_pairs_hook=unique)


def safe(name):
    path = PurePosixPath(name)
    require(name and not path.is_absolute() and ".." not in path.parts
            and "\\" not in name and str(path) == name, "unsafe artifact name")
    return name


def bounded_bytes(path, limit):
    with path.open("rb") as source:
        raw = source.read(limit + 1)
    require(len(raw) <= limit, "artifact exceeds read budget")
    return raw


def verify(folder):
    raw = bounded_bytes(folder / "manifest.json", 8 << 20)
    require(sha(raw) == MANIFEST_SHA256, "fixed manifest changed")
    manifest = read_json(raw)
    archive = bounded_bytes(folder / "text-choice.zip", 32 << 20)
    require(sha(archive) == manifest["archive_sha256"], "fixed archive changed")
    # The same owned bytes survive verification and the later audit; a
    # concurrent path replacement cannot change which members get extracted.
    with zipfile.ZipFile(io.BytesIO(archive)) as bundle:
        members = bundle.infolist()
        require(len(members) <= 20000 and sum(m.file_size for m in members) <= 128 << 20,
                "expanded evidence exceeds budget")
        names = [safe(m.filename) for m in members]
        require(len(names) == len(set(names)) and set(names) == set(manifest["files"]),
                "artifact inventory changed")
        for member in members:
            require(not member.is_dir() and member.file_size <= 8 << 20
                    and member.external_attr >> 16 == 0o100644
                    and member.compress_type == zipfile.ZIP_DEFLATED, "invalid archive member")
            require(sha(bundle.read(member)) == manifest["files"][member.filename],
                    "artifact bytes changed")
        decode = lambda name: read_json(bundle.read(name))
        require(not decode("audit.json")["findings"], "whole evidence audit failed")
        require(manifest["route"] == "classifier/openai/openai-decisions/host-decisions"
                and manifest["live"] == "NOT_RUN", "synthetic evidence mislabeled")
        green = 0
        for run in manifest["runs"]:
            prefix = safe(run["directory"]) + "/"
            recorded = decode(prefix + "manifest.json")
            counts = {}
            for case in recorded["cases"]:
                counts[case["status"]] = counts.get(case["status"], 0) + 1
                case_prefix = prefix + safe(case["dir"]) + "/"
                for name, digest in case["artifacts"].items():
                    require(sha(bundle.read(case_prefix + safe(name))) == digest,
                            "case artifact changed")
                assertions = decode(case_prefix + "assertions.json")
                require(assertions["status"] == case["status"]
                        and assertions["replay"] == case["replay"], "case verdict changed")
                if case["status"] == "PASS":
                    require(all(a["pass"] for a in assertions["assertions"]),
                            "failed assertion labeled PASS")
            require(counts == run["statuses"] and len(recorded["cases"]) == run["cases"],
                    "run verdict changed")
            require(not decode(prefix + "audit.json")["findings"], "run audit failed")
            if run["role"] == "green":
                require("FAIL" not in counts and counts.get("PASS", 0) > 0,
                        "green run contains failures")
                green += counts.get("PASS", 0)
            else:
                require(run["role"] == "red" and counts.get("FAIL", 0) > 0,
                        "recorded red failure lost")
        require(green > 0, "no green public evidence")
    return manifest, green, archive


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bundle", type=Path, default=Path(__file__).resolve().parent)
    parser.add_argument("--audit", action="store_true")
    args = parser.parse_args()
    try:
        manifest, cases, archive = verify(args.bundle)
        if args.audit:
            with tempfile.TemporaryDirectory(prefix="decisions-text-audit-") as directory:
                work = Path(directory)
                with zipfile.ZipFile(io.BytesIO(archive)) as bundle:
                    for item in bundle.infolist():
                        target = work / item.filename
                        target.parent.mkdir(parents=True, exist_ok=True)
                        target.write_bytes(bundle.read(item))
                root = Path(__file__).resolve().parents[3]
                result = subprocess.run(["go", "run", "./ai/internal/testkit/audit/cmd/auditbundle", str(work)],
                                        cwd=root, capture_output=True, timeout=60)
                require(result.returncode == 0, "fresh whole evidence audit failed")
    except (OSError, ValueError, KeyError, zipfile.BadZipFile):
        # Fixed text avoids echoing host paths or potentially corrupt content.
        print("FAIL: missing, changed or invalid Decisions text evidence")
        return 1
    print(f"PASS: {cases} green public cases and preserved red runs; no live execution")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

"""One-off archive of this approved delivery; not a catalog CLI API."""
from pathlib import Path
from datetime import datetime, timezone
import hashlib
import json
import re
import shutil
import subprocess
import sys
import time

root = Path.cwd()
operation_path = Path(sys.argv[1])
info = json.loads(operation_path.read_text())
work = root / info["work"]
delivery = root / info["delivery"]
version = info["catalogVersion"]
assert re.fullmatch(r"catalog-[0-9a-f]{64}", version)
assert not delivery.exists()
delivery.mkdir(parents=True)


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def write_json(path, value):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("x") as out:
        json.dump(value, out, ensure_ascii=False, indent=2)
        out.write("\n")


def copy_new(source, target):
    source, target = Path(source), Path(target)
    assert source.is_file() and not source.is_symlink() and not target.exists()
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, target)
    assert digest(source) == digest(target)


def inventory(directory):
    result = {}
    for path in sorted(Path(directory).rglob("*")):
        assert not path.is_symlink()
        if path.is_file():
            result[path.relative_to(directory).as_posix()] = digest(path)
    return result


sys.path.insert(0, str(root / "tools/verification"))
from run_process import environment, isolate_go_home

commands = []


def run(label, argv):
    out = work / "commands" / label
    out.mkdir(parents=True)
    env, _ = environment("e2e")
    env.update(GOWORK="off", GOPROXY="off", GOSUMDB="off")
    isolate_go_home(env, "e2e", out)
    started = datetime.now(timezone.utc).isoformat()
    clock = time.monotonic()
    with (out / "stdout").open("xb") as stdout, (out / "stderr").open("xb") as stderr:
        process = subprocess.run(argv, cwd=root, env=env, stdout=stdout, stderr=stderr, timeout=180)
    record = {
        "argv": argv, "cwd": str(root), "startedAt": started,
        "finishedAt": datetime.now(timezone.utc).isoformat(),
        "elapsedSeconds": time.monotonic() - clock, "exitCode": process.returncode,
        "goenv": env["GOENV"], "goflags": env["GOFLAGS"], "gowork": env["GOWORK"],
        "stdoutSHA256": digest(out / "stdout"), "stderrSHA256": digest(out / "stderr"),
    }
    write_json(out / "command.json", record)
    for name in ("command.json", "stdout", "stderr"):
        copy_new(out / name, delivery / "verification/commands" / label / name)
    commands.append({"label": label, "record": f"verification/commands/{label}/command.json", "exitCode": process.returncode})
    if process.returncode:
        print((out / "stdout").read_text(), (out / "stderr").read_text())
        raise SystemExit(process.returncode)
    return (out / "stdout").read_bytes()


raw = (root / "ai/catalog_builtin.json").read_bytes()
catalog = json.loads(raw)
assert catalog["catalog"]["version"] == version
snapshot_name = version + ".json"
copy_new(root / "ai/catalog_builtin.json", delivery / snapshot_name)
(delivery / (snapshot_name + ".sha256")).write_text(digest(delivery / snapshot_name) + "  " + snapshot_name + "\n")
assert raw == (root / "local-catalog.json").read_bytes() == (root / "catalog/data/baseline/snapshot.json").read_bytes()

store = root / "artifacts/catalog/work/store"
assert not store.exists()
original_store = inventory(root / ".catalog")
shutil.copytree(root / ".catalog", store)
assert inventory(store) == original_store == inventory(root / ".catalog")
write_json(work / "store-copy.json", {
    "source": ".catalog", "destination": "artifacts/catalog/work/store",
    "files": original_store, "status": "exact-copy",
})

binary = work / "bin/catalogctl"
binary.parent.mkdir()
run("01-build-catalogctl", ["go", "build", "-o", str(binary), "./cmd/catalogctl"])
versions = json.loads(run("02-copied-store-versions", [str(binary), "versions", "--store", str(store)]))
assert versions["revision"] == "sha256:4047a1f3e757fd6d1f2a0378341478e2f034df05e40031e628112a0d9e78b5dc"
assert versions["versions"]["openai"] == 2
args = [str(binary), "export", "--store", str(store), "--repo", str(root), "--review-dir", str(delivery / "review")]
run("03-review-preview", [*args, "--dry-run"])
assert not (delivery / "review").exists()
run("04-review-export", [*args, "--confirm"])
assert inventory(delivery / "review") == inventory(root / "catalog/data/baseline")
assert (delivery / "review/snapshot.json").read_bytes() == raw
run("05-versioned-snapshot-check", [
    "go", "run", "./cmd/cataloggen", "--builtin", "--out", str(delivery / snapshot_name), "--check",
])

original_run = root / ".evidence/catalog-gpt61-terra/maintenance/20261102T003945Z-dfba5cb9"
for name in ("official-prompt-caching.md", "official-gpt6-guide.md"):
    copy_new(original_run / name, delivery / "provenance" / name)
    copy_new(original_run / (name + ".capture.json"), delivery / "provenance" / (name + ".capture.json"))
    metadata = json.loads((original_run / (name + ".capture.json")).read_text())
    assert digest(original_run / name) == metadata["sha256"]
copy_new(Path(__file__), delivery / "provenance/assembly.py")
copy_new(operation_path, delivery / "provenance/operation.json")
copy_new(work / "store-copy.json", delivery / "provenance/store-copy.json")

base, head = info["runtimeComparisonBase"], info["baselineCommit"]
changes = []
for line in subprocess.check_output(["git", "diff", "--name-status", base, head], text=True).splitlines():
    status, path = line.split("\t", 1)
    changes.append({"status": status, "path": path})
core = [
    x for x in changes
    if x["path"].endswith(".go") and not x["path"].endswith("_test.go") and x["path"].startswith("ai/")
    and not x["path"].startswith(("ai/e2e/", "ai/examples/", "ai/live/", "ai/release/", "ai/internal/testkit/"))
]
dependencies = [x for x in changes if x["path"] in ("go.mod", "go.sum")]
example_code = [x for x in changes if x["path"].endswith(".go") and x["path"].startswith("ai/examples/")]
maintenance_code = [
    x for x in changes if x["path"].endswith(".go")
    and x["path"].startswith(("catalog/", "cmd/catalogctl/", "cmd/cataloggen/"))
]
assert core == [] and dependencies == [] and maintenance_code == []
assert [x["path"] for x in example_code] == ["ai/examples/chineseeval/cmd/chineseeval/main.go"]
old_check = root / ".evidence/catalog-gpt61-terra/check-inputs/20261102T004145.325113000Z-2642263489/check.json"
checked = json.loads(old_check.read_text())
assert checked["measuredCommit"] == base and checked["conclusion"] == "code_compatible"
assert checked["snapshotDigest"] == "sha256:" + digest(delivery / snapshot_name)
copy_new(old_check, delivery / "verification/old-program-check.json")

result_path = root / ".evidence/catalog-gpt61-terra-builtin/verification/result-v1.json"
result = json.loads(result_path.read_text())
assert digest(result_path) == "4f2332c9b42251c21b3591d1312f516c617d4f58f8fae2577da7b512fa0316bd"
copy_new(result_path, delivery / "verification/result.json")
copy_new(result_path.parent / "report.md", delivery / "verification/report.md")
copy_new(root / result["ordinary"]["receipt"], delivery / "verification/receipt.json")
originals = []


def reference(path, role):
    path = Path(path)
    originals.append({"path": path.relative_to(root).as_posix(), "sha256": digest(path), "role": role})


reference(old_check, "original-code-compatibility-check")
reference(result_path, "final-verification-index")
reference(root / result["ordinary"]["receipt"], "ordinary-receipt")
for bundle in result["ordinary"]["bundles"]:
    reference(root / bundle["path"] / "manifest.json", "native-bundle-manifest")
    reference(root / bundle["path"] / "audit.json", "native-bundle-audit")
for record in result["buildChecks"]:
    reference(root / record["record"], "build-check-command")
for record in result["ordinaryScript"]["reconstructedVersions"]:
    reference(root / record["path"], "reconstructed-script-matched-to-receipt-digest")
reference(root / result["ordinaryScript"]["path"], "actual-final-orchestration-script")
reference(root / "docs/barness-ai/CHANGELOG.md", "source-changelog")
write_json(delivery / "verification/originals.json", {
    "selfContainedEvidence": False, "pathBase": "repository root", "files": originals,
    "note": "Original case capsules remain at immutable repository-relative paths; summaries are not complete capsules.",
})

code_report = {
    "comparisonBase": base, "comparisonHead": head, "testedCurrentCommit": result["testedCommit"],
    "coreRuntimeCodeChanged": "NO", "coreRuntimeFiles": core,
    "runtimeDependenciesChanged": "NO", "dependencyFiles": dependencies,
    "maintenanceImplementationChanged": "NO", "maintenanceFiles": maintenance_code,
    "exampleCodeChanges": example_code, "otherChanges": changes,
    "externalJson": {
        "decision": "JSON_ONLY_FOR_TESTED_EXTERNAL_LOADER", "checkedProgramCommit": base,
        "evidence": "verification/old-program-check.json",
        "evidenceKind": "Actual historical code compatibility measurement on exact snapshot bytes",
        "conditions": ["Host explicitly loads the file and injects Config.Catalog", "Reload and create a new Client"],
        "arbitraryOlderReleasesVerified": False, "sdkUpgradeRequiredForThisModelOnCheckedProgram": False,
    },
    "builtinDefault": {
        "decision": "REBUILD_AND_DISTRIBUTE_PROGRAM",
        "reason": "go:embed defaults are compiled into existing binaries even when runtime Go code is unchanged",
    },
    "exampleImpact": "ChineseEvaluation defaults select v4; only config version and catalog pins change; historical configs remain immutable.",
    "newVendorQualification": False, "softwareVersionPublished": False,
}
write_json(delivery / "code-and-distribution.json", code_report)
notes = f"""# GPT-6.1 Terra catalog delivery

Catalog.Version: {version}.

The approved directory contains 119 identities (113 chat, 3 image, 3 classifier).
Only chat/openai/openai-responses/gpt-6.1-terra is added; the prior 118 definitions
and their relative order are unchanged. OpenAI content version is 2.

Runtime Go code and dependencies are unchanged between {base} and {head}.
The only non-test Go change is the ChineseEvaluation example default config
selection from v3 to v4. Its previous configs and snapshots remain immutable.

External JSON: the exact snapshot passed code compatibility on {base}.
A Host with external load/injection can distribute JSON and create a new Client;
this does not establish compatibility with arbitrary older releases.
Builtin default: rebuild and distribute the program to replace embedded data.

28 required native cases passed; full ordinary verification, generate, build and
vet exited 0. All four bundle audits have zero findings. Live/race/pressure are
not applicable for this scope. No vendor qualification or software release is claimed.

This batch applies the Catalog CLI delivery SOP: versioned immutable runtime JSON,
exact review inventory, changelog, code/distribution decisions and verification
references are collected together. Original capsules remain at the locations
listed in verification/originals.json.
"""
(delivery / "CHANGELOG.md").write_text(notes)
(delivery / "delivery.md").write_text(f"""# 代码变化与分发判定

- 目录版本：{version}。
- 核心运行代码变化：NO；依赖变化：NO。
- 比较基准：{base}；比较目标：{head}。
- 其他 Go 变化：ChineseEvaluation 示例 CLI 默认配置 v3 → v4；测试另行修订。
- 外部 JSON：已在上述基准代码实际检查相同原字节；Host 有外部加载/注入能力时，
  可只分发 {snapshot_name}，重新加载并创建新 Client；任意旧发布版未验证。
- 编译内置默认：需要重新构建并分发程序，因为目录由 go:embed 编入二进制。
- 示例默认、JSON 分发、软件发布与厂商真实资格分别记录，本次没有软件发布或厂商资格。

精确文件差异、条件及证据见 code-and-distribution.json；28 项必需场景、完整普通
回执与原件摘要见 verification/。当前目录不包含所有原生 case 胶囊。
""")

files = inventory(delivery)
assert "delivery.json" not in files
manifest = {
    "schema": 1, "status": "complete", "createdAt": datetime.now(timezone.utc).isoformat(),
    "catalogVersion": version, "catalogHash": catalog["hash"],
    "maintenanceRevision": versions["revision"], "providerVersions": versions["versions"],
    "snapshot": {"path": snapshot_name, "sha256": digest(delivery / snapshot_name), "bytes": len(raw)},
    "review": {"path": "review/review.json", "sha256": digest(delivery / "review/review.json"), "selfContainedSources": True},
    "changelog": {
        "path": "CHANGELOG.md", "kind": "Self-contained summary of the unreleased entry",
        "source": "docs/barness-ai/CHANGELOG.md", "sourceSHA256": digest(root / "docs/barness-ai/CHANGELOG.md"),
    },
    "codeAndDistribution": {
        "path": "code-and-distribution.json", "coreRuntimeCodeChanged": "NO", "runtimeDependenciesChanged": "NO",
        "externalJson": "JSON_ONLY_FOR_TESTED_EXTERNAL_LOADER", "builtinDefault": "REBUILD_AND_DISTRIBUTE_PROGRAM",
    },
    "verification": {
        "result": "verification/result.json", "originals": "verification/originals.json", "selfContainedEvidence": False,
        "ordinaryStatus": result["ordinary"]["status"], "ordinaryExitCode": result["ordinary"]["exitCode"],
        "requiredNativeCasesPassed": result["ordinary"]["requiredCaseCount"], "applicability": result["applicability"],
    },
    "commands": commands, "files": files,
}
write_json(delivery / "delivery.json", manifest)
assert set(inventory(delivery)) == set(manifest["files"]) | {"delivery.json"}
assert all(digest(delivery / name) == expected for name, expected in manifest["files"].items())
summary = {
    "status": "PASS", "delivery": info["delivery"], "manifestSHA256": digest(delivery / "delivery.json"),
    "snapshotSHA256": manifest["snapshot"]["sha256"], "files": len(files) + 1,
    "workStoreExactCopy": True, "publicSnapshotCheckExitCode": 0,
}
write_json(work / "assembly-result.json", summary)
print(json.dumps(summary))

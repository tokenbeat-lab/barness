#!/usr/bin/env python3
"""Validate the proposed ticket documents; never runs product tests."""
from pathlib import Path
import hashlib
import json
import re

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
FEATURE = HERE.parent
proposal = json.loads((HERE / "proposal.json").read_text())
assert proposal["status"] == "DRAFT_PENDING_APPROVAL"
assert proposal["baseline"] == "8ea61b8d70de6c84fe27be34dc7d21c5c40841d9"

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

for rel, expected in proposal["originalHashes"].items():
    assert digest(ROOT / rel) == expected, ("original document changed", rel)

spec = (FEATURE / "spec.md").read_text()
acceptance = (FEATURE / "acceptance.md").read_text()
required_r = set(re.findall(r"^### (R\d{2})", spec, re.M))
required_od = set(re.findall(r"^\| (OD\d{2}) \|", acceptance, re.M))
assert required_r == {f"R{i:02d}" for i in range(1, 21)}
assert required_od == {f"OD{i:02d}" for i in range(1, 59)}

tickets = proposal["tickets"]
assert [t["id"] for t in tickets] == [f"{i:02d}" for i in range(1, 14)]
assert len({t["file"] for t in tickets}) == len(tickets)
actual_files = {str(p.relative_to(HERE)) for p in (HERE / "drafts").glob("*.md")}
assert actual_files == {t["file"] for t in tickets}
ids = {t["id"] for t in tickets}
graph = {t["id"]: t["blockedBy"] for t in tickets}
covered_r, covered_od, sources = set(), set(), set()
for t in tickets:
    path = HERE / t["file"]
    body = path.read_text()
    assert path.name.startswith(t["id"] + "-")
    assert body.startswith("# " + t["id"] + ": " + t["title"] + "\n")
    assert "**Status:** needs-triage" in body
    assert "**Publication:** draft" in body
    assert "**What to build:**" in body
    assert "## Responsibility" in body
    assert "## Failure scenarios first" in body
    assert "## Acceptance criteria" in body and body.count("- [ ]") >= 5
    assert "## Completion" in body
    assert body.index("## Failure scenarios first") < body.index("## Acceptance criteria")
    assert re.search(r"先|生产变更前", body)
    assert "manifest" in body.lower() and re.search(r"审计|audit", body, re.I)
    assert not re.search(r"(?:^|[\s`])(?:ai/|docs/|\.scratch/)", body, re.M), path
    blockers = re.search(r"^\*\*Blocked by:\*\* (.+)$", body, re.M).group(1)
    if t["blockedBy"]:
        assert re.findall(r"(\d{2}):", blockers) == t["blockedBy"]
        for b in t["blockedBy"]:
            assert b in ids and int(b) < int(t["id"]), (t["id"], b)
            title = next(x["title"] for x in tickets if x["id"] == b)
            assert b + ": " + title in blockers
    else:
        assert blockers == "None (can start immediately)"
    assert len(t["blockedBy"]) == len(set(t["blockedBy"]))
    assert len(t["requirements"]) == len(set(t["requirements"]))
    assert len(t["cases"]) == len(set(t["cases"]))
    assert set(t["requirements"]) <= required_r and set(t["cases"]) <= required_od
    expected_fields = {
        "Source tickets": t["sourceTickets"],
        "Requirements": t["requirements"],
        "Acceptance cases": t["cases"],
    }
    for key, values in expected_fields.items():
        field = re.search(r"^\*\*" + key + r":\*\* (.+)$", body, re.M).group(1)
        assert field.split(", ") == values, (t["id"], key)
    covered_r.update(t["requirements"])
    covered_od.update(t["cases"])
    sources.update(t["sourceTickets"])
assert covered_r == required_r
assert covered_od == required_od
assert sources == {f"{i:02d}" for i in range(1, 10)}

def ancestors(node, seen=None):
    seen = set() if seen is None else seen.copy()
    assert node not in seen, ("cycle", node)
    seen.add(node)
    result = set(graph[node])
    for dep in graph[node]:
        result.update(ancestors(dep, seen))
    return result

for node, deps in graph.items():
    ancestors(node)
    for dep in deps:
        other_deps = [b for b in deps if b != dep]
        assert all(dep not in ancestors(b) for b in other_deps), ("redundant blocker", node, dep)
frontier = [t["id"] for t in tickets if not t["blockedBy"]]
assert frontier == ["01"]
report = (HERE / "README.md").read_text()
assert set(re.findall(r"^\| (R\d{2}) \|", report, re.M)) == required_r
assert set(re.findall(r"^\| (OD\d{2}) \|", report, re.M)) == required_od

docs = sorted(HERE.rglob("*.md"))
for path in docs:
    body = path.read_text()
    assert len(body.splitlines()) <= 500, path
    prose = re.sub(r"~~~[^\n]*\n.*?\n~~~", "", body, flags=re.S)
    for target in re.findall(r"\]\(([^)]+)\)", prose):
        if re.match(r"^[a-z][a-z0-9+.-]*://", target):
            continue
        target = re.sub(r":\d+$", "", target.split("#", 1)[0])
        assert not target or (path.parent / target).resolve().exists(), (path, target)

manifest_paths = docs + [HERE / "proposal.json", HERE / "verify.py"]
result = {
    "result": "PASS: ticket documentation checks only",
    "date": proposal["date"],
    "baseline": proposal["baseline"],
    "publication": proposal["status"],
    "tickets": len(tickets),
    "requirements": len(required_r),
    "acceptance_cases": len(required_od),
    "original_documents_unchanged": len(proposal["originalHashes"]),
    "original_ticket_count": 9,
    "dag_acyclic": True,
    "redundant_direct_edges": 0,
    "initial_frontier": frontier,
    "product_tests": "NOT_RUN",
    "paid_api": "NOT_RUN",
    "sha256": {str(p.relative_to(ROOT)): digest(p) for p in manifest_paths},
}
(HERE / "validation.json").write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({k: v for k, v in result.items() if k != "sha256"}, ensure_ascii=False, indent=2))

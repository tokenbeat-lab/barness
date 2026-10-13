#!/usr/bin/env python3
"""Verify published ticket documents and approval provenance; no product tests."""
from pathlib import Path
import hashlib
import json
import re

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
FEATURE = HERE.parent
proposal = json.loads((HERE / "proposal.json").read_text())
publication = json.loads((HERE / "publication.json").read_text())
before = json.loads((HERE / "originals/ticket-review/validation.json").read_text())
assert proposal["status"] == "DRAFT_PENDING_APPROVAL"  # Immutable proposal record.
assert publication["status"] == "APPROVED_PUBLISHED"
assert publication["authorization"] == "批准 13 张草案并替换旧票"
assert proposal["baseline"] == "8ea61b8d70de6c84fe27be34dc7d21c5c40841d9"

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

assert digest(HERE / "proposal.json") == before["sha256"][str((HERE / "proposal.json").relative_to(ROOT))]
assert set(publication["originalArchives"]) == set(proposal["originalHashes"])
assert set(publication["protectedDocuments"]) == {
    str((FEATURE / "spec.md").relative_to(ROOT)),
    str((FEATURE / "design.md").relative_to(ROOT)),
}
for source, expected in proposal["originalHashes"].items():
    archive = HERE / publication["originalArchives"][source]
    assert digest(archive) == expected, ("original archive changed", source)
for source, expected in publication["protectedDocuments"].items():
    assert expected == proposal["originalHashes"][source]
    assert digest(ROOT / source) == expected, ("protected contract changed", source)
for name in ["README.md", "verify.py"]:
    old_source = str((HERE / name).relative_to(ROOT))
    assert digest(HERE / "originals/ticket-review" / name) == before["sha256"][old_source]

spec = (FEATURE / "spec.md").read_text()
acceptance = (FEATURE / "acceptance.md").read_text()
plan = (FEATURE / "implementation-plan.md").read_text()
for header in ["Problem Statement", "Solution", "User Stories", "Implementation Decisions",
               "Testing Decisions", "Out of Scope", "Further Notes"]:
    assert "## " + header in spec
stories = spec.split("## User Stories\n")[1].split("## Implementation Decisions")[0]
assert len(re.findall(r"^\d+\. As an? .+, I want .+, so that .+", stories, re.M)) == 56
required_r = set(re.findall(r"^### (R\d{2})", spec, re.M))
required_od = set(re.findall(r"^\| (OD\d{2}) \|", acceptance, re.M))
assert required_r == {f"R{i:02d}" for i in range(1, 21)}
assert re.findall(r"^\| (OD\d{2}) \|", acceptance, re.M) == [f"OD{i:02d}" for i in range(1, 59)]
tickets = proposal["tickets"]
assert [t["id"] for t in tickets] == [f"{i:02d}" for i in range(1, 14)]
ids = {t["id"] for t in tickets}
assert set(publication["ticketFiles"]) == ids
actual_issues = {p.name for p in (FEATURE / "issues").glob("*.md")}
assert actual_issues == {Path(t["file"]).name for t in tickets}
assert {str(p.relative_to(HERE)) for p in (HERE / "drafts").glob("*.md")} == {t["file"] for t in tickets}
assert set(publication["approvedDraftHashes"]) == {str((HERE / t["file"]).relative_to(ROOT)) for t in tickets}
graph = {t["id"]: t["blockedBy"] for t in tickets}
covered_r, covered_od, sources = set(), set(), set()

def field(body, name):
    found = re.search(r"^\*\*" + re.escape(name) + r":\*\* (.+)$", body, re.M)
    assert found, name
    return found.group(1)

for t in tickets:
    draft_path = HERE / t["file"]
    draft_key = str(draft_path.relative_to(ROOT))
    assert publication["approvedDraftHashes"][draft_key] == before["sha256"][draft_key]
    assert digest(draft_path) == publication["approvedDraftHashes"][draft_key]
    draft = draft_path.read_text()
    path = (HERE / publication["ticketFiles"][t["id"]]).resolve()
    assert path.parent == (FEATURE / "issues").resolve()
    assert path.name == Path(t["file"]).name
    body = path.read_text()
    assert body.startswith("# " + t["id"] + ": " + t["title"] + "\n")
    assert field(body, "Status") == "ready-for-agent"
    assert field(body, "Type") == "task"
    assert field(body, "What to build") == field(draft, "What to build")
    for key, expected in {
        "Source tickets": t["sourceTickets"],
        "Requirements": t["requirements"],
        "Acceptance cases": t["cases"],
    }.items():
        assert field(body, key).split(", ") == expected, (t["id"], key)
    assert field(body, "Blocked by") == field(draft, "Blocked by")
    if t["blockedBy"]:
        assert re.findall(r"(\d{2}):", field(body, "Blocked by")) == t["blockedBy"]
        for dep in t["blockedBy"]:
            assert dep in ids and int(dep) < int(t["id"])
            title = next(x["title"] for x in tickets if x["id"] == dep)
            assert dep + ": " + title in field(body, "Blocked by")
    else:
        assert field(body, "Blocked by") == "None (can start immediately)"
    assert len(t["blockedBy"]) == len(set(t["blockedBy"]))
    approved_core = draft.split("## Responsibility\n", 1)[1].strip()
    published_core = body.split("## Responsibility\n", 1)[1].split("\n## Comments\n", 1)[0].strip()
    assert published_core == approved_core, ("approved scope changed", t["id"])
    assert body.index("## Failure scenarios first") < body.index("## Acceptance criteria")
    assert body.count("- [ ]") >= 5
    assert "manifest" in body.lower() and re.search(r"审计|audit", body, re.I)
    for old in t["sourceTickets"]:
        old_path = next(HERE / archive for source, archive in publication["originalArchives"].items()
                        if source.startswith(str(FEATURE.relative_to(ROOT)) + "/issues/" + old + "-"))
        original = old_path.read_text()
        if "\n## Comments\n" in original:
            assert original.split("\n## Comments\n", 1)[1].strip() in body
        assert "### 原票 " + old in body
    assert set(t["requirements"]) <= required_r and set(t["cases"]) <= required_od
    covered_r.update(t["requirements"])
    covered_od.update(t["cases"])
    sources.update(t["sourceTickets"])
assert covered_r == required_r and covered_od == required_od
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
        assert all(dep not in ancestors(other) for other in deps if other != dep), ("redundant blocker", node, dep)
frontier = [t["id"] for t in tickets if not t["blockedBy"]]
assert frontier == ["01"]

ticket_section = plan.split("## 2. 工单与阻塞关系")[1].split("## 3. 阶段门槛")[0]
plan_rows = [line for line in ticket_section.splitlines() if re.match(r"^\| \d{2} \|", line)]
assert len(plan_rows) == 13
for t, line in zip(tickets, plan_rows):
    parts = [part.strip() for part in line.split("|")]
    assert parts[1] == t["id"]
    assert "](issues/" + Path(t["file"]).name + ")" in parts[2]
    assert parts[3] == (", ".join(t["blockedBy"]) or "无")
requirement_rows = {re.match(r"^\| (R\d{2}) \|", line).group(1): line
                    for line in plan.splitlines() if re.match(r"^\| (R\d{2}) \|", line)}
assert set(requirement_rows) == required_r
for requirement, line in requirement_rows.items():
    owners = line.split("|")[2].strip().split(", ")
    assert owners == [t["id"] for t in tickets if requirement in t["requirements"]]
original_acceptance = (HERE / "originals/acceptance.md").read_text()
original_rows = {re.match(r"^\| (OD\d{2}) \|", line).group(1): line
                 for line in original_acceptance.splitlines() if re.match(r"^\| (OD\d{2}) \|", line)}
for line in acceptance.splitlines():
    match = re.match(r"^\| (OD\d{2}) \|", line)
    if not match:
        continue
    case = match.group(1)
    parts, old_parts = line.split("|"), original_rows[case].split("|")
    assert parts[:-2] == old_parts[:-2], ("case behavior changed", case)
    requirement, owner_text = parts[-2].strip().split(" / ")
    assert requirement == old_parts[-2].strip().split(" / ")[0]
    assert owner_text.split(",") == [t["id"] for t in tickets if case in t["cases"]]
review = (HERE / "README.md").read_text()
assert set(re.findall(r"^\| (R\d{2}) \|", review, re.M)) == required_r
assert set(re.findall(r"^\| (OD\d{2}) \|", review, re.M)) == required_od
assert "approved and published" in review and "## 批准与发布记录" in review

all_docs = sorted(FEATURE.rglob("*.md"))
for path in all_docs:
    assert len(path.read_text().splitlines()) <= 500, path
active_docs = [path for path in all_docs if "originals" not in path.relative_to(FEATURE).parts]
for path in active_docs:
    prose = re.sub(r"(?:~~~|`{3})[^\n]*\n.*?\n(?:~~~|`{3})", "", path.read_text(), flags=re.S)
    for target in re.findall(r"\]\(([^)]+)\)", prose):
        if re.match(r"^[a-z][a-z0-9+.-]*://", target):
            continue
        target = re.sub(r":\d+$", "", target.split("#", 1)[0])
        assert not target or (path.parent / target).resolve().exists(), (path, target)
manifest_paths = sorted(path for path in FEATURE.rglob("*") if path.is_file() and path != HERE / "validation.json")
result = {
    "result": "PASS: published ticket documentation checks only",
    "date": publication["date"],
    "baseline": proposal["baseline"],
    "publication": publication["status"],
    "tickets": len(tickets),
    "requirements": len(required_r),
    "acceptance_cases": len(required_od),
    "archived_original_documents": len(proposal["originalHashes"]),
    "protected_contracts_unchanged": len(publication["protectedDocuments"]),
    "approved_drafts_unchanged": len(publication["approvedDraftHashes"]),
    "original_comments_preserved": True,
    "plan_and_acceptance_owners_match": True,
    "dag_acyclic": True,
    "redundant_direct_edges": 0,
    "initial_frontier": frontier,
    "product_tests": "NOT_RUN",
    "paid_api": "NOT_RUN",
    "sha256": {str(path.relative_to(ROOT)): digest(path) for path in manifest_paths},
}
(HERE / "validation.json").write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({key: value for key, value in result.items() if key != "sha256"}, ensure_ascii=False, indent=2))

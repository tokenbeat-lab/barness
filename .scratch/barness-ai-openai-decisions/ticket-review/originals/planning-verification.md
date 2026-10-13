# 设计与计划的文档核验

Updated: 2026-10-12
Scope: planning documents only

已确认测试 Seam 为公共 Client.Classify / HookedClient.Classify。此次交付是 spec/design/验收/工单及拟议 ADR，没有生产代码、正式 Catalog、支持矩阵或真实 API 的实施。

核验项目：to-spec 七个模板章节；56 条完整 User Story；R01–R20 到实施计划全部映射；OD01–OD58 连续唯一；01–09 独立 ready-for-agent 工单及无环依赖；相对文件链接存在；单文档不超过 500 行；spec 的 Implementation Decisions 无具体文件路径或代码片段。

人工一致性审查还检查：Jev 的 scope 错误顺序和旧构造范围、公共 2–10 与 API 上限的区分、图片三条件及中间阶段失败关闭、逐题/全拒答、native 回调冻结、usage/cost、candidate/正式报告两次门槛和旧九组合保留。未给 runtime/live/质量或功能完成填 PASS。

以下从 repo root 重放的脚本只读取文档，输出检查数和 SHA-256 manifest；manifest 不包括本报告，避免自引用。它不能替代实际公共 E2E、差分、race 或发布 gate。

~~~sh
python3 - <<'PY'
from pathlib import Path
import hashlib, json, re

root = Path.cwd()
base = root / ".scratch/barness-ai-openai-decisions"
docs = sorted(p for p in base.rglob("*.md") if p.name != "planning-verification.md")
docs += [root / "docs/adr/0026-barness-ai-openai-decisions-classification.md", root / "GLOSSARY.md"]
for p in docs:
    text = p.read_text()
    assert len(text.splitlines()) <= 500, p
    fence = chr(96) * 3 + "|~~~"
    prose = re.sub(r"(?:" + fence + r")[^\n]*\n.*?\n(?:" + fence + r")", "", text, flags=re.S)
    for target in re.findall(r"\]\(([^)]+)\)", prose):
        if re.match(r"^[a-z][a-z0-9+.-]*://", target):
            continue
        target = re.sub(r":\d+$", "", target.split("#", 1)[0])
        assert not target or (p.parent / target).resolve().exists(), (p, target)
spec = (base / "spec.md").read_text()
headers = ["Problem Statement", "Solution", "User Stories", "Implementation Decisions",
           "Testing Decisions", "Out of Scope", "Further Notes"]
assert all("## " + h in spec for h in headers)
stories = spec.split("## User Stories\n")[1].split("## Implementation Decisions")[0]
assert len(re.findall(r"^\d+\. As an? .+, I want .+, so that .+", stories, re.M)) == 56
impl = spec.split("## Implementation Decisions\n")[1].split("## Testing Decisions")[0]
assert not re.search(r"ai/|docs/|\.scratch/|~~~", impl) and chr(96) * 3 not in impl
requirements = set(re.findall(r"^### (R\d{2})", spec, re.M))
plan = (base / "implementation-plan.md").read_text()
assert len(requirements) == 20
assert set(re.findall(r"^\| (R\d{2}) \|", plan, re.M)) == requirements
cases = re.findall(r"^\| (OD\d{2}) \|", (base / "acceptance.md").read_text(), re.M)
assert cases == [f"OD{i:02d}" for i in range(1, 59)]
tickets = sorted((base / "issues").glob("*.md"))
assert len(tickets) == 9
for i, p in enumerate(tickets, 1):
    text = p.read_text()
    assert p.name.startswith(f"{i:02d}-")
    assert "Status: ready-for-agent" in text
    dependency = re.search(r"^Depends on: (.+)$", text, re.M).group(1)
    assert dependency == ("none" if i == 1 else f"{i-1:02d}")
print(json.dumps({
    "result": "PASS: documentation checks only",
    "documents": len(docs), "stories": 56, "requirements": 20, "cases": 58, "tickets": 9,
    "sha256": {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in docs}
}, ensure_ascii=False, indent=2))
PY
~~~

实施状态仍以各工单、实际 E2E artifact、live/effect 报告和正式门禁为准。本报告只确认计划结构、引用和基本一致性可重放。

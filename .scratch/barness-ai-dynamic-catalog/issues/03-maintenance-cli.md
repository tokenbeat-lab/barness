# 确定性维护 CLI、来源映射和生命周期

Type: task
Status: claimed
Blocked by: 02

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责 cmd/catalogctl/、catalog/maintenance/ 的来源冻结、四层映射、候选存储、预览确认及新增/批量改价/下架/恢复/回退/版本历史。公共规则和 Provider 映射明确且保留缺失/有损/冲突/确认项；人工价格不被覆盖，缺价不补零。所有变更有只读 dry-run/help、JSON 与明确确认，原子保存且候选/历史不覆盖。实现 check/apply 的公共接口供 04 完成，不以自报 PASS 放行。

## Acceptance and verification

T02、T03、T04、T06、T07；真实 CLI 进程与固定来源、文件前后对照。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。

## Failure modes and agreed verification seams

Before implementation: compiled `catalogctl` process → immutable candidate/store/query boundary. Confirmed by spec Testing Decisions. Root owns aggregate current status. Baseline f8e234c928c5d901361707322739ed3425e29563; worktree initially clean.

Failures: help/query/dry-run creates files; missing confirmation writes; unknown arguments return success; unknown source/provider/type/modality or fractional limits silently map; canonical supplies another serving price; missing price becomes zero; duplicate identity overwritten; manual override lost; price refresh changes capabilities; missing/deprecated removes records; failed candidate save leaves partial object; preview/query advances versions; rollback rewrites old version; tampered candidate accepted; output/save failure advances current. Check/apply execution/receipt/cancellation authority is ticket04, with no public boolean bypass.

Ordinary checks: new CLI public E2E via tools/verification/run.py, explicit inputs for catalog and cmd code/data. No live/race/pressure. Each red and green attempt retained.

### Tested base slice (2026-11-01)

Implemented read-only help/query/preview, explicit confirmation, immutable candidate freezing, deterministic field diffs, pinned serving-source mapping and structured rejection. Does not yet complete cost/lifecycle/rollback integration; ticket remains claimed until public check/apply supports the complete workflows.

- Red absent CLI: `.evidence/barness-ai-dynamic-catalog/e2e/20261101T094405.670751Z-ovbigd2v/receipt.json` FAIL, actual exit 1.
- First read-only green: `20261101T094450.553690Z-ebcsvk8d` PASS/0.
- Freeze red/green: `20261101T094525.172282Z-e0d7ow6e` FAIL/1 → `20261101T094542.729160Z-3fs3bbz_` PASS/0.
- Mapping red/green: `20261101T094616.077064Z-wl0__980` FAIL/1 → `20261101T094718.733120Z-a8fqg1hm` PASS/0.
- Field diff red/green: `20261101T094828.330841Z-4bokcpiw` FAIL/1 → `20261101T094851.275222Z-ujj5zm3u` PASS/0.
- Latest command: `go test -json ./ai/e2e -run '^TestCatalogMaintenanceCLI(ReadOnly|Freeze|SourceMapping)$' -count=1 -timeout=5m`, through run.py with each catalog/cmd Go/JSON input explicitly bound. Actual required IDs CATALOG-T04-cli-read-only, CATALOG-T06-cli-freeze-immutable, CATALOG-T02-cli-serving-mapping completed; native manifest/audit retained under that attempt. Actual tested HEAD f8e234c928c5d901361707322739ed3425e29563 plus captured uncommitted source/input inventories; this is not a final frozen delivery receipt.

### Explicit revision and approved export slice

- Export/generator red `20261101T095007.116625Z-3r1dm3b7` FAIL/1 → green `20261101T095043.269036Z-90aa2rlk` PASS/0.
- Manual revise red `20261101T095103.178533Z-chrhdnha` FAIL/1 → five-group green `20261101T095123.587497Z-y8wwtrcq` PASS/0.
- Candidate object substitution red `20261101T095209.404824Z-42pw1k45` FAIL/1; following six-group receipt retained below.
- `catalogctl export` outputs selected approved Snapshot bytes or State (`--state`) without writes. `cataloggen` accepts `--state`, `--store`, or explicit `--builtin`; writes require explicit source/confirmation and atomic replacement. `go generate ./ai` now verifies reviewed checked-in source with `--check`, so it cannot reset an applied artifact.

Six-group command `go test -json ./ai/e2e -run '^TestCatalogMaintenanceCLI(ReadOnly|Freeze|SourceMapping|Export|Revise|CandidateTamper)$' -count=1 -timeout=5m` via collector: `.evidence/barness-ai-dynamic-catalog/e2e/20261101T095220.314644Z-e32kloj8/receipt.json` PASS/actual exit0; explicit catalog/cmd input inventory and native audit retained.

### Source edge coverage

- Projected tier gap: red `.evidence/barness-ai-dynamic-catalog/e2e/20261101T095330.594947Z-znsu4ts_/receipt.json` FAIL/1; source/tier green `20261101T095350.379015Z-54es5fgw` PASS/0.
- Broader boundary receipt `20261101T095425.364277Z-vagd4bc9` PASS/0 verifies explicit moonshotai/moonshotai-cn/zhipuai association, regional mismatch refusal, missing type chat-only confirmation, explicit independent API associations, decision→classifier capability confirmation and source-unavailable category.
- New evidence uses unique per-invocation artifact names even for identical argv, preserving all observations after changed input files. Original older attempts remain unchanged.

### Paused at integration checkpoint (ticket still claimed)

- Source-edge code commit `a69c77a1`; all eight CLI tests matched by `^TestCatalogMaintenanceCLI` PASS/0 at `.evidence/barness-ai-dynamic-catalog/e2e/20261101T095546.467337Z-h6sn0c6p/receipt.json` (tested committed source a69c77a1).
- Empty-overlay regression red `20261101T095646.010173Z-3llwiz5t` FAIL/1 → green `20261101T095702.081540Z-pgw4iet1` PASS/0; code fix commit `72737058` avoids candidates for property-order-only rewrites.
- Peer-merged ticket04 checkpoint `3bbb2c05` for subsequent public approved-state workflows; combined integration is not claimed fully verified by prior receipts.
- Remaining03: actual check/apply→cost-update/manual ownership protection/missing price and capability preservation; withdraw→restore→rollback new versions/history; current revision Snapshot tamper rejection; optional per-Provider reviewed source export. Root coordinates next turn after merger.

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

### Approved-state workflows and raw-input boundaries

All paths below are independent `.evidence/barness-ai-dynamic-catalog/e2e/<attempt>/receipt.json`; all ordinary green commands actual exit0, red actual exit1. Source inventories bind current worktree plus explicit catalog/cmd inputs; aggregate final ordinary verification remains root-owned.

- Cost update red `20261101T100335.484204Z-rbu_tona` → green `20261101T100507.474439Z-he352m97`: two actual checked/applied identities, manual and missing rates preserved, capability unchanged, Provider independent versions.
- Lifecycle red `20261101T100702.764917Z-eu12t271` → green `20261101T100745.759263Z-ki1kvwbl`: checked withdraw/restore/rollback, new versions and unchanged historical bytes.
- Current revision Snapshot red `20261101T100854.309415Z-buj3d2oo` → green `20261101T100952.556383Z-zleuidxk`: export/versions/history reject state/snapshot divergence.
- Raw nonnullable null/trailing/duplicate JSON red `20261101T101057.384977Z-g5zxkool` → green `20261101T101153.608420Z-x1s5e2gw`: reused loader JSON-shape validation in catalog/internal/jsonshape, raw manual overlays and definitions checked before canonicalization; no new protocol table.
- Local definition/decision input originals red `20261101T101320.525562Z-6b5aexj_` → green `20261101T101410.573325Z-64iemplh`: exact role/name/digest/bytes, no synthetic fetch metadata.
- Added model → rollback → restore expanded green `20261101T101453.542764Z-90jpm4q5`: retained withdrawn recovery definition and source, restored after original source deletion, versions monotonically advance.
- Misspelled decision path red `20261101T101740.653551Z-jlslx8e_`; following green retained below. Existing field/object paths use most-specific decision, then enclosing-object/default ownership; `/` is whole-definition historical fallback. Explicit leaf decisions are recommended for price overrides.

### Final source-gap and coherent export slices

- Official captured-source/manual full definition red `20261101T102136.875644Z-aoh5szy6` → green `20261101T102217.860527Z-4mvths7d`.
- Known typed unsupported future Kimi definition red `20261101T102306.145681Z-mng7p0ru` → green `20261101T102357.668332Z-a4xo9al0`: frozen candidate, retained code_gap, queryable failure, apply refused, current untouched. Private candidate encoding defers ai acceptance only; public Compose/Generate/ProviderHash/current/export remain strict. Proposed version/typed snapshot bytes are independent of future acceptance changes.
- New coherent review bundle red `20261101T102503.668573Z-ece2s7bh`, independent retained compile failure `20261101T102604.803140Z-t7zdau_d`, then green `20261101T102657.182473Z-6mlp9dsl`. Public export preview/confirmation/new-directory rule, exact approved snapshot, actual source/manual/historical captures, external reference availability, source adoption and fresh compiled absent-store bootstrap.
- Full review bundle and exact external reference bytes are retained in final native case artifacts, so replay does not depend on deleted test temporary directories. Local input role/name/digest records never invent URL/fetch metadata.
- Source decisions reject unresolvable paths; leaf decisions take priority over enclosing-object/default decisions. `/` is the retained historical whole-definition fallback. Price ownership should use explicit leaf paths when overriding prior specific decisions.

### Final combined ordinary and universal tier presence

- Combined ordinary at frozen `bdd46af879247c6ce5e653bdd81498d071ae53ef` retained FAIL/1 in `20261101T103206.295452Z-mdfno1ir`: only test artifact retention failed because Fixture does not create nested directories. Product source and all other selected cases passed; this failed stage remains failed.
- Artifact mapping fix commit `19619b51`; same combined ordinary PASS/0 in `.evidence/barness-ai-dynamic-catalog/e2e/20261101T103717.515935Z-3t3w9yez/receipt.json`, including fresh public BuiltinCatalog consumer and complete review/reference bytes retained under flat artifacts plus explicit restoration maps.
- Additional required target-tier presence regression red `20261101T104301.813484Z-2pjmy32o` FAIL/1 → green `20261101T104334.908684Z-9q21mias` PASS/0. Shared target-tier presence validation now applies to manually introduced tiers and full official definitions, independent of projected-source diagnostics; threshold plus all four rates required, explicit zero accepted. No runtime pricing/guard behavior changed.

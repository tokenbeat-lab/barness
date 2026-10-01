# 02: 冻结 pi 差分判定器

**What to build:** 维护者可以用同一份逻辑输入与响应脚本，分别驱动冻结 pi-ai（`0.87.1`、commit `898ab804050730e9dcefb4443875d5a932aa6a32`）与 barness-ai，自动比较两者实际发出的请求和得到的事件/结果，并输出首个差异位置与分类。它是兼容判定器，复用 E2E 场景而非第二套测试体系（spec Testing Decisions §2、§5、§6）。

**Blocked by:** 01

**Status:** ready-for-agent

- [x] 在 barness 自身可重建的位置建立冻结 pi 的独立副本：补齐 `src/providers/data` 生成数据与依赖，记录来源、哈希与重建步骤；不修改原研究目录，CI 不依赖开发者本机研究仓库
- [x] pi 侧运行器只连接本地受控 Provider，不读取环境 key、不访问线上服务；不执行整套 pi 测试
- [x] 比较维度：请求路径、headers、认证来源、JSON 字段 presence；Provider/API/model、消息块、工具关联、原生状态；事件类型/块索引/delta/顺序；StopReason、errorMessage、Usage 与成本
- [x] 只对 JSON 对象键做语义排序；数组与事件不排序；时间与生成 ID 使用明确字段的一一映射并保持后续引用；不得统一删除 null、零值、空数组、未知字段或全部 ID
- [x] 租户元数据、Code/Phase、计量完整性、资源限制、同步访问与脱敏等扩展不进入 pi golden
- [x] 差分记录至少包含 case_id、pi_commit、sdk_versions、model_catalog_hash、原始请求与帧哈希、首个不同 JSON path / 事件位置、分类（已修复 / 规范定义或批准的扩展 / 待处理）、处理决定
- [x] “待处理”差异使对应协议的差分门禁失败
- [x] 以 01 的 Responses 文本场景作为首个差分用例跑通，结果写入证据包

## Comments

**2026-10-01 — implemented** (oracle in `ai/internal/testkit/pioracle`, frozen copy in its `node/`, cases in `ai/e2e/pidiff_test.go`, decisions in `ai/e2e/testdata/pidiff/ledger.json`).

- Frozen copy = npm `@earendil-works/pi-ai@0.87.1` pinned by lockfile. The release was published from `f07218c4`; within `packages/ai` and `packages/telemetry` it differs from `898ab804` only in CHANGELOG. A source build of `898ab804`, with the published data filled into `src/providers/data`, is byte-identical to the published dist (770/770 files). Details and rebuild steps are in `node/PROVENANCE.md`. The runner checks the version and model-data hash on every run.
- Opt-in: `BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run TestPiDifferential` (needs Node.js plus `npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node`). Without it the cases are recorded NOT_RUN, so `go test ./...` stays offline.
- The runner gets an empty environment and refuses credential/endpoint variables. Its sockets are restricted to loopback before pi loads, and the API key comes only from the case.
- Comparison: no field is dropped. barness results are projected onto pi's JSON shape, keeping every serialized barness field and adding only pi's `role`/`type` discriminators. The barness-only `ErrorEvent.Err` (Code/Phase) stays out of the golden. Only the explicit message time paths (`result`, `events[*].message|error|partial` `.timestamp`) are mapped, one-to-one with consistent references.
- Records carry `first_diff` (any class), `first_pending` and `first_event_index` (the first difference inside events). Request hashes cover the full redacted capture. Replay commands include `BARNESS_AI_PIDIFF=1`.
- Ledger semantics: an unledgered diff is pending; a `fixed` decision that matches again counts as a regression and is also pending; a pattern cannot cover a whole section.
- Catalog re-check (from 01): gpt-4.1 / gpt-4.1-mini / gpt-4o-mini match the 0.87.1 data exactly. Only the source comment changed.

**First run: the Responses differential gate FAILs, as designed.** The request body matches pi exactly on `stream`. Every remaining difference is classified:

- pending → 04: `events[*].partial`
- pending → 05: `rawStopReason`
- pending → 09: `max_output_tokens` from simple-entry maxTokens defaulting (+ Content-Length)
- pending → 15: `usage.cost`, `usage.reasoning`
- open decisions: all resolved, see the maintainer decisions below

**2026-10-01 — maintainer decisions**

- The HTTP-stack and OpenAI SDK fingerprint headers (Accept-Encoding, Accept-Language, Sec-Fetch-Mode, Connection, X-Stainless-Lang/Package-Version/Runtime/Runtime-Version) are approved as extensions.
- User-Agent: barness sends its own `barness-ai`, with no host OS details. It is set in the Responses adapter and asserted by the offline P01 text E2E. The difference from pi's UA is an approved extension.
- `timestamp`: added. `AssistantMessage.Timestamp` is the call's creation time in Unix ms, the same on the terminal event and the Result, failures included. It is asserted by the offline E2E; the ledger entries are now `fixed` and guard against regression.
- Gate wiring: the differential stays opt-in (`BARNESS_AI_PIDIFF=1`) for now; ADR-0004 is accepted with this decision. It moves into the regular gate when CI is set up.
- No open decisions remain. Pending differences are tracked by tickets 04/05/09/15.

# 24: 发布门禁：压力验证的资源策略示例与脱敏审计

**What to build:** 维护者得到可发布的 barness-ai：附带经压力场景验证的有限资源策略示例及数值依据、完整的公共契约/错误/装配说明、支持矩阵和差异登记，且所有门禁绿灯（spec I10 后两条、Testing Decisions §6）。

**Blocked by:** 12, 13, 23

**Status:** ready-for-human

- [x] 正式资源策略示例在可重复压力场景下跑通，记录各数值依据（部署策略，不构成 pi 兼容承诺）
- [x] 对离线与 live 证据包、日志、观测、错误做脱敏审计，无 key、Authorization、非合成正文泄漏
- [x] 交付公共契约、错误分类与装配说明、支持矩阵、模型/价格快照、fixture、差异登记（含 D1/D2 与 DeepSeek Responses 扩展）
- [x] 用例报告可追溯：研究 T/C 条目 → E/P 场景 → 实际证据；追溯表中的“已映射”不被填为 PASS
- [x] 门禁全部通过：`go test ./...`、`-race`、`go vet`、全部 P0、无待处理差分、六组合 live PASS（或明确 UNSUPPORTED 的能力）
- [x] 报告分别列出离线、差分与 live 结果

## Comments

**2026-10-02 — implemented** (gate `go run ./ai/release/cmd/releasegate`; logic `ai/internal/testkit/release`; audit `ai/internal/testkit/audit`; pressure `ai/e2e/policy_pressure_test.go`; docs `docs/barness-ai/`; decisions ADR-0017, status proposed). Status is ready-for-human: **the gate does not pass, and only live blocks it.** No combination has ever run its real-API smoke (issue 23), so the support matrix's six rows are NOT_RUN. The unchecked box above stays unchecked until that happens.

- **Final gate run** (`.evidence/barness-ai-release/gate2`, commit abfc9c8 plus this change):

  | Gate | Result | Notes |
  | --- | --- | --- |
  | go vet (with and without the live tag) | PASS | |
  | `go test ./...` | PASS | 2m30s, with the differential and pressure switches on |
  | `go test -race ./...` | PASS | 2m38s |
  | Offline P0 | PASS | 2374 cases, all PASS |
  | Differential | PASS | 432 cases; 0 pending findings; 0 pending ledger decisions |
  | Redaction audit | PASS | offline and race bundles: 0 findings |
  | Traceability | FAIL | 40 of 41 items PASS; V6-live is NOT_RUN |
  | Snapshots | PASS | |
  | Live | FAIL | all six combinations NOT_RUN |

- **Pressure (opt-in `BARNESS_AI_PRESSURE=1`; the gate turns it on).** `E08-policy-pressure-<policy>-design-load` runs each example policy at the load its comments state:
  - Every permit is busy. Each call sends a request of the stated size (history plus images) and streams a turn of the stated size, or a tool call, through Responses.
  - Every byte and queue limit must be used at most 75%. The reachable heap must stay within the assumed budget: 1 GiB local, 2 GiB cloud.
  - `-stalled-consumer` checks that a reader who stops ends with resource_limit in the event queue.
  - The numbers are in `docs/barness-ai/README.md` and in the policy comments.
  - **The scenario changed three values:**
    - `MaxToolJSONBytes`: 512 KiB → 128 KiB (cloud) and 1 MiB → 256 KiB (local). Streamed tool arguments are re-parsed on every delta, as pi does, so CPU grows with the square of their size: 23 s for one 512 KiB call. Filed as **issue 32**.
    - Batch `MaxOutputBytes`: 32 → 64 MiB. A 128K-token turn used 88% of the old value.
    - The cloud comment's request-memory bound is doubled: about two copies of each body are held while an attempt runs.
- **Redaction audit.** It runs when every offline and live run finishes (`evidence.Run.Audit` writes `audit.json`; any finding fails the process). It replaces live's key-only check. The gate re-audits every bundle and requires each bundle's recorded audit to be clean. What it checks:
  - registered secrets;
  - the environment's own credential values, home directory and host name;
  - vendor key shapes;
  - credential headers and query keys;
  - Observer records, against a whitelist of fields;
  - live response headers, against the recorder's whitelist.

  A mutation check that turned off source redaction failed the E2E run through the audit.
- **Traceability** (`ai/release/traceability.json`). Each item's status is computed from evidence only. An item that is mapped but has no evidence is NO_EVIDENCE, never PASS. Live items come from the support matrix.
- **Live gate requirements.** For each combination:
  - the matrix shows a complete pass;
  - no capability is failing now;
  - the SDK version recorded for that pass equals go.mod's;
  - the combination's own live bundle is audited through `-live`.

  Adapter changes cannot be seen from the matrix. After one, the maintainer must rerun the affected combinations (ADR-0017).
- **Deliverables:**
  - contract, error classes and assembly: `docs/barness-ai/contract.md`;
  - difference register: `docs/barness-ai/differences.md`;
  - support matrix and fixtures: indexed in the README;
  - catalog and price snapshot: `ai/release/catalog-snapshot.json`, which the gate checks against the code.
- **Also:** issue 08 is closed; its last criterion was done in 19. One gofmt fix in `ai/chat_stream.go`.
- **Code review (two axes):**
  - Fixed:
    - goroutine cleanup when a pressure run times out;
    - a missing race bundle is now reported, not skipped;
    - the recorded `audit.json` is checked;
    - live staleness and live-audit requirements;
    - required differential record fields;
    - the `callback_failed` row in the error table;
    - overclaims in the policy comments.
  - Not changed: the 3-line findings-print loop repeated in the e2e and live `TestMain`.
- **For the maintainer:**
  1. Run the six live combinations (`ai/live/doc.go`).
  2. Merge their reports into the matrix.
  3. Rerun the gate with `-live <bundle>` for each combination.
  4. Confirm ADR-0017: the gate's composition, the opt-in pressure scenario and the 75% headroom rule, the three policy value changes before issue 32, the audit rules, and the traceability map.

**2026-10-02 — live run; the gate passes.** After the six real-API runs (see issue 23), the release gate passes all nine gates (`.evidence/barness-ai-release/gate4`, commit c976dc6 plus this change). That covers offline (2374 cases), the differential (432 cases, nothing pending), race, vet, live (6/6), the audit (8 bundles, 0 findings), all traceability items and the snapshot. Status stays ready-for-human until ADR-0017 is confirmed.

Run 3 hit two one-off failures. Neither recurred: run 4 passed, and so did the isolated reruns.

- **`PIDIFF-P03-usage-replaced-stream`.** Under full-suite load, pi's live partial at `events[1]` held the chunk's usage and barness's did not yet. Both apply usage after the chunk's parts, but pi serializes each event only after the whole chunk has run. This is the approved live-partial category in the other direction, so the Gemini ledger gains the `only_pi` entry for `partial.usage.reasoning`. 30 isolated runs matched pi.
- **`E08-policy-pressure-local-design-load`.** Two calls failed with `transport` on loopback before any response arrived. This did not reproduce in 5 isolated runs or in a full run. TIME_WAIT peaked near 1.9K of 16K ports, which rules out port exhaustion. barness does not expose the underlying error, so the pressure world now records every failed round trip's network error (`transport-errors.json`) for diagnosis if it recurs.

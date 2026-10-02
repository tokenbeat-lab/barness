# 15: 用量、成本与价格版本（E11）

**What to build:** 成本管理开发者获得与冻结 pi 一致的 Usage 与估算成本（缓存读写、1h 写入、reasoning、阶梯价格），同时在 Result 元数据中看到每次尝试的用量完整性与价格版本，未上报用量不会被当成免费（spec I10 前三条、User Stories 40–41）。

**Blocked by:** 01, 02

**Status:** resolved

- [x] Usage 保留 input/output/cacheRead/cacheWrite/totalTokens/cost、可选 cacheWrite1h/reasoning、初始化零值与 adapter 更新规则；reasoning 为 output 子集，不重复加入总量
- [x] 移植冻结 pi 的阶梯价格、缓存读写、1h 写入与 adapter 专用调整；模型目录与价格快照带版本与哈希
- [x] Result/观测元数据区分未上报、部分上报、完整上报，不修改兼容消息的数字字段
- [x] 每次尝试分别记录用量；不把全部尝试合计反写为最终消息 Usage；失败请求的零值不被解释为免费
- [x] Responses 成功、失败、重试路径上 usage 缺失 / null / 零值三态用例
- [x] 兼容字段接入 pi 差分无待处理差异；完整性与价格版本作为扩展单独断言

## Comments

**2026-10-02 — implemented** (E2E `ai/e2e/usage_test.go` with fixture `testdata/responses/usage.json`; isolated `ai/cost_test.go`; decisions in ADR-0010, status proposed).

- **API:**
  - `Usage` (moved to `ai/usage.go`) gains `Cost UsageCost` (always serialized), `Reasoning` and `CacheWrite1h` (both `Nullable[int64]`, omitted when unset).
  - `Model.Cost ModelCost` with `Tiers []CostTier`, in pi's JSON shape.
  - `Catalog.Hash()`.
  - `Attempt.UsageReporting` (`unreported`/`partial`/`complete`) and `Attempt.Usage`.
  - `CallMetadata.CatalogVersion`/`CatalogHash`.
  - `NewClient` refuses a catalog with a negative or non-finite price (`ConfigError{Field: "Catalog"}`).
- **pi compatibility:**
  - `calculateCost` is ported in pi's operation order: tiers strictly above the threshold, counting input, cache reads and cache writes; one-hour writes at twice the input rate.
  - So is the Responses service-tier adjustment (`response.service_tier ?? options.serviceTier`; flex ×0.5, priority ×2, gpt-5.5 ×2.5).
  - Cost is computed only at a completed/incomplete terminal, as in pi. `reasoning` is set only when usage was reported. A `response.failed` leaves the message usage zero.
  - The built-in catalog is now `2026-10-02.1`: every model's frozen price, plus `gpt-5.5-pro`, the one listable model with tiers.
- **Differential:**
  - 24 new cases `PIDIFF-P01-E11-*` cover success, failure and retry with usage missing, null, zero, partial and reported, plus reasoning, cache writes, tiers, service tiers and incomplete. The full run (`BARNESS_AI_PIDIFF=1 go test ./ai/...`) has **0 pending**.
  - The seven `usage.cost`/`usage.reasoning` ledger entries are now `fixed`.
  - Two extension entries were widened to the new optional field, with their existing rationale:
    - the live partial may already hold `partial.usage.reasoning`;
    - the over-limit cases lack pi's `result.usage.reasoning`.
  - The `terminated` runtime-text extension also lists the new stream-cut case.
- **Extensions** (asserted offline, not part of the pi golden):
  - Each attempt records its own reporting and priced usage. Failed HTTP attempts are `unreported`.
  - A failed response's usage is kept on its attempt only.
  - The message usage equals the streamed attempt's and is never a sum (retry scenarios).
  - Observer AttemptFinished records carry the same values as the Result (`checkCallRecords`).
  - The built-in catalog's version and hash are pinned in the fixture, so a content change without a version bump fails.
- **Isolated test:** `TestCostEstimate` checks one-hour writes and multi-tier selection, which Responses cannot produce. With the oracle on, it also checks them bit-for-bit against pi's `calculateCost`, and every built-in price against pi's model data. A new runner entry `costs` serves it. `pioracle.Oracle.direct` now holds the runner call shared with `ThinkingBudgets`.
- **Fixtures:** the expected usage in `text-basic`, `interleaved-blocks`, `failures` and `tool-round-trip` gains pi's `reasoning` and `cost`, computed with the frozen copy. `gpt-5-mini` and `gpt-5.5-pro` join the test binding's allowed models.
- **Mutation checks** (each breaks `TestUsageAndCost`):
  - the failed-response usage not recorded on its attempt;
  - partial read as complete;
  - tiers ignored;
  - a tier applied at its threshold;
  - the service tier ignored;
  - the request tier winning over the response's;
  - a failed attempt not marked unreported;
  - missing usage read as complete;
  - the catalog not validated;
  - the hash missing from the metadata;
  - reasoning added to the total.
- **Open for the maintainer** (ADR-0010 "待维护者确认"):
  1. `cache_write_tokens` is not required for `complete`;
  2. a failed response's usage is recorded on its attempt;
  3. the price snapshot sits on the call metadata, not on each attempt.
- **Deferred:** an E2E for one-hour cache writes comes with the Anthropic adapter (issue 16). The matching isolated cases can then go.

**2026-10-02 — maintainer decisions; ADR-0010 accepted.**

- **Missing counts:** count as 0; a missing `cache_write_tokens` still allows `complete` (as implemented).
- **Failed responses:** their usage is dropped as in pi. The attempt is now `unreported` with zero usage instead of carrying the priced usage. `failed-usage-zero` and `failed-usage-reported` in `usage.json` now expect `unreported`. Recording the usage later is tracked as issue 25.
- **Price snapshot:** stays on the call metadata (as implemented).
- Offline suite and the full differential (`BARNESS_AI_PIDIFF=1`) pass with 0 pending.

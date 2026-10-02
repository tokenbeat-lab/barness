# 30: 协议横向 E2E 套件参数化（E06–E09 资源、回调、观测）

**What to build:** 让 Anthropic 与 Gemini 的资源限额、准入/时限、回调、观测/预检/释放 E2E 套件共用一份按协议参数化的实现，协议只提供数据（回复脚本、错误体、长输出、被拒模型等）和协议差异（有无 timeoutMs、超时是否重试、回调时点）。

**Blocked by:** 19

**Status:** needs-triage

- [ ] `ai/e2e/anthropic_{limits,admission,callbacks,observe}_test.go` 与 `ai/e2e/gemini_{limits,admission,callbacks,observe}_test.go` 合并为按 `fixtureProtocol` 运行的套件，用例 ID 前缀（P02/P03）保持不变
- [ ] 协议差异（Gemini 无 timeoutMs、头超时不重试、不调用 onResponse、REST 回调体与拒绝字段）作为协议描述的显式字段或协议专属子用例，而非复制整套
- [ ] 合并前后用例集合与证据包 case ID 一致；pi 差分不受影响

## Comments

**2026-10-02 — opened from the issue 19 code review.** Gemini is the second protocol on the scenario-fixture format; its fixtures, E01–E05/E11 harness, entries and text helpers were made protocol-parametric there (`scenario_fixture_test.go`, `scenario_entries_test.go`, `scenario_pidiff_test.go`), but these four suites were written per protocol on the shared helpers because each case embeds protocol data and several cases deliberately differ (ADR-0012 决策三、四). The copies are ~70% parallel; this issue removes them.

**2026-10-02 — Chat Completions joins (issue 20).** `ai/e2e/chat_{limits,admission,callbacks,observe}_test.go` are a third copy, parallel to the Anthropic suites (Chat shares the Stainless retry/timeout rules and runs onResponse); the merge should cover P04 too, with Chat's differences (long retention for a cache key off OpenAI's endpoint, blocks closed before an aborted terminal) as protocol fields.

**2026-10-02 — DeepSeek Responses (issue 21).** `ai/e2e/deepseek_limits_test.go` runs only the request body, frame and output bounds on P05 (the rest is the shared Responses adapter's); a parametric suite should cover P05 with them.

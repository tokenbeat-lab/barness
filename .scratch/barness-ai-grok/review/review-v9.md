# Ticket16独立两轴复核 v9

固定差分 `3a44c514414f8d1753ae2f66a07814a44bd27cef...5f933af973d1e0b3784df149f489c70d53973a0b`；commit5f933af9。只读审查，未执行测试/构建/厂商调用。

## Standards

## Standards

No actionable documented-standard breaches or baseline smells found in `git diff 3a44c514414f8d1753ae2f66a07814a44bd27cef...5f933af973d1e0b3784df149f489c70d53973a0b` (commit `5f933af9`).

- `ai/internal/testkit/release/chat_capture.go:29–45` replaces materialization of polymorphic `ai.Result` with a consumer-local persisted projection. The change leaves public Result/Message decoding untouched, adds no compatibility layer or dependency, and preserves existing attribution, terminal Observer, wire-target/model, negative usage and aggregate-budget checks. The explanatory comment records the boundary decision (AGENTS.md principles 1, 3, 7, 8, 10).
- `ai/e2e/generic_capture_decoder_test.go:17–99` records its failure inventory and exercises both public AuditBundle consumption and the compiled releasegate CLI. The intact text/thinking/tool fixture passes audit; five separate mutations fail; all synthetic routes must retain exit 1 and qualification false. Processes use the existing isolated environment helper. Files remain below 500 lines (AGENTS.md principles 2, 4, 5, 9 and testing principles; verification SOP §2).
- Independently inspected retained RED/GREEN receipts: identical test hash `74c889f6…`; decoder hash changes from `6f2e1393…` to current `1eb847b9…`; RED fails the nonempty acceptance/CLI assertions, GREEN exits 0 with six PASS cases and empty audit findings. All four indexed RED/GREEN/manifest/audit hashes match their originals. This supports the test-first claim without rerunning tests (SOP §2, §8).
- `generic-capture-decoder-v1.json:53–59` explicitly retains offline status and identifies native source invalidation; `RegionalExecutionSource` includes this decoder. No historical paid qualification or stopped batch is renewed by the diff (ADR-0051/0053; SOP §3–4).

Read-only review only; no tests, builds or vendor calls executed. Broader ordinary verification and current-state publication remain root-owned; existing partial live/race/pressure status is not a finding against this change.

Findings: 0 hard breaches; 0 heuristic smells; worst severity: none.

## Spec

## Spec

1. **[P2] Complete the required typed-capture counterexamples.** In `ai/e2e/generic_capture_decoder_test.go:33–59`, the rejection matrix covers binding attribution, terminal kind, negative usage, budget totals and inventory, but not the retained Message identity/stop/error fields. The fixture gives every scenario the IDs `text`, `thinking` or `tool` (`:131–144`), so none enters `grokhistory.Scenario`, the only branch that checks those fields in `ai/internal/testkit/release/chat_capture.go:58–61`. Consequently, deleting or mistagging the new DTO's `API`, `Provider`, `Model`, `StopReason` or `ErrorMessage` fields would leave all six new cases green while breaking the mandatory history consumer. Issue 16:23 explicitly requires “并保留typed捕获所需身份/stop/error/Usage/CallMetadata约束的反例.” Add a valid nonempty mandatory-history capture and independent identity/stop/error counterexamples through the compiled consumer, preserving the existing refusal rules.

The consumer-local DTO otherwise matches the requested scope: no public Result JSON contract change, no qualification-rule or authorization change, and synthetic inputs explicitly remain ineligible for real live qualification. Full ordinary validation, source refreeze and current-state closure are pending integration work, not findings against this implementation.

Read-only review of `git diff 3a44c514414f8d1753ae2f66a07814a44bd27cef...5f933af973d1e0b3784df149f489c70d53973a0b`; no tests, builds or vendor requests run.

Findings: 1 P2; worst severity P2.

Standards 0项（无最高级别）；Spec 1项P2（mandatory-history消息身份/终态/错误反例缺失）。同一16实现者补齐后另行独立复核；本报告及已生成原件保持。

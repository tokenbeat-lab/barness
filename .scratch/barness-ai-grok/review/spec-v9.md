## Spec

1. **[P2] Complete the required typed-capture counterexamples.** In `ai/e2e/generic_capture_decoder_test.go:33–59`, the rejection matrix covers binding attribution, terminal kind, negative usage, budget totals and inventory, but not the retained Message identity/stop/error fields. The fixture gives every scenario the IDs `text`, `thinking` or `tool` (`:131–144`), so none enters `grokhistory.Scenario`, the only branch that checks those fields in `ai/internal/testkit/release/chat_capture.go:58–61`. Consequently, deleting or mistagging the new DTO's `API`, `Provider`, `Model`, `StopReason` or `ErrorMessage` fields would leave all six new cases green while breaking the mandatory history consumer. Issue 16:23 explicitly requires “并保留typed捕获所需身份/stop/error/Usage/CallMetadata约束的反例.” Add a valid nonempty mandatory-history capture and independent identity/stop/error counterexamples through the compiled consumer, preserving the existing refusal rules.

The consumer-local DTO otherwise matches the requested scope: no public Result JSON contract change, no qualification-rule or authorization change, and synthetic inputs explicitly remain ineligible for real live qualification. Full ordinary validation, source refreeze and current-state closure are pending integration work, not findings against this implementation.

Read-only review of `git diff 3a44c514414f8d1753ae2f66a07814a44bd27cef...5f933af973d1e0b3784df149f489c70d53973a0b`; no tests, builds or vendor requests run.

Findings: 1 P2; worst severity P2.

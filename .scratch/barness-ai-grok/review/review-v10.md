# Ticket16独立两轴复核 v10

固定完整差分 `3a44c514414f8d1753ae2f66a07814a44bd27cef...fd02b56aa6177106ae1ef348fec0c3a4920b8a57`；commits5f933af9、fd02b56a。上一轮v9 Spec P2已补齐，生产DTO未追加修改，新增历史分支回归直接GREEN；未伪造RED。两位独立审查者均只读，未执行测试/构建/厂商调用。

## Standards

## Standards

No actionable documented-standard breaches or baseline smells found in the full pinned diff `3a44c514414f8d1753ae2f66a07814a44bd27cef...fd02b56aa6177106ae1ef348fec0c3a4920b8a57` (commits `5f933af9`, `fd02b56a`).

- `ai/internal/testkit/release/chat_capture.go:29–45` retains the consumer-local persisted projection from v9; no public Result decoder, compatibility layer, dependency or production relaxation is added. Existing attribution, terminal, wire/model, usage, budget and mandatory-history checks remain intact (AGENTS.md principles 1, 3, 7–10).
- `ai/e2e/generic_capture_decoder_test.go:18–120` expands the explicit failure inventory with six independent mandatory-history provider/API/model/stop/message-error/call-error mutations. The intact positive fixture must pass AuditBundle and reach native qualification through the compiled CLI; each mutation must fail at audit. Every synthetic record must retain exit 1 and `Pass=false` (AGENTS.md testing principles; verification SOP §2, §7).
- Lines 183–208 reuse `grokHistoryWorld`, the public Client and its existing loopback-only transport, which has cleanup registered. The fixture verifies source envelopes, fixed Host input, actual encoded wire and CrossModel=2 before capture audit. This reuses stable semantics without adding speculative infrastructure; persisted inputs contain synthetic data. The changed test is 256 lines (AGENTS.md principles 2, 4, 5, 7, 8).
- Independently verified v2 summary SHA `d5a832ab…`, receipt SHA `1ca4f185…`, manifest/audit hashes and pinned test SHA `d833d5f6…`. The retained command exited 0, with 12 native PASS cases, 13 Go test names and empty audit findings. It accurately labels the added cases as direct GREEN against already-correct production code and preserves the original decoder RED/GREEN rather than fabricating failure history (SOP §2, §8).
- Offline qualification remains false; source applicability from v1 remains explicit. The test-only follow-up renews no historical paid evidence, grant or claim (ADR-0051/0053; SOP §3–4).

Read-only review; no tests, builds or vendor calls executed. Pending root-owned full ordinary verification, freezing and live/race/pressure selection are not findings against this diff.

Findings: 0 hard breaches; 0 heuristic smells; worst severity: none.

## Spec

## Spec

No findings. The v9 P2 is substantively resolved.

Issue 16:23 requires “并保留typed捕获所需身份/stop/error/Usage/CallMetadata约束的反例.” The positive fixture now includes the exact `grokhistory.Scenario` ID (`ai/e2e/generic_capture_decoder_test.go:152,172`), thereby entering `chat_capture.go:58–68`. Its Host request and provenance are retained, and the public Client sends the history through the loopback provider; the actual encoded request and CrossModel=2 are verified (`generic_capture_decoder_test.go:185–208`).

The six new cases mutate only the corresponding mandatory-history result: provider, API, model, stopReason, message.errorMessage and outer error (`:56–73`). They reseal inventory, then require rejection both through native audit and the compiled public CLI (`:81–116`); they do not merely add case names. The unchanged positive case passes capture audit before the unknown synthetic route is refused for qualification. Existing attribution, Observer-terminal, negative usage, budget and inventory counterexamples remain.

Read-only inspection of the retained GREEN attempt confirmed 12 native cases PASS, numeric exit 0, audit findings empty, and source hashes matching the reviewed test and production files. Receipt SHA256 `1ca4f1854937211f53b1854e49f21cdfa491f6e0a4ca07d401c3e8c4227c1946` and summary-v2 SHA256 `d5a832ab208396896fddd28922a3654b98628e3c09b59d96a84b63077d1efeb3` match their originals. The summary accurately distinguishes this directly GREEN coverage extension from the original decoder RED/GREEN.

The complete diff remains scoped to the consumer-local DTO, offline regression and evidence summaries. Public JSON contracts, authorization and formal qualification rules remain unchanged. Pending full ordinary validation/refreeze/live are not claimed complete.

Reviewed `git diff 3a44c514414f8d1753ae2f66a07814a44bd27cef...fd02b56aa6177106ae1ef348fec0c3a4920b8a57`. No tests, builds, vendor requests or source writes performed.

Findings: 0; worst severity: none.

Standards 0项（最高级别无）；Spec 0项（最高级别无）。完整普通、新冻结来源与正式live资格仍独立判定。

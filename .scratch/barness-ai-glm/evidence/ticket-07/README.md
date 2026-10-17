# Ticket 07 — GLM Messages tools and native thinking replay

Date: 2026-10-17 (Asia/Shanghai). Confirmed seam: spec T01 public ai.Client,
controlled Provider and trusted Host. Implementation commits `c00012c` and
`5606fda`; integration `f0fb0de` merged at `85b6fa9` before final verification.
[Failure modes](failure-modes.md) preceded implementation. [Index](index.json)
pins source files, manifests, case outcomes, audit results and ZIP hashes.
All raw runs and archives are durable under repo-root
`.evidence/barness-ai-glm/ticket-07/`; no temporary worktree is needed.

| Public behavior / check | Result |
| --- | --- |
| Exact 5.2 enabled/budget1024/max4096 | 3 RED → 3 PASS; full Stream/Complete and exact trusted payload replacement |
| Tools and native replay | 12 unsigned RED with 28 existing signed/tool PASS → 43 PASS; later 28 native origin/signature combinations |
| Illegal SSE sequence | 18 RED → PASS; two additional multi-index interleaving scripts PASS |
| Unverified native form / final controls | Empty-signature and redacted replay RED → PASS; named/simple/payload boundaries preserved |
| Text-only final callback image injection | 6 RED → PASS, user/system/tool-result for both models; schema-property control PASS |
| Canonical budget arithmetic | 1 RED → PASS; a Host map and budget cannot accidentally enable a guessed canonical mapping |
| Unverified tool definition extensions | 3 RED → PASS; strict/defer_loading/type:custom never sent |
| Final committed Messages suite | **193 PASS**, including 66 predecessor text/default/guard/terminal cases |
| Race | **190 PASS** at merged implementation + **3 PASS** final tool-field guard cases |
| Native Anthropic tools/thinking/history/cache/beta/options/usage | **104 PASS** |
| Native Anthropic runtime/callback/environment/resource | **168 PASS** |
| Frozen pi Messages differential | **70 PASS**, zero pending, on final implementation |
| Vet and staged/unstaged whitespace checks | PASS |

`green-sequence` retained 63 PASS / 4 FAIL because a newly strengthened full-wire
baseline assertion expected auto choice while that harness used omission. The
harness now sends the captured auto choice. `red-final-boundary` and `green-all`
also retained two simple-off phase expectation failures: off is not a public
ThinkingLevel and correctly fails at capability, not request. These are harness
iterations, not production regressions or claimed passing evidence. Two earlier
compile iterations created no behavior bundle. All runtime bundles completed
TestMain redaction audits with zero findings, including expected RED runs.

## Independent own-protocol evidence

[Messages exploration](../messages-exploration/README.md) and its index pin raw
`03-tools-thinking` and `04-default-unsigned` captures. Original request/SSE/
events/Result/Attempt/usage/source/provenance/hash records remain under
`.evidence/barness-ai-glm/protocol-exploration/messages/`. This ticket made
**zero paid calls**, read no `.env`, and does not claim formal production live
acceptance, built-in catalog registration, pricing or entitlement verification.

- Each model's first tool and continuation SSE and final request are copied
  unchanged to `ai/e2e/testdata/glm-messages/`; the fixture pins SSE SHA-256,
  independently captured expected content, IDs, stop reason and usage. Both
  models complete Host ValidateToolCall → fixed local execution → ToolResultText
  → final BRN42 through Stream/Complete/StreamSimple/CompleteSimple, default and
  auto. No Chat/Responses facts supply Messages capabilities.
- Own 5.3 tool output includes thinking/signature; own 5.2 tool output does not.
  The separately captured exact 5.2 enabled/budget1024/max4096 request produces
  signed thinking, accepted by its original continuation. Candidate-only
  ModelCompat.MessagesThinkingBudgetTokens=1024 opts in that exact shape; zero
  opts out, and provider implementation cannot be broadened by metadata. This
  proves neither a budget range/enforcement nor adaptive/effort/intensity mapping.
- Public continuation preserves signed state directly and after persisted Host
  restoration. Deliberate unsigned records preserve thinking text/type/order and
  omit ONLY signature relative to each successful baseline. Own probe 04 had
  already accepted those exact field-absence requests. This is not observed
  unsigned provider output or empty-string signature evidence.
- Unsigned thinking now requires existing provenance checks to retain its native
  wire kind. Missing envelope or cross-tenant/account/API/model downgrades to
  visible text, without native fields/account references; redacted opaque data
  drops on untrusted downgrade. Host storage integrity remains its responsibility
  before TrustNativeState. No actual redacted GLM output was observed; native
  redacted replay is explicitly unimplemented and rejected before any Attempt.
- Unproven named/simple thinking, arbitrary budget/effort, forced/any/none tools,
  cache/beta/display and final unsupported controls fail before send. Exact
  integer guards also reject decimal values float64 could round to accepted
  integers. Authorization retains precedence. Final native images cannot bypass
  the two models' text-only input constraint. Tool schema/input data is not
  inspected as protocol extensions.
- Controlled invalid-order and multi-index scripts are explicitly synthetic
  decoder fault/stress cases; they do not claim the service emitted malformed
  streams or supports parallel tool execution. Native Anthropic behavior remains
  pinned and unaffected.

[ADR-0035](../../../../docs/adr/0035-barness-ai-zhipu-messages-basic.md) records
these provider/native-state decisions; [Host notes](../../../../ai/internal/testkit/glm/messages.md)
show the bounded candidate. Disabling the Host Binding removes the route; there
is no compatibility shim, fallback endpoint, new runtime or environment reader.
Ticket 11 still owns formal production-route live acceptance and ticket 12
built-in registration. Ticket 07 does not mark the full release matrix PASS.

## Replay

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-07/replay" \
  go test ./ai/e2e -run '^TestGLMMessages' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-07/replay-race" \
  go test -race ./ai/e2e -run '^TestGLMMessages' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-07/replay-native" \
  go test ./ai/e2e -run '^Test(Anthropic|ManagedEffort)' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-07/replay-runtime" \
  go test ./ai/e2e -run '^TestProtocol/^anthropic$' -count=1
# With the frozen oracle's locked dependencies installed:
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/^anthropic-messages$' -count=1
go vet ./ai/...
```

Every bundle manifest supplies individual case replay commands. Temporary
oracle node_modules symlinks were removed; raw SSE retains the exact final
empty delimiter and existing narrowly scoped whitespace attribute.

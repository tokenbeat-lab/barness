# Ticket10 — frozen GLM route and extension acceptance

2026-10-17; implementation `af49ce029f8993cb5a5a7ac15bfdadab9d509266`.
[Failure plan](failure-plan.md) preceded tests/data changes. No production adapter,
frozen oracle, SDK pin, built-in catalog, formal route registration or support
matrix was changed by this ticket. No credential or environment file was read,
and no paid call was made. The approved spec T01 Client remains the sole
business acceptance seam; the source inventory and ledger are provenance gates.

## Source decision

All three China standard `zhipu` routes are **ExtensionRoute**, for **both**
`glm-5.3` and `glm-5.2`. The existing pi-ai 1.1.0 oracle first verifies its full
827-file dist SHA, 43-file data SHA, SDK lock and commit/version pins; then the
read-only source case inventories every native provider data file. No standard
China identity or base exists. Its same-spelled GLM models are different
services: `zai` international Coding (both models) and `zai-coding-cn` China
Coding (5.3 present, 5.2 absent), Chat only. All four native China-host models
use the Coding base. [Frozen source hashes](sources/frozen-source.json) and
[unchanged baseline inputs/317 decisions/3 old routes](sources/preservation.json)
record the exact scope; source E2E saves full data inventory and model identities.

The [central ledger](../../../../ai/e2e/testdata/pidiff/ledger.json) adds exactly
three extension registrations. [Own fixture/skip explanation](../../../../ai/e2e/testdata/glm-pidiff/README.md)
records why no same-route native pi comparison is possible. No custom pi model
or modified endpoint is used to manufacture parity. GLM own acceptance never
counts as `piParity=PASS`; native differential evidence remains separate.

## Bounded behavior and independent evidence

The [own case/path ledger](../../../../ai/e2e/testdata/glm-pidiff/ledger.json)
contains **43 records, 752 case/path pairs, 420 distinct public cases** and zero
pending. This is a source-backed own-extension behavior ledger, not fabricated
Pi/Barness observations. Public Result-envelope and captured-request leaf paths
are exact and GLM scoped. Unknown case/path/native protocol stays pending;
fixed content-loss and ID-collision recurrence remains a regression. There is no whole-body,
result, event-list or protocol exemption. The verification helper checks that
every referenced public case passed and every artifact retains its hash.

Both models independently prove official object/array/done Responses content,
original native JSON versus documented InputReasoning request conversion,
Chat reasoning_content/tool_stream/history, actual Messages signatures and
explicit restored signature absence, valid tools/results, usage presence,
errors/retries and cross-protocol opaque-state isolation. Full/simple entries
and hooks are exercised by their own public GLM suite. Native OpenAI,
Anthropic or DeepSeek success supplies no substitute GLM artifact.

[Responses schema excerpts](sources/responses-contract.json) retain original
OutputReasoning/InputReasoning and function IDs; [Chat excerpts](sources/chat-contract.json)
retain the independent request controls. Messages sources are the unchanged
model-specific observed captures and bounded probes referenced in 07, with raw
request/SSE and source SHA retained by their own fixtures. Synthetic object,
cache/error mutations and 5.3 native-history variants remain explicitly
synthetic; none is described as observed service behavior. Candidate rates
remain synthetic, and no current source fact is promoted to a paid billing fact.

The predecessor Responses tool captures have delimiter-free IDs. The pinned
InputFunctionCall.call_id and OutputFunctionCall.{call_id,id} are arbitrary
strings with no alphabet excluding `|`. Their success therefore does not prove
delimiter collision safety. The coordinated 05 correction now has four fixed
call/result ID leaf-path records for 58 exact public edge cases, covering pipe,
percent, Unicode, ambiguous concatenation, Host restoration and cross-origin
conversion. Additional cross-API collision and empty-origin-ID cases are retained
in the same separate post-merge run. See [05 follow-up evidence](../ticket-05-followup-ids/README.md).

## Verified runs

| Independent run | Result |
| --- | --- |
| Own GLM public batch before ID follow-up | 1,733 PASS |
| Existing DeepSeek Responses extension public suite | 64 PASS |
| Source/extension/finite ledger gate with frozen opt-in | 5 PASS |
| OpenAI Responses frozen differential | 335 PASS, zero pending |
| OpenAI Chat frozen differential | 203 PASS, zero pending |
| DeepSeek Chat frozen differential | 118 PASS, zero pending |
| Native Anthropic frozen differential | 70 PASS, zero pending |
| After integration merge: 90 ID cases + 5 provenance cases | 95 PASS |
| go vet ./ai/e2e; whitespace | PASS |

The combined offline public run also has one explicit NOT_RUN source-inventory
case because the oracle opt-in is absent there; its separate frozen run passes.
Thus 1,797 public PASS + 726 native differential PASS + 5 provenance PASS are
separate counts, not one GLM parity total. The full batches used integration `8839138` production bytes; this ticket
adds only tests/ledger/source records. After merging stable integration
`2222b9476f107843f23797ae6a094f08f231b71f` at
`2b8c1a69dc672d0e8a661d2e8cef118f9e386537`, the separate post-merge batch
passes all 90 new ID cases and five provenance cases. Upstream05 also retains
its own 288 related regression, 124 race and 335 frozen Responses PASS at code
`0f33a39`; its scope is documented in the linked follow-up, independently of
the prior full 726 differential. The initial ledger run contained 39 records;
the post-merge gate validates the final 43. New files were staged before
inventory-bearing tests; manifests record each run's actual Git HEAD and
pinned catalog/SDK versions.

The first source-gate harness called the Models query with JSON null rather
than an empty array; it failed before scanning. That diagnostic is preserved
separately and corrected without changing the frozen oracle. Missing-route
RED (3 failures), missing-ledger RED (1 failure), and this harness diagnostic
are never counted as successful acceptance.

## Raw evidence and replay

[Index](index.json) points to durable root `.evidence/barness-ai-glm/ticket-10/`:
raw runs, separately retained diagnostics, logs, exact official schema snapshots,
frozen source excerpts and case/path proof. Every completed bundle has an empty
redaction audit and verified artifact hashes. Archive index records ZIP SHA,
member count and byte-for-byte verified member contents. Prior issues' evidence is
referenced with its original scope rather than copied into new claims.

```sh
# Rebuild only if the unchanged frozen oracle's dependencies are absent:
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestGLM(Extension|FrozenRoute)' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^Test(GLM(Extension|FrozenRoute)|GLMResponses(OpaqueNativeIDs|CrossOriginOpaqueIDs|OriginIDCollision)|GLMChatOriginNonemptyResponseID)' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-10/replay-public" \
  go test ./ai/e2e -run '^Test(GLM|DeepSeekResponses)' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/(openai-responses|openai-completions|deepseek|anthropic-messages)$' -count=1
python3 .scratch/barness-ai-glm/evidence/ticket-10/verify-evidence.py \
  "$PWD/.evidence/barness-ai-glm/ticket-10" --out /tmp/glm10-index.json
go vet ./ai/e2e
```

Each manifest contains an exact one-case replay and complete hashes for requests,
scripted frames, events, Result, attempts, usage, observations and assertions.
Formal catalog/live/matrix and complete release status remain tickets11–13;
rollback is an ordinary code/evidence rollback or disabling the Host Binding.

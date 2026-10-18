# GLM Responses capture failure: read-only recovery assessment

Scope: inspected source and existing actual artifacts only at the parent's frozen source `89000346b48480d94c3d6f6523629870f4e7b1b0`. No repository edits, tests, HEAD changes, `.env`/credential reads, vendor calls or retries. This note is the only new file. Messages collection was still an independent parent-owned job; no assumption is made about its final counters.

## Finding

**There is no existing legitimate one-scene continuation/merge seam that can turn this run into current formal PASS within its two remaining logical calls.** The report correctly remains FAIL. The underlying response also ended at the output limit with no final text, independently of the capture defects. Spare arithmetic budget is not an implemented or approved corrective producer/importer contract.

Original bundle:
`/Users/cyber/RestoX/barness/.evidence/minimax-global-formal-glm/glm-responses/20261018T030609.279690000Z`.
Its formal report has 22 calls / 22 reserved Attempts / 22 actual Attempts, 21 PASS and only `glm-5.3/control-low` FAIL. Audit passed; complete=false; final failure is `inventory read failed`. Source/catalog/actual model are not an authentication issue.

Observed safe metadata:

| Layer | Actual fact | Cause / consequence |
| --- | --- | --- |
| Transport | HTTP200, exactly 524,288 retained response bytes, Truncated=true | Shared recorder hard cap512KiB. Actual total bytes read are not counted; exact missing-wire-byte count cannot be recovered. |
| Retained SSE | 2,539 complete `response.reasoning_text.delta` events, plus four initial events; no retained terminal | Prefix capture cannot prove original wire's complete terminal/usage/native shape. |
| Public events | 4,098 envelopes: start1, thinking-start1, delta4,094, thinking-end1, done1 | Public Client consumed beyond capture prefix. These are normalized public DTOs, not original SSE bytes. |
| Public content | Thinking6,369 characters; no Text part. Complete retained SSE deltas account for4,017 characters | 2,352 final thinking characters are not accounted for by complete retained SSE frames; a possible partial last frame prevents treating this as an exact lost-byte measurement. |
| Public terminal | Error empty; StopReason=length; raw stop `incomplete.max_output_tokens` confirmed by parent | A repaired recorder alone would still fail required Stop/final-text assertions. Do not relabel length as stop or success. |
| Usage | Input49, output4096 including reasoning4094; complete known Attempt usage;61,600ms | Known consumption is preserved. Reasoning is already included in output, not an added bill; no claim about actual invoice. |
| Events artifact | `events.json`=17,715,984 bytes | Greater than16MiB=16,777,216, the only oversized regular file in this bundle. This directly explains final inventory refusal. |
| Payload comparison | Actual model `glm-5.3`, reasoning.effort=`low`, max_output_tokens4096, stream=true; only reasoning differs from own full-stream baseline | Intended low-effort control reached wire. No basis to claim an adapter effort mapping defect, nor to change prompt/budget as an invisible recovery. |

The recorder truncates only its copy and continues feeding original bytes to the Client (`recorder_test.go:145–160`). Thus a successful HTTP read and complete public usage coexist with incomplete raw evidence. `glm_harness_test.go:318–343` serializes a complete Partial snapshot into each streamed envelope, giving quadratic duplication as thinking grows;4094 deltas produced the17.7MB event artifact. `finishGLM` (`glm_finish_test.go:48–70`) refuses any >16MiB file, rather than accepting an incomplete integrity inventory.

The first early-return at `glm_harness_test.go:363–365` masks later failures. If it were absent, `:391` requires Stop (ToolUse only for tool scenes), and `:394–401` requires nonempty final Text. This observation is a real bounded-output failure in addition to two measurement defects; do not fix the gate to accept it.

Retained artifact SHA256 (diagnostic pins only, not rewriting any evidence):
- `attempt-1-exchanges.json`: `9e3278fc3c1a105ced14190a8c6322999ff6a176572ed57f472f34b4579b8dc6`
- `attempt-1-result.json`: `38e8e23a8ed384251dfc7795e439cf3384aef21c2f8db03d0d3acbe2a7a02b93`
- `events.json`: `ae9b82db402044e93de4bbc20b032dc24fdc077600b9b984c96fbfa30bd13bfc`

## Why existing mechanisms do not recover certification

1. `ai/live/main_test.go:116–119` refuses every narrowed selector before running tests; formal requires exactly `^TestGLMFormal$`. That fixed driver (`glm_harness_test.go:107–123`) has one22-scene loop and no resume, saved-session import, failed-scene retry, prior-budget ledger or completion component. `glmReserve` only increments the newly initialized process-local budget. Starting another process would reset it; that is not a legitimate use of the two remaining calls.
2. `glmSession` keeps baseline requests/results in memory. `control-low` requires its own full-stream baseline (`glm_harness_test.go:255–263`), and the single-variable check requires both final wire bodies (`:433–480`). A subtest filter alone would not provide those prerequisites even without the explicit selector guard.
3. `supportmatrix/glm.go:94–106` requires planned22, exact ordered22 scenarios and per-scenario calls<=1/attempts<=calls; `:128` requires sums exactly equal report totals. It cannot represent an original failed Attempt plus a later correction, whether hidden in one scene or added as a23rd scene. `glm_artifact_contract.go:13–31` additionally requires exactly23 manifest cases, the fixed per-case files and only attempt-1 artifacts. A hand-merged report, dropped failed call or retroactive reseal would break those contracts and omit real consumption.
4. `paid.py:156–173` requires a fresh output and creates a common-root exclusive `glm-responses` intent. The original failure must not be deleted/renamed to recreate that intent. `pin_glm_capture(:141–153)` requires a successful reviewed full child before publishing external pins; the failed run has no successful capture authority to lend to an edited bundle.
5. `renewal_glm.py:31–65` requires unchanged own PASS/complete captures, exact current catalog/source, successful canonical launcher receipt and original report/seal/support/source hashes. It is an archive/derivation tool, not a correction merger. Existing `corrective.py` is for protocol114/visual cases, not GLM (parent independently confirmed).
6. Existing GLM capture replay requires a complete original actual PASS and rejects Truncated exchanges (`glm_replay_test.go:142–155`). It never makes network calls or publishes support. Public events/Result cannot reconstruct byte-exact missing SSE fields, frame boundaries and terminal wire; synthesizing those is forbidden by the capture contract.

Relevant original contract remains `.scratch/barness-ai-glm/spec.md:222–224` (24 calls/32 Attempts per combination run,4096 tokens,120s/1800s, at most one scene fault retry accounted in the same budget, complete own report) and `:272` (truncation cannot certify byte-exact evidence). The optional scene retry allowance was not implemented as a resumable/composable producer. This implementation gap is not authority to bypass exact existing importer and launcher boundaries.

## Source blast radius: preserve the three completed M3 formals

| Proposed change class | Exact current source applicability |
| --- | --- |
| Shared `ai/live/recorder_test.go`, `main_test.go`, runtime `ai/*.go`, evidence/audit helpers, or any `ai/internal/testkit/supportmatrix/` non-test file | MiniMax `ExecutionSource` includes these (`minimaxlive/source.go:12–59`). Changes invalidate all current M3 actual formal proofs. Increasing only GLM's16MiB importer bound in supportmatrix is therefore **not** isolated from M3. |
| Only `ai/live/glm_*.go` / `ai/internal/testkit/glm/` generation files | MiniMax's actual source set excludes these; its current three formal reports can remain eligible. GLM's own source set includes all such files across routes (`supportmatrix/glm_source.go:39–69`), so **all three GLM current captures** become stale, including successful Chat and any successful Messages. |
| Pure documentation/evidence metadata | Does not change those execution sets, but exact launcher HEAD/source approval still needs preparation under its current contract. It cannot create the missing actual proof. |

A clean shared response-limit option would improve module design but cannot be made free of its actual M3 source blast radius. Conversely, a truly GLM-local capture implementation could preserve M3 applicability under the *existing* source rules, but must be reviewed as such; do not rename/exclude actual measurement code or edit source manifests simply to preserve hashes. Legal output bounds and all event/native/usage data still need lossless bounded capture. This assessment does not claim a GLM-only implementation is already designed/tested or a new compact event format already accepted.

M3 formal current spend reported by parent is39 calls/36 Attempts. At three exact20-call slots, each already spent13 calls and has only7 left; a complete fresh13-scene formal per identity cannot fit. Shared formal stage has only21 calls remaining, short of a fresh3x13=39. Preserve current M3 evidence and do not trigger source changes that require those reruns without an explicit new finite budget/contract decision.

## Compliant decision options, no implementation or call authorization implied

- **Current authorization:** finish only the parent's already authorized independent frozen-source jobs. Keep GLM Responses FAIL and final release incomplete. Do not use the spare2calls for an unreviewed diagnostic that still cannot qualify the mandatory22-scene proof. Authentication repair is not indicated.
- **If a future corrected collection is authorized:** first failure-first offline tests for >512KiB legal SSE, lossless bounded event artifacts beyond the old17.7MB example, and truthful length/no-text failure. Select a genuinely isolated GLM-only design if M3 must remain eligible; keep existing source rules, strict current formal certification, immutable originals and finite limits. Freeze/review the complete new generation code, execute exact keyless gates, then pre-register new distinct explicit collection intents with all prior actual consumption carried forward. Under existing source applicability, a GLM-only generation fix requires three new full22-scene GLM collections, not a one-scene merge.
- The new collection needs **66 nominal additional calls/Attempts**, or **72 calls/96 Attempts reserved** for three current24/32 components. After the original three jobs settle at22/22 each, GLM lifetime actual+new reserved ceilings would be138 calls/162 Attempts (66 old +72/96 new). If preserving the other302 legacy call/Attempt capacity unchanged, the total legacy ceilings would need440 calls/464 Attempts, rather than the currently approved374/398. These are conditional calculations; use actual settled Messages counters if it differs, and include any further authorized consumption. Distinct intents and explicit increased aggregate/provider accounting are required; do not reset or borrow another provider's allowance.
- A **new one-scene corrective/composition contract** could only be a separately specified/reviewed project change, with retained original failures, source/measurement lineage for every contributing scene, a canonical external origin receipt, all23actualcalls accounted, complete22 qualifying capabilities, safe immutable publication and importer support. No such current seam exists. Implementing it in supportmatrix currently invalidates M3 source too; avoiding that impact by changing the source definition would be a freshness exemption, not a proven recovery. I cannot certify this as a currently compliant smaller path.

Until the parent explicitly decides on a new finite source/collection/budget scope, this is a concrete external acceptance blocker plus capture-harness defects. There is no justified mutation, retry, report promotion or new approval to infer from this diagnosis.

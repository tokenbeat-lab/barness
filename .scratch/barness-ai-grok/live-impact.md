# Final Grok live impact selection

Read-only static analysis, 2026-10-26. Base `2f58c68d8fbb30a07e71b9666feed9979c791d01`; requested integration `5a04b34d` / code merge `5561d76d` (fix `51c95922`, merged through `e3ad302d`). Observed HEAD `86a255b746bc3688e93dac191811629f35dbce97` adds only acceptance documentation after that reviewed code. No build, tests, provider requests or repository edits performed.

## Decision

The xAI-only branches do **not** require an automatic 29-route repurchase. However, the final fix contains one genuinely shared behavior change: `opaqueHistoryIDSource` recognizes xAI Responses as a collision-allocation source. That can change the actual tool call/result IDs sent to old chat destinations when Host explicitly supplies xAI history. Under the current broad cross-source history contract (R12/R13/E18), this prevents declaring *every* old route unaffected.

Conservative SOP-compliant **full-feature** selection: `xai-responses` plus the 22 old chat routes below that execute the allocator. The two GLM Responses targets bypass it, and five nonchat identities never run this path. This is a behavior-derived set, not a set obtained from changed source hashes. An initial, separately authorized candidate run can still select only `xai-responses`; that run alone must not close the shared cross-destination validation or full release.

## Impact table

| Route(s), exact native names | Direct behavior delta | Paid native disposition for full feature |
| --- | --- | --- |
| `xai-responses` | Entire new global Grok route: request subset, native replay, image projection, SSE authority, token/cost conversion | Required own candidate run: existing 18 scenes / 24 calls, then actual captures/replay and formal qualification. No authority currently exists. |
| `openai-responses` | xAI source pairs previously normalize to common `xar1` call prefix; new allocator produces distinct target IDs, also handles occupied legacy IDs / separate turns | Required. Concrete public red/green exists in `TestXAIHistoryDowngradesToOpenAIResponses`; retained controlled HTTP cannot establish OpenAI native acceptance. Native scenario must exercise downgraded multi-call history, not only ordinary OpenAI-origin tools. |
| `deepseek-responses` | Target sanitizer/truncation can collapse long xAI opaque IDs; new allocator emits replacement call/result pairs | Required for affected collision replay. Same normalizer does not imply another Provider's live acceptance. |
| `openai-chat`, `deepseek-chat`, `zhipu-chat`, `zai-chat`, `minimax-chat`, `minimax-cn-chat`, `kimi-cn-chat`, `kimi-intl-chat` | Shared Chat ID conversion folds/truncates/hashes source IDs. Recognizing xAI now activates allocation for collisions and identical source-ID reuse across turns | Required for affected collision/reused-ID replay. Native target codecs still own distinct service contracts. |
| `anthropic-messages`, `zhipu-messages`, `zai-messages`, `minimax-messages`, `minimax-cn-messages`, `kimi-cn-messages`, `kimi-intl-messages` | Shared Messages sanitizer retains at most 64 characters. Long valid xAI opaque IDs with common prefixes can collide; allocator now preserves pair correlation | Required for affected multi-call/long-ID replay. No other Messages request/response behavior is changed by the xAI branches. |
| `google-gemini` | Gemini 3+ (published smoke `gemini-3.8-flash`) uses the same 64-character sanitizer; new xAI sources can activate collision replacement. Gemini models without call-ID normalization do not enter allocator | Required for published Gemini 3 route's affected tool replay. This is not a need to retest image generation. |
| `minimax-responses`, `minimax-cn-responses`, `kimi-cn-responses`, `kimi-intl-responses` | Destination pair wrapping is injective for distinct xAI IDs, so the usual two-call case is unchanged. But same full xAI pair reused in separate assistant turns has distinct provenance owners; classification now causes replacement allocation where previously it did not | Required if retaining the complete cross-source/repeated-ID contract exposed by the current implementation. Narrow targeted scenario can use repeated source IDs across turns; do not claim these four have the same truncation bug as OpenAI. |
| `zhipu-responses`, `zai-responses` | **Unchanged old-target wire behavior.** Allocator returns immediately for responsesGLM; existing GLM input codec already wraps the complete foreign ID as opaque call association | No new paid native validation for this feature solely from this diff. Ordinary regression and current source applicability must still be recorded. |
| `typesafe-classifier`, `typesafe-classifier-jev-1.14.0`, `openai-decisions`, `openai-images`, `google-interactions-image` | **Unchanged operation paths.** Unary image/classifier entries do not prepare chat Transcript; ProviderXAI binding/catalog guards early-return for their old identities; no native/usage/target changes | Not affected live routes for this feature. No repurchase solely because broad native inventories include added runtime files. |

The 22 old routes above are **source-conditioned** changes: old destinations with histories containing no xAI assistant are behaviorally unchanged. Do not describe this as an across-the-board wire change to ordinary old-provider calls.

## Source evidence for the selection

- `ai/opaque_history_ids.go:16–28`: skips xAI target; skips nil normalization and GLM Responses target; otherwise only enters for opaque source histories. `:139–147`: sole changed source classification includes xAI. `:46–114`: complete-transcript owner/collision inventory and replacements; owner includes turn at `:156–159`. A repeated full source ID across turns can therefore change the old destination's emitted IDs, even for injective destination codecs.
- `ai/responses_history.go:25–61`: old OpenAI splits the source ID at `|`; non-OpenAI standard targets sanitize/truncate. Kimi/MiniMax target codecs `responses_{kimi,minimax}_history.go` wrap complete foreign ID into their own call pair. GLM `responses_glm_ids.go:27–31` wraps whole foreign ID and bypasses collision allocator.
- `ai/chat_request.go:364–381`, `:416–423`: all Chat target normalizers. `ai/anthropic_stream.go:464–470`: all Messages sanitizer. `ai/gemini_stream.go:333–338` and `ai/gemini_request.go:366–369`: Gemini 3+ call-ID condition.
- All other changed shared files are xAI-gated: call.go history validation/prebuild, history.go alternate settling, preflight target canonicalization, hooks_run.go alternate cache header, openai_sdk.go request-ID header, retry.go initial cost status, Responses profile/request/options/assistant/parser/error branches. Catalog builtin data, old target native codecs, modules, and their wire usage remain unchanged. Retry/cost changes do not alter other Providers' existing branches.
- `ai/live/main_test.go` new selector is active only when xAI combo/key/exact test entry is selected; ordinary old selectors retain prior initialization/report behavior. Mixed selector refusal belongs to controlled process regression, not a reason to buy all live services.

## Proposed `Impact.LiveRoutes` for full-feature completion

```json
[
  "xai-responses",
  "openai-responses",
  "deepseek-responses",
  "openai-chat",
  "deepseek-chat",
  "zhipu-chat",
  "zai-chat",
  "minimax-chat",
  "minimax-cn-chat",
  "kimi-cn-chat",
  "kimi-intl-chat",
  "anthropic-messages",
  "zhipu-messages",
  "zai-messages",
  "minimax-messages",
  "minimax-cn-messages",
  "kimi-cn-messages",
  "kimi-intl-messages",
  "google-gemini",
  "minimax-responses",
  "minimax-cn-responses",
  "kimi-cn-responses",
  "kimi-intl-responses"
]
```

This list records unresolved requirements; it is not authorization, a plan to run all routes now, or a grant to renew any closed native ledger. Prepare exact bounded cases / native consumers / budgets before asking for any additional paid stage. Existing old-route smoke can be useful but does not cover the new xAI-origin collision trigger by itself. Keep each outstanding route NOT_RUN until its actual own evidence is valid.

There is a narrower engineering alternative, but it is **not** the current code: confine the new collision recognition to explicitly supported old destinations, with an explicit public contract for all remaining cross-destination histories, then repeat red/green and independent review. Do not silently select only OpenAI while leaving the broad new behavior in place. Current R13/E18 and existing cross-Provider API support make a blanket zero-old-route assertion unjustified.

## Historical native source applicability is a separate fact

`supportmatrix.RegionalExecutionSource`, `supportmatrix.GLMExecutionSource`, and `minimaxlive.ExecutionSource` hash every `ai/*.go` runtime file (plus their declared native harness/consumer/module sources). New xAI runtime files and shared gated edits change those inventories, including for wire-unaffected identities. Their old formal reports cannot be repinned or relabeled current, even if no paid rerun is selected for Grok. Record historical PASS as historical; a standalone global catalog/current-source renewal remains separate work with new authority, not a prerequisite to claim every unrelated operation newly retested.

Feature release scope and global catalogue renewal must not be conflated. This Grok selection has 23 routes because of the actual allocator behavior plus xAI itself, not all 30 including Grok because of broad source fingerprints. The currently implemented xAI candidate report is explicitly nonformal and native release consumption remains closed, so even a successful 24-call candidate run cannot by itself set `release.pass=true`, activate BuiltinCatalog, or advance to race/pressure.

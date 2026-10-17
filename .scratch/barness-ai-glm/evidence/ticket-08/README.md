# Ticket 08 — GLM failures, bounded retry and usage

Date: 2026-10-17 (Asia/Shanghai). The approved spec T01 seam is the public
Client with controlled Provider/Host, separately versioned protocol candidates
and synthetic credentials. No `.env` was read and no paid request or deliberate
real authentication/billing/rate failure was sent. [Failure modes](failure-modes.md)
preceded implementation. [Hashed index](index.json) pins fixtures, sources,
completed run manifests, assertions, audit and durable ZIP archives.

## Protocol and source boundaries

- [Official numeric error table snapshot](sources/api-code.md), retrieved from
  `https://docs.bigmodel.cn/cn/api/api-code.md`, supplies Chat meanings. Besides
  1113/1308/1309/1311 it documents 1310 and 1313–1321 as persistent account,
  subscription, policy or spend-limit restrictions. All known permanent codes
  suppress retries even under HTTP429 and `x-should-retry:true`; 1302/1305 and
  unknown429 retain the existing HTTP/Binding policy. Chat accepts numeric and
  quoted numeric codes because its OpenAPI Error declares string while the
  business table uses numbers. Unknown codes are not reinterpreted as billing.
- [Responses own schema](sources/responses-errors.json) explicitly states its
  **named string codes are a different set**: insufficient_quota, quota_exceeded,
  usage_limit_reached and usage_not_included are permanent; rate_limit_exceeded,
  server_is_overloaded, overloaded and slow_down are transient candidates.
  Chat1113 is not used to certify a Responses billing wire. The initial
  `red-http` diagnostic used quoted numeric Responses codes; the corrected
  `red-http-protocols` run is the protocol-accurate behavior red.
- [Chat own Error schema](sources/chat-errors.json) and the error-code table are
  independent of the Responses code list. Chat sensitive/network_error/
  model_context_window_exceeded preserve RawStopReason and fail through the
  public Result; none maps to ordinary successful stop.
- Messages nested `{type:error,error:{type,message,code?}}` errors, including
  billing codes, are **synthetic compatibility fixtures**, not actual observed
  GLM billing/error wire. Type-only auth/permission/invalid/not-found objects
  are permanent, rate_limit/overloaded use HTTP fallback; unexpected envelopes
  remain safe generic diagnostics. No live Messages failure was induced to
  manufacture evidence. The official China guide has no error-event schema.
- Request-ID source is `X-Request-Id`, independently confirmed in real successful
  captures for all three protocols (implementation-research and existing
  protocol exploration indices). Tests also supply a conflicting DeepSeek ID
  and an absent GLM ID; neither DeepSeek nor local IDs can fill the absence.
  Existing Messages header selection from ticket06 is reused.

All GLM public errors omit upstream free-form messages and metadata echoes;
only documented code/type values or bounded canonical numeric codes enter
safe diagnostics. HTTP Code/Phase/Status, requested retry delay and provider ID
remain available; GLM stream errors retain actual HTTP200. Observer remains its
existing metadata-only schema. [ADR0006](../../../../docs/adr/0006-barness-ai-retry-policy-on-binding.md)
records the explicit exception to generic retry-header precedence.

## Verified caller behavior

Both glm-5.3 and glm-5.2 independently cover all three protocols. Defaults send
one HTTP attempt; explicit bounded Binding retries are pre-stream only. A
credential/Binding rotation during the first response cannot change later
attempts; OnPayload runs once and final byte-identical payload and the same
credential, model and endpoint are retained. SDK retries are disabled. Stream
errors, early EOF and body-read interruption never replay the generation;
every body closes and Attempt retains its actual failure/usage metadata.

Normal, complete-zero, absent, partial, cached, reasoning and interrupted usage
are protocol-specific fixtures. Input excludes cache already included in
Chat/Responses totals; reasoning is inside Output and is not added or priced
twice. Messages input is already exclusive of separately reported cache.
Actual Messages initial input/output0 is replaced by message_delta counts;
absent cache_creation stays partial and no separate reasoning count is
invented. GLM one-hour cache metadata stays unset even if a synthetic native-shaped
TTL detail arrives; aggregate CacheWrite uses only the configured rate, without
Anthropic’s 2×Input premium. Native Anthropic TTL pricing is unchanged.
Explicit cache-read4/cache-creation0 fixtures are synthetic arithmetic
variants, not observed cache hits. Responses missing total becomes its known
sum while reporting stays partial; response.failed usage remains unreported
under the preexisting contract. Candidate rates are synthetic and do not
certify pricing or account billing. See [ADR0010](../../../../docs/adr/0010-barness-ai-usage-reporting-and-price-snapshot.md).

Cancellation ordering/resource scenarios remain ticket09; the explicit
[handoff](handoff-09.md) distinguishes preterminal known Chat/Messages usage
from Responses terminal-only usage. Existing generic/native cancellation
regressions ran; this ticket does not invent post-terminal cancellation cases.

## Runs and replay

| Slice | Outcome |
| --- | --- |
| Correct own-protocol HTTP red → green | 260 FAIL → 260 PASS |
| Valid SSE usage/stream red → green | 24 FAIL / 32 PASS → 56 PASS |
| Messages type-only error red | 12 FAIL before the corresponding fix |
| Initial full new public Client slice | 346 PASS |
| New slice under race | 346 PASS |
| Affected GLM/native/http/retry/usage/callback regressions | 2,538 PASS |
| Frozen three-protocol differential | 608 PASS, zero pending |
| Messages TTL absence/premium red → green (plus native usage regression) | 14 FAIL → 136 PASS |
| Latest merged GLM suite (ticket07 integration included) | 1,108 PASS, including all 348 current ticket08 cases |
| Final GLM usage + native Anthropic usage race | 50 PASS |
| `go vet ./ai/...`, staged/unstaged whitespace | PASS |

The first stream harness run emitted indented JSON over single data lines, so
Messages parsed invalid frames. It is retained as a harness diagnostic;
`red-stream-usage-valid` compacted each event before the behavior red. All
completed runs contain their own audit, including intentional-red runs.
The final regression and race runs were performed after staging the new source
inventory; manifests/cases pin the actual source/request/response/events/Result
and assertion artifacts. No global release gate/support matrix is changed.

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-08/replay" \
  go test ./ai/e2e -run '^TestGLM(HTTPFailures|UsagePresence|StreamFailures|RetrySnapshot|FailureRequestIDAbsent)$' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-08/replay-race" \
  go test -race ./ai/e2e -run '^TestGLM(HTTPFailures|UsagePresence|StreamFailures|RetrySnapshot|FailureRequestIDAbsent)$' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-08/replay-regressions" \
  go test ./ai/e2e -run '^Test(GLM|Protocol|ExplicitRetry|UsageAndCost|FailureTerminals|Anthropic|Chat|DeepSeek|TrustedCallback)' -count=1
go vet ./ai/...
# Requires the locked oracle's dependencies (npm ci --ignore-scripts in its node dir):
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/(openai-responses|openai-completions|anthropic-messages)$' -count=1
```

Durable audited bundles/archives are at repo-root
`.evidence/barness-ai-glm/ticket-08/`, copied out of the temporary worktree.
The temporary oracle node_modules symlink is never committed. Source snapshot
changes require a new source/fixture index; rollback is the ordinary integrated
code rollback or disabling the Host GLM Binding, with no fallback or shim.

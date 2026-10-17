# Ticket 05 shared lifecycle evidence

Mode: offline synthetic public Client E2E; no real key or vendor request.
`failure-modes.md` predates implementation. Only external HTTP/SSE, Host identity,
Admission and Observer boundaries are controlled. Every kept full bundle retains
requests, response scripts, events/Result/Attempt/Usage, Observer records, resource
probes, manifest replay commands and a completed redaction audit.

Implemented fixes:

- Chat uses explicit ThinkingSwitchOnly, independently of mandatory thinking or
  a Host narrowing M3.1 to a single high depth. The final wire keeps that effort.
- MiniMax header/payload/response callback panics become private classified
  callback failures, with one lifecycle completion and released resources.
- MiniMax HTTP and SSE errors never expose free-form response bodies. HTTP facts
  and the finite documented Messages error.type remain diagnostic. Permanent
  HTTP request/auth/permission failures and documented Messages permanent types
  cannot be made retryable by x-should-retry. Unknown 429 and transient 408/409/5xx
  retain existing Host-bounded policy; the SDKs still retry nothing themselves.
- MiniMax post-send callback/stream failures retain actual HTTP status200.
- Exact output-budget predicate is now provider-neutral outputBudgetNumber;
  the old glm name/file was deleted. The public protocol wire-pair inspector is
  similarly shared by GLM/MiniMax without provider capability inference.

Known Messages categories come from the independently documented compatible
ErrorResponse in the 2026-10-18 official OpenAPI snapshot:
https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json
SHA-256 57111d5dddd2ada7f481e5c373d88c745eba0b852b60bf1ca20096c61f28b77c.
The source enumerates invalid_request_error, authentication_error, permission_error,
not_found_error, request_too_large, rate_limit_error, api_error and overloaded_error.
It does not establish compatible billing/MPlan codes. Native base_resp 1008 in a
synthetic transient error does not block its bounded retry; no message words are
used to guess billing semantics. Actual compatible quota shapes and safe trace/
request-id header/body priority remain evidence obligations for live ticket 10.
Only headers actually present in each own synthetic response are asserted here.

The 202 own cases cover both models and three protocols, all 13 resource/time
modes, parsed-prefix cancellation and independent sibling tool continuation,
queue/frame repeated concurrent release, tenant admission and bounded waiters,
precredential authorization, callback guards/panics, frozen input/options/catalog/
policy/credentials and retry bytes, native storage/trust/downgrades, all cross-
protocol conversions, tenant/purpose/account cache isolation, Observer failures
and payload privacy. The own resource cases finish Result without consuming Next
and then require one terminal plus zero bodies/calls/permits/waiters/events.

Mixed load preserves the original TestMixedPressure local 4/cloud 8 saturation
batches unchanged, including 128 KiB history/16 KiB chat output, full image/classifier
loads and 256/512 MiB budgets. Those baseline permits intentionally reach100%.
Additional rotating local 3/cloud 6 batches hold two old routes plus one MiniMax
protocol per tenant at once. All original old routes and all MiniMax protocols
appear, with unchanged policy and sizes. Every additional reported bound stays
<=75%, including process/tenant permits. Complete runs create no stream event
queue; its measured zero is stated, while separate unread-stream cases enforce
queue backpressure. Memory records include reachable heap and conservative total
allocations; every batch ends with all resource gauges at zero.

Initial fixture corrections are retained where material: the shared resource
world initially inherited image/classifier subpolicies incompatible with tiny
isolated text limits, then switched to the repository's existing bounded text
policy for resource cases only; load still selects unchanged mixed policies.
A credential ownership denial expectation was corrected from credential phase to
its existing consistency phase. Neither correction weakened a production guard.

Replay:

```sh
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run '^TestMiniMaxShared' -count=1
BARNESS_AI_PRESSURE=1 go test -race ./ai/e2e -run '^TestMiniMaxShared' -count=1
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run '^Test(MiniMaxSharedDesignLoad|MixedPressure)$' -count=1
go vet ./ai ./ai/internal/testkit/minimax
```

Meaningful RED captures: narrowed depth omission; callback panic escape; HTTP/SSE
private-text exposure/status loss; documented Messages permanent type retried;
permanent HTTP status retried despite its semantics. Full current GREEN/race and
committed-source acceptance artifacts are bound in SHA256SUMS. Broad regression
records 3311 passing cases, including 202 own cases and all impacted old protocol
families; no full slow release gate or live acceptance is claimed here.


Final committed-source acceptance: **270 PASS** cases at `4ca7a11` (202 own,
66 existing GLM native/cross-protocol regression, two unchanged original design
loads). All successful retained bundles have no redaction findings. Own race:
202 PASS at the same final runtime; vet PASS. `green-summary.json` records each
of the six added load batches' permit peak, <=75% maximum bound, unchanged
baseline4/8 identity and actual allocation against256/512MiB. Earlier broad
3311-case regression metadata is compacted in regression-metadata.tar.gz; its
summary states the later finite SSE-type-only diagnostic change separately.

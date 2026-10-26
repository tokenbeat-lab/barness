# Grok live/qualification implementation notes

Read-only exploration, 2026-10-26. Sources: repo AGENTS.md, GLOSSARY.md, docs/agents/verification.md; `.scratch/barness-ai-grok/{spec,design}.md`, issues 04–06. No tests or provider requests executed; no repository files edited.

## Main conclusion

Deliver the candidate adapter and ordinary public E2E first. Keep 05/06 open until the separate authorizations and real evidence exist. A generic combo registration alone is insufficient and potentially gives false confidence. The most coherent extension is a small isolated xAI candidate runner reusing existing public Client assembly, recorder, evidence Run, and supportmatrix scenario/report shapes. Keep formal native acceptance fail-closed until xAI's own captures and full capability proof exist. Do not add xAI to the regional identity table simply to reuse its contract: the regional routes have independent authentication, model/price and proof semantics.

## Useful existing implementation seams

- `ai/live/main_test.go`: TestMain chooses isolated runners before generic legacy config; `newClientWithPolicy` reuses Host resolver, fixed Binding/credential snapshot, recorder/Observer, candidate Catalog. New candidate runner can use this assembly directly. `livePolicy()` is finite and includes 2 images and resource limits, but xAI frame/output bounds need account for encrypted items.
- `ai/live/recorder_test.go`: explicit direct transport, redacted bearer capture, bounded complete SSE capture, terminal Observer collection. Reuse rather than a second capture implementation. Request capture cap 256KiB and response cap 512KiB: a truncated encrypted/native capture must fail qualification, not be called a full replay fixture.
- `ai/live/stream_test.go`: `drain`, event/block lifetime checks. Reuse without protocol internals.
- `ai/live/scenarios_test.go`: `redSquarePNG`, `weatherTool`, `nativeArtifacts`, `sentStrings` reusable; some generic scenarios need xAI-specific replacements below.
- `ai/live/session_test.go`: generic scenario budget/retries do not provide a new one-process grant/claim or USD cap; generic retries can replay a whole two-turn scenario. Prefer explicit bounded xAI scenario plan and zero automatic retries for a new paid run.
- `ai/live/regional_config_test.go`, `ai/internal/testkit/supportmatrix/regional_{grant,reservation}.go`: examples of source/binary/catalog/identity-bound single-process grant and durable reservation in common Git storage. These are reference shapes, not reusable xAI authorization by claiming a regional route.
- `tools/verification/run.py` and `live_environment.py`: generic selector `BARNESS_AI_LIVE_COMBO=xai-responses` already forwards its exact derived key plus account/region/permissions/proxy fields. If a new xAI-specific grant selector/env is introduced, explicitly register/forward only that native contract and add public process regression. No all-env forwarding.

## Why appending a generic combo is insufficient

1. `combos_test.go::responsesOptions` named forced choice uses `ResponsesToolChoice{Function: weatherTool.Name}`. xAI only allows required/none/auto; use required plus exactly one declared function.
2. `scenarios_test.go::toolRoundTrip` computes wire ID with `strings.Cut(call.ID,"|")`. This fails xAI's `xar1|base64url([call,item])`. Inspect captured `function_call`/`function_call_output` pairs instead; never put the opaque local ID directly on wire.
3. Generic `textComplete` only covers CompleteSimple and `textStream` only Stream. E17 requires all four entry points explicitly: Stream, Complete, StreamSimple, CompleteSimple.
4. Generic reasoning-history only proves some native string replayed. xAI requires nonzero reasoning arithmetic, entire reasoning item/encrypted content exact semantic replay, no downgrade, and default vs each low/medium/high/xhigh.
5. Generic image only tests user PNG. E17 also requires JPEG and tool-result image projection/order/call pairing; provide finite own scenarios.
6. `release/chat_qualification.go` assumes normalized output <= calls*max_output_tokens. xAI budget only bounds visible output, so cannot reuse that assumption for reasoning/tool consumption.
7. Global `catalog = BuiltinCatalog()` excludes candidates. Candidate runner must explicitly select its candidate snapshot; don't activate production Catalog merely to make tests callable.
8. `release/record_live.go::qualifiedLive` requires BuiltinCatalog version/hash. That is proper for formal proof but must not accept a candidate as already formal. Implement xAI candidate qualification separately, or keep the formal consumer rejecting it until promotion requirements are met.
9. Generic `verifyChatQualification` default currently returns nil for unknown combo, but `ValidIdentity` and expectedCapabilities reject xAI later. Registering identity without own complete fixed contract would weaken this fail-closed behavior.
10. `session.observe` searches arbitrary id-like captured headers if public ProviderRequestID is empty. Keep this separate as capture fact; xAI public request-ID adoption awaits real source confirmation. Do not fabricate a public ID from that search.

## Minimal files/components for ticket 04

- Add `ai/live/xai_{config,harness,scenarios,...}_test.go` (<500 lines each) with exact `^TestXAICandidate$` selection, explicit candidate catalog, isolated official endpoint, account/region aliases, and finite call/time/input/output reservation scope.
- Hook isolated selection/report finalization into TestMain using a few clear branches; prevent generic TestLive from also selecting another route.
- Reuse existing `combo` and newClientWithPolicy internally if useful; a private xAI `combo` need not join published generic `combos` until formal support.
- Candidate report should record `mode`, `stage`, source/catalog/SDK/binary hashes, exact expected scenario IDs, actual calls/HTTP attempts, nonzero reasoning evidence, and grant/claim IDs. State candidate identity explicitly; candidate/controlled/replay must never produce formal PASS.
- Budget grant authorization is a validation harness boundary, not a new Host authentication framework. Reuse existing trusted Host resolver for real calls. Spec does not authorize execution or provide a numeric spend cap: prepare concrete grant scope for the user; do not invent a paid authorization.
- New process/public E2E entry such as `TestXAILiveHarnessProcess`: compile the live package under isolated environment, then run only bounded no-key/invalid-grant/controlled entries. Do not launch all live-tag tests under inherited environment. Ordinary process evidence retains argv/exit/manifest/audit.
- Process failure scenarios first: no live flag/key/grant => NOT_RUN and zero HTTP; wrong identity/model/endpoint/catalog/source/expiry/alias, mixed route credentials, duplicate grant use, invalid entry => refusal and zero HTTP; nonzero child exit/cancel/timeout/evidence final write failure cannot yield PASS; any fixture response is controlled, not real qualification.
- Host doc should show candidate explicit Catalog, fixed `https://api.x.ai/v1`, same public resolvers, versioned Binding, resource policy/timeouts, TrustNativeState after persistence, tool results, cost statuses, and disabling Binding for rollback. Keep 05/06 tracked as pending.

## Finite E17 scenario plan for a future authorized run

Four entry-point texts; default and low/medium/high/xhigh; nonzero reasoning and usage arithmetic; full encrypted reasoning continuation (JSON roundtrip + Host TrustNativeState); auto/required/none tool behavior; parallel call/result pairing; PNG/JPEG user images; tool-result image roundtrip; cache key shape and observed zero/nonzero cache reporting; actual ResponseModel. Cancel after first frame only if explicitly in grant. Keep test conversations synthetic and captures complete, sanitized and audited. No 200k real prompt is needed to prove the price threshold: synthetic public E2E plus fixed official price source are sufficient per design. No unbounded attempts to manufacture a cache hit.

## Formal/native qualification integration (05/06, after evidence)

- `ai/internal/testkit/supportmatrix/identity.go`: exact xai-responses / chat / xai / openai-responses identity.
- Own fixed xAI expected capabilities in supportmatrix and `release/live.go::expectedCapabilities`; prohibit UNSUPPORTED for every required xAI capability (R21).
- Own check in `supportmatrix` validates fixed global endpoint/model/SDK, full nonduplicated scenarios, actual bounded calls, native capture/result/Observer triplets, source pin and consumed grant/claim. Reuse common report merge shape; do not infer support from another provider's report.
- `release/chat_qualification.go` needs xAI-specific qualification or independent branch that does not apply generic total output cap; `release/chat_capture.go` can supply attribution/call-triplet checks, augmented with xAI encrypted replay/tool/image/usage facts.
- `release/record_live.go` must continue requiring formal snapshot. If preactivation candidate proof is consumed for promotion, label it candidate and build an explicit reviewed activation from source-backed captures; do not silently pretend candidate hash equals BuiltinCatalog.
- `ai/live/support-matrix.json`: a NOT_RUN xAI row may be added independently of formal activation, but do not mark support without own report. `ai/catalog_*` production activation is separate, only after required evidence and source requirements.
- Native consumer command remains `consume_live.py` -> compiled `releasegate -live-bundle ...`, preserving outer schema-2 capture and actual numeric exit. Need process tests for candidate/controlled rejection, missing capability, substituted identity, unknown IDs, corrupt/truncated captures, and source/catalog drift before opening acceptance.
- Race and pressure require fresh human selections after valid live. Missing stages remain NOT_RUN, `release.pass=false`; do not close ticket06 or call formal support shipped.

## Old route/regression impact

The direct behavior surface is Responses: OpenAI, DeepSeek, Zhipu/ZAI, MiniMax/MiniMax-CN, Moonshot/Moonshot-CN. Provider-specific zero/default branches should preserve them; ordinary regression should cover all these routes plus shared native/history/ID/usage/resource tests and complete `./ai/e2e ./ai` per current plan. If only additive xAI branch semantics are changed, non-Responses wire behavior has no direct delta, but test the shared Client lifecycle with a mixed old provider.

Important source applicability: `supportmatrix.RegionalExecutionSource()` hashes every runtime `ai/*.go`, all live Go, internal non-test Go, catalog_kimi.json and modules; GLM similarly hashes every runtime Go and its capture dependencies. Adding xAI runtime files changes these native source sets even when their runtime branches are untouched. Thus historical regional/Kimi/GLM formal evidence does not become current automatically. List source invalidation separately from semantic behavior; assess affected live requirements in final review and leave them pending rather than copying historical PASS. `tools/verification` execution source is broader still. No reason to rebuild the frozen pi oracle.

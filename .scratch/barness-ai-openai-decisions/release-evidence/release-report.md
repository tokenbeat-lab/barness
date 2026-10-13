# barness-ai release gate: PASS

Commit `807ccbaa53b6764e2d03299cc582ef560fe87165`; offline bundle `.evidence/issue13-final-gate/offline/20261013T034156.007664000Z`.

| Gate | Result | Problems |
| --- | --- | --- |
| go-test — go test ./... (offline, differential on) | PASS | — |
| go-test-race — go test -race ./... | PASS | — |
| go-vet — go vet ./... (with and without the live tag) | PASS | — |
| p0-offline — every offline P0 case passes | PASS | — |
| differential — frozen pi differential: no pending difference | PASS | — |
| live — every combination fully passed its live smoke (PASS or explicit UNSUPPORTED) | PASS | — |
| redaction-audit — redaction audit: no key, Authorization or real content in any bundle | PASS | — |
| traceability — traceability: research item → scenario → evidence | PASS | — |
| snapshots — checked-in snapshots match the code | PASS | — |
| required-artifacts — mixed image design loads and independent Chinese business evaluation | PASS | — |

### Commands

| Command | Result | Duration |
| --- | --- | --- |
| `go vet ./...` | PASS | 0s |
| `go vet -tags live ./...` | PASS | 0s |
| `BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -count=1 ./...` | PASS | 6m2s |
| `BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race -count=1 -timeout=20m ./...` | PASS | 20m25s |

## Offline

6755 cases: 6755 PASS, 0 FAIL, 0 NOT_RUN, 0 UNSUPPORTED, differential cases excluded.

| P0 scenario | Pass | Fail | Not run | Unsupported |
| --- | --- | --- | --- | --- |
| D1 | 35 | 0 | 0 | 0 |
| D2 | 24 | 0 | 0 | 0 |
| E01 | 138 | 0 | 0 | 0 |
| E02 | 710 | 0 | 0 | 0 |
| E03 | 871 | 0 | 0 | 0 |
| E04 | 1528 | 0 | 0 | 0 |
| E05 | 236 | 0 | 0 | 0 |
| E06 | 175 | 0 | 0 | 0 |
| E07 | 421 | 0 | 0 | 0 |
| E08 | 936 | 0 | 0 | 0 |
| E09 | 57 | 0 | 0 | 0 |
| E10 | 119 | 0 | 0 | 0 |
| E11 | 154 | 0 | 0 | 0 |
| P01 | 960 | 0 | 0 | 0 |
| P02 | 293 | 0 | 0 | 0 |
| P03 | 293 | 0 | 0 | 0 |
| P04 | 574 | 0 | 0 | 0 |
| P05 | 339 | 0 | 0 | 0 |
| P06 | 351 | 0 | 0 | 0 |
| P07 | 433 | 0 | 0 | 0 |
| P08 | 267 | 0 | 0 | 0 |
| P09 | 300 | 0 | 0 | 0 |
| P10 | 1032 | 0 | 0 | 0 |
| PRESSURE | 15 | 0 | 0 | 0 |

## Differential

861 cases: 861 PASS, 0 FAIL, 0 NOT_RUN, 0 UNSUPPORTED against frozen pi; 0 pending finding(s). Ledger: 278 extension, 39 fixed, 0 pending decision(s); extension routes outside pi: google × google-interactions deepseek × openai-responses openai × openai-images.

| Protocol | Pass | Fail | Not run |
| --- | --- | --- | --- |
| P01 | 335 | 0 | 0 |
| P02 | 70 | 0 | 0 |
| P03 | 77 | 0 | 0 |
| P04 | 203 | 0 | 0 |
| P06 | 118 | 0 | 0 |
| P07 | 12 | 0 | 0 |
| P10 | 46 | 0 | 0 |

## Live

From the support matrix (ai/live/support-matrix.json).

| Combination | Operation | Status | Model | SDK | Account | Last complete pass | Unsupported | Failing |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| openai-responses | chat | PASS | gpt-5-mini | github.com/openai/openai-go/v3 v3.66.0 | local-openai@region-permissions-unconfirmed | 2026-10-13T03:20:36Z | — | — |
| anthropic-messages | chat | PASS | claude-haiku-4-5 | github.com/anthropics/anthropic-sdk-go v1.75.0 (HTTP; SSE decoded by barness-ai, ADR-0011) | local-anthropic@region-permissions-unconfirmed | 2026-10-13T03:21:06Z | — | — |
| google-gemini | chat | PASS | gemini-3.8-flash | direct HTTP (ADR-0012) | local-google@region-permissions-unconfirmed | 2026-10-13T03:21:53Z | — | — |
| openai-chat | chat | PASS | gpt-5-mini | github.com/openai/openai-go/v3 v3.66.0 | local-openai@region-permissions-unconfirmed | 2026-10-13T03:22:08Z | reasoning-history: OpenAI Chat Completions returns no reasoning content or signature to replay; reasoning is only counted in usage | — |
| deepseek-responses | chat | PASS | deepseek-flash | github.com/openai/openai-go/v3 v3.66.0 | local-deepseek@region-permissions-unconfirmed | 2026-10-13T03:22:16Z | — | — |
| deepseek-chat | chat | PASS | deepseek-flash | github.com/openai/openai-go/v3 v3.66.0 | local-deepseek@region-permissions-unconfirmed | 2026-10-13T03:22:26Z | — | — |
| typesafe-classifier | classifier | PASS | jev-1.13.0 | direct HTTP (ADR-0021; typesafe-system-one v1) | local-typesafe@region-permissions-unconfirmed | 2026-10-13T03:22:28Z | — | — |
| openai-images | image | PASS | gpt-image-2.5-sunburst-2026-09-08 | direct HTTP (OpenAI Images JSON; ADR-0022) | local-openai@region-permissions-unconfirmed | 2026-10-13T03:23:02Z | — | — |
| google-interactions-image | image | PASS | gemini-nano-banana-2.1 | direct HTTP (Google Interactions v1beta; ADR-0023) | local-google@region-permissions-unconfirmed | 2026-10-13T03:23:40Z | — | — |
| openai-decisions | classifier | PASS | gpt-6-luna | direct HTTP (OpenAI Decisions; ADR-0026) | live-unverified-account | 2026-10-13T03:23:44Z | — | — |

## Traceability

Status comes from evidence only; an item that is only mapped is NO_EVIDENCE, never PASS.

| Item | Source | Requirement | Scenarios | Status | Offline | Differential | Live | Artifacts | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| OD07-frozen-decisions | OpenAI Decisions issue 07 / R02 R08 R15 R19 / OD14 OD48 OD49 OD50 | The unique frozen Pi 1.1.0 Decisions classifier/catalog differential compares complete wire and named basic answers; finite case/path differences retain full captures, source/request/result hashes, manifest and audit. Existing Jev evidence remains required. | P10 PIDIFF | PASS | — | 46/46 | — | — | PIDIFF-P10-OD14-numeric-order-hooked-classify; PIDIFF-P10-OD14-numeric-order-plain-classify; PIDIFF-P10-OD48-catalog; … 3 more |
| C11-ChineseEffect | issue 16 / US87 / ADR-0024 | Independent fixed synthetic Chinese dataset and real jev-1.13.0 COMPLETE report, full denominators, confusion/calibration and attempts; integrity and audit verified separately from protocol smoke, without an accuracy threshold | ARTIFACTS | PASS | — | — | — | 1/1 | artifact:chinese-effect-jev=PASS |
| H5-design-load-artifacts | issue 15–17 / ADR-0017 | Actual local and cloud mixed image design-load reports with measured whole JSON/base64, concurrency, allocation upper bounds, all limit ratios <=75% and released resources | ARTIFACTS | PASS | — | — | — | 2/2 | artifact:mixed-pressure-local=PASS; artifact:mixed-pressure-cloud=PASS |
| T03-mixed-operation-isolation | issue 15 / ADR-0001/0003/0020 | Concurrent chat/image/classifier with same-named tenant bindings and model IDs, legal same-account credential sharing, preflight operation/options refusals, fixed retry snapshots and independent route disabling | E07 | PASS | 49/49 | — | — | — | E07-mixed-binding-disabled-classifier; E07-mixed-binding-disabled-google-image; E07-mixed-binding-disabled-openai-image; … 3 more |
| C08-mixed-lifecycle | issue 15 / ADR-0008 | Mixed cancellation/deadlines/failures and tenant/process/waiter/time caps release only own resources; complete body and validation remain under admission | E08 | PASS | 38/38 | — | — | — | E08-mixed-admission-chat-process-cancel; E08-mixed-admission-chat-tenant-cancel; E08-mixed-admission-chat-wait-timeout; … 3 more |
| C09-mixed-observer | issue 15 / ADR-0009 | Mixed attributed metadata and usage only; slow/congested/error/panic Observer cannot change model success or failure and statistics identify loss | E09 | PASS | 4/4 | — | — | — | E09-mixed-observer-congested; E09-mixed-observer-error; E09-mixed-observer-panic; … 1 more |
| H5-mixed-design-load | issue 15 / ADR-0002/0017 | Explicit image/classifier policies with measured whole JSON/base64, atomic limits, reachable heap budgets, Provider copies, throughput and at most 75% byte-limit use; opt-in NOT_RUN | PRESSURE E08 | PASS | 18/18 | — | — | — | E08-mixed-image-budget-google-image-above-frame; E08-mixed-image-budget-google-image-count; E08-mixed-image-budget-google-image-input-image; … 3 more |
| C10-mixed-hosts | issue 15 / ADR-0001/0003 | Local and cloud public examples with independent bindings, legal shared credentials, authenticated identity, cancellation, typed results and host-owned confidence/routing | E10 | PASS | 11/11 | — | — | — | E10-mixed-host-cloud; E10-mixed-host-local; E10-mixed-local-guard-disabled; … 3 more |
| V6-google-images-live | issue 14 / ADR-0016/0023 | Own fixed Nano Banana 2.1 on official v1beta; required generation/reference editing, four-call/image 1K budget, complete synchronous inline evidence, honest absent identity, atomic own-report matrix merge; separate live gate requires its own audited bundle | P09 | PASS | 29/29 | — | — | — | P09-E01-live-replay-generation-captured; P09-E01-live-replay-generation-partial; P09-E01-live-replay-generation-unreported; … 3 more |
| H2-google-images-catalog | issue 14 / ADR-0010/0018/0023 | Own live before builtin inclusion; verified counts/1K/1:1, dated modality rates; image-only output detail excludes text thought, Output contains Reasoning, native TotalTokens retained, nullable partial/unreported without invented missing text/cache or per-image pricing | P09 | PASS | 7/7 | — | — | — | P09-E01-live-replay-generation-captured; P09-E01-live-replay-generation-partial; P09-E01-live-replay-generation-unreported; … 3 more |
| C04-google-images-delivery-omission | issue 14 / maintainer-approved ADR-0023 | Always omit refused delivery field, forbid callback insertion, still reject URI/continuation/non-completed output without protocol/version/model fallback | P09 | PASS | 10/10 | — | — | — | P09-E01-live-replay-generation-captured; P09-E01-live-replay-generation-partial; P09-E01-live-replay-generation-unreported; … 3 more |
| V4-google-images-editing | issue 13 / design 8.1, 12.3–12.9 / ADR-0022 | Same public GenerateImages operation, prompt then ordered inline reference images, independent input snapshots, all output steps, atomic failures retaining usage and no replay | P09 | PASS | 152/152 | — | — | — | P09-E01-edit-ordered; P09-E01-live-replay-reference-edit-captured; P09-E01-live-replay-reference-edit-partial; … 3 more |
| H5-google-images-input-budgets | issue 13 / ADR-0002/0022 | Host/model/14 reference intersection; per-image bytes and final decimal 20 MB or smaller host budget including encoding overhead, checked before copies and again after callbacks | P09 | PASS | 29/29 | — | — | — | P09-E07-edit-budget-decimal-protocol-bytes; P09-E07-edit-budget-encoding-overhead; P09-E07-edit-budget-exact-request-bytes; … 3 more |
| C04-google-images-edit-final | issue 13 / design 13 / ADR-0022 | Every forbidden field/fixed-field deletion refused on editing; final input MIME/base64/count/bytes and model options revalidated, no foreign identity/URI/mask or interleaving | P09 | PASS | 110/110 | — | — | — | P09-E02-audio-edit-replay; P09-E02-bad-base64-edit-replay; P09-E02-bad-content-edit-replay; … 3 more |
| V4-google-images-generation | issue 12 / ADR-0012/0022 | Native v1beta synchronous stateless inline generation, ordered model_output collection, real terminal mapping, atomic images and extension registration without premature live claims | P09 | PASS | 300/300 | — | — | — | P09-E01-edit-ordered; P09-E01-generation; P09-E01-interleaved; … 3 more |
| C04-google-images-final | issue 12–13 / design 12–13 / ADR-0022 | Final callback model/fixed fields/schema/20 MB guard, protocol options, no continuation/external resources, ordered final inline references | P09 | PASS | 134/134 | — | — | — | P09-E04-allowed-options; P09-E04-bad-input; P09-E04-delete-format; … 3 more |
| H5-google-images-budgets | issue 12 / ADR-0002/0007/0008/0022 | Total JSON/image budgets, admission held through complete validation, cancellation/read timeout, body/permit release and no 2xx replay | P09 | PASS | 58/58 | — | — | — | P09-E08-edit-allocation-byte-data; P09-E08-edit-allocation-custom-key; P09-E08-edit-allocation-escaped-raw; … 3 more |
| T09-google-images-modalities | issue 12 / ADR-0009/0010/0022 | Shared complete/partial/unreported axis, 7/42/22/49 normalization, explicit thought inclusion pricing, sourced fixtures and independent metadata-only modality audit | P09 | PASS | 31/31 | — | — | — | P09-E09-modality-audit; P09-E11-duplicate-modality; P09-E11-duplicate-modality-edit-replay; … 3 more |
| V6-openai-images-live | issue 11 / ADR-0016/0022 | Own fixed-model edit-first JSON acceptance, generation and optional mask, four-image 1K budget, audited actual consumption and atomic own-route matrix merge; separate live gate requires its own audited bundle | P08 | PASS | 4/4 | — | — | — | P08-E01-live-replay-generation; P08-E01-live-replay-json-edit; P08-E01-live-replay-mask-edit; … 1 more |
| H2-openai-images-catalog | issue 11 / ADR-0010/0018/0022 | Fixed date snapshot and verified capabilities, sourced text/image token pricing, no cached/per-image rates, current catalog/price hash and actual usage replay | P08 | PASS | 4/4 | — | — | — | P08-E01-live-replay-generation; P08-E01-live-replay-json-edit; P08-E01-live-replay-mask-edit; … 1 more |
| V4-openai-images-editing | issue 10 / ADR-0022 | Fixed generation/edit endpoint, ordered inline JSON references and mask, sourced dimensions and model fidelity, atomic output and shared unary lifecycle | P08 | PASS | 132/132 | — | — | — | P08-E01-edit-mask; P08-E01-edit-multiple; P08-E01-edit-omit-fidelity; … 3 more |
| H5-image-input-budgets | issue 10 / design 8.1 / ADR-0022 | Reference and mask host/model/protocol count and byte budgets, known lengths before copy or encoding, callback allocation bounds | P08 | PASS | 33/33 | — | — | — | P08-E04-edit-host-count; P08-E04-edit-image-bytes; P08-E04-edit-mask-bytes; … 3 more |
| C04-image-edit-final | issue 10 / spec image callbacks / ADR-0022 | Final JSON edit schema, no foreign resource/operation/model, final MIME/count and mask validation, single response callback and lease/body release | P08 | PASS | 58/58 | — | — | — | P08-E04-edit-bad-response; P08-E04-edit-bounded-cycle; P08-E04-edit-bounded-deep; … 3 more |
| V4-openai-images-generation | issue 09 / ADR-0022 | Native synchronous generation, all-image atomic validation, protocol options, explicit extension registration and no premature live catalog claim | P08 | PASS | 267/267 | — | — | — | P08-D1-policy-image; P08-D1-policy-image-over-total; P08-D1-policy-input; … 3 more |
| H5-image-budgets | issue 09 / ADR-0002/0007/0008/0022 | Image subpolicy snapshots, bounded unary JSON, image count/byte limits, cancellation and zero resource gauges, callback size before encoding | P08 | PASS | 47/47 | — | — | — | P08-D1-policy-image; P08-D1-policy-image-over-total; P08-D1-policy-input; … 3 more |
| C04-image-final-payload | issue 09 / ADR-0005/0022 | Final model and allowed generation fields, frozen options/count/MIME, no external resources or continuation, callbacks once outside binding retries | P08 | PASS | 89/89 | — | — | — | P08-E04-all-options; P08-E04-edit-bad-response; P08-E04-edit-bounded-cycle; … 3 more |
| T09-image-usage | issue 09 / ADR-0009/0010/0022 | One reporting axis, nullable modality tokens, known-only modality prices without cache/per-image billing, metadata-only independent Observer snapshots | P08 | PASS | 18/18 | — | — | — | P08-E09-observer-isolation; P08-E11-cache-price-image; P08-E11-cache-price-null; … 3 more |
| V6-typesafe-live | issue 08 / ADR-0016/0021 | Official fixed classifier wire replay and explicit unconfirmed 422 shape; the separate live gate requires this route's own audited bundle | P07 | PASS | 8/8 | — | — | — | P07-E02-live-score-beyond-rounding; P07-E02-live-score-higher-precision-mismatch; P07-E02-live-score-observed-hundredths; … 3 more |
| H2-typesafe-catalog | issue 08 / ADR-0010/0018 | Only jev-1.13.0 with official capabilities, input rate 0.042 and free output; versioned hash snapshot | P07 | PASS | 4/4 | — | — | — | P07-E11-builtin-official; P07-E11-live-wire-context-422; P07-E11-live-wire-mixed-questions; … 1 more |
| C08-unary-lifecycle | issue 07 / design 7.5–7.6 / ADR-0003/0006/0008 | Unary cancellation, deadlines, transport timeouts, independent admission, frozen retries and subsequent permit acquisition | P07 | PASS | 41/41 | — | — | — | P07-E05-unary-backoff-cancel; P07-E05-unary-backoff-deadline; P07-E05-unary-frozen-snapshot; … 3 more |
| H5-unary-budgets | issue 07 / design 8 / ADR-0007/0008 | Inclusive unary output and error budgets, one overflow byte, no replay, body and lease release after parsing | P07 | PASS | 19/19 | — | — | — | P07-E08-unary-budget-bad-json; P07-E08-unary-budget-error-exact; P07-E08-unary-budget-error-over; … 3 more |
| C04-unary-callbacks | issue 07 / ADR-0006/0008 | One frozen request callback per logical call, invalid callback rejection and host cause Is/As preserved | P07 | PASS | 35/35 | — | — | — | P07-E04-unary-callback-custom-byte-element; P07-E04-unary-callback-encoding-pointer-array-element; P07-E04-unary-callback-encoding-pointer-json-element; … 3 more |
| T09-unary-observer | issue 07 / ADR-0009 | Trusted unresolved/resolved attribution, shared usage reporting axis, bounded asynchronous observer faults and redaction | P07 | PASS | 4/4 | — | — | — | P07-E09-unary-observer-congested; P07-E09-unary-observer-error; P07-E09-unary-observer-panic; … 1 more |
| V4-typesafe-mixed | issue 06 / spec TypeSafe / ADR-0021 | Three closed kinds, ordered scores, bool/noul, native JSON precision and strict final-question answers | P07 | PASS | 145/145 | 10/10 | — | — | P07-E01-bool-no-criteria; P07-E01-bool-no-criteria-parity; P07-E01-mixed; … 3 more |
| H2-typesafe-parity | issue 06 / design 9.2–9.3 / ADR-0021 | Frozen 1.0.0 classifier request/basic answer differential with explicit projection and scoped extension decisions | P07 PIDIFF | PASS | — | 12/12 | — | — | PIDIFF-P07-E01-bool-no-criteria-classify; PIDIFF-P07-E01-choice-parity-classify; PIDIFF-P07-E01-mixed-classify; … 3 more |
| T04-typesafe-final | issue 06 / ADR-0021 | Final callback questions frozen, original inputs independent, authority unchanged, invalid payloads refused | P07 | PASS | 13/13 | — | — | — | P07-E04-mixed-bool-capability; P07-E04-mixed-count-limit; P07-E04-mixed-final-kind-mismatch; … 3 more |
| V4-typesafe-choice | issue 05 / design sections 7–10 / ADR-0021 | TypeSafe choice unary public call, strict final-question answer validation, shared runtime, callbacks, usage, isolation and resource release | P07 | PASS | 433/433 | 12/12 | — | — | P07-D1-question; P07-D1-questions; P07-D1-state; … 3 more |
| T01 | research-traceability §2 T01 | Every entry point takes an explicit TenantID; refused calls send nothing | E07 E09 E10 | PASS | 12/12 | — | — | — | E07-reject-missing-request-id-complete-full; E07-reject-missing-request-id-complete-simple; E07-reject-missing-request-id-stream-full; … 3 more |
| T02 | research-traceability §2 T02 | Identity comes from the trusted host; forged tenants and history are refused at the host boundary | E10 | PASS | 42/42 | — | — | — | E10-example-host-merge; E10-example-host-refuses-claimed-tenant; E10-example-host-refuses-foreign-session; … 3 more |
| T03 | research-traceability §2 T03 | Bindings are located by (tenant, binding) and owned by the tenant | E06 E07 | PASS | 98/98 | — | — | — | E06-anthropic-both-succeed-complete-full; E06-anthropic-both-succeed-complete-simple; E06-anthropic-both-succeed-stream-full; … 3 more |
| T04 | research-traceability §2 T04 | Ordinary requests cannot override credentials or targets | E04 E07 | PASS | 70/70 | — | — | — | E07-request-surface-carries-no-authority; P01-E04-callbacks-deny-headers-alternate-key-header; P01-E04-callbacks-deny-headers-authorization-added-twice; … 3 more |
| T05 | research-traceability §2 T05 | No credential fallback (environment, other identity) | E06 E07 H1 | PASS | 78/78 | — | — | — | E06-anthropic-a-credential-missing-complete-full; E06-anthropic-a-credential-missing-complete-simple; E06-anthropic-a-credential-missing-stream-full; … 3 more |
| T06 | research-traceability §2 T06 | Each call keeps its configuration snapshot; shared implementation holds no tenant state | E05 E06 E07 | PASS | 190/190 | — | — | — | E06-anthropic-a-canceled-complete-full; E06-anthropic-a-canceled-complete-simple; E06-anthropic-a-canceled-on-arrival-complete-full; … 3 more |
| T07 | research-traceability §2 T07 | History, results and native state carry and check their attribution | E03 E10 | PASS | 80/80 | — | — | — | E10-example-host-merge; E10-example-host-refuses-foreign-session; E10-example-host-refuses-unknown-session; … 3 more |
| T08 | research-traceability §2 T08 | Cancellation is per call and releases only its own resources | E06 E08 E10 | PASS | 81/81 | — | — | — | E06-anthropic-a-canceled-complete-full; E06-anthropic-a-canceled-complete-simple; E06-anthropic-a-canceled-on-arrival-complete-full; … 3 more |
| T09 | research-traceability §2 T09 | Observations attribute tenant, call and attempt and hold no secret | E09 E10 | PASS | 102/102 | — | — | — | E09-mixed-observer-congested; E09-mixed-observer-error; E09-mixed-observer-panic; … 3 more |
| T10 | research-traceability §2 T10 | Updates and revocations take effect for new calls; in-flight calls keep their snapshot | E07 | PASS | 97/97 | — | — | — | E07-mixed-binding-disabled-classifier; E07-mixed-binding-disabled-google-image; E07-mixed-binding-disabled-openai-image; … 3 more |
| T11 | research-traceability §2 T11 | Every credential-dependent operation goes through the one authorized path | E07 | PASS | 421/421 | — | — | — | E07-mixed-binding-disabled-classifier; E07-mixed-binding-disabled-google-image; E07-mixed-binding-disabled-openai-image; … 3 more |
| T12 | research-traceability §2 T12 | Usage is recorded per attempt and per account; zero is never free | E05 E09 E11 | PASS | 86/86 | 21/21 | — | — | P01-E05-attempts-keep-the-snapshot; P01-E09-observe-retry; P01-E11-cache-write-priced-complete-full; … 3 more |
| C01 | research-traceability §4 C01 | Call lifecycle | E01 | PASS | 138/138 | 21/21 | — | — | E01-close-after-terminal-keeps-result; E01-close-while-resolver-blocked-stream-full; E01-close-while-resolver-blocked-stream-simple; … 3 more |
| C02 | research-traceability §4 C02 | Failures and truncation | E02 | PASS | 710/710 | 26/26 | — | — | E02-parity-anthropic-claude-sonnet-5-5-stream-complete; E02-parity-anthropic-claude-sonnet-5-5-stream-stream; E02-parity-anthropic-claude-sonnet-5-5-streamSimple-complete; … 3 more |
| C03 | research-traceability §4 C03 | History and tools | E03 | PASS | 871/871 | 21/21 | — | — | E03-catalog-builtin-models-are-pis; E03-catalog-inclusion-listed; E03-managed-effort-listed; … 3 more |
| C04 | research-traceability §4 C04 | Options and callbacks | E04 | PASS | 1528/1528 | 58/58 | — | — | E04-model-sampling-reject-openai-completions-bad-json; E04-model-sampling-reject-openai-completions-model; E04-model-sampling-reject-openai-completions-prompt_cache_key; … 3 more |
| C05 | research-traceability §4 C05 | Retry | E05 | PASS | 236/236 | 26/26 | — | — | P01-E05-attempts-keep-the-snapshot; P01-E05-auth-401-not-retried-complete-full; P01-E05-auth-401-not-retried-complete-simple; … 3 more |
| C06 | research-traceability §4 C06 | Tenant concurrency isolation | E06 | PASS | 175/175 | — | — | — | E06-anthropic-a-canceled-complete-full; E06-anthropic-a-canceled-complete-simple; E06-anthropic-a-canceled-on-arrival-complete-full; … 3 more |
| C07 | research-traceability §4 C07 | Authorization and updates | E07 | PASS | 421/421 | — | — | — | E07-mixed-binding-disabled-classifier; E07-mixed-binding-disabled-google-image; E07-mixed-binding-disabled-openai-image; … 3 more |
| C08 | research-traceability §4 C08 | Resources and release | E08 | PASS | 936/936 | 13/13 | — | — | E08-admission-burst-pressure-cloud-batch; E08-admission-burst-pressure-cloud-interactive; E08-admission-burst-pressure-cloud-interactive-waiters-x4; … 3 more |
| C09 | research-traceability §4 C09 | Observability and secrets | E09 | PASS | 57/57 | — | — | — | E09-mixed-observer-congested; E09-mixed-observer-error; E09-mixed-observer-panic; … 3 more |
| C10 | research-traceability §4 C10 | Routing and host contract | E10 | PASS | 119/119 | — | — | — | E10-example-host-merge; E10-example-host-refuses-claimed-tenant; E10-example-host-refuses-foreign-session; … 3 more |
| C11 | research-traceability §4 C11 | Usage and cost | E11 | PASS | 154/154 | 31/31 | — | — | E11-chat-duration-complete-full-failure; E11-chat-duration-complete-full-success; E11-chat-duration-complete-simple-failure; … 3 more |
| V4-responses | research-traceability §4 V§4 OpenAI Responses | P01 protocol specifics | P01 | PASS | 960/960 | 335/335 | — | — | P01-E01-config-read-only-after-construction; P01-E01-stream-returns-before-resolution; P01-E01-text-complete-full; … 3 more |
| V4-anthropic | research-traceability §4 V§4 Anthropic | P02 protocol specifics | P02 | PASS | 293/293 | 70/70 | — | — | P02-E01-block-start-content-stream; P02-E01-interleaved-stream; P02-E01-repair-and-trailing-events-stream; … 3 more |
| V4-gemini | research-traceability §4 V§4 Gemini | P03 protocol specifics | P03 | PASS | 293/293 | 77/77 | — | — | P03-E01-function-call-ids-generated-stream; P03-E01-interleaved-stream; P03-E01-signature-on-empty-text-stream; … 3 more |
| V4-chat | research-traceability §4 V§4 OpenAI Chat | P04 protocol specifics | P04 | PASS | 574/574 | 203/203 | — | — | P04-E01-after-done-ignored-stream; P04-E01-interleaved-stream; P04-E01-parallel-tool-calls-stream; … 3 more |
| V4-deepseek | research-traceability §4 V§4 DeepSeek | P05/P06: two bindings accepted separately | P05 P06 | PASS | 690/690 | 118/118 | — | — | P05-E01-reasoning-simple-stream; P05-E01-reasoning-stream; P05-E01-reasoning-tool-call-stream; … 3 more |
| V6-live | research-traceability §4 V§6 | Nine combinations supply their own current complete audited live smoke; required classification and image generation/editing cannot be UNSUPPORTED | LIVE | PASS | — | — | 10/10 | — | support-matrix:openai-responses=PASS; support-matrix:anthropic-messages=PASS; support-matrix:google-gemini=PASS; … 7 more |
| H1 | research-traceability §5 H1 | Per-call credential, endpoint and header isolation | H1 E06 E07 | PASS | 603/603 | — | — | — | E06-anthropic-a-canceled-complete-full; E06-anthropic-a-canceled-complete-simple; E06-anthropic-a-canceled-on-arrival-complete-full; … 3 more |
| H2 | research-traceability §5 H2 | Field presence, signatures, tools and pi differential | PIDIFF E03 E04 | PASS | 2399/2399 | 861/861 | — | — | E03-catalog-builtin-models-are-pis; E03-catalog-inclusion-listed; E03-managed-effort-listed; … 3 more |
| H3 | research-traceability §5 H3 | Correct terminals and partial messages | E01 E02 | PASS | 848/848 | 47/47 | — | — | E01-close-after-terminal-keeps-result; E01-close-while-resolver-blocked-stream-full; E01-close-while-resolver-blocked-stream-simple; … 3 more |
| H4 | research-traceability §5 H4 | pi-equivalent retry | E05 | PASS | 236/236 | 26/26 | — | — | P01-E05-attempts-keep-the-snapshot; P01-E05-auth-401-not-retried-complete-full; P01-E05-auth-401-not-retried-complete-simple; … 3 more |
| H5 | research-traceability §5 H5 | Resource limits before reading or allocating | E08 | PASS | 936/936 | 13/13 | — | — | E08-admission-burst-pressure-cloud-batch; E08-admission-burst-pressure-cloud-interactive; E08-admission-burst-pressure-cloud-interactive-waiters-x4; … 3 more |
| D1 | spec D1 / ADR-0002 | Explicit finite resource policy, refused when missing or invalid | D1 E08 | PASS | 971/971 | 13/13 | — | — | D1-construct-accept-boundaries; D1-construct-reject-admission-wait-exceeds-call; D1-construct-reject-connect-exceeds-call; … 3 more |
| D1-examples | spec I9 正式示例 / issue 24 | The formal policy examples pass their repeatable pressure scenarios | PRESSURE | PASS | 15/15 | — | — | — | E08-admission-burst-pressure-cloud-batch; E08-admission-burst-pressure-cloud-interactive; E08-admission-burst-pressure-cloud-interactive-waiters-x4; … 3 more |
| D2 | spec D2 / ADR-0003 | An inconsistent snapshot fails before sending; no internal re-resolution | D2 | PASS | 24/24 | — | — | — | E07-reject-binding-disabled-between-reads-complete-full; E07-reject-binding-disabled-between-reads-complete-simple; E07-reject-binding-disabled-between-reads-stream-full; … 3 more |
| S6-1-3 | research-traceability §1 S§6 验收 1–3 | Protocol capabilities, A/B isolation, zero requests before failure | P01 P02 P03 P04 P05 P06 E06 E07 | PASS | 3304/3304 | 803/803 | — | — | E06-anthropic-a-canceled-complete-full; E06-anthropic-a-canceled-complete-simple; E06-anthropic-a-canceled-on-arrival-complete-full; … 3 more |
| S6-4-7 | research-traceability §1 S§6 验收 4–7 | Updates/revocation, EOF/errors, limits/release, redaction | E02 E05 E06 E07 E08 E09 | PASS | 2535/2535 | 65/65 | — | — | E02-parity-anthropic-claude-sonnet-5-5-stream-complete; E02-parity-anthropic-claude-sonnet-5-5-stream-stream; E02-parity-anthropic-claude-sonnet-5-5-streamSimple-complete; … 3 more |
| S6-8-10 | research-traceability §1 S§6 验收 8–10 | Full differential, simple/full/reasoning/usage, routing | PIDIFF E03 E04 E05 E10 E11 | PASS | 2908/2908 | 861/861 | — | — | E03-catalog-builtin-models-are-pis; E03-catalog-inclusion-listed; E03-managed-effort-listed; … 3 more |
| TS-R01 | .scratch/barness-ai-thinking-sampling/spec.md#r01 | Model profile metadata, seven canonical keys, optional maps | E03 E04 | PASS | 646/646 | — | — | — | E03-thinking-sampling-config-anthropic-empty-key; E03-thinking-sampling-config-anthropic-empty-map; E03-thinking-sampling-config-anthropic-empty-profile; … 3 more |
| TS-R02 | .scratch/barness-ai-thinking-sampling/spec.md#r02 | Shared canonical clamp independent of wire; supported, null, extended and nonreasoning matrices | E04 | PASS | 544/544 | — | — | — | P01-E04-thinking-sampling-S02-supported-high-stream-complete; P01-E04-thinking-sampling-S02-supported-high-stream-complete-hooks; P01-E04-thinking-sampling-S02-supported-high-stream-stream; … 3 more |
| TS-R03 | .scratch/barness-ai-thinking-sampling/spec.md#r03 | Full/simple effort boundaries, summary-only, DeepSeek disabled/enabled | E04 | PASS | 152/152 | — | — | — | P01-E04-thinking-sampling-S06-all-null-stream-complete; P01-E04-thinking-sampling-S06-all-null-stream-stream; P01-E04-thinking-sampling-S06-all-null-streamSimple-complete; … 3 more |
| TS-R04 | .scratch/barness-ai-thinking-sampling/spec.md#r04 | Named/general/profile/call/hook top-level merge preserves all valid JSON values | E04 | PASS | 296/296 | — | — | — | P01-E04-thinking-sampling-S01-merge-stream-complete; P01-E04-thinking-sampling-S01-merge-stream-complete-hooks; P01-E04-thinking-sampling-S01-merge-stream-stream; … 3 more |
| TS-R05 | .scratch/barness-ai-thinking-sampling/spec.md#r05 | Four consumers; native maps ignored; extension route explicitly skipped | P01 P02 P03 P04 P05 P06 | PASS | 936/936 | 192/192 | — | — | P01-E03-thinking-sampling-snapshot-call-stream-complete; P01-E03-thinking-sampling-snapshot-call-stream-stream; P01-E03-thinking-sampling-snapshot-call-streamSimple-complete; … 3 more |
| TS-R06 | .scratch/barness-ai-thinking-sampling/spec.md#r06 | NewClient/Hash validate all levels, keys, JSON and reserved fields safely | E03 | PASS | 521/521 | — | — | — | E03-thinking-sampling-config-anthropic-empty-key; E03-thinking-sampling-config-anthropic-empty-map; E03-thinking-sampling-config-anthropic-empty-profile; … 3 more |
| TS-R07 | .scratch/barness-ai-thinking-sampling/spec.md#r07 | Callback observes final sampling; no reselection; final authorization/limits maintained | E04 E05 | PASS | 120/120 | — | — | — | P01-E04-thinking-sampling-callback-stream-complete-hosted; P01-E04-thinking-sampling-callback-stream-complete-in-place; P01-E04-thinking-sampling-callback-stream-complete-limit; … 3 more |
| TS-R08 | .scratch/barness-ai-thinking-sampling/spec.md#r08 | Deep snapshots, tenant/client ownership, cancellation and retry frozen bytes | E03 E05 E06 | PASS | 89/89 | — | — | — | E03-thinking-sampling-snapshots; P01-E03-thinking-sampling-snapshot-call-stream-complete; P01-E03-thinking-sampling-snapshot-call-stream-stream; … 3 more |
| TS-R09 | .scratch/barness-ai-thinking-sampling/spec.md#r09 | Verified catalog contents, stable nested map hashing and upgraded price snapshot | E03 | PASS | 2/2 | — | — | — | E03-catalog-builtin-models-are-pis; E03-thinking-sampling-snapshots |
| TS-R10 | .scratch/barness-ai-thinking-sampling/spec.md#r10 | Single 1.1.0 oracle with two source equivalence proofs and zero pending full differential | E03 PIDIFF | PASS | 3/3 | 861/861 | — | — | E03-thinking-sampling-baseline-evidence-b; E03-thinking-sampling-baseline-evidence-c; E03-thinking-sampling-baseline-oracle-guards; … 3 more |
| TS-R11 | .scratch/barness-ai-thinking-sampling/spec.md#r11 | Current nine-route release contract, independent Chinese effect and design-load artifacts | LIVE ARTIFACTS | PASS | — | — | 10/10 | 10/10 | support-matrix:openai-responses=PASS; support-matrix:anthropic-messages=PASS; support-matrix:google-gemini=PASS; … 9 more |
| TS-R12 | .scratch/barness-ai-thinking-sampling/spec.md#r12 | Complete documented Host configuration and public protocol fixtures | E03 E04 | PASS | 48/48 | — | — | — | P01-E03-thinking-sampling-snapshot-call-stream-complete; P01-E03-thinking-sampling-snapshot-call-stream-stream; P01-E03-thinking-sampling-snapshot-call-streamSimple-complete; … 3 more |
| OD-R01 | OpenAI Decisions spec R01 | 独立型号身份和协议 | P10 | PASS | 56/56 | — | — | — | P10-OD03-preflight-bool-hooked; P10-OD03-preflight-bool-plain; P10-OD03-preflight-catalog-hooked; … 3 more |
| OD-R02 | OpenAI Decisions spec R02 | System One/Jev 不回归 | P07 | PASS | 142/142 | 3/3 | — | — | P07-E01-choice-parity-parity; P07-E01-mixed; P07-E01-mixed-parity; … 3 more |
| OD-R03 | OpenAI Decisions spec R03 | 公共请求和选项 | P10 | PASS | 44/44 | 2/2 | — | — | P10-OD04-jev-options-*ai.OpenAIDecisionsOptions; P10-OD04-jev-options-*ai.TypeSafeOptions; P10-OD04-jev-options-ai.OpenAIDecisionsOptions; … 3 more |
| OD-R04 | OpenAI Decisions spec R04 | 格式与容量 | P10 | PASS | 37/37 | 2/2 | — | — | P10-OD08-images-binding-; P10-OD08-images-binding-chat; P10-OD08-images-binding-image; … 3 more |
| OD-R05 | OpenAI Decisions spec R05 | 模型能力与安全默认 | P10 | PASS | 75/75 | — | — | — | P10-OD03-preflight-bool-hooked; P10-OD03-preflight-bool-plain; P10-OD03-preflight-catalog-hooked; … 3 more |
| OD-R06 | OpenAI Decisions spec R06 | 图片资源策略 | ARTIFACTS | PASS | — | — | — | 2/2 | artifact:mixed-pressure-local=PASS; artifact:mixed-pressure-cloud=PASS |
| OD-R07 | OpenAI Decisions spec R07 | Binding 图片授权 | P10 | PASS | 45/45 | — | — | — | P10-OD19-images-byte-encoding-text-input-boundary; P10-OD19-images-byte-encoding-text-input-limit; P10-OD19-images-byte-encoding-typed-text-boundary; … 3 more |
| OD-R08 | OpenAI Decisions spec R08 | 请求转换 | P10 | PASS | 15/15 | 2/2 | — | — | P10-OD27-mixed-all-refusal-hooked; P10-OD27-mixed-all-refusal-plain; P10-OD27-mixed-metadata-hooked; … 3 more |
| OD-R09 | OpenAI Decisions spec R09 | native 回调和最终请求 | P10 | PASS | 111/111 | — | — | — | P10-OD25-duplicate-name-hooked; P10-OD25-duplicate-value-hooked; P10-OD25-images-message-type-alias; … 3 more |
| OD-R10 | OpenAI Decisions spec R10 | 拒答是类型化结果 | P10 | PASS | 271/271 | — | — | — | P10-OD31-answer-answer-unknown-field-hooked; P10-OD31-answer-answers-null-hooked; P10-OD31-answer-choice-boolean-hooked; … 3 more |
| OD-R11 | OpenAI Decisions spec R11 | 严格结果发布 | P10 | PASS | 357/357 | — | — | — | P10-OD31-answer-answer-unknown-field-hooked; P10-OD31-answer-answers-null-hooked; P10-OD31-answer-choice-boolean-hooked; … 3 more |
| OD-R12 | OpenAI Decisions spec R12 | 用量和价格 | P10 | PASS | 298/298 | — | — | — | P10-OD35-38-usage-basic-hooked; P10-OD35-38-usage-basic-plain; P10-OD35-38-usage-broken-answer-hooked; … 3 more |
| OD-R13 | OpenAI Decisions spec R13 | 调用生命周期 | P10 | PASS | 59/59 | — | — | — | P10-OD40-runtime-429-ms; P10-OD40-runtime-503-seconds; P10-OD40-runtime-504-extension; … 3 more |
| OD-R14 | OpenAI Decisions spec R14 | 隔离和不可变状态 | P10 | PASS | 3/3 | — | — | — | P10-OD44-runtime-independent-accounts; P10-OD44-runtime-shared-account-and-key; P10-OD45-runtime-pinned-snapshot |
| OD-R15 | OpenAI Decisions spec R15 | 固定 Pi 差分 | PIDIFF | PASS | — | 46/46 | — | — | PIDIFF-P10-OD14-numeric-order-hooked-classify; PIDIFF-P10-OD14-numeric-order-plain-classify; PIDIFF-P10-OD48-catalog; … 3 more |
| OD-R16 | OpenAI Decisions spec R16 | 目录纳入 | P10 | PASS | 38/38 | — | — | — | P10-OD52-formal-catalog; P10-OD53-formal-batch-candidate; P10-OD53-formal-batch-narrowed; … 3 more |
| OD-R17 | OpenAI Decisions spec R17 | 独立支持与发布 | LIVE | PASS | — | — | 10/10 | — | support-matrix:openai-responses=PASS; support-matrix:anthropic-messages=PASS; support-matrix:google-gemini=PASS; … 7 more |
| OD-R18 | OpenAI Decisions spec R18 | 宿主示例和效果 | ARTIFACTS | PASS | — | — | — | 2/2 | artifact:chinese-effect-jev=PASS; artifact:chinese-effect-luna=PASS |
| OD-R19 | OpenAI Decisions spec R19 | 证据和观测 | ARTIFACTS | PASS | — | — | — | 5/5 | artifact:mixed-pressure-local=PASS; artifact:mixed-pressure-cloud=PASS; artifact:authorization-revocation=PASS; … 2 more |
| OD-R20 | OpenAI Decisions spec R20 | 文档和回滚 | ARTIFACTS | PASS | — | — | — | 1/1 | artifact:authorization-revocation=PASS |

## Capacity: normal and race design loads

- mixed-pressure-local: PASS; `.evidence/issue13-final-gate/offline/20261013T034156.007664000Z`; 4 concurrent calls, heap growth 7586808 / budget 268435456, allocation upper bound 142150808; all limit use <=75%; classifier-image design/cancel/recovery verified; resources zero
- mixed-pressure-cloud: PASS; `.evidence/issue13-final-gate/offline/20261013T034156.007664000Z`; 8 concurrent calls, heap growth 19649336 / budget 536870912, allocation upper bound 284722232; all limit use <=75%; classifier-image design/cancel/recovery verified; resources zero
- race-mixed-pressure-local: PASS; `.evidence/issue13-final-gate/race/20261013T034800.523245000Z`; 4 concurrent calls, heap growth 9000696 / budget 268435456, allocation upper bound 151864608; all limit use <=75%; classifier-image design/cancel/recovery verified; resources zero
- race-mixed-pressure-cloud: PASS; `.evidence/issue13-final-gate/race/20261013T034800.523245000Z`; 8 concurrent calls, heap growth 17288000 / budget 536870912, allocation upper bound 302694896; all limit use <=75%; classifier-image design/cancel/recovery verified; resources zero

## Business effect: independent Jev and Luna

COMPLETE verifies the fixed task; no quality threshold. Null means unavailable, not zero.

- chinese-effect-jev: PASS; `.evidence/issue13-current-live/evaluation/jev/bundles/20261013T032345.376597000Z`; independent Chinese task; 24/24 answered; accuracy 1; conditional accuracy 1; ECE 0.0004166666666666208; Brier 0.00000833333333333334; cost USD 0.000539406; no quality threshold
- chinese-effect-luna: PASS; `.scratch/barness-ai-openai-decisions/chinese-live-evidence/live`; independent Chinese task; 24/24 answered; accuracy 1; conditional accuracy 1; ECE 0.014583333333333393; Brier 0.0006916666666666667; cost USD 0.000623800; no quality threshold

## Authorization revocation: normal and race

- authorization-revocation: PASS; `.evidence/issue13-final-gate/offline/20261013T034156.007664000Z`; six initial/final-Hook revocations: new calls denied, pinned attempts and reported cost preserved, no fallback, resources zero
- race-authorization-revocation: PASS; `.evidence/issue13-final-gate/race/20261013T034800.523245000Z`; six initial/final-Hook revocations: new calls denied, pinned attempts and reported cost preserved, no fallback, resources zero

## Race artifact verification

- race-offline: PASS; `.evidence/issue13-final-gate/race/20261013T034800.523245000Z`; 
- race-differential: PASS; `.evidence/issue13-final-gate/race/20261013T034800.523245000Z`; 

## Redaction audit

- `.evidence/issue13-final-gate/offline/20261013T034156.007664000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/openai-responses/bundles/20261013T032015.505978000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/anthropic-messages/bundles/20261013T032036.840263000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/google-gemini/bundles/20261013T032106.833565000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/openai-chat/bundles/20261013T032154.096377000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/deepseek-responses/bundles/20261013T032208.754632000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/deepseek-chat/bundles/20261013T032216.397782000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/typesafe-classifier/bundles/20261013T032226.401666000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/openai-images/bundles/20261013T032228.227090000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/google-interactions-image/bundles/20261013T032302.962640000Z`: 0 finding(s)
- `.evidence/issue13-current-live/smoke/openai-decisions/bundles/20261013T032340.988798000Z`: 0 finding(s)
- `.evidence/issue13-current-live/evaluation/jev/bundles/20261013T032345.376597000Z`: 0 finding(s)
- `.scratch/barness-ai-openai-decisions/chinese-live-evidence/live`: 0 finding(s)
- `.evidence/issue13-final-gate/race/20261013T034800.523245000Z`: 0 finding(s)

## Snapshots

- catalog: current
- release-plan: current

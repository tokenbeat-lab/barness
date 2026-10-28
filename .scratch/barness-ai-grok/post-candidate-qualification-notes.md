# Grok complete candidate: offline qualification path

Read-only code investigation, 2026-10-26. No tests, builds, vendor calls, credentials or body/ciphertext output. Only this file outside the repository is written. Canonical repository: `/Users/cyber/RestoX/barness`.

## Recommendation

Do the remaining software work before asking for another paid run: (1) sanitize and replay the successful own captures through the public Client, (2) implement the fixed xAI native formal qualification/report path and formal live harness mode, (3) prepare the final Builtin catalog wiring/version as an unreleased integration-branch implementation artifact, (4) review and freeze the entire final payload. Then request a fresh explicitly **formal** grant and, after valid required live evidence, obtain the separate human race and pressure choices. Do not change the catalog/code after these final gates.

The successful v3 candidate remains immutable historical evidence. It cannot be renamed, repinned, copied into a formal report, or given `formalQualified=true`. New parser/catalog source needs a new formal source/binary/catalog-bound run. Do not buy another candidate run piecemeal while the formal software is still changing.

The integration-branch final catalog wiring recommendation is an interpretation of the promotion boundary, discussed below. It is **not** a claim that the literal post-SOP promotion wording is already satisfied. No change has been made by this investigation.

## What actually exists now

- Own complete capture: `.evidence/barness-ai-grok/live/xai-responses/20261026T050450.562258Z-tlbttr1o/bundles/20261026T050451.200594000Z`.
- Report is `xai-candidate-report.json`, kind `xai-candidate-only`, mode `live`, stage `candidate`, 18 scenarios, `formalQualified=false`.
- Root files are the candidate report, unfinalized prior report, grant, claim, manifest and audit. There is **no** `live-report.json` (`supportmatrix.ReportFile`).
- Per-call artifacts are actual `*-http.json`, `*-result.json`, `*-observer.json`, and for streaming `*-events.json`. Examples: `text-*`, `first-*`/`continued-*`, `calls-*`/`results-*`, `first-*`/`repeat-*`. Encrypted continuation additionally retains `persisted-turn.json`.
- Actual outer formal consumer refusal is preserved under `.evidence/barness-ai-grok/live-consumers/attempt3-wrapper-v1`; this is expected with the current contract, not a failure of the 18-scenario candidate result.
- Parent's safe usage observation index is `.evidence/barness-ai-grok/candidate-usage-observations-v3.json`: all 24 observed usage shapes are inclusive `T=P+O`, all have nonzero reasoning, 23 have positive cached input. These are observations of this run, not a justification to delete the alternative valid synthetic arithmetic path.

## Minimum offline replay work

Use existing public seams, not a new adapter or private-helper unit tests:

| Existing seam | Location | Use |
| --- | --- | --- |
| Official endpoint retained, loopback transport injected | `ai/e2e/xai_fixture_test.go:27`, `newXAIWorld` | Replay against public Client with xAI identity and current candidate/custom catalog; no endpoint exception |
| Arbitrary captured SSE chunks | `ai/internal/testkit/provider/provider.go:172`, `provider.SSE` | Feed sanitized observed SSE without rebuilding event sequences from expected Client output |
| World queue and public calls | `ai/e2e/xai_fixture_test.go`, existing xAI acceptance/history/tools tests | Cover Stream, Complete and Simple entry points, native persistence/trust, tools and images |
| Own observed native replay analog | `ai/e2e/minimax_responses_observed_test.go:12` | First response -> public ValidateToolCall -> persisted/restored native message -> next request assertions |
| Source-pinned own request/response fixtures analog | `ai/e2e/regional_zai_chat_source_test.go:15` | Index origin/files SHA256, read request + `.sse`, replay exact own shape; do not reuse another Provider's pass |

Suggested ownership/files: `ai/e2e/testdata/xai-responses-observed/` with an origin/normalization/hash index and the selected request/response pairs, plus one or more `ai/e2e/xai_observed_*_test.go` files under 500 lines. Existing `xaiTransport`, `newXAIWorld`, provider replies and public Client are enough; no new dependency is needed.

Retain actual captured scenarios rather than hand-inventing a second protocol. A finite table covering the 18 logical scenarios / 24 response pairs is straightforward; at minimum the fixture set must preserve every distinct observed shape required by E17: four entry points; omitted vs explicit effort; native encrypted continuation; auto/required/none/parallel function results; PNG/JPEG user and tool-result image order; cache/usage. Do not narrow required capabilities to make replay or qualification easier.

Sanitization requirements:

- Copy only the replayable protocol subset to tracked fixtures. Keep the original ignored bundle unchanged.
- Remove auth/credential/account/tenant identifiers. Replace opaque ciphertext with deterministic synthetic tokens consistently across response snapshots, persisted state and subsequent requests. Preserve JSON type, field presence, array/order, whole-item equality, opaque ID pair correspondence and event structure. Label replacements in the fixture index.
- Preserve actual raw usage numbers and their presence/null/zero facts. Expected normalized usage and estimated costs should be independently calculated, not loaded from the same Result as the response under test.
- Record original source bundle/manifest/capture hash, normalization rules and derived file hashes; do not present a normalized fixture as a new live acceptance. No need to print the original body/ciphertext while extracting it.
- An observed fixture may pass existing adapter code on first use: do not manufacture a runtime defect. A real public replay-entry/fixture absence can be the initial red for wiring, and corruption/missing-pair/usage/identity mutations must be separately labeled synthetic negatives. Follow SOP's red-before-wiring requirement and retain actual commands/exit/artifacts.

## Formal qualification software still missing

`releasegate -live-bundle` is already the public native consumer entry (`ai/release/cmd/releasegate/main.go`); keep it. The minimum change is an xAI-specific native contract behind that existing entry, not another global acceptance bypass.

| Existing code | Current behavior / required xAI work |
| --- | --- |
| `ai/internal/testkit/release/record_live.go:16` | Loads `live-report.json`, dispatches isolated native consumers for regional/MiniMax/Jev/GLM. Add own xAI formal report/consumer dispatch. |
| `ai/internal/testkit/supportmatrix/identity.go:5` | xai-responses is absent. Add exact chat/xai/openai-responses identity; do not register xAI as a regional route. |
| `ai/internal/testkit/release/live.go:14` | xAI expected capabilities absent. Add fixed complete 18-scenario registry, independent of report claims; do not permit required xAI capabilities as UNSUPPORTED. |
| `ai/internal/testkit/release/chat_qualification.go:14` | Generic budget is 32 calls/one scenario retry/4096 max output and incorrectly caps all normalized output by visible output budget for xAI. Own xAI qualification must validate the xAI 24-attempt, zero-retry, visible/reserved budgets and observed estimate contract. |
| `ai/internal/testkit/release/live_manifest.go:13` | Generic chat expects `attempt-*-exchanges/result` and `observations-*`, and a different live package string. xAI existing capture names/package are different. Add own native manifest/capture verification branch; do not rename old artifacts or relax old routes. |
| `ai/internal/testkit/release/chat_capture.go:19` | Valuable attribution/Result/terminal Observer/target/inventory pattern, but filename assumptions differ. Implement xAI's own finite triplet mapping for text, continuation, tool and cache labels. |
| `ai/live/xai_config_test.go:142` | Only builds CandidateCatalog, fixed candidate kind/stage and candidate grant. Add explicit formal mode selected by a fresh formal grant, fixed full Builtin catalog identity, exact endpoint/model and same bounded 18/24 contract. Wrong stage/catalog/source/binary must refuse before HTTP. |
| `ai/live/xai_finish_test.go`, `xai_finalize_test.go` | Preserve fail-closed finalization/audit, but formal reports need own schema/provenance and `live-report.json` projection. Controlled/candidate remain distinct and cannot qualify formal. |
| `ai/catalog.go:52` and runtime `ai/catalog_xai.go` | Builtin currently omits xAI; final formal consumer requires exact current full Builtin version/hash. Runtime package cannot import `internal/testkit/xai` because that package imports `ai`; use reviewed runtime model facts/literal or existing data loader pattern, not that import cycle. |

Native verification should independently validate source/catalog/SDK/model/endpoint/account-region aliases, exact complete scenario set, consumed grant/claim and binary, 24 distinct actual call IDs and HTTP attempts, timestamps, artifact hashes/audit/assertions, full non-truncated captures, Result/terminal Observer correspondence, stateless request, native continuation equality, tools/results/parallel pairing, PNG/JPEG/tool-image ordering, raw-to-normalized usage arithmetic, and fixed catalog cost estimates. All these are concrete current contracts/captures, not speculative extra abstractions.

Process negatives come first: current candidate/controlled evidence cannot pass formal; missing/duplicate/substituted scenarios; wrong stage/identity/SDK/catalog/source/binary/grant; changed/truncated/missing triplets; altered usage or Observer; late finalization/audit failure. Use the existing public CLI/process E2E style (`xai_live_process_test.go`, `xai_option_evidence_process_test.go`) and retained local captures. Do not start vendor/race/pressure during these checks.

A separate positive historical candidate-readiness inspection may be useful to record provenance for promotion preparation, but it must have an explicitly non-formal result and verify the original frozen origin. It is not necessary to add another public CLI flag just to make current `-live-bundle` succeed. In particular, do not change its current-commit/full-Builtin expectations so v3 can masquerade as formal.

## Two source boundaries and what invalidates what

| Change | xAI native hash | Outer execution source / formal record |
| --- | --- | --- |
| Only prose/indices outside selected inputs | Unchanged | Runtime files unchanged; current-state hash/index still needs new immutable record and actual read-only gate |
| Add `ai/e2e/*.go` or `ai/e2e/testdata/*` | **Unchanged** by current `xaiSource` filter | **Changed**: outer hashes every non-MD/cache/executable file below `ai/` |
| Add/change own qualifier under `ai/internal/testkit/**` | **Changed** (all non-test internal Go included) | Changed |
| Change `ai/live/*.go` including `_test.go` | **Changed** | Changed |
| Change runtime `ai/*.go`, modules or candidate.json | **Changed** | Changed |
| Add a runtime catalog JSON | Included in outer source; native inclusion must be explicit/catalog-bound | Changed; cannot assume a catalog-only edit is exempt |

Exact code:

- `ai/live/xai_config_test.go:124`: xaiSource = RegionalExecutionSource + independently embedded xAI candidate JSON.
- `ai/internal/testkit/supportmatrix/regional_source.go:15`: native set includes runtime `ai/*.go`, every `ai/live/*.go`, internal non-test Go, regional/kimi facts and specified capture/evidence/audit/support dependencies, modules; it excludes E2E test code/testdata. This does **not** confer regional identity on xAI.
- `tools/verification/receipt_files.py:73` and `ai/internal/testkit/release/record_files.go:125`: outer inventories `ai/`, tools, modules; markdown excluded unless selected, fixture `.json`/`.sse` included.
- `ai/internal/testkit/release/record_verify.go:116`: every passing stage must have exact selected source file contents **and the same Source.Commit** as the record, plus exactly equal selected input sets/content.
- `ai/internal/testkit/release/live_consumer.go:124-129`: native formal consumption requires report.GitCommit == current checked-out commit.
- `ai/internal/testkit/release/record_live.go:83`: selected live report commit must match the record and exactly one valid outer schema-2 native consumer is required.

Therefore: E2E replay additions preserve the historical v3 native run hash but prevent treating its outer receipt as a passing stage of a new outer source freeze. New qualifier/harness/catalog code invalidates current native applicability too. None erases the actual historical PASS. Source changes need review + a new complete ordinary freeze; next live uses a new explicitly authorized grant and binary. Never rewrite old pins/timestamps/status.

Commit hygiene matters: choose the final source/input freeze after new authorization/grant/config inputs are fixed, then do not commit between stage runs if the code expects one Source.Commit across all stages. Pure closing commits are permitted later only via current-source/content equality, not by editing old stage receipt commits. Stage-specific human choices and argv still need actual preserved selection; do not manufacture future choices or stamp old commits into new commands.

## The Builtin ordering conflict

There is a real distinction, not an implementation trick:

1. R21/Solution prohibit a candidate PASS from automatically turning on Builtin or claiming formal support. Ticket06 says promotion needs own evidence and current SOP.
2. `qualifiedLive` requires the report catalog version/hash to equal the **current full BuiltinCatalog**. `ConsumeNativeLive` also requires current source commit. Thus keeping Builtin absent through the alleged final formal/race/pressure runs, then editing catalog code/data afterward, cannot give a valid final release record. It creates new source/catalog and invalidates prior final-stage applicability.
3. Existing regional implementation distinguishes independently consumed candidate proof/activation intent from a later formal report: `VerifyRegionalQualification` vs `NewRegionalReport` (`regional.go:41-45`), `RegionalPromotionProposal` preserving candidate lineage (`regional_promotion.go:20-48`), and stage-specific catalog/grant (`regional_grant.go:57-80`). Its controlled proposal explicitly cannot activate production (`ai/e2e/regional_promotion_controlled_test.go`). This supports separate candidate/promoted/formal facts; it does not turn a candidate into formal evidence.

Smallest practical interpretation for this task: complete the final xAI Builtin wiring **as an unreleased integration-branch proposal** after its own candidate proof and replay/ordinary checks are in hand. Deliberately review that final proposed payload, freeze it, and test the actual proposed catalog via a fresh formal live grant followed by race/pressure. Keep support/publication/release false until all required own routes and stages pass. Do not merge/publish it as formally supported early, and do not advertise the branch wiring as already completed promotion.

This needs a short explicit design/ticket clarification because the literal reading “no Builtin bytes may change until after every SOP gate” is cyclic with the current consumer. The clarification can preserve the intended release boundary without a new controller/tag/compatibility layer. It is a scope/order interpretation; no code currently grants that exception automatically. If the maintainer instead insists on the literal byte-order prohibition, the honest alternative is a deliberately separate candidate-promotion consumer and reviewed activation step, then a fresh full final-source ordinary/formal/race/pressure cycle after activation. That is more work/runs, not something fixed by relabeling v3.

The other 22 affected routes still require their own planned scopes, account/region eligibility, credentials and explicit paid authorizations. The candidate grant cannot cover them. Formal xAI preparation can proceed offline independently, but xAI-only success will not make the feature's all-route release pass.

## What needs human choice vs what is already specified

Specified/inferable: one global xAI route/model, SDK, 18 capabilities, 24 calls/attempts, zero retries, raw usage arithmetic/cost rules, candidate/formal/controlled separation, complete capture replay and public seams, fixed source/catalog/binary, no unsupported-capability escape, no new provider or endpoint fallback, no new general controller.

Human selections still required before actual paid/manual stages: fresh formal (and if chosen, fresh current-source candidate) grant scope/budget/aliases/credentials, each other affected route's own grant, then actual race scope/argv and later pressure loads. Do all permitted offline code/review/freeze preparation first so these approvals concern concrete artifacts. Do not automatically spend another grant or start the next manual stage.


## Follow-up: doc/index commits after successful native capture

**Yes: later pure doc/index closing commits can pass the outer record Gate.** The current-HEAD equality belongs to *capturing a new native consumer*, not to replaying its already captured successful schema-2 receipt.

Exact call path:

1. `EvaluateRecord`, `release/record_verify.go:41-44`, reads current source and uses `sameRecordSource` to compare it with `record.Source`. `record_files.go:202` defines this as equality of `Files` only, not `Commit`. The current commit is retained separately as `ReleaseCommit` (`record_verify.go:42`). Therefore closing commit H2 is accepted against tested/paid source commit H if all included file bytes and paths still agree.
2. Every stage receipt must still say H and equal the record's immutable selected input set: `record_verify.go:116` checks both file-source equality and `record.Source.Commit == receipt.Source.Commit`, and selected input map equality.
3. The live route passes from `verifyStageRecord` (`record_verify.go:231-232`) to `verifySelectedLive` (`record_live.go:78`). It re-runs **qualifiedLive** at line79. That function (`record_live.go:16-75`) rechecks native contract/captures/audit/full Builtin version+hash/SDK/scenarios, but has no current-HEAD equality itself. Existing regional own verification also compares original manifest commit to report commit and current native file hash, not report commit to current HEAD.
4. `verifySelectedLive` compares `report.GitCommit` to the **passed-in record source H**, not current H2 (`record_live.go:83`).
5. `verifyLiveConsumer` verifies the retained outer schema-2 and native schema-1 receipt/output against H: `live_consumer.go:174-195`. Both source commits must equal the record's H, included source files must equal, bundle inventory/report hash/argv/exit/log hashes must match. It does **not** call `ConsumeNativeLive` or start a subprocess.
6. In contrast, a **new** `ConsumeNativeLive` at `live_consumer.go:124-127` first sets `proof.Source` from the current checkout, then requires `report.GitCommit == proof.Source.Commit`. Starting this after H2 would fail for an H report even when only docs changed. The outer Python wrapper also captures and requires the exact same source object throughout its one invocation (`tools/verification/consume_live.py:29,49`).

Valid final sequence without commit-only invalidation:

- Finish/review all runtime, harness, qualifier, fixtures and proposed catalog code. Build the exact formal binary against the native hash.
- Create/fix the actual authorization/grant/config/selected inputs, then commit H. A binary built just before this documentation-only authorization commit remains usable if actual source hash, binary SHA and catalog identity match its grant; native grant source is a content hash, not the build commit label.
- Run complete ordinary, formal live, and immediately run `consume_live.py` at H. The emitted formal report/manifest/outer stage source must say actual **paid H**, not the earlier build commit. Retain successful schema-2 and nested native receipts intact.
- Keep HEAD H for subsequent stage captures that must coexist in one release record. Human race/pressure choices still require their real explicit selection. Do not change the selected input bytes/set or included source files. The final record demands one shared Source.Commit and selected input map across passing stages.
- After applicable stages are done, save the previous current state, write the new closing current state/index, and commit doc/index-only H2. Keep `record.Source` equal to H's actual captured source, bind the *new* current-state hash, retain unchanged selected inputs, and reference the already captured H native consumer.
- Save source-file/input-set equality H vs H2, then run compiled `releasegate -record ...` at H2. It reports `Commit=H`, `ReleaseCommit=H2` and can pass if every required actual stage qualifies. Do not rerun `consume_live.py` merely because H2 exists, and never edit H receipt commits to H2.

For the forthcoming xAI formal parser, preserve this architecture: validate report/manifest/grant/binary/source-content provenance, and leave `report.GitCommit == current HEAD` at the existing new-consumer capture boundary. Adding that equality inside xAI's `qualifiedLive` branch would accidentally break the deliberately supported doc-only closure path. This is not a reason to weaken code/catalog/fixture source drift rejection.

# MiniMax global postcandidate execution map

Read-only source audit at `/Users/cyber/RestoX/barness`, 2026-10-18. No tests, credentials, vendor requests or checkout edits. The commands below are future commands after three complete actual candidates PASS. Names identify fresh output collections, never replacement attempts.

## Resolve one source blocker before candidate freeze

The current implementation has a circular postactivation preflight:

- `.scratch/barness-ai-minimax/tools/paid.py:56–66` requires both vets and the **complete** credential-free `go test -tags live -count=1 -timeout=20m ./ai/live` PASS. `paid_core.py:101–110` refuses every paid route without that check.
- `ai/live/glm_replay_process_test.go:59–61,98–114` deliberately starts three positive **formal capture replay** children and requires their separate replay reports to PASS, 22 replay attempts, zero HTTP attempts.
- `ai/live/glm_formal_test.go:13–31` selects **current** `BuiltinCatalog()` even for replay. `glm_replay_test.go:142` then requires the original immutable report's catalog hash to equal that current hash. M3 activation necessarily changes it. The current original GLM source is `34853063396ad58476e77dfe52463b3a9154ee77`, catalog version `2026-10-17.2`, hash `sha256:81d6d9f3b69e9231985967e1ec8e13640aa7f4fdfd9e7ccef72f7d05320a6942`.
- Thus postactivation `paid.py check` fails before GLM3 can renew the raw captures. Collecting GLM first at the old catalog does not solve current-catalog applicability. A new output directory, old checks/approval, narrowed test selection or hand-written `checks.json` would bypass the existing boundary.

The smallest compliant seam is **replay-only historical catalog selection**, preserving the existing full prerelease checks. Authenticate the original report through the already pinned capture index/archive/report hashes; obtain its complete catalog from that trusted source revision's `ai/release/catalog-snapshot.json` (the above commit contains it); recompute its hash and version and require exact agreement with the original report. A bounded, timed `git show <validated trusted commit>:ai/release/catalog-snapshot.json` or an independently pinned retained snapshot can supply those bytes. No arbitrary submitted catalog or current-report relabeling. The replay Client, report and manifest must all use that real historical snapshot so captured Result/Observer metadata remains exact.

Merely removing the hash comparison is insufficient: `glm_replay_compare_test.go:26–49,84` intentionally retains `catalogHash`/`catalogVersion` during exact Result/Observer comparisons. Do not normalize those fields away. Actual formal/live continues using current BuiltinCatalog. `publishGLMFormal` (`glm_formal_test.go:51–53`) already excludes replay; `supportmatrix/glm.go:88–91` must remain current-only. Historical replay must remain `mode=replay`, `kind=glm-capture-replay-only`, never publish own support, and never certify current service.

Give this source correction to the sole implementer before paid candidate freeze, with failure-first isolated executable/public-process evidence: unrelated catalog activation changes current hash; authenticated old capture still replays exactly with original catalog metadata and zero HTTP; current importer rejects that old actual report; missing/tampered catalog, wrong source/hash/version and existing replay tamper cases refuse. This is a correction to historical replay semantics, not a new paid-check exemption. No sanctioned existing wrapper option presently resolves the circle.

## Collections and applicability

Use three paid collections plus separate keyless outputs, all under the real repository (no symlink aliases):

| Symbol | Example fresh relative root | Paid jobs |
|---|---|---|
| C | `.evidence/minimax-global-candidate-final` | M3 candidates, after existing research |
| M | `.evidence/minimax-global-formal-glm` | M3 formal three + GLM three |
| R | `.evidence/minimax-global-glm-renewal` | none; raw ZIPs/selectors/usage/views |
| L | `.evidence/minimax-global-legacy` | ten generic + protocol114 + four effects |
| A | `.evidence/minimax-global-final-assembly` | none; public verification/merge/gate |

`paid_core.guard` pins HEAD as well as the whole selected source set, binaries, plan and auth. Every commit invalidates the old collection even when only documentation changes. Preserve C/M as immutable records; do not refresh their source IDs. Narrow report applicability is different: MiniMax ExecutionSource includes runtime Go + own capture/helpers/sources, **not** GLM formal JSON selectors or catalog_minimax.json; its formal catalog hash is checked separately. GLMExecutionSource includes runtime and GLM capture Go, not the renewed JSON selector data. JSON-only raw/view updates therefore do not expire those already measured M3/GLM reports if runtime/importer bytes and BuiltinCatalog remain identical. Jev seals every ai Go file, visual seals all non-test ai Go, and final Chinese evaluation has its own selected runtime/evaluation source set. Finish every Go/helper/importer fix before C; finish GLM JSON renewal before L.

## Exact order after C candidates PASS

1. Verify each C own session without keys: `go run ./ai/live/cmd/minimaxverify <C/mm-global-candidate-m3-ROUTE/SESSION>`. Save original captures, successful canonical launcher receipts and candidate-ledger prefix. Activate once:

   `go run ./ai/live/cmd/minimaxpromote -receipt <C/promotion.json> <chat-own-session> <responses-own-session> <messages-own-session>`

   The receipt is exclusive and outside all three own session directories. Activation changes only `ai/catalog_minimax.json`, requires its exact inactive hash and three independently verified actual candidates, and cannot be rerun over an active target.

2. `python3 .scratch/barness-ai-minimax/tools/renewal.py catalog-pins .evidence/minimax-global-catalog-pins`

   Commit activated catalog, generated `ai/release/catalog-snapshot.json`, and only the new catalog version/hash in `ai/examples/chineseeval/testdata/{jev,luna}-route-v1.json`. Keep route/model/dataset/call limits intact. All Go must still equal reviewed C execution bytes.

3. Prepare M, then check M using the unchanged public launcher commands. Parent writes its new exact APPROVED template, authorizing only `mm-global-formal-m3-{chat,responses,messages}` and `glm-{chat,responses,messages}`; M3 scope remains `MINIMAX_KEY: global-standard`.

   `python3 .scratch/barness-ai-minimax/tools/paid.py prepare .evidence/minimax-global-formal-glm`

   `python3 .scratch/barness-ai-minimax/tools/paid.py check .evidence/minimax-global-formal-glm`

   Sequentially call `paid.py run-selected <M> <each-six-exact-job>`. Do not edit source/HEAD until all six complete. M3 formal still owns unchanged 13-case inventory and full built-in catalog. GLM each owns 22 fixed model-qualified cases, 24 calls/32 attempts, 4096 output, 120s call/1800s process. The launcher's `-formal -publish` step creates M3 own `live-report.json` only after successful origin verification; never fabricate it.

4. Once all M paid jobs finish, publish portable origins (keyless, fresh output):

   `python3 .scratch/barness-ai-minimax/tools/renewal.py origins .evidence/minimax-global-origins --bundle mm-global-candidate-m3-chat=<C-chat-own> --bundle mm-global-candidate-m3-responses=<C-responses-own> --bundle mm-global-candidate-m3-messages=<C-messages-own> --bundle mm-global-formal-m3-chat=<M-chat-own> --bundle mm-global-formal-m3-responses=<M-responses-own> --bundle mm-global-formal-m3-messages=<M-messages-own>`

5. `python3 .scratch/barness-ai-minimax/tools/renewal.py glm-archive .evidence/minimax-global-glm-renewal --chat <M/glm-chat/SESSION> --responses <M/glm-responses/SESSION> --messages <M/glm-messages/SESSION>`

   Helper requires all three successful canonical `glm-ROUTE.json` receipts and exact original four capture pins, same actual source commit, current catalog/source and public validation. It writes R raw ZIPs plus `ai/live/testdata/glm-formal/{live-captures.json,live-captures.sha256,known-usage.json}`. The raw ZIP paths are ignored evidence inputs: preserve them before cleanup. Known usage is independently measured lower-bound Attempt sums; incomplete/cancelled stays unknown. Commit these selectors, any origin index and evidence metadata. This intentionally ends M's freeze; no more M paid jobs.

6. `python3 .scratch/barness-ai-minimax/tools/renewal.py glm-views .evidence/minimax-global-glm-renewal`

   This requires committed raw selector bytes and importer source, reuses original raw bytes, creates R `support-views/{chat,responses,messages}` plus three ZIPs, and updates `support-views.json`. Commit that selector and final metadata **before L prepare**. No Go or catalog changes.

7. Prepare/check L and obtain a fresh parent exact-source approval authorizing the ten `GENERIC` job names, `protocol114`, `text-jev`, `text-luna`, `visual-jev`, `visual-decisions`. Run those 15 exact selected jobs sequentially, without source/HEAD edits. Do not run already consumed `glm-*` jobs from L. Canonical common-root intents prevent that across collections. Component limits stay302/302 generic/effects +72/96 GLM =374/398. `summarize L` carries canonical M GLM receipts automatically; it does not reset them.

   Before final collection ends, `renewal.py m31-not-run <L>` may record three current zero-key/zero-call/zero-attempt missing-entitlement outcomes. It creates no paid intent or stage reservation. M3.1 remains absent from formal catalog/matrix.

8. Derive Jev support once from L's new own protocol114 session and its independent visual-jev effects:

   `go run ./ai/live/cmd/jevformal -protocol <L/protocol114/SESSION> -effects <L/visual-jev> -out <L/support> -require-live`

   `go run ./ai/live/cmd/jevformal -verify <L/support> -require-live`

   Verify both Chinese effects with L/chineseeval `-route ROUTE -verify <L/text-ROUTE/SESSION>`, both visuals with L/visualeval `-verify <L/visual-ROUTE> -require-live`, and M3 own formal with public minimaxverify `-formal`.

## Final 17 inputs across collections

Existing `paid.py keyless-plan` cannot assemble this schedule: `paid_plan.live_inputs(out)` assumes GLM and M3 formal evidence are both under the same collection as legacy/effects, and its `guard(out)` rejects a stale earlier source. This is a planner limitation, not a public gate limitation. Do not rerun paid producers, relabel captures, or copy/rewrite manifests to satisfy it. Direct public commands accept independent paths. Build exactly this ordered list:

1–10: unique `L/live/JOB/SESSION` for `openai-responses`, `openai-chat`, `anthropic-messages`, `google-gemini`, `deepseek-responses`, `deepseek-chat`, `openai-images`, `google-interactions-image`, `openai-decisions`, `typesafe-classifier`.

11: `L/support` (Jev1.14 own combined protocol/effects support; distinct from generic Jev1.13).

12–14: `R/support-views/chat`, `R/support-views/responses`, `R/support-views/messages`. Original M GLM raw own session directories are also valid public inputs if complete; derived views are preferable because the current offline selectors consume exactly these audited views. They retain the original measured timestamp/source.

15–17: unique `M/mm-global-formal-m3-ROUTE/SESSION` for chat/responses/messages, each already containing own published `live-report.json`.

Execute keyless `go run ./ai/live/cmd/supportmatrix <all17-directories>` **once** after every independent validator passes. The CLI stages all input validation and writes atomically; remerging identical timestamps correctly refuses. Matrix mutation is excluded from paid source manifests, but committing it changes HEAD: finish any L paid/NOT_RUN jobs before that commit, or defer the commit until after gate.

Then keyless real-repository command:

`go run ./ai/release/cmd/releasegate -out <A/gate> -evaluation-jev <L/text-jev/SESSION> -evaluation-luna <L/text-luna/SESSION> -visual-evaluation-jev <L/visual-jev> -visual-evaluation-decisions <L/visual-decisions> -live <INPUT1> ... -live <INPUT17>`

Repeat `-live` exactly17 times. No `-bundle`/`-race-bundle`: the actual gate runs both vets, ordinary20m, race75m, frozen pi and pressure, full differential and audits. Use the launcher's small allowlisted environment (e.g. paid_core.execute with source=None), finite outer6600s, fresh A output, and require exit0 + fresh release-report.json + pass=true. It must include every old mapping, 17 current own objects/16 routes, four effects, load/resource/revocation checks. Fullbase review and ticket closure follow actual gate success.

## Portable origins and immutable retention

`renewal.py origins` uses authority only from common-root `.evidence/minimax-launches/mm-global-{candidate,formal}-m3-ROUTE.json`: COMPLETED, exit0, credentials acquired, clean launcher audit, exact global identity/source, original report/seal/ledger hashes and public verifier. It writes curated `ai/live/testdata/minimax-formal/origins.json` keyed by report hash; six entries cover own candidates and formal reports. The index is intentionally excluded from both paid generation sources and MiniMax ExecutionSource, but committing it still invalidates prepared HEAD. Finish six paid M jobs before commit.

Retain every original own bundle byte, including report, seal, full ledger/predecessor evidence, exact requests/SSE/events/Observer/results/sources/catalog/manifest/audit and formal `live-report.json`. Never add archive metadata inside sealed bundles. Retain collection source/binary manifests, plans, exact APPROVED reviews/checks, command receipts/logs and original canonical successful launcher receipts, both append-only stage ledgers and immutable research/CN history; pin their hashes in an external archive index. Preserve original source commits in Git. Keep all R raw/view ZIPs at selector paths (and archive them portably with hash/restore mapping) before deleting worktrees or ignored evidence. Derivation copies original GLM bytes rather than changing them.

Local public origin validation prefers a present canonical receipt and requires the captured stage ledger to be an exact canonical prefix. A broken existing receipt cannot fall back to origins.json. In a portable checkout **without** those canonical receipts, only the curated repository index supplies authority; a bundle-supplied receipt/index never does. Verify portability in a separate checkout/common Git root, not a worktree sharing canonical receipts. Research failures and historical CN captures are retained accounting/provenance, never current actual PASS.

MiniMax candidate120/144 and formal60/72 ledgers, every per-route/model/purpose20/24/40000/120000 slot, preexisting research/CN spend, and all one-shot intents remain common-Git scoped across C/M/L. New collection preparation creates no extra spend allowance.

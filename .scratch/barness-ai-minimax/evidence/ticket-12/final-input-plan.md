# Final mixed-source release input plan

Read-only narrow follow-up at `/Users/cyber/RestoX/barness`; inspected root HEAD `b862249da6510d7bd23ab06f0470ce758529965c`. This note writes only `/tmp`. No tests, code changes, credentials, network or vendor calls. The prior MiniMax execution-source boundary audit is accepted, not repeated. The corrective-1 implementation is still a separate worker responsibility and is not treated as already reviewed/passed here.

## Decision and ordering

**Wait for the worker merge, successful corrective GLM collection, and both raw/view selector commits before preparing the final legacy collection L.** Then freeze L once, check/review its exact bytes, execute only its 15 unused original legacy/effect jobs, and keep root source and HEAD unchanged until every paid child finishes.

Do not merge during paid execution. The launcher requires the entire selected source set and HEAD to match at entry and again after the child. That full set includes every `ai` Go/JSON file, tool Python files and selected MiniMax source documents, even when a particular evidence family's narrower execution map excludes them. Adding a Python composition helper or committing GLM selector JSON after preparing L invalidates L. [paid_core.py:68](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_core.py:68), [paid_core.py:88](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_core.py:88), [paid.py:156](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid.py:156).

Recommended stages:

1. Finish/review/merge the strictly GLM-only capture and prospective corrective-1 Python work. Confirm the worker changed no M3-selected Go, shared recorder, runtime, catalog or importer source; keep every old failed capture and receipt.
2. Execute the separately approved corrective-1 workflow; establish three coherent new actual GLM complete PASS captures under its reviewed contract. This note does not supply that new paid command or change its budget.
3. Publish fresh GLM raw archives and known usage through the reviewed corrective workflow. Commit the raw selector/index/pin/known-usage JSON before deriving views. Derive all three public support views and commit the view selector. Their source identity remains the immutable measured source and current importer bytes, not a claim that the selector commit reran models. [renewal_glm.py:93](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:93), [renewal_glm.py:100](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:100), [glm_derivation_source.go:45](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_derivation_source.go:45).
4. Finish any optional committed composition helper, its failure-first checks and final source review now. No helper is technically required; see below.
5. Prepare a **new** L, e.g. `.evidence/minimax-global-legacy-final`, then check and approve the exact L template. Restrict `authorizedRoutes` to the original 15 jobs below. Preserve all already spent original/corrective GLM and M3 markers; do not authorize them again from L.
6. Hold source/HEAD steady throughout L paid execution and any source-sensitive Jev support derivation. Paid evidence output under `.evidence` is outside the source set. Re-run keyless `summarize L` after all children settle; concurrent intermediate summary files are not the final accounting authority.
7. Verify the 17 explicit input paths, atomically merge exactly them once, then run the real full release CLI with all 17 `-live` arguments and the four effect packages. No source corrections after collecting Jev protocol114; they would require re-evaluating evidence applicability before spending again.

Preparation commands (run later by parent, not executed in this audit):

```sh
python3 .scratch/barness-ai-minimax/tools/paid.py prepare .evidence/minimax-global-legacy-final
python3 .scratch/barness-ai-minimax/tools/paid.py check .evidence/minimax-global-legacy-final
```

After parent exact-source APPROVED receipt, each paid child retains the existing command form:

```sh
python3 .scratch/barness-ai-minimax/tools/paid.py run-selected .evidence/minimax-global-legacy-final <one-exact-unused-job-name>
```

The 15 names are: `openai-responses`, `openai-chat`, `anthropic-messages`, `google-gemini`, `deepseek-responses`, `deepseek-chat`, `openai-images`, `google-interactions-image`, `openai-decisions`, `typesafe-classifier`, `protocol114`, `text-jev`, `text-luna`, `visual-jev`, `visual-decisions`. Their selectors/models/endpoint/limits remain the original ones in [paid_plan.py:24](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_plan.py:24). They do not include another GLM or MiniMax run.

## What could run before the merge, and why waiting is better

| Family | Actual freshness boundary | Effect of stated worker-only `ai/live/glm_*.go` + tool Python changes |
| --- | --- | --- |
| Ten generic support reports | Manifest equals its original report commit, current catalog, current SDK; no comparison to current HEAD or a full Go execution map. Adapter applicability still requires maintainer judgment. | Completed pre-merge captures could remain applicable if runtime/catalog/SDK stay byte-identical; this is not blanket authority to ignore future runtime changes. |
| Jev protocol114 and derived Jev1.14 support | **Every `ai/**/*.go`, including tests**, go.mod/go.sum, fixed source JSON and the independent Python verifier. | A pre-merge protocol capture becomes stale as soon as any GLM test file changes. Must wait. |
| Two visual effects | **Every non-test `ai/**/*.go`**, go.mod/go.sum, independently verified exact source bytes and set. | The currently stated `_test.go` and task Python changes alone do not expire completed visual runs; any new non-test Go helper does. |
| Two Chinese text effects | Recorded full code snapshot stays immutable; release applicability compares a narrower map of non-test runtime/testkit + Chinese CLI/request-id source, excluding live/E2E/release and unrelated example programs. | The stated GLM live test and task Python changes alone do not expire completed text runs. |

Citations: generic manifest [live_manifest.go:14](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_manifest.go:14), current catalog/matrix run matching [live.go:141](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live.go:141), explicit runtime-rerun responsibility [live.go:183](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live.go:183); Jev all-Go selector and exact comparison [jevformal/verify.go:148](/Users/cyber/RestoX/barness/ai/internal/testkit/jevformal/verify.go:148); visual selector [visualeval/bundle.go:76](/Users/cyber/RestoX/barness/ai/examples/visualeval/bundle.go:76) and current comparison [visualeval/verify.go:168](/Users/cyber/RestoX/barness/ai/examples/visualeval/verify.go:168); Chinese recorded snapshot [provenance.go:28](/Users/cyber/RestoX/barness/ai/examples/chineseeval/cmd/chineseeval/provenance.go:28) versus current release applicability [evaluation_source.go:31](/Users/cyber/RestoX/barness/ai/internal/testkit/release/evaluation_source.go:31).

Thus there is a technically safe pre-merge subset **only if root stays frozen for each entire reviewed child and the eventual code delta stays exactly as stated**. It requires another prepare/check/approval boundary and splitting source attribution. It saves no calls, cannot include protocol114, and risks wasted evidence on an expanded worker fix. The recommended single final L avoids that risk and simplifies review.

After the final L freeze, distinct unused jobs have separate process environments, selected secrets, exclusive route intents and evidence roots. Bounded concurrent dispatch is structurally possible while root stays immutable; use distinct route outputs and retain per-child completion/receipts. Do not run a full suite or merge/renew selector data concurrently. No provider load increase, retries, budget changes or new route names are implied. Exact source guards and individual one-shot markers still apply to every child.

## Legitimate mixed-source final composition

Let:

- **M** = existing `.evidence/minimax-global-formal-glm`, retaining the original three M3 complete formal own packages at source `89000346…`.
- **R** = the fresh successful corrective GLM renewal directory, whose three `support-views/<route>` packages derive from the new original raw captures/index and are current under the unchanged public importer.
- **L** = `.evidence/minimax-global-legacy-final`, the final 15-job collection after worker/selector commits.

The final 17 own inputs are exactly:

| Role | Path family |
| --- | --- |
| Ten original generic identities | `L/live/<exact-generic-name>/<unique-session>/` |
| Independent Jev1.14 identity | `L/support/` |
| GLM Chat / Responses / Messages | `R/support-views/chat`, `R/support-views/responses`, `R/support-views/messages` |
| M3 Chat | `M/mm-global-formal-m3-chat/20261018T030234.929339000Z` |
| M3 Responses | `M/mm-global-formal-m3-responses/20261018T030308.589357000Z` |
| M3 Messages | `M/mm-global-formal-m3-messages/20261018T030335.626044000Z` |

The four separate effect paths are `L/text-jev/<unique-session>`, `L/text-luna/<unique-session>`, `L/visual-jev`, and `L/visual-decisions`.

Mixed commits are legitimate **only through each family's existing current-source/catalog verification**. Neither supportmatrix nor releasegate requires all 17 reports to claim one Git commit. The matrix CLI loads arbitrary report/directory arguments and merges them oldest first, publishing only after all validate: [supportmatrix/main.go:42](/Users/cyber/RestoX/barness/ai/live/cmd/supportmatrix/main.go:42). The release CLI dispatches M3/GLM/Jev verification at the manifest boundary and accepts repeatable `-live` paths: [live_manifest.go:14](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_manifest.go:14), [releasegate/main.go:56](/Users/cyber/RestoX/barness/ai/release/cmd/releasegate/main.go:56). The M3 family needs byte-identical selected execution set and catalog, not relabelled source; R must retain its own original measured commit and derivation proof.

### Keyless commands after L outputs exist

Resolve each `<unique-session>` by requiring exactly one direct child with the stated leaf (`live-report.json`, `multimodal-report.json`, or `manifest.json`). Refuse missing/ambiguous outputs; do not select newest by timestamp or substitute historical bundles.

First derive Jev support using its **own L protocol114** and independent L visual Jev task:

```sh
go run ./ai/live/cmd/jevformal -protocol <L/protocol114/unique-session> -effects <L/visual-jev> -out <L/support> -require-live
```

Verify Chinese effects with the frozen L binary (`-route jev|luna -verify <own-text-bundle>`), visual effects (`L/visualeval -verify <own-visual-bundle> -require-live`), Jev support (`go run ./ai/live/cmd/jevformal -verify <L/support> -require-live`), and each M3 package (`go run ./ai/live/cmd/minimaxverify -formal <own-M-package>`). Validate the complete mixed 17 against a temporary **copy** of the current matrix first:

```sh
go run ./ai/live/cmd/supportmatrix -matrix <temporary-matrix-copy.json> <ten-L-generic-packages> <L/support> <three-R-support-views> <three-M-formal-packages>
```

Only after all verifiers and temporary import succeed, run the same 17 arguments once against the actual default matrix (omit `-matrix`). Original failures remain in their raw captures/history; no report, source map, timestamp or producer receipt is rewritten.

The actual release argv is:

```text
go run ./ai/release/cmd/releasegate
  -out L/gate
  -evaluation-jev L/text-jev/<unique-session>
  -evaluation-luna L/text-luna/<unique-session>
  -visual-evaluation-jev L/visual-jev
  -visual-evaluation-decisions L/visual-decisions
  -live <each of the exact 17 own packages above, once per role>
```

These are normal argv elements, not shell wildcard expansions. Do not add `-bundle` or `-race-bundle`. The public CLI runs both vets, actual ordinary20m/race75m suites with frozen pi + pressure, and independent evidence audits/rollback checks. Preserve the original 13 protocol routes / 14 legacy identities, 317 decisions, 6 extensions and 262 old trace objects; include all old effect obligations. Composition changes paths only, not selectors, catalog, matrix schema, trace/ledger entries or gate behavior.

## Is a composition helper needed before paid freeze?

**No new production/helper implementation is required.** Existing public CLIs already support the correct mixed-path composition. A parent-owned explicit argv manifest under `.evidence` or a `/tmp` keyless planning script can resolve/package paths and write command receipts after L completes without changing frozen execution sources. It must not read credentials or execute paid children. It should pin the 17 identities/paths/report hashes and four effect roles, preserve original Git/derivation claims, refuse duplicates/ambiguity/symlinks, and rely on public verifiers for authority.

Current `paid.py keyless-plan L` cannot be used unchanged: `live_inputs(out)` expects all GLM and M3 inputs physically under the same `out`, and chooses raw GLM process bundles rather than R's derived support views. It is a convenience planner limitation, not a release CLI restriction. [paid_plan.py:118](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_plan.py:118), [paid_plan.py:129](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_plan.py:129). Do not create symlinks, transplant reports or copy captures into L to satisfy that layout; Jev evidence rejects symlink paths and copying would obscure source ownership unnecessarily.

If the team prefers a committed reusable mixed-source planner, add/review/test its failure modes **before preparing L**, since tool Python files enter the paid source freeze. It needs no Go/importer/gate change and must not create paid authorization, bypass L guard, rehash measured evidence, reset original/corrective budgets or change the 17-role set.

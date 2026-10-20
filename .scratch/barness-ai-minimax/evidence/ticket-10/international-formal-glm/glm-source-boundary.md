# GLM capture source and recovery boundary

Read-only audit at `/Users/cyber/RestoX/barness`, HEAD `89000346b48480d94c3d6f6523629870f4e7b1b0`, 2026-10-18. No repository/evidence mutations, tests, credentials, network or vendor calls. Only this report is written outside the repository. Conclusions concern the existing contract and implementation, not a proposed waiver.

## Conclusion

A correction confined to `ai/live/glm_*.go` or the three non-test Go files under `ai/internal/testkit/glm/` can preserve existing MiniMax formal **execution-source applicability**, provided every MiniMax-selected byte and catalog identity remains unchanged. It invalidates the current GLM execution-source snapshots, including the already successful Chat/Messages reports. Changing the shared recorder, runtime, evidence/audit helpers, or any non-test `supportmatrix/` file instead changes the MiniMax source set and expires the current three M3 formal captures too.

There is **no existing admissible one-scene recovery/reconciliation path** for the observed truncated GLM Responses capture. Neither a free budget remainder nor the no-paid support-view derivation permits replacing one case, changing a source claim, clearing truncation, or turning the failed process into a launcher-pinned successful complete report.

## Independent current observations

Read the actual complete source maps from `.evidence/minimax-global-formal-glm/mm-global-formal-m3-chat/20261018T030234.929339000Z/minimax-report.json` and `.evidence/minimax-global-formal-glm/glm-chat/20261018T030407.471080000Z/glm-execution-source.json`, then independently SHA-256 hashed every mapped current file:

- MiniMax: **263 entries**, no current byte drift.
- GLM: **211 entries**, no current byte drift.
- Shared: **193 entries**; GLM-only 18; MiniMax-only 70.
- Actual GLM Chat and Messages report `PASS`, `complete:true`, 22 calls / 22 actual HTTP attempts each, actual Git commit `89000346…`.
- Actual GLM Responses report `FAIL`, `complete:false`, 22 calls / 22 actual attempts; its sole non-PASS scene is `glm-5.3/control-low`, failure `assertion`, `truncated:true`.

These are metadata/source observations; I did not reproduce the recorder implementation diagnosis or reinterpret model output.

## Exact source scope

The actual MiniMax function is named `minimaxlive.ExecutionSource` (the task's “MiniMaxCapturedExecutionSource” refers to that map). It walks `ai`, skips `node_modules` and `.evidence`, and considers additions/deletions material independently of Git revision: [source.go:11](/Users/cyber/RestoX/barness/ai/internal/testkit/minimaxlive/source.go:11).

| Selected scope | MiniMax | GLM |
| --- | --- | --- |
| Direct `ai/*.go`, excluding `_test.go` | Yes | Yes |
| `go.mod`, `go.sum` | Yes | Yes |
| All `ai/live/minimax_*.go`, including test files | Yes | No |
| All `ai/live/glm_*.go`, including test files | No | Yes |
| Seven shared live files listed below | Yes | Yes |
| `ai/examples/requestid/` | All non-test files | Non-test `.go` only |
| `ai/internal/testkit/evidence/`, `audit/` | All non-test files | Non-test `.go` only |
| `ai/internal/testkit/minimax/`, `minimaxlive/`, `supportmatrix/` | All non-test files | No |
| `ai/live/testdata/minimax/` | All non-test files | No |
| `ai/internal/testkit/glm/` | No | Non-test `.go` only |

Seven shared live files: `main_test.go`, `recorder_test.go`, `proxy_test.go`, `combos_test.go`, `decisions_config_test.go`, `decisions_sources_test.go`, `session_test.go`, all under `ai/live/`.

Exact selectors are at [MiniMax source.go:31](/Users/cyber/RestoX/barness/ai/internal/testkit/minimaxlive/source.go:31) and [GLM glm_source.go:40](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_source.go:40); dependency manifests are added at their lines 54 and 63 respectively. MiniMax rejects a non-identical current map at [verify.go:59](/Users/cyber/RestoX/barness/ai/internal/testkit/minimaxlive/verify.go:59). GLM independently rejects current map drift at [glm_capture.go:132](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_capture.go:132).

The 18 currently GLM-only entries are:

```text
ai/internal/testkit/glm/catalog.go
ai/internal/testkit/glm/messages.go
ai/internal/testkit/glm/responses.go
ai/live/glm_assembly_test.go
ai/live/glm_capture_test.go
ai/live/glm_config_test.go
ai/live/glm_controlled_test.go
ai/live/glm_finish_test.go
ai/live/glm_formal_process_test.go
ai/live/glm_formal_test.go
ai/live/glm_formal_validation_test.go
ai/live/glm_harness_test.go
ai/live/glm_process_test.go
ai/live/glm_replay_cancel_test.go
ai/live/glm_replay_catalog_test.go
ai/live/glm_replay_compare_test.go
ai/live/glm_replay_process_test.go
ai/live/glm_replay_test.go
```

Documentation, selector JSON and Python launch/renewal plumbing are outside both measured execution maps, except the explicit MiniMax non-test prefix scopes above. This **does not** preserve an existing prepared paid collection after changing them: the broader launcher freeze selects all `ai` Go/JSON files (with two stated mutable output exclusions), MiniMax source documents, the pricing source, and all tool Python files; it also requires HEAD equality. [paid_core.py:68](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_core.py:68), [paid_core.py:88](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_core.py:88). A new prepared collection is distinct from authority to reuse a spent route marker.

## Complete GLM report and raw renewal requirements

1. The source-pinned plan always supplies 22 model-qualified scenes. The harness iterates all 22, without a recovery selector or rerun loop: [glm_harness_test.go:110](/Users/cyber/RestoX/barness/ai/live/glm_harness_test.go:110). Every attempt is written under `attempt-1-*`; actual truncation immediately fails the scene before complete request/response fixtures are emitted: [glm_harness_test.go:341](/Users/cyber/RestoX/barness/ai/live/glm_harness_test.go:341), [glm_harness_test.go:359](/Users/cyber/RestoX/barness/ai/live/glm_harness_test.go:359).
2. `NewGLMReport` requires live/formal mode, source and current catalog, exact ordered expected22/scenes22, at most one call/attempt per scene, no replay, no truncation, and sums equal process counters. [glm.go:90](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm.go:90), [glm.go:97](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm.go:97), [glm.go:104](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm.go:104), [glm.go:128](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm.go:128).
3. The artifact contract permits exactly **23 manifest cases** (source + 22 scenes), exactly the fixed per-case artifact names and no hidden attempts or extra root/case artifacts. [glm_artifact_contract.go:10](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_artifact_contract.go:10). Assertions, manifest status, file hashes and actual wire must agree; the own exchange must be unique and untruncated: [glm_capture.go:140](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_capture.go:140), [glm_capture.go:199](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_capture.go:199).
4. `renewal.py glm-archive` requires all three routes PASS/complete/audited with current catalog/source; a canonical completed successful launcher receipt with key-acquisition/audit/source agreement; exact four original capture hashes; and **one actual Git commit for all three routes**. It then invokes the public Go importer. [renewal_glm.py:39](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:39), [renewal_glm.py:48](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:48), [renewal_glm.py:57](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:57), [renewal_glm.py:76](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:76).
5. Only a successful exact-source own live child receives `capturePins`. [paid.py:141](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid.py:141), [paid.py:187](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid.py:187). A one-shot canonical marker is created exclusively across output directories/worktrees; old output also refuses reuse. [paid.py:156](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid.py:156). Changing directories does not provide a new allowance or a new original-process identity.

Consequently even a strictly GLM-only source correction cannot mix new Responses evidence with current Chat/Messages raw reports: the latter fail current GLM map equality, and the renewal requires a single actual commit. Relabelling their Git commit or execution maps would destroy the original evidence claims.

## Existing recovery mechanisms and why they do not apply

- GLM spec R21 allows **one scene-level retry for a vendor failure**, counted in the same 24-call/32-attempt budget; fixed budgets must be changed in the spec and plan before execution. [GLM spec:222](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/spec.md:222). Current Responses has two logical-call slots and ten attempt slots numerically remaining. That contract does not implement a post-process splice path; the observed failure is local capture truncation, not established vendor failure. Current fixed importer would reject an added second attempt/case. No such recovery was executed by this audit.
- `recovery` is one of the original 22 scenarios (a normal post-cancellation generation), not a facility to rerun a failed scene or reconcile reports. See [glm.go:34](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm.go:34) and the fixed harness loop above.
- Capture replay requires an original live **PASS**, complete22, own route/source/catalog and exact untruncated captures. It explicitly stays replay and performs zero vendor traffic. It cannot turn incomplete recording into original live evidence. [glm_replay_test.go:141](/Users/cyber/RestoX/barness/ai/live/glm_replay_test.go:141).
- The no-paid support view corrects generic known-usage aggregation while preserving original measured bytes/source. Both its producer and consumer call `NewGLMReport(original)` first, retaining all above checks. It cannot restore missing SSE bytes, reverse FAIL, waive truncation or renew measured source. [glm_derivation.go:14](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_derivation.go:14), [glm_derivation.go:29](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_derivation.go:29), [glm_derivation.go:59](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_derivation.go:59), [glm_derivation.go:169](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_derivation.go:169).

The current paid plan limits GLM to 24/32 per route and Zhipu72 calls/96 attempts overall, with automatic process reruns disabled; the canonical launcher remains one-shot. [paid_plan.py:12](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_plan.py:12), [paid_plan.py:50](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid_plan.py:50), [paid.py:227](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/paid.py:227). A full corrected three-route repeat would be additional 66 calls, beyond the remaining fixed allowance; a full Responses repeat would already exceed its 24-call cumulative limit. Neither is authorized by the existing inventory.

**Admissible present status:** retain MiniMax three formal PASS records while their exact selected bytes/catalog remain unchanged; retain GLM Chat/Messages current PASS and Responses original FAIL with consumption counted. Complete old-route release renewal is blocked by that GLM Responses capture. Any new correction/recovery contract must be explicit and prospective, retain all failures/spend, and be reviewed before paid execution; this audit offers no metadata or source waiver.

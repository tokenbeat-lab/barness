# GLM renewal archive portability before final legacy freeze

Read-only inspection of `/Users/cyber/RestoX/barness`, 2026-10-18. No repository/evidence mutations, tests, credentials, network or vendor calls. This note alone is written under `/tmp`.

## Finding

New GLM archives are **not copied or hydrated automatically** into another checkout. The current-root final release gate will find them while the exact selector paths remain present in that root, but committing only selector JSON leaves a fresh Git checkout/worktree without the six referenced current ZIPs. Retain the three raw ZIPs and three derived support-view ZIPs as tracked immutable data at their original generated paths before final L prepare.

This is a portability/retention requirement, **not an immediate new dependency of the fixed historical rollback subprocesses**. The exact isolated paths are distinguished below; no Go change is needed.

## Requested `.scratch` R root is currently rejected

The proposed new root is:

```text
.scratch/barness-ai-minimax/evidence/ticket-12/current-glm
```

`renewal_common.output()` does **not** accept any arbitrary repository-relative fresh directory. It first validates containment/no symlink, then explicitly requires first component `.evidence` and `minimax` in the second component. The `.scratch` root is rejected before directory creation. [renewal_common.py:41](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_common.py:41), [renewal_common.py:53](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_common.py:53).

Two prospective choices avoid relocation and index drift:

1. **No code changes:** choose R directly as `.evidence/minimax-ticket12-current-glm`, generate both raw and views there, and force-track the exact six ZIPs and safe keyless checkpoints at those original paths. The existing output validator accepts that shape. This is the minimal currently executable choice.
2. **If the requested `.scratch/.../ticket-12/current-glm` location is required:** before R is created, have the sole worker add/review a narrowly constrained Python-only output allowlist for that exact task-owned root, preserving containment, no-symlink, new-directory and existing-checkpoint rules. Then use that root directly for both commands and force-track the ZIPs there. No Go/runtime/importer/catalog change is necessary. This is a source change to the paid tooling and must finish before L prepare; this audit does not implement or approve it.

Do not generate R elsewhere and move it later, rewrite `raw-published.json`, or patch selector SHA fields after derivation. Views pin the original raw index hash. Generating directly in the final retention location avoids all these adjustments.

## Exact generation and retention order

After the reviewed corrective workflow supplies three current own successful raw captures, invoke its approved raw renewal command against the chosen fresh R root. The existing standard command shape is:

```text
python3 .scratch/barness-ai-minimax/tools/renewal.py glm-archive R
  --chat <successful own corrective Chat package>
  --responses <successful own corrective Responses package>
  --messages <successful own corrective Messages package>
```

The corrective workflow may provide a separately reviewed command; do not pass original failed captures to the unchanged canonical-success boundary.

`renewal_glm.capture` generates exactly `R/glm-chat-<actual-source12>.zip`, `R/glm-responses-<actual-source12>.zip`, and `R/glm-messages-<actual-source12>.zip`. Each archive path is written directly relative to the repository root, with byte hash and original report path/hash. [renewal_glm.py:71](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:71).

First retention commit:

- Force-track those **three exact raw ZIPs**, retaining their original bytes and generated paths; do not repack them.
- Commit `ai/live/testdata/glm-formal/live-captures.json`, `live-captures.sha256`, and `known-usage.json` exactly as generated.
- Retain R's keyless `raw-published.json` checkpoint and `raw-selector-before.json` history, plus the reviewed keyless command/audit receipts needed to repeat validation. Force-add only explicit owned files, not the whole output tree indiscriminately.
- Independently confirm the three index archive hashes equal the retained regular-file bytes and the selector pin equals the exact index bytes. Tracking or committing a file changes neither archive bytes nor its hash.

Then derive in the **same R**, with no directory move or checkpoint modification:

```text
python3 .scratch/barness-ai-minimax/tools/renewal.py glm-views R
```

The helper requires the raw selector bytes to match the checkpoint and be committed. It generates original directories `R/support-views/{chat,responses,messages}` plus `R/support-chat.zip`, `R/support-responses.zip`, and `R/support-messages.zip`; the current support-view selector points directly to these ZIPs. [renewal_glm.py:100](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/tools/renewal_glm.py:100).

Second retention commit:

- Force-track the **three exact generated support ZIPs**, preserving bytes and paths.
- Commit `ai/live/testdata/glm-formal/support-views.json` exactly as generated.
- Retain R's safe keyless `views-published.json`, `support-views-before.json` and validation/derivation receipts. The three expanded view directories can remain generated artifacts because the committed ZIPs reconstruct their exact bytes.
- Verify all six currently referenced ZIPs exist, are regular files, are listed by `git ls-files`, and match their respective trusted index hashes. Keep the original failed collection and its immutable capsule unchanged.

Only after those data commits and all tooling changes are complete should the parent prepare/check/review final L. This preserves the intended original source attribution: the second data commit does not claim it reran model calls.

Both `.evidence` and `.scratch` ZIP paths need explicit force-add under current ignore rules: [`.gitignore:1`](/Users/cyber/RestoX/barness/.gitignore:1). Do not remove broad ignore rules or add compatibility copies. Use exact paths from the generated JSON rather than a wide ZIP glob.

## Actual consumers and isolated restoration behavior

### Current-root gate and GLM tests

The final release CLI runs ordinary/race suites from the caller's repository, not a new full clone: [releasegate/main.go:111](/Users/cyber/RestoX/barness/ai/release/cmd/releasegate/main.go:111). GLM E2E opens each index's repository-relative archive literally with `os.ReadFile("../../"+archive)` and checks its hash before extraction: [glm_formal_usage_test.go:18](/Users/cyber/RestoX/barness/ai/e2e/glm_formal_usage_test.go:18). The public derived-view verifier independently opens the raw archive at `repositoryRoot / entry.Archive` and rejects absence/hash mismatch: [glm_derivation_source.go:84](/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/glm_derivation_source.go:84). No current archive fallback, common-root lookup or hydration exists.

Therefore root-only success with untracked ignored ZIPs is possible, but it does not establish that a clean checkout contains the same required evidence. Six tracked originals close that GLM-specific gap without changing verification code.

### Fixed GLM and MiniMax software rollbacks

GLM rollback uses `git archive` of fixed baseline `f5cf536…`, verifies every baseline blob/mode, then runs a no-GLM public probe and only `TestLocalAssemblyExample`: [GLM rollback.py:91](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/rollback.py:91), [GLM rollback.py:108](/Users/cyber/RestoX/barness/.scratch/barness-ai-glm/evidence/ticket-13/rollback.py:108).

MiniMax rollback likewise restores fixed `ecfd4b4…`, verifies that historical tree and runs the no-MiniMax probe plus only `TestLocalAssemblyExample`: [MiniMax rollback.py:95](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-08/rollback.py:95), [MiniMax rollback.py:109](/Users/cyber/RestoX/barness/.scratch/barness-ai-minimax/evidence/ticket-08/rollback.py:109).

Neither restoration copies current ignored `.evidence` ZIPs, neither needs the newly selected current GLM raw/view archives, and neither executes the current GLM formal-report E2E inside the restored historical module. Do not alter fixed baseline pins, historical archives or rollback scripts to inject current GLM data.

### Current-source private tracker clone

`TestCurrentReleaseTrackerTransition` copies only `git ls-files -- ai go.mod go.sum <one tracker>`. It would omit both ignored files and even force-tracked archives outside that path set. Its child runs only `TestCurrentReleaseTraceAndBlocker`, whose public CLI invocation uses a synthetic incomplete offline bundle and no GLM live inputs. It therefore does not need the new GLM archives. [current_catalog_release_test.go:81](/Users/cyber/RestoX/barness/ai/e2e/current_catalog_release_test.go:81), [release_gate_process_test.go:21](/Users/cyber/RestoX/barness/ai/e2e/release_gate_process_test.go:21). The previous Node dependency/setup issue is separate; retaining ZIPs does not change that clone's bounded purpose.

A full new Git checkout/worktree, unlike this deliberately filtered test clone, will obtain force-tracked `.scratch` or `.evidence` files at their exact committed relative paths. The six ZIPs only address current GLM portability; other historical ignored ZIP/Node prerequisites remain their own existing setup obligations.

## Source-set and applicability impact

- R's own JSON/ZIP/checkpoint files under either proposed external tree are outside `ai` and outside production Go/runtime captures. Force-tracking these bytes does not change M3 or GLM measured execution maps.
- The four selector/known-usage files under `ai/live/testdata/glm-formal/` are data only. They are outside `minimaxlive.ExecutionSource` and outside the GLM Go-only measured scope; unchanged runtime, shared capture source, importer and catalog preserve the measured family applicability already established.
- The importer independently pins its own non-test supportmatrix Go source and original capture/index bytes; keeping R in place and using raw then views in the same R preserves those claims. A support-view data commit is not a source waiver or a model rerun.
- The broad paid launcher freeze **does include** these `ai` selector JSON files, any Python output-allowlist change and exact HEAD. Thus both retention/data commits and any optional Python change must precede final L prepare. An already prepared collection cannot continue through those commits.
- Jev protocol's all-Go and visual's non-test-Go source scopes are unaffected by pure JSON/ZIP retention, but collect them only after the final source/HEAD freeze as planned.

No source-set exclusions, report edits, archive repacks, index relocation, raw checkpoint rewrites, historical capture relabelling or Go changes are needed for the unchanged-output-root option.

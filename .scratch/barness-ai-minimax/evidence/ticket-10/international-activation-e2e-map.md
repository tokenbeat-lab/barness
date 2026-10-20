# Data-only MiniMax activation: E2E selector map

Read-only audit, 2026-10-18, `/Users/cyber/RestoX/barness`. No tests, checkout changes, credentials or vendor calls. Findings below follow source paths; they are predicted activation failures, not execution results. Sole implementer: `/root/live_harness`.

## Deterministic activation dependencies to fix before paid source freeze

| Public test selector | Current dependency / expected failure after activation | Minimal responsibility |
|---|---|---|
| `^TestOpenAIImagesBuiltinCatalog$` | `images_builtin_test.go:36` requires current whole-catalog version `2026-10-17.2`; sourced fixed image model/rates remain unchanged. | Keep official retrieved date/model equality; assert current global snapshot/version without hardcoding a past unrelated catalog revision. Independent complete content pin remains in usage fixture. |
| `^TestOpenAIImagesLiveWireReplay$` | `images_builtin_test.go:131` hardcodes same global version on new current Client metadata. | Compare attribution with selected current catalog/version and current independently pinned hash; immutable original image wire/usage checks remain. |
| `^TestGoogleImagesBuiltinCatalog$` | `google_images_builtin_test.go:35` same fixed current global version. | Same separation of unchanged official image facts from evolving global snapshot. |
| `^TestGoogleImagesLiveWireReplay$` | `google_images_builtin_test.go:149` same metadata global version. | Preserve exact original wire, raster, usage and reporting checks; current attribution follows current catalog. |
| `^TestClassifierBuiltinOfficial$` | `classifier_builtin_test.go:34` same global version. Provider-scoped TypeSafe count2 remains correct. | Remove past global-version coupling while keeping official model/rates/capability equality. |
| `^TestDecisionsFormalCatalog$` | `decisions_formal_catalog_test.go:43` hardcodes old global version; `:47–49` subtracts later GLM models from current catalog but leaves new MiniMax models, so prior exact catalog/hash comparison also fails. | Compare independent existing classifiers/facts and current content pin; historical subtraction must remove both independently added GLM and MiniMax subsets. Keep exact prior-model equality/hash and old snapshot bytes. |
| `^TestJevFormalCatalog$` | `jev_formal_catalog_test.go:39–42` subtracts GLM but leaves MiniMax before exact historical equality/hash; both plain/hooked fail. | Also exclude MiniMax from the later-provider subset in this historical comparison. Keep old snapshot exact and all Box/image checks. |
| `^TestUsageAndCost$/^builtin-catalog-pinned$` | `usage_test.go:84–92` compares current complete catalog to old literal `testdata/responses/usage.json.catalog.{version,hash}`. It intentionally catches unversioned content change. | **Data pin renewal**, not dynamic expectations: extend existing keyless `renewal.py catalog-pins` to update only those two fields after actual activation. Preserve every source/prefix/scenario/model/expected cost/usage value. Source helper fix + controlled failure-first coverage happen before paid freeze; output data changes stay postactivation. |
| `^TestMiniMaxHost(Example\|Tools\|LocalAssembly\|LocalAndCloud\|Disable\|PurposeDenial)$` (see exact regex below) | `minimax_host_fixture_test.go:59–62` appends six synthetic candidate models to actual BuiltinCatalog. After activation the three M3 keys duplicate active built-in keys; NewClient rejects the fixture before feature assertions. | Controlled Host fixture deliberately replaces the MiniMax provider subset with its six synthetic candidate identities; retain other providers. Do not change runtime duplicate validation, use active prices as fake candidate rates, or alter actual activation. |

Exact Host regex (the table's backslashes are only Markdown readability):

`^TestMiniMaxHost(Example|Tools|LocalAssembly|LocalAndCloud|Disable|PurposeDenial)$`

`TestMiniMaxHostBindings` does not construct that fixture and should remain unchanged.

**Correction to my preliminary message:** `newMiniMaxSharedWorld` is safe. It appends candidates to `newMixedWorld` → `mixedCatalog()` → explicit synthetic `operationCatalog()`. The mixed fixture copies built-in **image/classifier** models only; it never copies built-in chat models. No Shared* fixture change is required for activation.

## Postactivation data pins and Chinese selectors

Existing `renewal.py catalog-pins` currently rewrites the release snapshot and only Chinese `{jev,luna}-route-v1.json`. Two additional current-global pin owners must participate in the same transactional keyless data operation:

- `ai/e2e/testdata/responses/usage.json`: nested `catalog.version` and `catalog.hash` only. Preserve the independent content assertion in `TestUsageAndCost/builtin-catalog-pinned`; reading the current catalog twice instead would remove its intended protection.
- `ai/examples/chineseeval/testdata/jev-1.13.0-v1.json`: `catalogVersion` and `catalogHash` only. This is the still-current legacy default evaluation configuration, not an immutable old evidence package. It has **timeoutSeconds180**, whereas route-v1 has600; preserve that difference and all instructions/dataset/24-call/attempt/question/zero-retry/15s-call fields. `chinesePlan()` and default CLI load it against current BuiltinCatalog; `plan.go:102` correctly rejects its stale full catalog pins.

Without that second data update, these tests fail before meaningful behavior: `TestChineseEvaluationFixedWorkflow`, `Failures`, `NotRunAndArtifactGuards`, `Command`, `PreCancelled`, `ImportedEvidence`, and `TestChineseEvaluationRoutesInputProcess/old-default`; `HistoricalProcess` also calls current `chinesePlan` while checking old evidence is inapplicable. InputGuards may still pass for the unrelated stale-catalog reason, so include positive workflows.

Minimal exact Chinese smoke after data pin renewal:

`^TestChineseEvaluation(FixedWorkflow|Failures|NotRunAndArtifactGuards|Command|PreCancelled|ImportedEvidence|HistoricalProcess|RoutesInputProcess)$`

Keep archived Chinese report/config/catalog bytes under historical `.scratch/.../chinese-evaluation-evidence` unchanged. Historical verification through `LoadRecorded` is intentionally distinct from current release applicability.

## Recommended activation clone selectors

Run in the implementer's isolated keyless activation clone, with its own evidence directory, after synthetic **data-only** activation plus current pin renewal. Do not run concurrently with its rollback/full source checks or mutate integration checkout. The audit did not run these commands.

1. Tight deterministic current catalog/fixture selector (Go regex, one command):

`go test -count=1 -timeout=10m ./ai/e2e -run '^(TestOpenAIImagesBuiltinCatalog|TestOpenAIImagesLiveWireReplay|TestGoogleImagesBuiltinCatalog|TestGoogleImagesLiveWireReplay|TestClassifierBuiltinOfficial|TestDecisionsFormalCatalog|TestJevFormalCatalog|TestUsageAndCost|TestMiniMaxHostExample|TestMiniMaxHostTools|TestMiniMaxHostLocalAssembly|TestMiniMaxHostLocalAndCloud|TestMiniMaxHostDisable|TestMiniMaxHostPurposeDenial|TestMiniMaxCandidateCatalog|TestGLMFormalCatalog)$'`

2. Chinese pin/behavior positives: use the exact Chinese regex above with `go test -count=1 -timeout=10m ./ai/e2e -run ...`.

3. Current release/control boundaries after pins (expected **test PASS**, while their deliberately incomplete/controlled release reports remain non-PASS):

`^(TestCurrentCatalogRelease|TestFourteenIdentityReleaseCLI|TestJevFormalEffectProcess|TestCurrentReleaseVisualEffects|TestCurrentReleaseVisualCorruption|TestMiniMaxAdditiveReleaseContract|TestMiniMaxPromotionBoundary)$`

4. MiniMax behavior smoke/full existing selector `^TestMiniMax` covers standalone candidate protocol worlds, Shared*, Host*, M3.1 exclusion, promotion boundary, frozen pi and rollback. Existing `TestMiniMaxCandidateCatalog:52–60` already supports either inactive or active M3 while requiring M3.1 absent; do not rewrite it to demand active M3 before real acceptance. `TestMiniMaxFullRollback` restores the fixed ecfd4b4 baseline in its own checkout and checks caller bytes unchanged; activation should not require changing that baseline.

5. Soleworker's corrected `^TestGLMFormalCaptureReplayProcess$` under `-tags live ./ai/live` must replay historical original catalog exactly, remain replay-only and produce no support certification. Existing decisions-process hardcoded global-version correction remains in scope (`ai/live/decisions_process_test.go:293`). Full no-key live must pass before postactivation paid renewal.

## Expected historical failures before GLM3 renewal

The known six current-applicability checks must remain failures until fresh GLM captures + raw/usage/view selectors exist:

- `^TestGLMFormalKnownUsage$/^(chat|responses|messages)$`: three original raw reports fail current execution source/catalog validation.
- `^TestGLMFormalReports$/^(chat|responses|messages)$`: three derived views/raw originals fail current source/catalog validity.

Do not skip, rehash, loosen source checks, replace report identities, or call the isolated activation clone a complete release PASS. After paid GLM3 renewal and data-only selector commits, those same six must PASS. `TestGLMSupportViewBounds` checks owned shape before stale consumer source and should still reject invalid shape for its stated reason. `TestGLMFormalExtraCaptures` remains a negative test; run it again after genuine renewal so source drift cannot be the accidental reason for rejection.

Actual old14 and four effects remain unrenewed after synthetic activation-clone testing; the real final release gate must still reject missing/stale current evidence. Release row counts17/route count16 are already correct and **do not change** after catalog activation. The frozen pi catalog test and `TestCatalogInclusion/builtin-models-are-pi's` already treat MiniMax separately; the latter explicitly removes GLM/MiniMax before frozen-model comparison. No total model-count literal needing revision was found in ./ai/e2e.

# MiniMax whole ordinary suite timing investigation

Read-only investigation of `/tmp/barness-minimax-10`, 2026-10-18. No source or repository artifact changes, suite reruns, vendor access, or credential reads. The existing command receipt refers to source `9481a6ad8028025a7f35da19a3e488935c6cd897`; checkout HEAD was `120530041de78c26e803719dd5c856d321d6aa34` during inspection.

## Finding

There is no evidence of an E03 or decimal-parser hang. The test body finished. The CPU-bound tail most likely belongs to the final redaction audit of a large evidence bundle. The cancellation itself does not establish a test failure, but the already written manifest independently records **52 actual failing cases**.

Evidence root: `.evidence/minimax-ticket10-prepay-final/whole-ordinary/20261017T235628.944008000Z/`.

- Receipt: command started `2026-10-17T23:56:25.448076Z`, cancelled after 536.7489 seconds with SIGTERM/-15 for the separately known harness fix.
- Manifest: `started_at=23:56:28.944008Z`, `finished_at=00:04:32.805397Z`; 12,356 cases, 12,304 PASS and 52 FAIL. It does not contain the `m.Run()` return code.
- Final assertion report was written at `00:04:23.221148Z`, for `E08-classifier-images-import-cloud-delete-batch.resources`; manifest file mtime was `00:04:33.028837Z`. No `audit.json` exists.
- The 558 `outcome.json` files are an optional artifact type, not a completion counter. There are **12,356 `assertions.json` reports**, 131,505 files total, and 1,291,614,660 bytes (~1.20 GiB).
- `ai/e2e/main_test.go:44–49` executes `m.Run()`, `run.Finish()`, then `auditPassed(run)`. `Run.Finish` writes the manifest after test cleanup. `audit.Bundle` sequentially reads every file, scans forbidden values and three regex rules, decodes and validates JSON, then recursively walks its fields (`ai/internal/testkit/audit/audit.go:83–152`). The manifest proves `m.Run()` returned; absent audit output places the remaining work at the audit boundary, subject to the small unprofiled interval immediately before it.
- Earlier recorded progress continued throughout the run: largest gap between completed assertion writes was 21.11 seconds (oracle retry/backoff), not an indefinite gap at E03. Pressure loads deliberately include large histories, images and concurrent output. The audit has no numeric `big.Rat` conversion.

## Actual E03 scenarios and numeric interactions

`TestThinkingSamplingValidation` (`ai/e2e/thinking_sampling_validation_test.go:15`) has 476 scenarios named `E03-thinking-sampling-validation-{protocol}-{level}-{key}`. Protocols are responses, chat, deepseek-responses, deepseek-chat, anthropic and gemini; levels are off/minimal/low/medium/high/xhigh/max. They test reserved keys, malformed JSON and empty JSON. Values are a fixed synthetic string or invalid/empty bytes, not huge numeric exponents. **All 476 PASS**. The last was `gemini-max-empty-json` at `00:03:10.516860Z`; the three `E03-thinking-sampling-non-object-*` assertions completed immediately afterward.

Other E03 families are `TestCatalogThinkingSampling` (config key/type validity), `TestThinkingSamplingBaseline` (retained source proofs/oracle guards), and `TestThinkingSamplingSnapshots` (snapshot identity). E03 labels also appear inside protocol history fixtures, so an unqualified E03 filename search is misleading.

- Shared `applySamplingParams` uses the existing JavaScript-compatible `parseJSON` → `strconv.ParseFloat` path (`ai/sampling.go:22`, `ai/tool_json.go:86,150`), not rational expansion. MiniMax then restores raw sampling bytes for exact final wire guards. This is provider-specific.
- At the actually executed commit, `miniMaxSamplingNumber` already uses sign/nonzero checks and the float fast path; `big.Rat.SetString` is reached only at the upper endpoint and after an exponent bound of `len(s)+16` (`ai/minimax_sampling_number.go:13–41`). Tiny exponents such as `1e-100000000000000000000` do not expand a rational.
- `outputBudgetNumber`, also used on GLM final output budgets, rejects invalid/out-of-range floats and bounds the exponent before rational expansion (`ai/output_budget.go:15–35`).
- The oracle comparator has an unbounded `big.Rat.SetString` in `canonicalNumber` (`ai/internal/testkit/pioracle/compare.go:275`), but that file is unchanged from base `ecfd4b4`. It is a potential pre-existing concern for adversarial numeric comparisons, not an explanation supported by this completed test manifest.
- The audit/evidence implementation, E03 validation source and oracle comparator have no changes from `ecfd4b4` to the executed `9481a6a`. MiniMax added meaningful suite work (9,838 MM-prefixed artifact files, ~85.3 MB); no quantified base timing comparison was run.

## The 52 FAIL entries are real Go failures

`Case.Check(false)` calls `t.Errorf` (`evidence.go:286`), and `Case.finish` reports FAIL only when `t.Failed()` is true (`:294–300`). Expected-negative fixtures should report PASS when rejection works. These 52 are distinct failing test names; they are not just parent statuses counted again.

| Test family | Cases | First recorded failure or setup boundary |
| --- | ---: | --- |
| TestGLMFormalReports | 3 | `own full current-hash report accepted`: `GLM consumer/importer source drift`. |
| TestGLMFormalKnownUsage | 3 | `captured formal source accepted`: `GLM execution source drift`. |
| TestReleaseRollbackArtifacts | 8 | No assertions/import output; common `releaseRollbackBundle` loads required `runtime-evidence/runtime.zip` before checks (`release_rollback_import_test.go:20–22`). That archive is absent in this checkout. |
| TestReleaseRevocationSemanticIntegrity | 28 | Same common helper/setup failure before mutation assertions. |
| TestJevBaselineEvidence | 4 | Intact: `fixed evidence verifies and preserves release failure` → `FAIL: baseline input missing or invalid`. Other three have null assertions and require copying `jev-baseline/baseline.zip`, which is absent. |
| TestDecisionsEvidence | 4 | Intact: `recorded evidence intact` → `FAIL: missing, changed or invalid Decisions text evidence`. Other three have null assertions and require copying absent `text-choice-evidence/text-choice.zip`. |
| TestDecisionsEvidenceSnapshot | 1 | Null assertions; setup copies the same absent text-choice ZIP. |
| TestJevFormalMatrixHistory | 1 | Three `other model history preserved exactly` checks fail for MiniMax Chat/Responses/Messages. Captured before/after rows differ only because `capabilities: []` is omitted on re-encoding and decodes as nil; `reflect.DeepEqual` distinguishes empty and nil slices. |

The missing ZIPs live under `.scratch/barness-ai-openai-decisions/`. The exact historical `t.Fatal` strings for null-assertion cases were not retained in their reports; the command log is empty because the interrupted `go test` process had not emitted buffered package output. Thus missing-archive attribution is strongly supported by current filesystem state and source order, not a claim to possess the missing stderr. Nine FAIL cases have explicit false assertions; the other 43 fail during setup.

## Recommendation

Do not treat E03 as stalled or spend another run on numeric-hang diagnosis. Before the fresh whole suite, restore/verify the three retained ZIP prerequisites, resolve GLM source-identity evidence drift honestly, and address the MiniMax empty-capability normalization mismatch. For progress use completed assertion/manifest status and the final audit receipt, rather than `outcome.json` count. Allow the full sequential audit tail under a bounded launcher; a written manifest alone is not a successful process exit or completed redaction audit. No passing-suite claim is justified by this cancelled run.

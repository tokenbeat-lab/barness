# Kimi final review fixes — ordinary evidence

Sole implementation branch `codex/kimi-review-fixes`; product/test commit `6757fb2`, merged integration `a4dd7df`, frozen and tested source **`0300710ad027a7f86fb8a55d4c81367c9da39620`**. No root Git mutation. All five hard review findings and the private naming observation addressed. Both read-only follow-up reviews closed every finding at the frozen source with no new actionable findings; final primary ordinary acceptance is owned by the orchestrator.

- Messages validates nonempty final tools for `any` and derives the terminal requirement from the actual post-hook payload. Text/empty ordinary `end_turn` fails; valid tool, length and hook-to-auto paths succeed. Native-history test now requests auto for its deliberate plaintext continuation.
- Narrowed Chat/Responses Host output caps become explicit defaults when omitted. Final hooks cannot remove or exceed the cap; the standard 131072 server-default omission remains valid.
- Chat validates the exact raw tool index before float conversion, with an inclusive safe-integer ceiling of 2^53−1; old Providers keep their existing parsing contract.
- Responses completed text and summary parts must agree with part/item/terminal snapshots, including presence and index boundaries.
- Qualification independently checks actual captured scenario effort/presence, budget and tool choice; compares saved-option controls/tools with captures; executes the captured native body through public final guards and verifies the no-dial transport sends those same bytes semantically. Resealed forged captures are rejected on all six routes.
- Private opaque-history file/function names now match their shared responsibility. Existing `glm_` generated wire spellings remain stable and explicitly carry no Provider attribution or replay authority.

## Red → green originals

Every command below ran through SOP `run.py --feature barness-ai-kimi-review-fixes --stage e2e`. Attempts remain independent and unmodified. The first ten completed attempts live in primary `.evidence/barness-ai-kimi-review-fixes/e2e/`.

| Attempt | Status / exit | Native argv |
| --- | --- | --- |
| `20261022T215824.899156Z-zhaxcj89` | FAIL / 1 | `go test -json ./ai/e2e -run ^TestKimiMessagesRequiredTools$ -count=1` |
| `20261022T215855.682558Z-usyfk3uk` | PASS / 0 | `go test -json ./ai/e2e -run ^TestKimiMessages(RequiredTools|ToolsAndNative)$ -count=1` |
| `20261022T215943.310726Z-pzx43gbc` | FAIL / 1 | `go test -json ./ai/e2e -run ^TestKimiNarrowOutputLimit$ -count=1` |
| `20261022T220006.770762Z-kidt5mbt` | PASS / 0 | `go test -json ./ai/e2e -run ^TestKimi(NarrowOutputLimit|ChatText|ResponsesText)$ -count=1` |
| `20261022T220021.979813Z-axsyngvd` | FAIL / 1 | `go test -json ./ai/e2e -run ^TestKimiChatExactToolIndex$ -count=1` |
| `20261022T220042.010771Z-rkpq8w01` | PASS / 0 | `go test -json ./ai/e2e -run ^TestKimiChat(ExactToolIndex|ToolsAndHistory|StreamTerminals)$ -count=1` |
| `20261022T220102.522352Z-lsgudkbw` | FAIL / 1 | `go test -json ./ai/e2e -run ^TestKimiResponsesCompletedTextConsistency$ -count=1` |
| `20261022T220125.195581Z-7p8pgfnc` | PASS / 0 | `go test -json ./ai/e2e -run ^TestKimiResponses(CompletedTextConsistency|PartConsistency|Terminals|Text)$ -count=1` |
| `20261022T220212.410728Z-kpe8q9bn` | FAIL / 1 | `go test -json ./ai/e2e -run ^TestKimiQualificationPipeline$ -count=1 -timeout=10m` |
| `20261022T220314.647843Z-_jn3ys49` | PASS / 0 | `go test -json ./ai/e2e -run ^TestKimiQualificationPipeline$ -count=1 -timeout=10m` |
| `20261022T220449.485533Z-asv4ohyx` | PASS / 0 | `go test -json ./ai/e2e ./ai ./ai/internal/testkit/supportmatrix -run ^Test(Kimi|Anthropic|Chat|Responses|GLM(Chat|Messages|Responses)|MiniMax(Chat|Messages|Responses)|RegionalProvider(Text|ToolsAndNativeHistory|PreflightAndFinalGuards|FailuresAndCancellation|ReleaseCapturePipeline)|CostEstimate$|ThinkingBudgets$|ImagesReport$|Merge|Classifier|ReportAndMatrix|CurrentCapabilitySet$) -count=1 -timeout=20m` |

The authoritative final primary copy is **`/Users/cyber/RestoX/barness/.evidence/barness-ai-kimi-review-fixes-final/e2e/20261022T220449.485533Z-asv4ohyx`**. Its single main bundle is `bundles/20261022T220451.333747000Z`; manifest SHA-256 `8e962120d0e646076d024c9152b80ae5a060b8219648261a2c8b62365c8f8329`, audit SHA-256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa` (findings empty). `review-fix-01.json` records exact source and receipt pointers.

Actual final inventory: E2E 5325 Go PASS records (includes parent tests), 178 SKIP records; main manifest 5117 PASS cases and 178 NOT_RUN cases. Required new groups are all PASS: Messages 64, output limits 48, indexes 44, completed text 36, captured controls 60. Auxiliary ai has 26 PASS records; supportmatrix has 79 PASS records, so neither auxiliary selection is empty. The broad Classifier auxiliary selector incidentally discovered the 178 pressure/import cases; normal gates skipped them. No pressure executed.

## Retention and source equality

All **38279 original files** compare byte-for-byte with authoritative primary destinations. Proof `/Users/cyber/RestoX/barness/.evidence/barness-ai-kimi-review-fixes/retention-20261022T-final.json`, SHA-256 `15a6e58663cb8148d741586e8ab8fd3ecf8adeefbd9bd7e5ec7c862f094a9981`. The full execution source contains 1346 files, matches the final receipt, and hashes as a sorted path/digest map to `b434a3ddeb2879b9aeb75727715b7b76c70dcf17020c4380c0bb9d64d6407bbc`; selected inputs are empty.

One retention preparation attempt stopped before overwrite because an incomplete interim copy of the active final attempt's Go telemetry cache counter differed after the process completed. That interim copy stays untouched at primary `.evidence/barness-ai-kimi-review-fixes/e2e/20261022T220449.485533Z-asv4ohyx/`; it is **not authoritative**. The new `barness-ai-kimi-review-fixes-final` destination preserves the completed originals exactly. This was a cache-copy timing failure, not a failing test or product reproduction. The implementation attempts themselves retain their original statuses.

Live, race and pressure remain NOT_RUN; candidate/formal paid qualification and issue 07's human gates remain pending. `release.pass=false`. No paid requests, race, pressure or historical archive suite ran. Rollback: revert implementation commit `6757fb2`; retain every receipt and capture. Final primary full ordinary acceptance remains separate evidence. Follow-up reports: `/tmp/barness-kimi-implementation-notes/final-standards-followup.md` and `/tmp/barness-kimi-implementation-notes/final-spec-followup.md`. Product source equality after the documentation commit is independently recorded at `/tmp/barness-kimi-implementation-notes/review-fix-delivery-equality.json`.

# Grok frozen ordinary required-case inventory

- Requested execution/test source: `5561d76dcd1a9ccee578835a10151ea1145e8d91`. Observed HEAD `5a04b34dbbedee1ee44c60d9a6a6534a577d6aa7` differs only by three review Markdown files; no selected source difference.
- **564 exact required cases = 514 XAI + 50 affected prior regression cases; 64 top-level tests =61 XAI +3 prior.**
- 66 families: 22 literals + 44 finite dynamic families.
- Derivation: `475 + (18+3+8+3+5+2) + (28+18+4) = 564`.
- Existing bounded-worker manifest:564 IDs /64 top-level tests, no missing, extra or duplicate selected IDs. This is read-only provenance, not a new/full ordinary execution receipt.
- No test/build/vendor/live/race/pressure run and no repository edits for this enumeration.
- Exact `requiredCases`, `expectedTests`, split `xaiCases` / `regressionCases`, per-test IDs, families, source hashes and mapping: `/tmp/barness-grok-freeze-cases.json`.

## Added source-derived families

| Function | Exact family | Count |
| --- | --- | ---: |
| `TestXAILiveHarnessProcess` | `E17-xai-process-{suffix}` | 29 |
| `TestXAIShorthandNativeAuthorization` | `XAI-review-shorthand-native-{suffix}` | 5 |
| `TestXAIDuplicateProtocolControls` | `XAI-review-duplicate-{suffix}` | 18 |
| `TestXAIDuplicateControlsAuthorityAndBusiness` | `XAI-review-duplicate-scope-{suffix}` | 3 |
| `TestXAICacheHeaderAuthority` | `XAI-review-cache-header-{suffix}` | 7 |
| `TestXAICacheHeaderAuthority` | `XAI-review-cache-header-other-provider` | 1 |
| `TestXAIHistoryDowngradesToOpenAIResponses` | `XAI-review-foreign-history-{suffix}` | 3 |
| `TestGLMResponsesOriginIDCollision` | `GLM05-origin-id-collision-{suffix}` | 28 |
| `TestKimiResponsesCrossTargetIDs` | `KIMI-{INTL|CN}-RESPONSES-K06-collision-TestKimiResponsesCrossTargetIDs/{market}/{source}/{destination}` | 18 |
| `TestKimiResponsesSequentialProvenanceIDs` | `KIMI-INTL-RESPONSES-K06-sequential-{suffix}` | 4 |

## Expected top-level tests

| Function | Source | Exact case count |
| --- | --- | ---: |
| `TestGLMResponsesOriginIDCollision` | `ai/e2e/glm_responses_ids_downgrade_test.go:13` | 28 |
| `TestKimiResponsesCrossTargetIDs` | `ai/e2e/kimi_responses_collision_test.go:13` | 18 |
| `TestKimiResponsesSequentialProvenanceIDs` | `ai/e2e/kimi_responses_collision_test.go:93` | 4 |
| `TestXAIAcceptanceCollidingOpaqueTuples` | `ai/e2e/xai_acceptance_test.go:153` | 1 |
| `TestXAIAcceptanceCrossAPIToolAssociations` | `ai/e2e/xai_acceptance_test.go:93` | 1 |
| `TestXAIAcceptanceExplicitEffortStateless` | `ai/e2e/xai_acceptance_test.go:194` | 1 |
| `TestXAIAcceptanceReturnedNativeToolRoundTrip` | `ai/e2e/xai_acceptance_test.go:33` | 2 |
| `TestXAIAcceptanceToolChoices` | `ai/e2e/xai_acceptance_test.go:11` | 2 |
| `TestXAIAcceptanceUserImageOrder` | `ai/e2e/xai_acceptance_test.go:127` | 1 |
| `TestXAICacheHeaderAuthority` | `ai/e2e/xai_header_guard_test.go:11` | 8 |
| `TestXAICacheProviderIsolation` | `ai/e2e/xai_cache_test.go:85` | 1 |
| `TestXAICacheScope` | `ai/e2e/xai_cache_test.go:13` | 1 |
| `TestXAICandidateSnapshot` | `ai/e2e/xai_foundation_test.go:101` | 1 |
| `TestXAICatalogBoundary` | `ai/e2e/xai_foundation_test.go:19` | 24 |
| `TestXAICompletedStatusShape` | `ai/e2e/xai_stream_structure_test.go:68` | 4 |
| `TestXAICostFailureSemantics` | `ai/e2e/xai_cost_failure_test.go:14` | 5 |
| `TestXAIDuplicateControlsAuthorityAndBusiness` | `ai/e2e/xai_duplicate_controls_test.go:65` | 3 |
| `TestXAIDuplicateProtocolControls` | `ai/e2e/xai_duplicate_controls_test.go:11` | 18 |
| `TestXAIFinalGuard` | `ai/e2e/xai_request_test.go:50` | 9 |
| `TestXAIHTTPFailures` | `ai/e2e/xai_cost_failure_test.go:50` | 7 |
| `TestXAIHeadersAndHooksOrder` | `ai/e2e/xai_cache_test.go:57` | 1 |
| `TestXAIHistoryDowngradesToOpenAIResponses` | `ai/e2e/xai_history_downgrade_test.go:12` | 3 |
| `TestXAIHistoryProjection` | `ai/e2e/xai_history_test.go:25` | 7 |
| `TestXAIHistoryRejectsBrokenAssociations` | `ai/e2e/xai_history_test.go:102` | 7 |
| `TestXAIHostConcurrentIsolation` | `ai/e2e/xai_host_isolation_test.go:13` | 1 |
| `TestXAIHostDisableAndSnapshots` | `ai/e2e/xai_host_disable_test.go:12` | 1 |
| `TestXAIHostNarrowsCandidate` | `ai/e2e/xai_foundation_test.go:123` | 1 |
| `TestXAIImageSizeBoundaries` | `ai/e2e/xai_images_test.go:119` | 5 |
| `TestXAIImagesAndToolOrder` | `ai/e2e/xai_images_test.go:31` | 1 |
| `TestXAIInputAndToolsSubset` | `ai/e2e/xai_payload_test.go:11` | 13 |
| `TestXAILiveHarnessProcess` | `ai/e2e/xai_live_process_test.go:30` | 30 |
| `TestXAILongContextCost` | `ai/e2e/xai_usage_test.go:87` | 4 |
| `TestXAINamedStreamEvent` | `ai/e2e/xai_stream_structure_test.go:58` | 1 |
| `TestXAINativeHookCannotChangeSource` | `ai/e2e/xai_payload_test.go:80` | 1 |
| `TestXAIOfficialTarget` | `ai/e2e/xai_route_test.go:77` | 4 |
| `TestXAIOptionsBeforeCredentials` | `ai/e2e/xai_options_test.go:13` | 14 |
| `TestXAIReasoningAndBudget` | `ai/e2e/xai_options_test.go:109` | 60 |
| `TestXAIReasoningAuthority` | `ai/e2e/xai_reasoning_authority_test.go:12` | 8 |
| `TestXAIReasoningPreservesExactJSON` | `ai/e2e/xai_history_test.go:175` | 1 |
| `TestXAIReasoningPreservesNullCiphertext` | `ai/e2e/xai_history_test.go:201` | 1 |
| `TestXAIRejectsInvalidImages` | `ai/e2e/xai_images_test.go:63` | 20 |
| `TestXAIRejectsOtherOperationCatalogs` | `ai/e2e/xai_foundation_test.go:84` | 2 |
| `TestXAIRequestContract` | `ai/e2e/xai_request_test.go:11` | 4 |
| `TestXAIRouteRejectedBeforeCredentials` | `ai/e2e/xai_route_test.go:12` | 19 |
| `TestXAIRuntimeBounds` | `ai/e2e/xai_runtime_resources_test.go:14` | 9 |
| `TestXAIRuntimeCallbackFailures` | `ai/e2e/xai_runtime_callbacks_test.go:14` | 3 |
| `TestXAIRuntimeCallbackPanicPropagation` | `ai/e2e/xai_runtime_callbacks_test.go:71` | 3 |
| `TestXAIRuntimeCancellation` | `ai/e2e/xai_runtime_resources_test.go:114` | 3 |
| `TestXAIRuntimeCompleteNoEventQueue` | `ai/e2e/xai_runtime_resources_test.go:102` | 1 |
| `TestXAIRuntimeHTTPFailures` | `ai/e2e/xai_runtime_retry_test.go:17` | 14 |
| `TestXAIRuntimeRedirectRefused` | `ai/e2e/xai_runtime_callbacks_test.go:55` | 1 |
| `TestXAIRuntimeRetryBounds` | `ai/e2e/xai_runtime_retry_test.go:116` | 2 |
| `TestXAIRuntimeRetrySnapshot` | `ai/e2e/xai_runtime_retry_test.go:63` | 1 |
| `TestXAISamplingSubset` | `ai/e2e/xai_options_test.go:45` | 88 |
| `TestXAIShorthandNativeAuthorization` | `ai/e2e/xai_payload_test.go:124` | 5 |
| `TestXAIStreamAuthoritativeSnapshots` | `ai/e2e/xai_stream_test.go:47` | 5 |
| `TestXAIStreamFaults` | `ai/e2e/xai_stream_faults_test.go:12` | 26 |
| `TestXAIStreamSnapshots` | `ai/e2e/xai_stream_test.go:20` | 2 |
| `TestXAIStreamStructure` | `ai/e2e/xai_stream_structure_test.go:12` | 10 |
| `TestXAIStreamTools` | `ai/e2e/xai_tools_test.go:21` | 10 |
| `TestXAIToolsCountBoundary` | `ai/e2e/xai_payload_test.go:102` | 2 |
| `TestXAITruncatedFunctionOutput` | `ai/e2e/xai_tools_test.go:83` | 1 |
| `TestXAIUsageAndCost` | `ai/e2e/xai_usage_test.go:12` | 27 |
| `TestXAIWholeOpaqueIDs` | `ai/e2e/xai_history_test.go:139` | 3 |

## Bounded worker provenance

- Manifest: `/Users/cyber/RestoX/barness/.evidence/barness-ai-grok/e2e/20261026T030721.568691Z-pzit2yuv/bundles/20261026T030725.014638000Z/manifest.json`.
- SHA-256: `f2ef5e24a68e5fb3332fe3b88a9cf2304fb68a962e197fe7d5575f1082ae1c85`.
- Source-derived expected-only:0; worker-only:0; duplicate IDs:0; expected test-only:0; worker test-only:0.

## Full ordinary comparison

- Full ordinary package manifest contains additional tests; require all64 expectedTests and564 exact requiredCases as a subset with no duplicate selected IDs.
- Expected sets split514 XAI and50 prior regression cases. XAI includes30 E17 process cases whose IDs do not start with XAI-.
- Select affected prior tests by exact function names; GLM/Kimi package prefixes also belong to unselected broader tests.
- Retain raw JSON punctuation, double slashes and full t.Name() portions; no ID normalization.
- Do not count child native/live bundles copied as fixtures as main ./ai/e2e IDs.
- Existing bounded worker manifest equality is provenance only; final full ordinary source/input binding and receipts remain separate.
- No source enumeration or controlled process PASS establishes real E17, race/pressure, formal support or release.pass.

## E01–E18 distinction

- Existing E01–E16 public controlled coverage and all prior five gap closures/explicit-effort assertion are retained. Review additions cover ambiguous protocol controls, cache header authority, shorthand native authorization, cross-target native downgrade and late report finalization.
- E17 main process coverage is30 cases with zero vendor HTTP; real `^TestXAICandidate$` remains a separate18-scenario/24-call source plan requiring explicit paid authorization and native qualification.
- E18 includes the50 exact selected GLM/Kimi collision/provenance regression cases. The final full affected package run and frozen-source/input receipt remain required; controlled proofs do not establish release.pass.

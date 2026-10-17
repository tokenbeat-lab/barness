# Responses red evidence

Selected original failing public Client runs, before their implementing slice. All inputs are synthetic official-schema cases; no live credential was accessed.

## options

Source SHA-256: `12ba5bec4ee69a0d990b27ff5efaaa8af2c683dea0465a98075d655d3779fdb9`

```text
--- FAIL: TestMiniMaxResponsesEntries (0.01s)
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3/stream (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/stream] default omits store: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"reasoning":{"effort":"none"}}
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/stream] default omits reasoning: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"reasoning":{"effort":"none"}}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3/complete (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/complete] default omits store: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"reasoning":{"effort":"none"}}
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/complete] default omits reasoning: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"reasoning":{"effort":"none"}}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3/stream-simple (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/stream-simple] default omits store: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"max_output_tokens":131072,"reasoning":{"effort":"none"}}
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/stream-simple] default omits reasoning: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"max_output_tokens":131072,"reasoning":{"effort":"none"}}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3/complete-simple (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/complete-simple] default omits store: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"max_output_tokens":131072,"reasoning":{"effort":"none"}}
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3/complete-simple] default omits reasoning: {"model":"MiniMax-M3","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"max_output_tokens":131072,"reasoning":{"effort":"none"}}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/stream (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/stream] default omits store: {"model":"MiniMax-M3.1-Flash-Preview","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/complete (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/complete] default omits store: {"model":"MiniMax-M3.1-Flash-Preview","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/stream-simple (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/stream-simple] default omits store: {"model":"MiniMax-M3.1-Flash-Preview","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"max_output_tokens":131072}
    --- FAIL: TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/complete-simple (0.00s)
        minimax_responses_options_test.go:38: [MM03-entries-TestMiniMaxResponsesEntries/MiniMax-M3.1-Flash-Preview/complete-simple] default omits store: {"model":"MiniMax-M3.1-Flash-Preview","input":[{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}],"stream":true,"store":false,"max_output_tokens":131072}
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221201.608534000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.640s
FAIL
```

## content

Source SHA-256: `7ceae65a47e861d50dbd70582cbae2ff8ae6d4d6d69e753276346ded614ea999`

```text
--- FAIL: TestMiniMaxResponsesAuthoritativeContent (0.02s)
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/terminal-only (0.00s)
        minimax_responses_content_test.go:50: [MM03-content-terminal-only] authoritative content retained once: thinking="" text=""
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/done-only (0.00s)
        minimax_responses_content_test.go:50: [MM03-content-done-only] authoritative content retained once: thinking="" text=""
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/done-corrects-delta (0.00s)
        minimax_responses_content_test.go:50: [MM03-content-done-corrects-delta] authoritative content retained once: thinking="" text=""
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/reasoning-length (0.00s)
        minimax_responses_content_test.go:50: [MM03-content-reasoning-length] authoritative content retained once: thinking="" text=""
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/object-content (0.00s)
        responses_transport_test.go:121: [MM03-content-object-content] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-content-object-content] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-content-object-content] message carries errorMessage: empty
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/duplicate-terminal (0.00s)
        responses_transport_test.go:121: [MM03-content-duplicate-terminal] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-content-duplicate-terminal] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-content-duplicate-terminal] message carries errorMessage: empty
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/missing-slot-delta (0.00s)
        responses_transport_test.go:121: [MM03-content-missing-slot-delta] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-content-missing-slot-delta] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-content-missing-slot-delta] message carries errorMessage: empty
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/wrong-delta-kind (0.00s)
        responses_transport_test.go:121: [MM03-content-wrong-delta-kind] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-content-wrong-delta-kind] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-content-wrong-delta-kind] message carries errorMessage: empty
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/invalid-terminal-status (0.00s)
        responses_transport_test.go:121: [MM03-content-invalid-terminal-status] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-content-invalid-terminal-status] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-content-invalid-terminal-status] message carries errorMessage: empty
    --- FAIL: TestMiniMaxResponsesAuthoritativeContent/after-terminal (0.00s)
        responses_transport_test.go:121: [MM03-content-after-terminal] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-content-after-terminal] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-content-after-terminal] message carries errorMessage: empty
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221328.408483000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.641s
FAIL
```

## history

Source SHA-256: `cc0ec0adfdb618fdb12e110689d546e28421c6587a2c6a7474dd8f565ffba9cf`

```text
--- FAIL: TestMiniMaxResponsesNativeRoundTrip (0.02s)
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/trusted/collision (0.00s)
        minimax_responses_history_test.go:72: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/trusted/collision] composite IDs do not collide: [same|call%7C|item% same|call%7C|item%]
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/trusted/collision] synthetic schema history continues: ai: invalid_request (request): MiniMax input reasoning requires summary parts
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/trusted/unicode-long (0.00s)
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/trusted/unicode-long] synthetic schema history continues: ai: invalid_request (request): MiniMax input reasoning requires summary parts
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/restored/collision (0.00s)
        minimax_responses_history_test.go:72: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/restored/collision] composite IDs do not collide: [same|call%7C|item% same|call%7C|item%]
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/restored/collision] synthetic schema history continues: ai: invalid_request (request): MiniMax input reasoning requires summary parts
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/restored/unicode-long (0.00s)
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/restored/unicode-long] synthetic schema history continues: ai: invalid_request (request): MiniMax input reasoning requires summary parts
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/untrusted/collision (0.00s)
        minimax_responses_history_test.go:72: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/untrusted/collision] composite IDs do not collide: [same|call%7C|item% same|call%7C|item%]
        minimax_responses_history_test.go:95: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/untrusted/collision] exact call IDs pair without collision: calls=[same_call_7C_item same_call_7C_item] results=[same_call_7C_item same_call_7C_item]
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/untrusted/unicode-long (0.00s)
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/untrusted/unicode-long] synthetic schema history continues: ai: invalid_request (request): MiniMax function calls require name, call_id and arguments
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/cross-model/collision (0.00s)
        minimax_responses_history_test.go:72: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/cross-model/collision] composite IDs do not collide: [same|call%7C|item% same|call%7C|item%]
        minimax_responses_history_test.go:95: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/cross-model/collision] exact call IDs pair without collision: calls=[same_call_7C_item same_call_7C_item] results=[same_call_7C_item same_call_7C_item]
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/cross-model/unicode-long (0.00s)
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/cross-model/unicode-long] synthetic schema history continues: ai: invalid_request (request): MiniMax function calls require name, call_id and arguments
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/cross-account/collision (0.00s)
        minimax_responses_history_test.go:72: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/cross-account/collision] composite IDs do not collide: [same|call%7C|item% same|call%7C|item%]
        minimax_responses_history_test.go:95: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/cross-account/collision] exact call IDs pair without collision: calls=[same_call_7C_item same_call_7C_item] results=[same_call_7C_item same_call_7C_item]
    --- FAIL: TestMiniMaxResponsesNativeRoundTrip/cross-account/unicode-long (0.00s)
        minimax_responses_history_test.go:76: [MM03-native-TestMiniMaxResponsesNativeRoundTrip/cross-account/unicode-long] synthetic schema history continues: ai: invalid_request (request): MiniMax function calls require name, call_id and arguments
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221434.966005000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.641s
FAIL
```

## order

Source SHA-256: `95cc218c4ab0e99885b0b8a6853dcd30da5211279cd6c5b4ed3ed4b642b7ea86`

```text
--- FAIL: TestMiniMaxResponsesPartsAndOrder (0.01s)
    --- FAIL: TestMiniMaxResponsesPartsAndOrder/multiple-done-parts (0.00s)
        minimax_responses_order_test.go:32: [MM03-order-multiple-done-parts] associated parts complete exactly: <nil> {Content:[{Text:hello  Signature:}] API:openai-responses Provider:minimax-cn Model:MiniMax-M3 ResponseID:r-minimax ResponseModel: ProviderThinkingLevel: Usage:{Input:3 Output:2 CacheRead:0 CacheWrite:0 CacheWrite1h:{state:0 value:0} Reasoning:{state:2 value:0} Modalities:<nil> TotalTokens:5 Cost:{Input:3e-06 Output:4e-06 CacheRead:0 CacheWrite:0 Total:7e-06}} StopReason:stop RawStopReason:completed ErrorMessage: Timestamp:1792275353843 DurationMs:{state:2 value:1} NativeState:{env:{TenantID:tenant-a AccountScopeID:acct-minimax-responses ProviderID:minimax-cn API:openai-responses ModelID:MiniMax-M3}}}
    --- FAIL: TestMiniMaxResponsesPartsAndOrder/delta-after-other-part-done (0.00s)
        minimax_responses_order_test.go:32: [MM03-order-delta-after-other-part-done] associated parts complete exactly: ai: protocol (stream): MiniMax Responses event violates the protocol {Content:[{Text: Signature:}] API:openai-responses Provider:minimax-cn Model:MiniMax-M3 ResponseID: ResponseModel: ProviderThinkingLevel: Usage:{Input:0 Output:0 CacheRead:0 CacheWrite:0 CacheWrite1h:{state:0 value:0} Reasoning:{state:0 value:0} Modalities:<nil> TotalTokens:0 Cost:{Input:0 Output:0 CacheRead:0 CacheWrite:0 Total:0}} StopReason:error RawStopReason: ErrorMessage:MiniMax Responses event violates the protocol Timestamp:1792275353845 DurationMs:{state:2 value:1} NativeState:{env:{TenantID:tenant-a AccountScopeID:acct-minimax-responses ProviderID:minimax-cn API:openai-responses ModelID:MiniMax-M3}}}
    --- FAIL: TestMiniMaxResponsesPartsAndOrder/duplicate-part-done (0.00s)
        responses_transport_test.go:121: [MM03-order-duplicate-part-done] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-order-duplicate-part-done] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-order-duplicate-part-done] message carries errorMessage: empty
    --- FAIL: TestMiniMaxResponsesPartsAndOrder/arguments-done-unassociated (0.00s)
        responses_transport_test.go:121: [MM03-order-arguments-done-unassociated] error is a classified *ai.Error: err=<nil>
        responses_transport_test.go:122: [MM03-order-arguments-done-unassociated] message ends with StopReason error: got "stop"
        responses_transport_test.go:123: [MM03-order-arguments-done-unassociated] message carries errorMessage: empty
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221553.839737000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.606s
FAIL
```

## input

Source SHA-256: `ad58113ddc30f91ce91a6de88811678723de2c98795fb7463ad6c8a20de47dbf`

```text
--- FAIL: TestMiniMaxResponsesInputAndToolBoundary (0.01s)
    --- FAIL: TestMiniMaxResponsesInputAndToolBoundary/invalid-call-id-type (0.00s)
        preflight_reject_test.go:223: [MM03-input-boundary-invalid-call-id-type] errors.Is matches code and phase: want invalid_request/request, got <nil>
        preflight_reject_test.go:236: [MM03-input-boundary-invalid-call-id-type] stream ends with only an error terminal: got ["start" "done:stop"]
        minimax_responses_guards_test.go:51: [MM03-input-boundary-invalid-call-id-type] zero attempts: {CallAttribution:{TenantID:tenant-a RequestID:TestMiniMaxResponsesInputAndToolBoundary/invalid-call-id-type ActorID:actor-7 JobID:job-42 BindingID:minimax-responses Operation:chat Resolved:true ProviderID:minimax-cn API:openai-responses ModelID:MiniMax-M3 AccountScopeID:acct-minimax-responses} BindingVersion:minimax-candidate-1 CredentialVersion:v1 CatalogVersion:minimax-six-candidate-2026-10-18.1-synthetic-prices CatalogHash:sha256:cd92963a43d390f1fa42cda9cef360486755a3ded6f6a1ebf49442a85f31e1cb NativeStateDowngrades:{NoEnvelope:0 AccountMismatch:0 CrossModel:0} Attempts:[{AttemptID:TestMiniMaxResponsesInputAndToolBoundary/invalid-call-id-type#1 HTTPStatus:200 ProviderRequestID: Code: RetryDelay:0s UsageReporting:partial Usage:{Input:3 Output:2 CacheRead:0 CacheWrite:0 CacheWrite1h:{state:0 value:0} Reasoning:{state:2 value:0} Modalities:<nil> TotalTokens:5 Cost:{Input:3e-06 Output:4e-06 CacheRead:0 CacheWrite:0 Total:7e-06}}}]}
    --- FAIL: TestMiniMaxResponsesInputAndToolBoundary/invalid-argument-type (0.00s)
        preflight_reject_test.go:223: [MM03-input-boundary-invalid-argument-type] errors.Is matches code and phase: want invalid_request/request, got <nil>
        preflight_reject_test.go:236: [MM03-input-boundary-invalid-argument-type] stream ends with only an error terminal: got ["start" "done:stop"]
        minimax_responses_guards_test.go:51: [MM03-input-boundary-invalid-argument-type] zero attempts: {CallAttribution:{TenantID:tenant-a RequestID:TestMiniMaxResponsesInputAndToolBoundary/invalid-argument-type ActorID:actor-7 JobID:job-42 BindingID:minimax-responses Operation:chat Resolved:true ProviderID:minimax-cn API:openai-responses ModelID:MiniMax-M3 AccountScopeID:acct-minimax-responses} BindingVersion:minimax-candidate-1 CredentialVersion:v1 CatalogVersion:minimax-six-candidate-2026-10-18.1-synthetic-prices CatalogHash:sha256:cd92963a43d390f1fa42cda9cef360486755a3ded6f6a1ebf49442a85f31e1cb NativeStateDowngrades:{NoEnvelope:0 AccountMismatch:0 CrossModel:0} Attempts:[{AttemptID:TestMiniMaxResponsesInputAndToolBoundary/invalid-argument-type#1 HTTPStatus:200 ProviderRequestID: Code: RetryDelay:0s UsageReporting:partial Usage:{Input:3 Output:2 CacheRead:0 CacheWrite:0 CacheWrite1h:{state:0 value:0} Reasoning:{state:2 value:0} Modalities:<nil> TotalTokens:5 Cost:{Input:3e-06 Output:4e-06 CacheRead:0 CacheWrite:0 Total:7e-06}}}]}
    --- FAIL: TestMiniMaxResponsesInputAndToolBoundary/input-extra-state (0.00s)
        preflight_reject_test.go:223: [MM03-input-boundary-input-extra-state] errors.Is matches code and phase: want invalid_request/request, got <nil>
        preflight_reject_test.go:236: [MM03-input-boundary-input-extra-state] stream ends with only an error terminal: got ["start" "done:stop"]
        minimax_responses_guards_test.go:51: [MM03-input-boundary-input-extra-state] zero attempts: {CallAttribution:{TenantID:tenant-a RequestID:TestMiniMaxResponsesInputAndToolBoundary/input-extra-state ActorID:actor-7 JobID:job-42 BindingID:minimax-responses Operation:chat Resolved:true ProviderID:minimax-cn API:openai-responses ModelID:MiniMax-M3 AccountScopeID:acct-minimax-responses} BindingVersion:minimax-candidate-1 CredentialVersion:v1 CatalogVersion:minimax-six-candidate-2026-10-18.1-synthetic-prices CatalogHash:sha256:cd92963a43d390f1fa42cda9cef360486755a3ded6f6a1ebf49442a85f31e1cb NativeStateDowngrades:{NoEnvelope:0 AccountMismatch:0 CrossModel:0} Attempts:[{AttemptID:TestMiniMaxResponsesInputAndToolBoundary/input-extra-state#1 HTTPStatus:200 ProviderRequestID: Code: RetryDelay:0s UsageReporting:partial Usage:{Input:3 Output:2 CacheRead:0 CacheWrite:0 CacheWrite1h:{state:0 value:0} Reasoning:{state:2 value:0} Modalities:<nil> TotalTokens:5 Cost:{Input:3e-06 Output:4e-06 CacheRead:0 CacheWrite:0 Total:7e-06}}}]}
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221717.759362000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.807s
FAIL
```

## output-image

Source SHA-256: `957aa1df57d51de72030b26b00f4eeeac64b5d4144bf087aed97cc5107503c4c`

```text
--- FAIL: TestMiniMaxResponsesInputAndToolBoundary (0.00s)
    --- FAIL: TestMiniMaxResponsesInputAndToolBoundary/tool-result-image (0.00s)
        preflight_reject_test.go:223: [MM03-input-boundary-tool-result-image] errors.Is matches code and phase: want invalid_request/request, got <nil>
        preflight_reject_test.go:236: [MM03-input-boundary-tool-result-image] stream ends with only an error terminal: got ["start" "done:stop"]
        minimax_responses_guards_test.go:54: [MM03-input-boundary-tool-result-image] zero attempts: {CallAttribution:{TenantID:tenant-a RequestID:TestMiniMaxResponsesInputAndToolBoundary/tool-result-image ActorID:actor-7 JobID:job-42 BindingID:minimax-responses Operation:chat Resolved:true ProviderID:minimax-cn API:openai-responses ModelID:MiniMax-M3 AccountScopeID:acct-minimax-responses} BindingVersion:minimax-candidate-1 CredentialVersion:v1 CatalogVersion:minimax-six-candidate-2026-10-18.1-synthetic-prices CatalogHash:sha256:cd92963a43d390f1fa42cda9cef360486755a3ded6f6a1ebf49442a85f31e1cb NativeStateDowngrades:{NoEnvelope:0 AccountMismatch:0 CrossModel:0} Attempts:[{AttemptID:TestMiniMaxResponsesInputAndToolBoundary/tool-result-image#1 HTTPStatus:200 ProviderRequestID: Code: RetryDelay:0s UsageReporting:partial Usage:{Input:3 Output:2 CacheRead:0 CacheWrite:0 CacheWrite1h:{state:0 value:0} Reasoning:{state:2 value:0} Modalities:<nil> TotalTokens:5 Cost:{Input:3e-06 Output:4e-06 CacheRead:0 CacheWrite:0 Total:7e-06}}}]}
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221738.505063000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.596s
FAIL
```

## reasoning-history

Source SHA-256: `b807affdfd0e1f95d6b823cc7e051d9a63847708d66e613b0f64c4f8272ddba4`

```text
--- FAIL: TestMiniMaxResponsesReasoningHistoryWithoutContent (0.01s)
    --- FAIL: TestMiniMaxResponsesReasoningHistoryWithoutContent/item-done (0.00s)
        minimax_responses_reasoning_history_test.go:47: [MM03-reasoning-history-item-done] no available reasoning is silently erased: [map[content:[map[text:Say hello. type:input_text]] role:user] map[summary:[] type:reasoning] map[content:[map[text:Continue. type:input_text]] role:user]]
    --- FAIL: TestMiniMaxResponsesReasoningHistoryWithoutContent/text-done (0.00s)
        minimax_responses_reasoning_history_test.go:47: [MM03-reasoning-history-text-done] no available reasoning is silently erased: [map[content:[map[text:Say hello. type:input_text]] role:user] map[content:[map[text:Continue. type:input_text]] role:user]]
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221822.754171000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.607s
FAIL
```

## native-raw

Source SHA-256: `dacd2198a61961ef5238b7c2d4a5a83b4bf4ade4d2bfd5aaeaeeadef691ec5a2`

```text
--- FAIL: TestMiniMaxResponsesRetainsOpaqueReasoningOutput (0.00s)
    minimax_responses_content_test.go:72: [MM03-raw-reasoning-output] original raw bytes retained before input projection: {"type":"reasoning","id":"rs","content":[{"type":"reasoning_text","text":"exact"}],"capture_marker":1}
FAIL
barness-ai evidence bundle: /private/tmp/barness-minimax-03/.evidence/barness-ai/20261017T221952.605094000Z
FAIL	github.com/tokenbeat-lab/barness/ai/e2e	0.656s
FAIL
```

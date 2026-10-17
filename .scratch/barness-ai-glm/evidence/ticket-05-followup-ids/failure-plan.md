# Scoped ticket 05 follow-up: opaque GLM function IDs

Written before implementation. Public seam: existing Client/Host controlled TLS Provider, as confirmed in spec T01. No private helper unit tests or paid calls.

Source: official Responses schema snapshot SHA256 `0914af477ff4c1fae0372f9c44ace23033d0bed0cc486796eaff2c2024e5f18c`, InputFunctionCall.call_id and OutputFunctionCall.call_id/id are plain strings; no alphabet excludes `|`, percent or Unicode. This is documented admissibility, not an observed live edge-ID claim.

- Both models return two native call IDs `same|a` / `same|b`: public tool IDs must stay distinct, Host validation/results must retain each pairing, and trusted replay must send the exact original call_id instead of truncating both to `same`.
- Percent-looking sequences, Unicode and delimiter combinations in both call_id and item id must remain distinct and round-trip without double decoding. Existing ordinary captured IDs must keep their public spelling and wire semantics.
- Host-restored, untrusted, cross-model and cross-account own Responses histories must retain exact function linkage while independently applying native-state downgrade rules.
- Cross-origin Chat histories whose IDs contain the delimiter must not collide when targeting GLM Responses. This proves only the GLM target normalization contract; OpenAI/DeepSeek target semantics remain frozen.
- Final wire must omit output item IDs, preserve both result texts under the correct calls, and contain no state references. Run relevant Responses/native/cross-route regression, race and frozen native Responses differential.

Parent-requested additional slice before final freeze: real GLM Responses-origin IDs can collide when another target replaces punctuation or truncates to 64 characters. Prove it through public two-model output→Host results→GLM Chat/Messages and native OpenAI/DeepSeek Responses/Chat/Messages targets. Scope any safeguard to colliding GLM-origin IDs, preserving ordinary existing GLM conversions and every history without GLM. Verify distinct wire calls and each result's pairing; never reuse one result for two calls.

Final parent-requested boundary: a single GLM Chat-origin Unicode call ID can become empty under legacy Responses normalization (punctuation replacement then trailing-underscore trimming). Reproduce through the real public Chat parser and Host tool result, targeting OpenAI and DeepSeek Responses in both models. If real, reserve a valid short ID for empty GLM-origin wire IDs even with no second call; histories without GLM keep their original policy.

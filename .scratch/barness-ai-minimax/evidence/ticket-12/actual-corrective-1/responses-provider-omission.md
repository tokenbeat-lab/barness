# GLM corrective-1 Responses native-history failure (read-only diagnosis)

Source frozen at `5bf2c16f2207a127723e4b19bfa0b59d23825cd4`. No test, process
replay, credential/environment read, network request, vendor retry, source edit
or capture mutation was performed for this diagnosis. Only existing artifacts
and current source were read. This note contains schema/length metadata, no
response IDs, credential/header values, full prompt/reasoning or raw body dump.

## Finding

`glm-5.2/native-history` is a genuine required-visible-text FAIL. The provider's
complete raw SSE contains no visible-text message or output_text event. The
adapter retains all five reasoning characters in the public Result. There is
no captured visible text for the adapter to have dropped. This is different
from the original capture-size failure: the complete response is only3,220
bytes, the recorder exchange body equals response.sse byte-for-byte, HTTP200,
elapsed583ms, no truncation, and the terminal is completed rather than length.

This observation establishes what was returned, not why the provider returned
reasoning-only output. It does not establish that a repeat would succeed or
authorize a repeat. The current formal report stays FAIL/complete=false,
auditPassed=true;22 calls/22 HTTP attempts, zero retries. Twenty-one scenarios
PASS; this is the sole failed case. The complete=false field is overall formal
acceptance failure here, not an incomplete/truncated raw response.

## Exact observed response shape

Artifact root: `$COMMON/.evidence/minimax-global-glm-corrective-1-reviewed/`
`glm-corrective-1-responses/20261018T042202.466861000Z/`. Failed scene directory:
`GLM11-glm-5.2-native-history/`.

The12 raw events in order are:

1. response.created
2. response.in_progress
3. response.output_item.added
4. response.content_part.added
5–8. response.reasoning_text.delta (lengths2,1,1,1 characters)
9. response.reasoning_text.done
10. response.content_part.done
11. response.output_item.done
12. response.completed

Created/in_progress output arrays are empty. The single output item is a
reasoning item. Its final keys are content, encrypted_content, id, status,
summary, type; type=reasoning, status=completed, summary=[],
encrypted_content=null, content=[{type:reasoning_text,text:<5 characters>}].
No output item has type=message/function_call; no output_text/refusal event or
part exists. Terminal response.output repeats exactly that one reasoning item.
Terminal model=glm-5.2, status=completed, error=null, incomplete_details=null.
Terminal response keys are conversation_id, created_at, error, id,
incomplete_details, instructions, metadata, model, object, output,
parallel_tool_calls, previous_response_id, status, usage.

Late terminal usage is input_tokens63, cached_tokens0, output_tokens2,
reasoning_tokens2, total_tokens65. These are retained known usage, not an
assertion of zero billing. Public Result: error empty, stopReason=stop,
rawStopReason=completed, one Thinking part of5 characters, no Text part, and
original native reasoning signature retained (178 characters). Public usage
retains63/2/2. The failed assertion is exactly
`required text entry produces final text` (pass=false, detail=no final text).
The own InputReasoning object replay assertion passes.

## Native-history request and controls

Final request top-level keys: input, max_output_tokens, model, reasoning,
store, stream. Values: model=glm-5.2, max_output_tokens4096,
reasoning={effort:high}, store=false, stream=true. No extra flags, unsupported
encrypted-content replay, server conversation dependency or hook workaround.

Input order and exact shapes:

1. message/user, content=[{type:input_text,text:<120 characters>}].
2. reasoning, content={type:reasoning_text,text:<992 characters>} (a single
   object, not an array; no output id/status/summary/encrypted fields).
3. message/assistant, content=[{type:output_text,text:<3 characters>}].
4. message/user, content=[{type:input_text,text:<40 characters>}].

The992-character reasoning equals this model's own successful full-stream
public Thinking byte-for-byte. The3-character assistant answer also equals
its own baseline public Text byte-for-byte. The native-history scene therefore
does not borrow another model's history or lose baseline text. Current
`ai/responses_glm_history.go:10–62` implements the independently documented
GLM InputReasoning single-object boundary, preserving the original output
signature. The frozen official excerpt is
`.scratch/barness-ai-glm/evidence/ticket-10/sources/responses-contract.json`
(InputReasoning); its provenance/research explanation is in
`.scratch/barness-ai-glm/research.md:90–100`.

The harness builds history from this model's trusted own full-stream result
(`ai/live/glm_harness_test.go:215–250`) and requires visible text for native
continuation (`:397–404`), independently of successful HTTP/terminal status.
The adapter validates/retains GLM original output content
(`ai/responses_glm_decode.go:10–129`) and finishes only message/reasoning parts
actually present (`ai/responses_glm_lifecycle.go:36–73`). No evidence supports a
request encoding or parser loss fix for this captured failure.

## Same-round comparisons

| Scene | Raw terminal output | Known usage input/output/reasoning | Result |
| --- | --- | --- | --- |
| glm-5.2/full-stream | reasoning992 chars + assistant output_text3 chars; completed/error null |49/520/518|PASS, visible3 chars retained|
| glm-5.2/native-history | reasoning5 chars only; completed/error null |63/2/2|FAIL required visible text|
| glm-5.3/native-history | reasoning4 chars + assistant output_text3 chars; completed/error null |63/7/3|PASS, visible3 chars retained|

All three use max_output_tokens4096, reasoning.effort=high, store=false,
stream=true. The5.3 history request has the same four-item input schema
(reasoning length435 instead of992). Its native-history output has the expected
second message item and output_text delta/done, which the same adapter retains.
This supports a provider-return omission localized to this5.2 continuation;
it is not proof about a service-wide guarantee or future outcomes.

## Integrity and disposition

The bundle audit has zero findings. Original report fields/source and raw
bytes are unchanged. Metadata-only hashes:

| Artifact | Bytes | SHA256 |
| --- | ---: | --- |
|formal-glm-report.json|27868|a42b5b12fa6b37bacf1f1efba5e4c9ae58eb373f67b5d9c5e684d5c45bc2b1e2|
|manifest.json|48422|cc4476b166af8f6d23a6c806597432701be6af5282f71de6a291a510fa77e09b|
|failed scene response.sse|3220|7278640e3dbbf3eb8b0e96325de4f67e5caf21be386491f2f8eda4f9f65222a4|
|failed scene request.json|1657|2c9eba367f44ff69429d508b3fb2a848199309c89e0b23c966b37e54c9634a8b|
|failed scene attempt-1-result.json|2270|c63199ab0ec5b94a62216c3b3dbe51c03ffc2328b781d7790b816f9945cf8403|

No recovery is performed or proposed as an existing entitlement: corrective-1
is the distinct fixed full22 round, with no automatic R2 or scene splice. The
remaining nominal2-call headroom is not a composable formal-report contract.
Preserve the FAIL and all known usage/receipts, do not relabel reasoning as
visible text, and do not rewrite source/catalog/report pins. Any future paid
recovery would need a separately reviewed concrete contract and explicit
additional authorization; there is no evidence-backed code change to make now.

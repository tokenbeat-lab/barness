# Deidentified Grok 4.7 Responses captures

These are all 24 successful responses from the explicitly authorized candidate
run `20261026T050451.200594000Z`: 18 scenarios, no automatic retries. The original
paid report has `complete=true`, `formalQualified=false`. This directory grants
no formal support or fresh paid permission.

`manifest.json` records the original report/manifest/audit and each source HTTP
capture SHA-256, the sanitized file digests, native source identity and the exact
transformation. The original account-bound material stays in the ignored source
bundle at the repository-relative `sourceBundle` path. Header values, credentials,
cookies and trace identifiers were not copied here.

Opaque response/item/call IDs and encrypted strings use deterministic stand-ins
shared across all captures. Thus a prior public Result can be persisted, restored
and passed through `TrustNativeState`; the resulting continuation must reproduce
the deidentified native item and call/result associations. The stand-ins are not
valid encrypted state for a vendor call. Scoped cache keys use a placeholder;
tests compare the newly derived key with a fixed ADR-0049 test vector for the
synthetic replay scope and check repeat stability. The request comparison only
replaces the captured placeholder with that vector; it drops no actual fields.

Synthetic PNG/JPEG bytes, protocol/event order, visible text, function schemas
and arguments, and raw token counts remain unchanged. JSON is serialized anew;
event/data framing remains ordered, SSE comments are omitted, and original
network chunk boundaries are not claimed. Replays inject complete captured SSE
through the existing loopback HTTP Provider, using the public Client entries.

Expected token counts are literals derived from these observed inclusive reports:
`total=input+output`, cached input is split from input, and reasoning is already a
subset of output. All 24 source responses reported nonzero reasoning. This is an
observation of this run, not a general promise about future service behavior.

The retained generator is
`.evidence/barness-ai-grok/replay-assembly-v1/sanitize.py`; its digest is in the
manifest. Reproduce into a fresh directory with `--bundle <original bundle>` and
`--out <new directory>`, then compare all JSON digests. It performs no network or
credential access. Public E2E cases are `TestXAIRealCaptureProvenance` and
`TestXAIRealCaptureReplay`.

---
status: accepted
date: 2026-10-17
---

# GLM Messages preserves provider omission and uses its own wire evidence

Zhipu shares the Messages adapter and its explicit SDK service, SSE decoder,
resource controls and lifecycle, but receives an empty automatic-feature profile
under ADR-0032. Both models' own public Client wire captures accepted basic
Messages at the China endpoint, including the locked SDK's `?beta=true` query
and version header. That query is an SDK routing fact, not beta-feature support.
Native Anthropic keeps its existing defaults and request-ID header; GLM uses
its actually reported X-Request-Id.

Simple omitted reasoning must remain omitted for GLM, because generating
Claude's `thinking.disabled` changes the provider default. The actual 5.3 basic
sample returned signed thinking, while 5.2 returned text only; another 5.2
complex same-prompt control did likewise. No Chat/Responses effort map is
copied. Explicit unverified thinking controls are rejected before payload
callbacks or attempts, and final callback payload/header validation prevents
unsupported extensions from reappearing. This favors a visible limited
contract over inferred Claude features.

A separate Messages Host candidate and unchanged raw SSE fixtures retain
scope and provenance. Existing usage replacement already handles final
message_delta counts; absent cache_creation remains partial. Basic text
replay cannot register built-in models, certify tools/native replay, change the
support matrix or count as formal live acceptance. Evidence and repeatable
commands are indexed in the [ticket 06 record](../../.scratch/barness-ai-glm/evidence/ticket-06/README.md).

Ticket 07 extends only own observed shapes. Both independent models accept
function definitions, omission/auto tool selection, fragmented tool arguments
and Host tool-result continuation. GLM-5.3 returned signed thinking by default;
GLM-5.2 returned it for the exact enabled/budget1024/max4096 request. A scalar
`MessagesThinkingBudgetTokens` ModelCompat constraint opts a model into one
verified budget value; zero opts into none. The provider implementation accepts
only 1024 with output 4096, so metadata cannot enable another value. Only the
independent 5.2 candidate opts in. This is acceptance of one request shape,
not a range, enforced ceiling, adaptive/effort behavior or canonical intensity.
All Messages canonical levels remain unsupported until separately evidenced.

Both models also accepted native thinking history with only the signature field
removed from a successful signed baseline. Trusted same-model GLM unsigned
thinking therefore stays a thinking block and omits `signature`; native
Anthropic keeps its prior unsigned-to-text behavior. Preserving GLM's native
wire kind requires existing Native State Envelope checks even though its
visible thinking text is not opaque. Untrusted, cross-tenant/account/API/model
history follows the existing visible-text downgrade and drops opaque redacted
data. Stored integrity remains the Host's responsibility before TrustNativeState;
there is no new trust constructor, stamp, signature or credential in history.
The evidence did not observe provider-emitted unsigned or redacted thinking,
nor test empty-string signatures. Empty signatures and native redacted replay
are rejected before send; no Claude signature is invented.

GLM stream order is checked before assembling events: one message start,
unique complete blocks with matching delta kinds, then final stop reason and
message stop. Independent tool indexes retain separate JSON assemblers. The
native Anthropic parser retains its pinned behavior. Final GLM payload checks
also enforce text-only model input after trusted callbacks, including nested
tool results, without treating a tool schema/input property as a protocol field.
The [ticket 07 record](../../.scratch/barness-ai-glm/evidence/ticket-07/README.md)
retains raw captures, public Client red/green runs and regressions. Formal
production-route live acceptance and built-in registration remain downstream.

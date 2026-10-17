# Ticket 03 failure modes — written before production changes

Approved seam: spec T01 public `ai.Client` + Host bindings/catalog/tool execution,
controlled TLS Provider. This is the already user-approved seam; no private helper
tests. Each vertical slice is exercised before its minimal production change.

1. Both independent model candidates: declared functions fail to request
   `tool_stream:true`; interleaved index/ID fragments become separate or mixed
   calls; empty or partial JSON is executed before completion; streamed end and
   Result disagree. Host validates complete calls and supplies `ToolResultText`,
   then a separate generation returns the final answer with actual Attempts.
2. Same-turn `reasoning_content`, visible content, call IDs or result order are
   dropped in a legitimate tool continuation; plain assistant turns get the
   DeepSeek-only empty-reasoning field; replay falsely sets preserved thinking.
3. Unsupported named full required/function/allowed-tools/none or simple none
   silently weakens the tool constraint. Sampling/default/profile or final
   OnPayload bypasses auto-only or changes tool_stream away from true. Legal
   auto/default and a repairing final hook remain available.
4. Sampling or OnPayload requests clear_thinking:false, malformed thinking
   controls, or disables mandatory model thinking. Explicit true is valid;
   omitted retains vendor clear_thinking:true default. Authorization failures
   precede capability failures; all pre-send failures have zero Requests/Attempts.
5. Untrusted persisted, cross-model, cross-provider/API/account native state
   is replayed without Host authority or valid tool pairs/visible text disappear.
   Trusted persisted history replays the exact required reasoning; raw JSON
   deserialization alone must not restore that authority.
6. Caller/hook maps mutate an in-flight request, stream/event/Result snapshots
   share mutable content, or the two models/legacy providers influence each other.
7. Missing terminal/incomplete JSON becomes executable success; existing OpenAI,
   DeepSeek, callback and sampling semantics change. Pin regression/differential
   evidence independently from synthetic GLM extension-route evidence.

Fixtures are synthetic, based on dated official Chat schema/thinking sources in
`implementation-research.md`; no fixture claims a paid service observation.

Later root review identified an exact-number budget gap before its corresponding
production fix: `1.00000000000000001` and `131072.00000000001` can round to legal
integers. The public precision test first captured callback red, then captured
model/profile/call red before preserving effective GLM budget JSON through the
existing sampling path. Exact final validation accepts integer-valued decimal
and exponent forms, rejects fractions/huge exponents with zero sends, and does
not change old-provider JavaScript sampling semantics.

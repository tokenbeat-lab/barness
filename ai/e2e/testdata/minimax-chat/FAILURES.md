# MiniMax Chat failure inventory (before implementation)

Public seam: Client Stream/Complete/StreamSimple/CompleteSimple and WithHooks,
with an HTTP/SSE external Provider only (spec T01). Every test writes replayable
request, fixture, event, Result and Attempt evidence through the existing runner.

- Original OpenAI defaults can send developer/store/strict/cache and wrong thinking.
- Omitted canonical sampling labels can accidentally force explicit wire effort.
- Unsupported full effort/tool choices can be silently clamped or dropped.
- Final sampling/hooks can inject disabled mandatory thinking, unknown effort,
  reasoning_split=false, invalid sampling, fractional/zero/oversized output budget,
  old max_tokens, unsupported tools, hosted/multimodal content, caching or priority.
- Guard recursion can falsely reject ordinary function schema and argument fields.
- Tool fragments can merge across indexes, lose complete assistant reasoning/text,
  or report success with a conflicting finish reason.
- Finish frames can hide late usage; clean EOF may lack a finish; content_filter,
  unknown finish, invalid JSON and in-band errors must remain failures.
- Missing/zero/partial usage and interrupted streams must preserve honest reporting.
- Untrusted or cross-account native history must not replay native reasoning.

Synthetic reasoning_content and typed reasoning_details exercise preservation only.
They do not resolve the official M3 split-field conflict. Real M3 international standard
pay-as-you-go capture and replay remain ticket 07 evidence obligations.

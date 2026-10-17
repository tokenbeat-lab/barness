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
contract over inferred Claude features; ticket 07 may extend exact proven
shapes without asserting a range or canonical-strength mapping.

A separate Messages Host candidate and unchanged raw SSE fixtures retain
scope and provenance. Existing usage replacement already handles final
message_delta counts; absent cache_creation remains partial. Basic text
replay cannot register built-in models, certify tools/native replay, change the
support matrix or count as formal live acceptance. Evidence and repeatable
commands are indexed in the [ticket 06 record](../../.scratch/barness-ai-glm/evidence/ticket-06/README.md).

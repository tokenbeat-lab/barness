# Issue 11 two-axis review

Baseline: `efe6d198e301e96e4ccfcd0226d0f3c17eff00a8`. Implement requires review before commit, so reviewers compared this baseline to staged work rather than a pre-existing implementation commit. Only issue 11 edits were reviewed; pre-existing glossary and planning changes were excluded. Standards and Spec ran independently through the code-review skill.

## Standards

Four findings addressed: Decisions nullable reasoning/chat-only usage import checks; recorded source limits before decode; route-specific request-ID header selection; producer-reachable capture limits/states. Failure-first artifact E2E reproduced the invalid usage cases and dual-header/HTTP fixtures cover producer paths. Final reviewer confirmation: no remaining findings within scope.

## Spec

Three findings addressed: coherent fixture-to-live mutation; missing schema 2 live captures; cancellation before Attempt creation falsely rejected by replay. Added coherent identity/missing-capture process counterexamples, a bounded shared transport collector and request/response replay, and a resolver cancellation process case with 1 Call/0 Attempt. Final reviewer independently replayed the formerly rejected zero-Attempt bundle and confirmed it verifies as FAIL/fixture. No remaining findings within scope.

Own review also reproduced coherently rehashed request-question and response-answer edits; public Client replay now rejects both. Replay uses recorded Attempts only and does not replace historical metrics or timings.

Final findings: Standards 0 unresolved (4 fixed); Spec 0 unresolved (3 fixed). Real quality remains NOT_RUN under issue 12.

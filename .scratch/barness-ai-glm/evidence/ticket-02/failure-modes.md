# Ticket 02: confirmed public Client seam and failure modes

The root implement-spec session confirmed ai.Client Stream / Complete / StreamSimple / CompleteSimple, including WithHooks, against the local controlled Provider and trusted Host. All cases emit request captures, events, Results, errors and a candidate Catalog hash. No private helper tests or paid calls.

Vertical slices:

1. Both independent Chat candidates: correct official resource suffix and explicit Bearer Credential; text plus reasoning_content; finish_reason before trailing usage; all four entries, nil/full options and legal hooks. Failures: inherited OpenAI fields, DeepSeek empty assistant reasoning, wrong account/API attribution, lost trailing usage.
2. Model rules: GLM-5.3 only low/high/max, omitted preserves default; unsupported simple levels clamp upward, explicit full levels fail. GLM-5.2 maps its independently documented aliases and allows thinking disabled. Failures: provider-wide always-thinking, canonical/wire confusion, omitted means off, unsupported named level surviving mapping.
3. Final payload: model general → canonical profile → call → hook; sampling and callbacks cannot disable mandatory thinking or inject illegal effort, malformed switch or unimplemented text fields. A hook may repair an intermediate sampling violation. Authorization wins over capability errors. Every local failure has zero sends and Attempts.
4. GLM output budget: named/model/profile/call/hook values cannot exceed the model cap or official Chat schema cap 131072; valid 1/cap remain unchanged. Final null/zero/noninteger/wrong-type values and max_completion_tokens are local refusals. Named null/zero retain omission semantics. Host metadata may narrow but never enlarge Provider capacity.
5. Binding/Options: wrong protocol, disabled Binding, wrong Operation, forbidden ModelID fail before credential read; auth/endpoint remain Host-owned; header authorization cannot be replaced.

Candidate metadata and fixture rates are synthetic test configuration, not formal support or provider prices. Later tickets own tools, advanced capability guards, Responses/Messages facts and paid reports.

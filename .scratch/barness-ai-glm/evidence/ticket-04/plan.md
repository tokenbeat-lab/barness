# Ticket 04 public Client failure plan

Seam: spec T01 approved public ai.Client Stream/Complete/StreamSimple/CompleteSimple + WithHooks; only the external HTTP Provider is controlled. No isolated/private tests.

Failures enumerated before implementation: official-schema single-object reasoning/message content silently disappearing; empty authoritative done replacing delta output; done-only text/reasoning events being ignored; corrected done text being duplicated; interleaved output_index binding to wrong block; terminal-only output being ignored; malformed object/array/text accepted despite successful SDK decoding; missing terminal EOF reported as success; incomplete/error/failed inconsistency; normalized reasoning contaminating native signature; OpenAI automatic summary/include/tier/cache fields sent to GLM; 1-token output raised to16; output over model/provider limits; unsupported named5.3 effort accepted; sampling/profile/finalHook bypass; unauthorized model masked by effort error.

Vertical slices: object loss red→raw boundary green; authoritative lifecycle red→GLM parser green; capability/budget/effort red→request green; malformed/terminal behavior red→boundary green. Each run writes audit-able artifacts under root .evidence/barness-ai-glm/ticket-04.

Provenance: object fixtures derive from official Responses schema (2026-10-17 snapshot), not live observations. Both authorized real basic streams used arrays. Live bundles and hashes remain under responses-exploration; this ticket makes zero paid requests.

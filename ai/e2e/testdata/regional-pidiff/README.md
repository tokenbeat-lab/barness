# Regional frozen comparisons

This inventory uses the existing pi-ai 1.1.0 / commit `abe508e1b89912adde45528136c3221eb69acdd7` oracle. Its 43-file model-data hash and 827-file source hash, package lock and SDK dependencies remain unchanged. The wrapper adds only existing `ZAI_MODELS` and `MINIMAX_CN_MODELS` exports to its selection table. Empty subprocess environment, package verification and controlled-loopback socket restrictions remain active.

| Route | Target models | Proof |
| --- | --- | --- |
| zai × openai-completions | glm-5.3, glm-5.2 | Actual two-sided frozen comparison |
| minimax-cn × anthropic-messages | MiniMax-M3 | Actual two-sided frozen comparison |
| zai × openai-responses | glm-5.3, glm-5.2 | Independent extension fixture; pi parity NOT_RUN |
| zai × anthropic-messages | glm-5.3, glm-5.2 | Independent extension fixture; pi parity NOT_RUN |
| minimax-cn × openai-completions | MiniMax-M3 | Independent extension fixture; pi parity NOT_RUN |
| minimax-cn × openai-responses | MiniMax-M3 | Independent extension fixture; pi parity NOT_RUN |

The two actual frozen routes are never registered as route exemptions. Each target runs text, fragmented tools and a separate native tool replay through public `Client.Stream` / `StreamSimple` and the pinned real pi API. Case IDs are `PIDIFF-RP06-<route>-<model>-<text|tools|tool-replay>-<stream|streamSimple>`; each model also has a complete `<prefix><model>-catalog` comparison. No modelPatch or content removal hides a difference. Captures retain complete requests, all events and results, usage, costs and package/catalog/source hashes.

The current independent candidates come from `ai/internal/testkit/regional`. They remain qualification candidates; source-backed input/output/cache-read estimates do not qualify unknown positive cache writes, subscription deductions, live credentials or catalog promotion. Fixtures report zero cache writes. Their synthetic tokens and signatures are not live captures.

The new difference decisions in the central `../pidiff/ledger.json` name finite case sets, exact paths and kinds. `regional_pidiff_assertions_test.go` checks literal values before classification. Request model, auth, path, schemas, tool pairing, native reasoning/signatures, final content and token counts stay compared. The recorded differences are:

- ZAI's frozen Chat sends `stream_options.include_usage:true`, omitted-input thinking disabled, enabled `clear_thinking:false` and tool `strict:false`. The own GLM profile preserves service thinking defaults, parses trailing usage without requesting this optional stream flag, uses enabled/high for explicit tool turns, and leaves unqualified strict capability absent. Each exact body is checked; unrelated fields remain pending.
- Frozen CN Messages uses generic Anthropic enabled thinking with a budget/display/beta, browser-access declaration, ephemeral boundaries and eager streaming. The own CN profile uses its documented adaptive binary switch, omits the unqualified optional fields and preserves omitted service defaults. Both complete bodies and exact Content-Length are checked.
- Frozen CN M3 has historical USD prices and a 512000 output ceiling, with image input metadata. Current China candidates deliberately support text only and retain a conservative 2048 qualification ceiling pending conflicting source-limit qualification. Retained CNY rates and the fixed 6.7220 CNY/USD conversion produce their independent USD rates. Metadata and worked token/cost calculations are exact; no vendor maximum or live availability is inferred.
- Frozen ZAI catalog uses a Coding Plan base; the own standard API candidate uses `/api/paas/v4`. Frozen CN data retains `api.minimaxi.com`; the current own China configuration names `api.minimax.cn`. Complete model comparisons retain these differences. Both clients use explicit local controlled endpoints for wire comparison, which certifies no real account purpose or endpoint eligibility.
- Protocol flags stored on pi models live in barness's immutable Provider configuration. Complete model compat/map differences remain visible. The own glm-5.2 mapping's extra canonical positions are metadata differences, not proof for unexecuted effort settings; the actual high turns are compared.
- Independent elapsed durations and the existing concurrent PartialView publication timing use only the established finite duration scopes. Terminal/Result duration validity and equality remain mandatory.

The existing original eight differential subsets and their historical decisions remain. Cache exceptions now state ADR-0049's exact v2 algorithm and are limited to six existing Responses cache/affinity cases and three existing Chat cache cases. This changes only derived values; it does not exempt other wire, identity, event or usage behavior.

`TestRegionalProviderExtensions` runs all six absent-route model identities through all four Client entries from `extensions.json` and the independent ZAI Messages SSE. These cases have no `pidiff.json` and cannot count as differential PASS. `TestRegionalProviderDifferentialLedgerBounds` checks wrong case, kind, field and protocol refusal plus the distinction between real frozen routes and true extensions.

Development receipts and immutable failed/green native bundles are indexed under `.evidence/provider-sites/differential`. Feature state and later formal release qualification remain in the feature spec.

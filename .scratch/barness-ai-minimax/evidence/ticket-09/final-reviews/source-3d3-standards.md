# Standards review: final integration and corrective retention

Target: `3d3a1fb28e143a2c29d684f78edcad94c0217243`; fixed base: `ecfd4b4d6ec54db30453f4b5112989e9fdbd0782`. Reused the full-base Standards review through5bf2c16 and inspected `git diff 5bf2c16...3d3a1fb`. Confidence: high.

Runtime/harness source: PASS, unchanged across all1,034 frozen generation files. Retention delta: one documented-standard finding; no actionable judgment-smell findings. Applied AGENTS.md, domain/issue-tracker/GLOSSARY/ADR context and all twelve smell heuristics; excluded tooling-enforced concerns.

**[P2] Bound inventory decompression before allocating it.** `.scratch/barness-ai-minimax/evidence/ticket-12/original-preflight-source-proof/verify_q.py:39–40` calls `gzip.decompress()` before checking the32MiB expanded limit and before verifying the inventory hash. An8MiB compressed input can expand to several GiB before refusal. The same allocation-before-limit pattern occurs at `actual-corrective-1/verify_actual.py:147–148`. This breaches AGENTS.md principles4/5: external inputs require boundary validation and resource use must be bounded. Read through `gzip.GzipFile` with a32MiB+1 limit, reject overflow before JSON decoding, and retain exact hash checks. These are portable retention utilities; this finding does not invalidate the unchanged runtime source maps.

The retained actual archive hash/length and original reports agree: Chat/Messages22PASS each; Responses21PASS/1FAIL at `glm-5.2/native-history`, incomplete. Original failures, approval/source bindings and132/132 GLM consumption remain explicit. Failed Responses has no canonical successful capture pins; renewal refusal and unchanged selectors are correctly recorded. No support/release PASS is minted by integrity verification.

Separate external acceptance blockers: the closed15-job L receipt records168/168, two DeepSeek PASS and thirteen FAIL; total legacy consumption300/300 versus440/464. Parent reports network/403/451 causes. Jev protocol is partial, no formal derived support exists, and the17-input temporary import was rejected without changing the actual matrix. GLM Responses/current renewal remain blocked. The real gate ordinary run is in flight; race remains pending. No completed gate verdict is claimed.

Read-only review: no repository edits, tests, environment/credential access, network or vendor calls; no mutable gate bundle audited as closed.

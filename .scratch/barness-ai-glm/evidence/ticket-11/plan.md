# Ticket 11 fixed candidate execution plan

Registered 2026-10-17, before any ticket-11 paid request. Public seams were approved by spec T01: `ai.Client` Stream/Complete/StreamSimple/CompleteSimple, Host `ValidateToolCall` + ToolResult continuation and trusted native history. No new seam or permission requirement is introduced. The user supplied root `.env` GLM_KEY for the requested China standard routes; account `local-glm` / region `cn-standard-endpoint` remain unverified anonymous Host source labels.

`ai/live/testdata/glm-candidate/plan.json` is the executable, exact 22-case plan for each route. `sources.json` pins that plan, six-entry sourced USD catalog and complete dated official source snapshots. Manifest SHA256: `2e7c1a4e82f4e218e2a973c2049d220837ff7f8b428d3ae5e22342a535b38140`. Existing synthetic offline catalogs retain their old identities and rates. Candidate authority never publishes `live-report.json`, BuiltinCatalog, support matrix or formal route registration.

Each process selects exactly one China official protocol: Chat `https://open.bigmodel.cn/api/paas/v4/chat/completions`, Responses `https://open.bigmodel.cn/api/v1/responses`, Messages `https://open.bigmodel.cn/api/anthropic/v1/messages?beta=true` (locked SDK suffix; no beta header). SDKs are the repository-locked Go SDKs; versions are captured in report and manifest. No endpoint/model/key fallback, no SDK/Binding retry, no transport request JSON replacement. Both live build tag and `BARNESS_AI_LIVE=1` plus the exact isolated `^TestGLMCandidate$` entry are required. Narrowed/broad entry and foreign credentials fail before calls. The core never reads dotenv.

Per protocol, both GLM-5.3 and GLM-5.2 each have nine base calls: full Stream, full Complete, simple Stream, simple Complete, automatic add(17,25), validated Host tool result continuation, trusted own-model/account reasoning history, cancellation after observable streamed content, fresh recovery. Every model has qualified case IDs. Native replay uses each model's own full-stream result. Responses must actually send its InputReasoning **object**; 5.3 cannot borrow 5.2 or a tool turn with no reasoning. Messages must possess its own signed thinking before the unsigned-signature control. Only 5.2 full-stream enables exact thinking budget1024/max4096; 5.3 always preserves default omission. Canonical Messages thinking mappings stay null.

Two additional calls per model change exactly one final request field against the named baseline:

| Route/model | Baseline | Sole change |
| --- | --- | --- |
| Chat / each | full-stream (high) | temperature=0 |
| Chat / each | full-stream (high) | stream_options.include_usage=true |
| Responses / each | full-stream (high) | reasoning.effort=low |
| Responses / each | full-stream (high) | reasoning.effort=max |
| Messages / each | native-history | omit actual thinking.signature |
| Messages / 5.3 | tool-call | explicit tool_choice={type:auto} |
| Messages / 5.2 | full-stream (enabled1024) | omit thinking field |

Responses 5.2's canonical Low maps high; therefore the **safe Host OnPayload** callback changes only final wire effort to literal low. Acceptance is no claim of service intensity equality. Chat base calls additionally capture omitted, low, high, max across the four entries. Every comparison serializes and checks equality of all other request fields. No new probes may fill unused reservations.

Hard budget per protocol: planned22; max24 Logical Calls, max32 Attempt reservations AND actual external requests, max4096 output tokens each, max120 seconds per call (shorter Host wins), max1800 seconds process. Twenty seconds inside the process ceiling are reserved for finalization. Reserve before every public call and gate actual HTTP independently before the real transport. Remaining two calls permit at most one vendor-fault rerun per scenario; this initial execution chooses **zero automatic scenario reruns**, preserving any real failure for diagnosis and root coordination before any source-version rerun. Assertion failure cannot trigger retry-until-pass. No dynamic budget raising or target removal. No deliberate account/billing errors.

Missing key/switch gives NOT_RUN with22 unexecuted outcomes and zero spend. Actual errors (including429), prerequisite failure, exhausted budgets/deadline, missing provenance/reasoning, truncated captures, failed audit/inventory/write give FAIL. Required capability cannot become UNSUPPORTED. Both model reports must pass their11 cases before route PASS. Candidate report kind is structurally distinct from formal report; failure archives remain durable.

Evidence per call: final request JSON; raw SSE bytes (intentional cancellation partial stream identified separately); copied event snapshots/envelopes; Result, Attempts and raw usage-presence fields; Observer metadata; actual request-ID header source or explicit unreported; truncation flag; budgets, timing and reported USD estimate. Evidence manifest uses the candidate's real version/hash. Fixed source snapshots support prices8/28/2 RMB/M and fixed CFETS2026-10-16 1USD=CNY6.7220, rounded10 decimals; cache-write is explicitly inferred new-content computation at input rate, not free/premium/storage pricing. Context1000000 and Messages output128000 are conservative barness bounds, not vendor-exact or load tested; Chat/Responses output131072 from their schemas plus independent model128K pages. Source metadata records these limits.

Exact selected key is registered with `evidence.RedactSecret` before artifacts. Host launcher reads only GLM_KEY and starts clean children; no other vendor keys, environment/config dumps or key fingerprints. Audits must pass before accepting candidate report and integrity inventory; launcher additionally checks exact secret absence without recording it. Raw bundles and ZIPs are retained at root `.evidence/barness-ai-glm/ticket-11/`; tracked receipts/index/README link hashes. Every failed attempt remains retained.

Controlled public child failure plan, executed before paid calls: absent key, switch off, narrowed entry, foreign route key, ambient vendor key, endpoint override, unknown route, actual source missing/changed, catalog/plan mutation, fixed call/reservation/HTTP ceilings, shorter Host deadline, upstream429, inventory failure. Successful synthetic peers independently cover all22 cases on each protocol. Their transports cannot network and do not alter production request JSON. Red/green and final race runs leave replayable bundles.

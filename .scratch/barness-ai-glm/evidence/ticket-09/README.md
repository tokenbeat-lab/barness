# Ticket 09 — shared Client isolation, callbacks and resource release

2026-10-17. The [approved spec T01](../../spec.md) public Client seam is reused.
[Failure plan](failure-plan.md) preceded tests and production changes. All calls
use controlled loopback Providers, synthetic credentials and own-model protocol
fixtures. No real key was read, no paid call was made, and no live/support-matrix
or built-in catalog claim changes. [Index](index.json) pins source snapshots,
fixtures, production files, audited manifests, case artifacts and verified ZIPs.
Durable bundles, raw diagnostics and archives are at repo-root
`.evidence/barness-ai-glm/ticket-09/`.

## Changes and red evidence

The final Chat native payload previously accepted callback-injected non-text
content, strict functions, Host-authorized hosted tools, developer roles,
stream:false and unintegrated cache/service/parallel/strict-format fields. A
Host Catalog could also cause automatic strict:false. The GLM-only final guard
now checks actual protocol containers after authorization; function schema
properties/arguments remain ordinary data. Both models cover sampling, hooks,
authorization priority and catalog compatibility. Safe documented controls and
the explicitly planned temperature=0 / stream_options.include_usage candidate
comparison remain available; defaults still omit unverified stream_options.
[Official Chat OpenAPI snapshot](sources/chat-boundary.json) records its URL,
retrieval date, source SHA and relevant independent schemas. This is an
integrated subset, not a claim that every omitted field is rejected upstream.

The refined `diagnostics/20261017T132556.774834000Z` red has **6 FAIL / 6 PASS**:
Host-authorized web_search_options, a tool namespace and object-shaped native
content reached the Provider, while unauthorized models still won. The initial
Chat red also captured real holes but included incorrect expected phases for
already-reserved sampling fields; those assertions were corrected to preserve
the existing contract. Their raw artifacts are retained as diagnostics, never
counted as additional defects or passing release evidence.

A Host header callback panic terminated the test process; payload and response
invocations had the same unguarded boundary. Recovery now wraps only the three
Host function invocations. It returns callback_failed, retains error causes for
Host diagnosis, never formats arbitrary panic values, and preserves context
priority. The crashing process is **INCOMPLETE**: its original directory has no
finished behavior manifest/audit. Its raw log and unfinished files are retained
inside the separately audited diagnostic-log bundle. No adapter-wide recovery
or compatibility shim was added. See [ADR0005](../../../../docs/adr/0005-barness-ai-payload-hosted-tools.md)
and [ADR0033](../../../../docs/adr/0033-barness-ai-zhipu-chat-candidate.md).

## Verified behavior

- Two tenants share one Client, same BindingID/model/session names and distinct
  account/credential scopes. Their concurrent logical calls use distinct
  RequestIDs. Each model and each protocol completes its own protocol-fixture tool
  round trip; captures prove credential, account, cache-key and event/result/
  Attempt/Observer attribution isolation.
- Native State covers in-process trust, restored Host attestation, untrusted or
  foreign-tenant input, account mismatch, actual cross-protocol tool history,
  provider mismatch and authorized cross-model/account conversion. Downgrades
  preserve visible content and valid tool pairs without replaying opaque state.
  The actual 5.2 Messages tool capture contains no thinking: its separate own
  thinking capture supplies that model's native-state scenario. No native
  fingerprint or new integrity promise replaces the existing Host contract.
- Hooks preserve header→payload→response, legal modifications, protected Key/
  target/model/resource authorization and failure priority. Preflight failures
  create zero requests/Attempts; OnResponse failure retains its real Attempt
  and unreported usage because the body has not yet been read. Ticket08's
  fixed credential/Binding/payload retry snapshot is rerun.
- Request/tool Raw JSON, sampling maps, constructor and returned catalogs,
  nested thinking maps, hook-retained maps, event snapshots, Result content
  and Attempt slices remain independently owned. Mutations happen after an
  explicit snapshot/serialization barrier, never concurrently with a Go map
  read. Observer remains metadata only and all completed bundles pass audit.
- Request/frame/error-body/tool/output/queue count/queue byte limits, host
  Admission denial, connect/header/read/call/Host deadlines and cancellation
  before send/headers/after legal content release all bodies and permits.
  Result waits independently of consumers and each stream has one terminal.
  A sibling tenant completes its tool round trip while a subject is held.
  Chat/Messages preserve legally reported usage; Responses before a legal
  terminal reports none. Cancellation while Responses awaits EOF retains known
  terminal usage; cancellation after successful settlement cannot replace it.
- Repeated concurrent failure rounds run **48 failed GLM calls** across both
  models, two tenants and three routes: unread queues and unfinished oversized
  frames settle without Close driving completion. Body/permit/waiter/event/
  active-call gauges reach zero and goroutines return to a bounded baseline.

## Design load and limits of the measurement

The controlled functional load is six simultaneous HTTP leases, two tenants ×
three protocols, repeated with both models. Each request adds 1 KiB synthetic
text; original own-protocol fixture replies are retained. This is not a new
production policy or a substitute for the existing formal pressure gate.
Observed maxima: request 1,804 bytes, reply/frame conservative bound 5,801
bytes, tool JSON 18 bytes, 39 complete events and 31,184 bytes of serialized
complete-event snapshots. The declared 4 KiB tool / 32 KiB wire ceilings are
conservative envelopes, not amounts claimed to have been saturated.

The bundle records headroom for every applicable byte/concurrency/queue limit,
zero image/error-body/waiter use, active leases, release gauges and goroutines.
Queue serialized size is a conservative external bound, not an internal
allocation metric. Memory is process allocation including fixture/Observer/
evidence work: the final run records about 24.4 MB total allocation and 5.0 MB
heap afterward; it is not a pure Client benchmark. Normal idle connection
pooling is bounded separately; the initial 16-idle-connection harness failure
is retained and corrected to a two-connection idle pool, rather than described
as a library goroutine leak.

## Validation and replay

| Run | Result |
| --- | --- |
| Current complete ticket09 public Client suite | 359 PASS |
| Current complete ticket09 under race | 359 PASS |
| Repeated concurrent failure release, separately under race | 2 PASS / 48 failed logical calls |
| Affected GLM + old callback/resource/Chat regressions | 2,794 PASS |
| Broader callback/usage/resource race | 523 PASS |
| Frozen OpenAI + DeepSeek Chat differential | 321 PASS, zero pending |
| Chat boundary after distinct RequestID literals | 116 PASS |
| go vet ./ai/... and whitespace | PASS |

Production code is fixed at `6e6cf45`; later commits adjust only cancellation
fixture order, headroom/identity evidence and repeated-failure public E2E. The
old-provider and differential runs therefore cover the same production bytes.
All current runs and earlier diagnostics remain separately identified in the
index; no diagnostic PASS total is added to a final total.

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-09/replay" \
  go test ./ai/e2e -run '^TestGLMShared' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-09/replay-race" \
  go test -race ./ai/e2e -run '^TestGLMShared' -count=1
go vet ./ai/...
# Requires the unchanged frozen oracle's locked dependencies:
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/(openai-completions|deepseek)$' -count=1
```

Each case manifest supplies its exact one-case replay command and hashed
requests, scripts, events, Result, Attempts, observations and assertions.
Rollback is an ordinary integrated code rollback or disabling the Host GLM
Binding. Formal/live/catalog/release evidence remains with tickets 10–13.

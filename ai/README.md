# barness-ai

English | [简体中文](README.zh-CN.md)

`github.com/tokenbeat-lab/barness/ai` is barness's model protocol module: one
trusted, concurrency-safe `Client` that runs **one generation turn** against a
model provider for **many tenants**. It is a Go port of
[pi-ai](https://github.com/earendil-works/pi) (frozen baseline `1.0.0`) with
tenant isolation, explicit resource limits and observability added on top.

This README is for two audiences:

- **People** who want to know what the module does, how to call it, and where
  its guarantees are written down.
- **Coding agents** changing the module, who need its invariants, layout,
  conventions and test commands before they edit anything (see
  [For coding agents](#for-coding-agents)).

Authoritative sources, in order of precedence: the spec
(`.scratch/barness-ai/spec.md`), the [ADRs](../docs/adr/), and the package
documentation (`go doc ./ai`). This README summarizes them and never relaxes
them. Terms are defined in [GLOSSARY.md](../GLOSSARY.md).

## What it does — and what it doesn't

| barness-ai does | The host (your program) does |
| --- | --- |
| Convert a host-selected transcript into a provider request, stream the response back as pi-ai events, and return the final `AssistantMessage` | Authenticate callers and build each `CallScope` from that — never from request content |
| Resolve the tenant's **binding** and **credential** through host resolvers, and pin that snapshot for the whole call | Store bindings, credentials and history; decide which history goes into each turn |
| Enforce a finite `ResourcePolicy` (bytes, queues, concurrency, time) | Choose the policy values for its workload |
| Replay provider native state (signatures, encrypted reasoning) only where its provenance matches | Vouch for stored native state with `TrustNativeState` |
| Report tool calls, validate their arguments on request | Execute tools and start every next turn as a new logical call |
| Classify errors, estimate cost, emit observations | Map errors to its own API (e.g. `admission_denied` → HTTP 429) |

Non-goals: no agent loop, no built-in memory or history trimming, no tool
execution, no logging, no reading of `OPENAI_API_KEY`-style environment
variables, no default resource limits.

## Supported combinations

| Provider × API | `ProviderID` / `API` | Live smoke |
| --- | --- | --- |
| OpenAI × Responses | `openai` / `openai-responses` | PASS |
| Anthropic × Messages | `anthropic` / `anthropic-messages` | PASS |
| Google × Gemini Developer API | `google` / `google-generative-ai` | PASS |
| OpenAI × Chat Completions | `openai` / `openai-completions` | PASS (no replayable reasoning by protocol) |
| DeepSeek × Responses (extension route, not in pi) | `deepseek` / `openai-responses` | PASS |
| DeepSeek × Chat Completions | `deepseek` / `openai-completions` | PASS |

The source of truth is [`live/support-matrix.json`](live/support-matrix.json).
A combination may be called supported only when its own row has fully passed;
"OpenAI-compatible" does not mean any compatible service is accepted.
Callable models are the intersection of the catalog (`BuiltinCatalog()`,
snapshot in [`release/catalog-snapshot.json`](release/catalog-snapshot.json))
and `Binding.AllowedModels`.

## Quick start (local)

A local program uses the same `Client` as a cloud host; it just states
explicitly the one tenant, binding and key a cloud host would resolve.
[`examples/localassembly`](examples/localassembly) does that assembly:

```go
asm, err := localassembly.Open(localassembly.Config{
	TenantID: "local",
	Binding: ai.Binding{
		BindingID:      "openai",
		ProviderID:     ai.ProviderOpenAI,
		API:            ai.APIOpenAIResponses,
		Endpoint:       "https://api.openai.com/v1",
		AccountScopeID: "my-openai-account",
		AllowedModels:  []string{"gpt-5-mini"},
	},
	// Exactly one source, chosen by the program. Nothing falls back to OPENAI_API_KEY.
	Key: localassembly.KeySource{EnvVar: "BARNESS_OPENAI_KEY"},
	// Policy: nil uses localassembly.LocalPolicy().
})
if err != nil { return err }

req := ai.Request{
	SystemPrompt: "You are concise.",
	Messages:     []ai.Message{ai.UserText("Say hello in French.")},
}
res, err := asm.Client.CompleteSimple(ctx, asm.Scope(), asm.Target("gpt-5-mini"), req,
	ai.SimpleOptions{Reasoning: ai.ThinkingLow})
if err != nil { return err } // *ai.Error; res is still a complete Result

for _, c := range res.Message.Content {
	if t, ok := c.(ai.Text); ok {
		fmt.Println(t.Text)
	}
}
```

## Core concepts

```
Host ──CallScope + Target + Request + Options──▶ Client
                                                  │ 1. validate scope, request, image bytes
                                                  │ 2. BindingResolver(tenant, bindingID)
                                                  │ 3. authorize model (catalog ∩ AllowedModels), map options
                                                  │ 4. prepare history; downgrade untrusted native state
                                                  │ 5. CredentialResolver → check snapshot consistency
                                                  │ 6. per attempt: admission → adapter → provider
                                                  ▼
                        Stream (events + EventEnvelope)  /  Result{Message, Metadata}
```

| Concept | Go type | Notes |
| --- | --- | --- |
| Call scope | `CallScope` | `TenantID` and `RequestID` required. `RequestID` is globally unique per logical call; retries reuse it, every new call (including the next tool turn) needs a new one — see [`examples/requestid`](examples/requestid). |
| Target | `Target` | `BindingID` + `ModelID`. Protocol, endpoint, account and key come from the binding, never from the request. |
| Binding | `Binding` | One tenant's service config: provider, API, endpoint, account, allowed models, allowed hosted tools, `Retry` (zero value = never retry). `Enabled` and `Version` are required; zero values fail closed. |
| Credential | `Credential` | Versioned snapshot; `BindingVersion` must equal the resolved binding's version, otherwise the call fails (`credential_unavailable`/`consistency`, ADR-0003). Keys are `ai.Secret` and never appear in output. |
| Request | `Request` | `SystemPrompt`, `Messages`, `Tools`. Copied on entry. The library never selects, trims or stores history. |
| Options | `ResponsesOptions`, `AnthropicOptions`, `GeminiOptions`, `ChatOptions` / `SimpleOptions` | Full options must match the binding's API. `Nullable[T]` distinguishes unset / `null` / zero. |
| Result | `Result` | `Message` (always valid, failures included) and `Metadata` (`CallMetadata`: attribution, versions, catalog hash, attempts, usage reporting, native state downgrades). |

### Entry points

```go
res, err := client.Complete(ctx, scope, target, req, opts)        // full protocol options
res, err := client.CompleteSimple(ctx, scope, target, req, simple) // protocol-neutral options
s := client.Stream(ctx, scope, target, req, opts)
s := client.StreamSimple(ctx, scope, target, req, simple)
hc := client.WithHooks(hooks) // trusted request callbacks; same four methods
```

Use `Complete*` when you only need the final message. A `Stream` queues every
event; one whose events are never read ends with `resource_limit`
(`PhaseEventQueue`) once its output outgrows `MaxQueuedEvents` /
`MaxQueuedEventBytes`.

```go
s := client.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{})
defer s.Close() // idempotent; cancels I/O and releases permits
for s.Next() {   // single consumer
	switch e := s.Event().(type) {
	case ai.TextDeltaEvent:
		fmt.Print(e.Delta)
	}
	_ = s.Envelope().Call // attribution of this event, for routing merged streams
}
res, err := s.Result() // may also be awaited in parallel with Next
```

Events are pi's: `start`; interleaved `text_*`, `thinking_*`, `toolcall_*`
(`start`/`delta`/`end`) with stable content indexes; exactly one terminal
`done` or `error`.

### Tool round trip

barness-ai never runs a tool. A turn ending with `StopReasonToolUse` hands the
calls to the host, which validates each with `ValidateToolCall`, runs it,
answers with `ToolResultText`, and starts the next turn as a new logical call.
`ToolCall.Arguments` is a best-effort parse for display only; never execute
it. Turns ending in `length`, `error` or `aborted` are never executed. See
[`examples/toolloop`](examples/toolloop).

### Native state

Messages produced in this process carry trusted provenance and replay as is.
Messages decoded from storage (JSON) are untrusted until the host calls
`TrustNativeState(scope, envelope)` after its own checks. Untrusted or
mismatched (tenant, account, model) native state is downgraded by pi's
cross-model rules, never refused; downgrades are counted in `Metadata`.

### Errors

A failed call returns a complete `Result` plus `*ai.Error` with a stable
`Code` and `Phase`:

```go
var e *ai.Error
switch {
case errors.Is(err, &ai.Error{Code: ai.CodeRateLimited}): // Phase empty = any phase
case errors.As(err, &e):
	log.Println(e.Code, e.Phase, e.HTTPStatus, e.ProviderRequestID, e.RetryAfter)
}
```

| Before the provider | At / after the provider |
| --- | --- |
| `invalid_request`, `tenant_denied`, `binding_not_found`, `credential_unavailable`, `admission_denied` | `upstream_auth`, `rate_limited`, `upstream_error`, `transport`, `protocol` |
| Any stage: `canceled`, `deadline_exceeded`, `resource_limit`, `callback_failed` | |

`Error.Message` quotes the provider's error body as pi does (key-shaped text
redacted) and may echo request content: return it only to the calling tenant,
never to shared logs. Use the `Observer` for logging. The full table is in
[contract.md §4](../docs/barness-ai/contract.md).

### Resource policy, admission, time limits

Every `Client` requires an explicit, finite `ResourcePolicy` — there are no
defaults and no "unlimited" mode (ADR-0002). Byte limits are enforced while
encoding and reading. Each attempt, retries included, takes a per-tenant and
per-process permit, then the host's `Config.Admission` if set (ADR-0008). A
call ends at the earliest of the context deadline, `CallTimeout`, and the
per-attempt connect / header / read-idle / protocol timeouts.

Annotated, pressure-tested starting points:
`localassembly.LocalPolicy`, `hostintegration.CloudInteractivePolicy`,
`hostintegration.CloudBatchPolicy`. Every value's rationale is in their source
comments; measurements are in
[docs/barness-ai/README.md](../docs/barness-ai/README.md).

### Observability, usage and cost

`Config.Observer` receives `call_started`, `attempt_started`,
`attempt_finished`, `call_finished` asynchronously through a bounded queue; a
slow observer never blocks a call, drops are counted in
`Client.ObserverStats()`. Records hold no keys, content, tool arguments or
error text (ADR-0009). `Usage` keeps pi's numbers; `Usage.Cost` is an
estimate from catalog prices, not billing. Whether usage was reported at all is
`Attempt.UsageReporting` — a zero `Usage` never means free (ADR-0010).

## Cloud host integration

[`examples/hostintegration`](examples/hostintegration) is the reference.
Checklist:

1. Authentication result → `CallScope` (new `RequestID`, `ActorID`). Refuse a
   self-claimed `TenantID` and another tenant's session.
2. Read history by `(tenant, session)`; vouch for stored native state with
   `TrustNativeState`.
3. Implement `BindingResolver` / `CredentialResolver`. Any cache must be keyed
   by tenant, credential ID and version, with documented TTL and revocation
   propagation.
4. Pick a cloud policy; answer `admission_denied` with 429 + `Retry-After`.
   Queueing stays outside barness-ai.
5. Inject `Admission` for per-account or distributed quotas; inject `Observer`
   and watch `ObserverStats`.
6. Configure retries on the binding (`Binding.Retry`, ADR-0006). Only the
   initial request is retried; a started stream is never replayed.
7. When the downstream disconnects, cancel the call's context or `Close` its
   stream. When merging streams, route by `EventEnvelope.Call`, not arrival
   order (`hostintegration.Merge`).

Never accept keys, credential refs, endpoints, `Authorization` headers, proxies
or callbacks from request JSON — they come only from trusted assembly.

---

## For coding agents

Read this section, [GLOSSARY.md](../GLOSSARY.md) and the ADRs touching your
area before editing. The repo-wide rules in [AGENTS.md](../AGENTS.md) apply
(500-line file limit, delete rather than keep compatibility shims, E2E-first
testing).

### Invariants — do not break

1. **Identity is trusted input only.** `CallScope` never comes from request
   content, env or `context.Context`. Every tenant-scoped lookup goes through
   the resolvers with the scope.
2. **No defaults for limits.** Never add a default `ResourcePolicy`, a
   "disable limits" switch, or an unbounded buffer, queue, retry or wait.
   Every new buffer must be covered by a policy limit and fail with
   `CodeResourceLimit`.
3. **Snapshot pinning.** After resolution, a call uses one binding +
   credential snapshot for all attempts; never re-resolve inside a call (D2,
   ADR-0003).
4. **Retries only on the initial request**, governed by `Binding.Retry`.
   Never retry a resource-limit failure or a started stream; never let an SDK
   retry on its own.
5. **Secrets never leave.** `Secret` values must not reach messages, events,
   errors, observations, evidence bundles or logs. barness-ai writes no logs.
6. **No ambient configuration.** No reading provider env vars, proxy env vars,
   credential files or default endpoints. Endpoints must be https except with
   `AllowLoopbackHTTP` (tests only).
7. **pi parity is the default.** Observable behavior matches frozen pi-ai
   `1.0.0`. Any intentional difference must be recorded in
   [`e2e/testdata/pidiff/ledger.json`](e2e/testdata/pidiff/ledger.json) and
   [differences.md](../docs/barness-ai/differences.md); a `pending` entry
   blocks release.
8. **Capabilities are per provider, not per protocol.** A shared adapter must
   not send another provider's fields (e.g. DeepSeek × Responses sends no
   `store`/`include`/cache fields). Derive request fields from the provider's
   capabilities.
9. **Observers never affect calls; hooks may fail but never widen** auth,
   target, model, native references or hosted tools (ADR-0005).
10. **Tool arguments are not executable** until `ValidateToolCall` passes on a
    finished message.

### Package layout

All production code is the single package `ai` (flat, files grouped by
prefix). Unexported types do not leak across these groups by convention.

| Area | Files |
| --- | --- |
| Public entry & stream | `client.go`, `entry.go`, `call.go`, `stream.go`, `event.go`, `assembler.go` (single production/merge process behind every entry), `partial.go`, `partial_json.go` |
| Domain types | `message.go`, `request.go`, `result.go`, `binding.go`, `options.go`, `nullable.go`, `reasoning.go`, `errors.go` |
| Resolution & security | `preflight.go`, `policy.go`, `admission.go`, `limits.go`, `timeouts.go`, `retry.go`, `transport.go` |
| History & native state | `transcript.go`, `history.go`, `replay.go`, `native.go`, `cache_key.go` |
| Tools | `tool.go`, `tool_validate.go`, `tool_coerce.go`, `tool_json.go`, `argument_text.go` |
| Catalog | `catalog.go` (catalog and lookup), `catalog_models.go` (model metadata), `catalog_builtin.go` (built-in data), `catalog_snapshot.go` (copy, hash and price validation) |
| Usage, cost | `usage.go`, `estimate.go`, `anthropic_effort.go` |
| Hooks & observability | `hooks.go`, `hooks_run.go`, `observer.go` |
| Adapter contract | `adapter.go` (`adapter` interface + `registry()` keyed by `API`) |
| OpenAI Responses | `responses*.go`, `openai_sdk.go` |
| Anthropic Messages | `anthropic*.go` |
| Gemini Developer API | `gemini*.go` (REST, no SDK) |
| Chat Completions | `chat*.go` (`chat_compat.go` holds per-provider differences) |
| Shared HTTP/SDK glue | `sdk_middleware.go`, `http_failures.go` |

Other directories:

| Directory | Purpose |
| --- | --- |
| `examples/` | `localassembly`, `hostintegration`, `toolloop`, `requestid` — minimal, each exercised by an offline E2E |
| `e2e/` | Offline acceptance tests against the public API, a local controlled provider and a host double; fixtures in `e2e/testdata/<protocol>/` |
| `live/` | Real-API smoke (build tag `live`), feeds `support-matrix.json` |
| `release/` | Release gate command, catalog snapshot, traceability map |
| `internal/testkit/` | Test-only: controlled provider, host double, pi oracle (Node), evidence, audit, release gate logic |
| `internal/probe`, `internal/clock` | Test hooks injected through `Config` (hosts cannot construct them) |

### Common changes

- **Add a model:** add it to the matching `builtin*Models()` in `catalog_builtin.go`
  following the inclusion criterion (ADR-0018), then regenerate the catalog
  snapshot with `go run ./ai/release/cmd/releasegate -write-snapshot` and
  review the diff.
- **Add a provider on an existing API:** add a `ProviderID`, its capability
  set (e.g. in `chat_compat.go` / `responses_capabilities.go`), fixtures under
  `e2e/testdata/`, a live combo, an ADR, and a ledger entry if pi has no such
  route.
- **Add a protocol:** implement `adapter` (`adapter.go`), register it in
  `registry()`, add an `API` constant, its `Options` type, E2E suites
  (`e2e/<protocol>_*_test.go`) and an ADR. It must enforce `call.limits`
  while encoding and reading.
- **Change observable behavior:** run the pi differential, then update the
  ledger and `differences.md`.

### Testing

Per AGENTS.md, E2E is the primary mechanism. Tests drive the public `Client`
only; nothing in `e2e/` reaches inside the package.

```sh
go test ./ai/...                               # offline E2E + package tests, no network
go test -race ./ai/...
BARNESS_AI_PRESSURE=1 go test ./ai/e2e         # policy pressure scenarios
BARNESS_AI_PIDIFF=1  go test ./ai/e2e          # differential against frozen pi
#   needs Node.js and: npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node

# Live smoke: one combo per process, only that combo's key
BARNESS_AI_LIVE=1 BARNESS_AI_LIVE_COMBO=deepseek-chat \
BARNESS_AI_LIVE_ACCOUNT_ALIAS=<alias> BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT=<key> \
go test -tags live -count=1 ./ai/live
go run ./ai/live/cmd/supportmatrix <bundle-dir>...   # merge reports into the matrix

go run ./ai/release/cmd/releasegate [-live <bundle-dir>]...   # full release gate
```

Every run writes a redacted evidence bundle under `.evidence/`
(`BARNESS_AI_EVIDENCE_DIR` overrides). New behavior needs an E2E case whose
evidence the release gate can trace (`release/traceability.json`).

## Further reading

| Document | Content |
| --- | --- |
| [docs/barness-ai/contract.md](../docs/barness-ai/contract.md) | Public contract, full error table, assembly notes |
| [docs/barness-ai/differences.md](../docs/barness-ai/differences.md) | Every difference from pi-ai and its disposition |
| [docs/barness-ai/README.md](../docs/barness-ai/README.md) | Release deliverables, gate status, policy measurements |
| [docs/adr/](../docs/adr/) | ADR-0001 … ADR-0021 |
| [GLOSSARY.md](../GLOSSARY.md) | Domain terms (Tenant, Binding, Logical Call, Native State, …) |
| `go doc -all ./ai` | Package and type documentation |

### TypeSafe classification

`Client.Classify` and `HookedClient.Classify` submit a state and a named set of
`ChoiceQuestion`, `ScoreQuestion` and `BoolQuestion` values in one unary call.
Enable the finite `ResourcePolicy.Classifier` and a classifier binding plus a
host classifier catalog. Native JSON descriptions and numbers retain precision;
score levels stay ordered and bool answers give the probability of yes.
All final answers are validated together, with usage retained on failure.
See [the classifier contract](../docs/barness-ai/contract.md) and
[ADR-0021](../docs/adr/0021-barness-ai-typesafe-unary-classification.md).
Built-in TypeSafe models/pricing and live support remain gated by issue 08.

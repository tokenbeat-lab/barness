# barness-ai

English | [简体中文](README.zh-CN.md)

`github.com/tokenbeat-lab/barness/ai` is barness's model protocol module: one trusted,
concurrency-safe `Client` that runs **one model operation** against a model provider for **many
tenants**. It is a Go port of [pi-ai](https://github.com/earendil-works/pi) (frozen baseline
`1.0.0`) with tenant isolation, explicit resource limits and observability added on top.

This README covers the public API and host responsibilities. Coding agents
should also read [the invariants, layout and testing guide](#for-coding-agents).

Authoritative sources, in order of precedence: the spec ([1.0 upgrade
spec](../.scratch/barness-ai-pi-1.0/spec.md) and its base spec), the [ADRs](../docs/adr/), and the
package documentation (`go doc ./ai`). The spec describes the target; support claims below follow
the implemented catalog and each route's own evidence. Terms are defined in
[GLOSSARY.md](../GLOSSARY.md).

## Model operations

| Operation | Entry points | Request → result | Delivery |
| --- | --- | --- | --- |
| `chat` | `Complete`, `CompleteSimple`, `Stream`, `StreamSimple` | `Request` → `Result` / `Stream` | Final assistant message or pi-ai events |
| `image` | `GenerateImages` | `ImagesRequest` → `ImagesResult` | Synchronous, ordered text/image blocks |
| `classifier` | `Classify` | `ClassifierRequest` → `ClassifierResult` | Synchronous, typed answers to named questions |

`Client` and `HookedClient` expose all six entry points. Image and classifier calls honor context
cancellation, emit no delta events, and accept protocol options or `nil` for defaults;
`SimpleOptions` is chat-only.

## What it does — and what it doesn't

| barness-ai does | The host (your program) does |
| --- | --- |
| Convert a host-selected transcript into a provider request, stream the response back as pi-ai events, and return the final `AssistantMessage` | Authenticate callers and build each `CallScope` from that — never from request content |
| Resolve the tenant's **binding** and **credential** through host resolvers, and pin that snapshot for the whole call | Store bindings, credentials and history; decide which history goes into each turn |
| Enforce a finite `ResourcePolicy` (bytes, queues, concurrency, time) | Choose the policy values for its workload |
| Replay provider native state (signatures, encrypted reasoning) only where its provenance matches | Vouch for stored native state with `TrustNativeState` |
| Report tool calls, validate their arguments on request | Execute tools and start every next turn as a new logical call |
| Generate/edit native images or return validated classifier answers in a unary response | Explicitly enable an operation binding, catalog model and finite child policy; choose business thresholds and human review |
| Classify errors, estimate cost, emit observations | Map errors to its own API (e.g. `admission_denied` → HTTP 429) |

Non-goals: no agent loop, no built-in memory or history trimming, no tool execution, no logging, no
reading of `OPENAI_API_KEY`-style environment variables, no default resource limits.

## Supported combinations

| Operation | Provider × API | `ProviderID` / `API` | Live smoke |
| --- | --- | --- | --- |
| chat | OpenAI × Responses | `openai` / `openai-responses` | PASS |
| chat | Anthropic × Messages | `anthropic` / `anthropic-messages` | PASS |
| chat | Google × Gemini Developer API | `google` / `google-generative-ai` | PASS |
| chat | OpenAI × Chat Completions | `openai` / `openai-completions` | PASS (no replayable reasoning by protocol) |
| chat | DeepSeek × Responses (extension route, not in pi) | `deepseek` / `openai-responses` | PASS |
| chat | DeepSeek × Chat Completions | `deepseek` / `openai-completions` | PASS |
| classifier | TypeSafe × System One, jev-1.13.0 | `typesafe` / `typesafe-system-one` | PASS (context probe returned 400; 422 shape unconfirmed) |
| image | OpenAI × Images, gpt-image-2.5-sunburst-2026-09-08 | `openai` / `openai-images` | PASS (generation, JSON edit and mask) |
| image | Google × Interactions v1beta, gemini-nano-banana-2.1 | `google` / `google-interactions` | PASS (generation and reference edit, 1K / 1:1) |

The table summarizes the nine routes' full live runs on **2026-10-08**; the source of truth is
[`live/support-matrix.json`](live/support-matrix.json). A combination may be called supported only
when its own row has fully passed; "OpenAI-compatible" does not mean any compatible service is
accepted. Callable models are the intersection of the catalog (`BuiltinCatalog()`, snapshot in
[`release/catalog-snapshot.json`](release/catalog-snapshot.json)) and `Binding.AllowedModels` within
`(Operation, Provider, API, ModelID)`. Both native image routes and DeepSeek Responses are
registered extensions; their evidence is independent of pi differential results. Release status and
subsequent offline fixes are in [the release notes](../docs/barness-ai/README.md).

## Quick start (local)

Requires **Go 1.26.2 or later** (see [go.mod](../go.mod)). Examples below are function-body snippets
using `context`, `fmt`, `encoding/json`, `ai`, and `ai/examples/localassembly`; callers supply `ctx`
and handle the returned error.

A local program uses the same `Client` as a cloud host; it just states explicitly the one tenant,
binding and key a cloud host would resolve. [`examples/localassembly`](examples/localassembly) does
that assembly:

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
	Policy: localassembly.LocalPolicy(), // chat-only example policy; review its load assumptions
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

### Image generation and classification

`OpenOperations` assembles several bindings on one Client. It requires an explicit policy and
preserves each binding's `Enabled` choice. The program reads only the key sources it names, indexed
by `CredentialRef`:

```go
asm, err := localassembly.OpenOperations(localassembly.OperationsConfig{
	TenantID: "local",
	Bindings: []ai.Binding{
		{
			BindingID: "images", Operation: ai.OperationImage, Enabled: true,
			ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIImages,
			Endpoint: "https://api.openai.com/v1", AccountScopeID: "my-openai-account",
			CredentialRef: "openai-key",
			AllowedModels: []string{"gpt-image-2.5-sunburst-2026-09-08"},
		},
		{
			BindingID: "classifier", Operation: ai.OperationClassifier, Enabled: true,
			ProviderID: ai.ProviderTypeSafe, API: ai.APITypeSafeSystemOne,
			Endpoint: "https://api.typesafe.ai", AccountScopeID: "my-typesafe-account",
			CredentialRef: "typesafe-key", AllowedModels: []string{"jev-1.13.0"},
		},
	},
	Keys: map[string]localassembly.KeySource{
		"openai-key":   {EnvVar: "BARNESS_OPENAI_KEY"},
		"typesafe-key": {EnvVar: "BARNESS_TYPESAFE_KEY"},
	},
	Policy: localassembly.MixedPolicy(),
})
if err != nil { return err }

images, err := asm.Client.GenerateImages(ctx, asm.Scope(),
	ai.Target{BindingID: "images", ModelID: "gpt-image-2.5-sunburst-2026-09-08"},
	ai.ImagesRequest{Prompt: "A small red sailboat on a calm lake."},
	ai.OpenAIImagesOptions{Size: ai.Value("1024x1024"), Quality: ai.Value("low")})
if err != nil { return err } // images still carries usage and call metadata
for _, block := range images.Content {
	if img, ok := block.(ai.ImageOutputImage); ok {
		// img.Data is strict base64; the host chooses how to save or display it.
		fmt.Println(img.MimeType)
	}
}

classified, err := asm.Client.Classify(ctx, asm.Scope(),
	ai.Target{BindingID: "classifier", ModelID: "jev-1.13.0"},
	ai.ClassifierRequest{
		State: json.RawMessage(`{"message":"I cannot sign in."}`),
		Questions: map[string]ai.ClassifierQuestion{
			"needs_help": ai.BoolQuestion{Instructions: json.RawMessage(`"Does the user need support?"`)},
		},
	}, nil)
if err != nil { return err } // classified retains reported usage on answer validation failure
fmt.Println(classified.Answers["needs_help"].(ai.BoolAnswer).Probability)
```

For OpenAI, an empty `ReferenceImages` generates; inline `ai.Image` references select JSON editing.
`OpenAIImagesOptions` covers quantity, size, quality, background, format, compression, mask, input
fidelity and moderation, subject to model capabilities. The built-in snapshot allows one
reference/output, 1024×1024, low/medium quality, mask and transparency; input fidelity is omitted.

For Google, use `OperationImage`, `ProviderGoogle`, `APIGoogleInteractions`, endpoint
`https://generativelanguage.googleapis.com/v1beta`, and model `gemini-nano-banana-2.1`.
`GoogleImagesOptions` exposes `AspectRatio` and `ImageSize`; the built-in model allows one
reference/output, `1:1` and `1K`. Calls are synchronous with `store=false`, no continuation, and
inline output; the request omits `delivery` (ADR-0023). Exact output count is not guaranteed.
Neither route reads files, downloads URLs or accepts provider file IDs.

Classification accepts JSON strings, objects or arrays as state and question descriptions.
`ChoiceQuestion` returns `ChoiceAnswer` (choice, distribution, confidence); `ScoreQuestion` returns
`ScoreAnswer` (expected score, confidence, optional distribution and legend); `BoolQuestion` returns
the probability of yes. Score criteria are ordered, indexed from zero. Answers are validated as one
collection against the final questions after hooks; invalid answers are rejected without
renormalization. `TypeSafeOptions{}` is empty. The [Chinese evaluation
host](examples/chineseeval/README.md) reports business results separately. The [24/24 synthetic
evaluation](../.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/README.md) does not establish
production accuracy.

## Core concepts

```
Host ──CallScope + Target + operation request/options──▶ Client
  1. Entry fixes Operation; validate scope and input bounds
  2. Resolve Binding; check operation, typed catalog and AllowedModels
  3. Validate options/input; chat only: prepare history and native state
  4. Resolve Credential; pin consistent binding + credential + catalog
  5. Freeze the request after hooks; per attempt: admission → protocol → provider
  6. chat: assemble events/message; unary: read and validate complete response
     → Result / Stream / ImagesResult / ClassifierResult, each with CallMetadata
```

| Concept | Go type | Notes |
| --- | --- | --- |
| Call scope | `CallScope` | `TenantID` and `RequestID` required. The entry sets `Operation`. `RequestID` is globally unique per logical call; retries reuse it, each new call needs a new one. It is attribution, not deduplication — see [`examples/requestid`](examples/requestid). |
| Target | `Target` | `BindingID` + `ModelID`. Protocol, endpoint, account and key come from the binding, never from the request. |
| Binding | `Binding` | Fixes one operation, provider, API, endpoint and account, plus model/tool allowlists and `Retry` (zero = never retry). `Operation` zero means **chat only**; unknown nonempty values fail. `Enabled` and `Version` remain required. |
| Credential | `Credential` | Versioned snapshot; `BindingVersion` must equal the resolved binding's version, otherwise the call fails (`credential_unavailable`/`consistency`, ADR-0003). Keys are `ai.Secret` and never appear in output. |
| Requests | `Request`, `ImagesRequest`, `ClassifierRequest` | Chat history/tools; image prompt/references; classifier state/questions. Mutable input is independently copied; unary checks known bounds before copying. |
| Options | `ResponsesOptions`, `AnthropicOptions`, `GeminiOptions`, `ChatOptions` / `SimpleOptions`; `OpenAIImagesOptions`, `GoogleImagesOptions`, `TypeSafeOptions` | Full options must match the binding's API. `Nullable[T]` distinguishes unset / `null` / zero; image options reject explicit null. |
| Results | `Result`, `ImagesResult`, `ClassifierResult` | Each carries `Metadata`: operation, attribution, versions, catalog hash, attempts and usage reporting. Chat retains a valid message on failure; unary content/answers are empty while reported usage survives. |
| Catalog | `Model`, `ImageModel`, `ClassifierModel` | `Catalog.Lookup` / `ModelsOf` are chat-only. Use `LookupImage` / `ImageModelsOf` and `LookupClassifier` / `ClassifierModelsOf` for other operations. `Client.Catalog()` returns an independent snapshot. Discovery grants no authorization. |

Using a binding with the wrong entry fails with `tenant_denied/capability` before reading
credentials or making an attempt. Bindings of the same provider and account may share a credential
reference while authorizing distinct operations.

### Entry points

```go
res, err := client.Complete(ctx, scope, target, req, opts)        // full protocol options
res, err := client.CompleteSimple(ctx, scope, target, req, simple) // protocol-neutral options
s := client.Stream(ctx, scope, target, req, opts)
s := client.StreamSimple(ctx, scope, target, req, simple)
hc := client.WithHooks(hooks) // trusted request callbacks; all six entry points
```

`hc` also exposes `GenerateImages` and `Classify`. Header/payload hooks run once per logical call;
retries reuse the frozen request. Unary response hooks run once after a successful HTTP response,
before reading its body. Gemini chat continues to omit the response hook.

Use `Complete*` when you only need the final message. A `Stream` queues every event; one whose
events are never read ends with `resource_limit` (`PhaseEventQueue`) once its output outgrows
`MaxQueuedEvents` / `MaxQueuedEventBytes`.

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

Events are pi's: `start`; interleaved `text_*`, `thinking_*`, `toolcall_*` (`start`/`delta`/`end`)
with stable content indexes; exactly one terminal `done` or `error`.

### Tool round trip

barness-ai never runs a tool. A turn ending with `StopReasonToolUse` hands the calls to the host,
which validates each with `ValidateToolCall`, runs it, answers with `ToolResultText`, and starts the
next turn as a new logical call. `ToolCall.Arguments` is a best-effort parse for display only; never
execute it. Turns ending in `length`, `error` or `aborted` are never executed. See
[`examples/toolloop`](examples/toolloop).

### Native state

Messages produced in this process carry trusted provenance and replay as is. Messages decoded from
storage (JSON) are untrusted until the host calls `TrustNativeState(scope, envelope)` after its own
checks. Untrusted or mismatched (tenant, account, model) native state is downgraded by pi's
cross-model rules, never refused; downgrades are counted in `Metadata`.

### Errors

A failed call returns its operation-specific result plus `*ai.Error` with a stable `Code` and
`Phase`:

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

`Error.Message` quotes the provider's error body as pi does (key-shaped text redacted) and may echo
request content: return it only to the calling tenant, never to shared logs. Use the `Observer` for
logging. The full table is in [contract.md §4](../docs/barness-ai/contract.md).

Unary reading, decoding and output validation fail in `PhaseResponse`. Successful HTTP responses are
never replayed after those failures; a started chat stream is never replayed either. Retries apply
only to the initial request when `Binding.Retry` explicitly allows them.

### Resource policy, admission, time limits

Every `Client` requires an explicit, finite `ResourcePolicy` — there are no defaults and no
"unlimited" mode (ADR-0002). Byte limits are enforced while encoding and reading. Each attempt,
retries included, takes a per-tenant and per-process permit, then the host's `Config.Admission` if
set (ADR-0008). A call ends at the earliest of the context deadline, `CallTimeout`, and the
per-attempt connect / header / read-idle / protocol timeouts.

Annotated, pressure-tested starting points: `localassembly.LocalPolicy`,
`hostintegration.CloudInteractivePolicy`, `hostintegration.CloudBatchPolicy`. Every value's
rationale is in their source comments; measurements are in
[docs/barness-ai/README.md](../docs/barness-ai/README.md).

`ResourcePolicy.Image == nil` disables images; `Classifier == nil` disables classification. Enabled
child capacities must be positive and are deep-copied: image input/output counts and per-image/total
output bytes; classifier question count, state bytes and question bytes. Global request/output
bytes, concurrency and timeouts apply to all operations. Unary JSON uses `MaxOutputBytes`, rather
than the SSE-only `MaxFrameBytes`; effective limits also respect protocol/model caps.

For concurrent chat, image and classifier calls on one Client, choose `localassembly.MixedPolicy()`
or `hostintegration.CloudMixedPolicy()`. They explicitly enable finite subpolicies at 4/8 concurrent
attempts, sized for complete unary JSON and base64 buffers. `OpenOperations` assembles independent
bindings and named key sources. Cloud `Host.GenerateImages` and `Host.Classify` build scope from
Principal and preserve downstream cancellation. The host handles typed results/errors, confidence
thresholds and human review. Run `BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run '^TestMixed' -count=1`
for the design load; without the flag pressure cases are NOT_RUN.

### Observability, usage and cost

`Config.Observer` receives `call_started`, `attempt_started`, `attempt_finished`, `call_finished`
asynchronously through a bounded queue; a slow observer never blocks a call, drops are counted in
`Client.ObserverStats()`. Records hold no keys, content, tool arguments or error text (ADR-0009).
`Usage` keeps pi's numbers; `Usage.Cost` is an estimate from catalog prices, not billing. Whether
usage was reported at all is `Attempt.UsageReporting` — a zero `Usage` never means free (ADR-0010).

Image calls add optional `Usage.Modalities` token counts and estimate cost by text/image rates;
missing details remain partial or unreported. Google thought tokens count once within output and
separately as reasoning. TypeSafe charges only input tokens. Observations include `Operation` and
usage, never prompts, images, classifier state or answers.

### Chat behavior at the pi-ai 1.0.0 baseline

- Unparseable/nonfinite retry delays use bounded exponential backoff with jitter.
- Responses cost uses the reported service tier, falling back to options: flex
  is 0.5×; priority/fast is 2× (2.5× for gpt-5.5).
- Anthropic's later 1-hour cache-write detail replaces the earlier value.
- Responses/Chat merge model `SamplingParams` defaults once, then call overrides;
  full/simple entries agree and reserved authorization/state fields stay protected.
- Responses rejects unfinished tool calls, including blocks replaced by duplicate
  output indexes. The host must still validate tool arguments before execution.

## Cloud host integration

[`examples/hostintegration`](examples/hostintegration) is the reference. Checklist:

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

Never accept keys, credential refs, endpoints, `Authorization` headers, proxies or callbacks from
request JSON — they come only from trusted assembly.

---

## For coding agents

Read this section, [GLOSSARY.md](../GLOSSARY.md) and the ADRs touching your area before editing. The
repo-wide rules in [AGENTS.md](../AGENTS.md) apply (500-line file limit, delete rather than keep
compatibility shims, E2E-first testing).

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
   Never retry a resource-limit failure, a started stream or a unary 2xx result; never let an SDK
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
9. **Observers never affect calls; hooks may fail but never widen** operation,
   auth, target, model, native references or hosted tools. Revalidate final unary input (ADR-0005).
10. **Tool arguments are not executable** until `ValidateToolCall` passes on a
    finished message.
11. **Operations remain separate.** Zero binding operation authorizes chat only;
    typed catalog lookup precedes credential resolution (ADR-0020).
12. **Unary output is atomic.** Publish all validated images/answers together;
    retain reported usage and release permits on every failure.

### Package layout

All production code is the single package `ai` (flat, files grouped by prefix). Unexported types do
not leak across these groups by convention.

| Area | Files |
| --- | --- |
| Shared call runtime | `client.go`, `call_runtime.go`, `unary*.go` (lifetime, pinned snapshot, unary HTTP/JSON) |
| Chat entry & stream | `entry.go`, `call.go`, `stream.go`, `event.go`, `assembler.go`, `partial*.go` |
| Domain types | `operation.go`, `message.go`, `request.go`, `result.go`, `binding.go`, `options.go`, `nullable.go`, `reasoning.go`, `errors.go` |
| Resolution & security | `preflight.go`, `policy.go`, `admission.go`, `limits.go`, `timeouts.go`, `retry.go`, `transport.go` |
| History & native state | `transcript.go`, `history.go`, `replay.go`, `native.go`, `cache_key.go` |
| Tools | `tool.go`, `tool_validate.go`, `tool_coerce.go`, `tool_json.go`, `argument_text.go` |
| Catalog | `catalog*.go` (typed models, queries, validation, built-in data, snapshots and prices) |
| Usage, cost | `usage.go`, `estimate.go`, `anthropic_effort.go` |
| Hooks & observability | `hooks.go`, `hooks_run.go`, `observer.go` |
| Chat adapter contract | `adapter.go` (`adapter` interface + fixed `registry()` keyed by `API`); unary routes dispatch explicitly by operation/protocol in `images.go` / `classifier.go` |
| OpenAI Responses | `responses*.go`, `openai_sdk.go` |
| Anthropic Messages | `anthropic*.go` |
| Gemini Developer API | `gemini*.go` (REST, no SDK) |
| Chat Completions | `chat*.go` (`chat_compat.go` holds per-provider differences) |
| Native images | `images.go`, `image*.go`, `openai_images*.go`, `google_images*.go` |
| TypeSafe classifier | `classifier*.go`, `typesafe*.go` |
| Shared HTTP/SDK glue | `sdk_middleware.go`, `http_failures.go` |

Other directories:

| Directory | Purpose |
| --- | --- |
| `examples/` | `localassembly`, `hostintegration` (chat/mixed operations), `toolloop`, `requestid`, `chineseeval` (independent business evaluation); exercised by offline E2E |
| `e2e/` | Offline acceptance tests against the public API, a local controlled provider and a host double; fixtures in `e2e/testdata/<protocol>/` |
| `live/` | Real-API smoke (build tag `live`), feeds `support-matrix.json` |
| `release/` | Release gate command, catalog snapshot, traceability map |
| `internal/testkit/` | Test-only: controlled provider, host double, pi oracle (Node), evidence, audit, release gate logic |
| `internal/probe`, `internal/clock` | Test hooks injected through `Config` (hosts cannot construct them) |

### Common changes

- **Add a model:** update `catalog_builtin.go`, `catalog_images.go` or `catalog_typesafe.go`
  under ADR-0018; regenerate with `go run ./ai/release/cmd/releasegate -write-snapshot` and review.
- **Add a provider on an API:** add its ID/capabilities, fixtures, live combo, ADR and
  ledger entry when pi has no route.
- **Add a protocol:** define the operation, API, typed options/results, E2E and ADR.
  Chat uses `adapter` / `registry()`; unary uses shared runtime and explicit dispatch.
  Enforce encoding/reading/validation limits; defer registration abstractions until needed.
- **Change observable behavior:** run the pi differential, update ledger and `differences.md`.

### Testing

Per AGENTS.md, E2E is the primary mechanism. Tests drive the public `Client` only; nothing in `e2e/`
reaches inside the package.

```sh
go test ./ai/...                               # offline E2E + package tests, no network
go test -race ./ai/...
BARNESS_AI_PRESSURE=1 go test ./ai/e2e         # policy pressure scenarios
BARNESS_AI_PIDIFF=1  go test ./ai/e2e          # differential against frozen pi
#   needs Node.js and: npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node

# Live smoke: one combo per process, only that combo's key
BARNESS_AI_LIVE=1 BARNESS_AI_LIVE_COMBO=deepseek-chat \
BARNESS_AI_LIVE_ACCOUNT_ALIAS='test-account' \
go test -tags live -count=1 ./ai/live
# Inject BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT into that process beforehand.
go run ./ai/live/cmd/supportmatrix <bundle-dir>...   # merge reports into the matrix

go run ./ai/examples/chineseeval/cmd/chineseeval       # NOT_RUN bundle; no network
go run ./ai/examples/chineseeval/cmd/chineseeval -verify <evaluation-bundle>
go run ./ai/release/cmd/releasegate -evaluation <Chinese-evaluation-bundle> [-live <bundle-dir>]...
```

Every E2E run writes a redacted evidence bundle under `.evidence/` (`BARNESS_AI_EVIDENCE_DIR`
overrides). New behavior needs an E2E case whose evidence the release gate can trace
(`release/traceability.json`). The full release gate requires each of the nine live routes and an
independent real Chinese evaluation; offline fixtures and NOT_RUN reports cannot replace them.

## Further reading

| Document | Content |
| --- | --- |
| [docs/barness-ai/contract.md](../docs/barness-ai/contract.md) | Public contract, full error table, assembly notes |
| [docs/barness-ai/differences.md](../docs/barness-ai/differences.md) | Every difference from pi-ai and its disposition |
| [docs/barness-ai/README.md](../docs/barness-ai/README.md) | Release deliverables, gate status, policy measurements |
| [docs/adr/](../docs/adr/) | ADR-0001 … ADR-0024 |
| [GLOSSARY.md](../GLOSSARY.md) | Domain terms (Tenant, Binding, Logical Call, Native State, …) |
| `go doc -all ./ai` | Package and type documentation |

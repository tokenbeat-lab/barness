# barness-ai

> Follow the [testing SOP](../docs/agents/verification.md) for current verification, backed by [ADR-0051](../docs/adr/0051-scoped-verification-receipts.md); see the [single current status](../.scratch/barness-ai-test-process-review/status.md). Historical acceptances below apply only to their measured source.

> **Catalog and release status (2026-10-19):** `2026-10-18.1-minimax` contains 96 chat identities, two image models and three classifiers. The release contract requires 16 protocol combinations and 17 independent own support identities. China standard GLM-5.3/5.2 keep separate Chat, Responses and Messages bindings; [Host examples](examples/glmhost/README.md) cover standard keys and full/simple/WithHooks. Jev1.14 images and Box retain experimental-image and explicit-authorization requirements. The current support matrix has 17 PASS identities; current-source qualification, the complete release gate and final reviews are **COMPLETE**. See the [delivery status](../.scratch/barness-ai-minimax/delivery-status.md).

> **MiniMax international:** `BuiltinCatalog()` includes `MiniMax-M3` on `minimax` × `openai-completions`, `openai-responses` and `anthropic-messages`, using `api.minimax.io/v1` or `api.minimax.io/anthropic`. The three own [current formal reports](../.scratch/barness-ai-minimax/evidence/ticket-19/actual/remaining/README.md) pass with source applicability verified at the final measured source. M3.1 remains an unpromoted candidate: its own online reports are NOT_RUN, with M Plan entitlement and production price unconfirmed. [Host examples](examples/minimaxhost/README.md) require an explicit account and credential purpose. The [final closure](../.scratch/barness-ai-minimax/evidence/ticket-21/final-closure/README.md) records eleven qualified source groups, the passing complete gate and exact matrix publication; original failures and older source maps remain historical.

English | [简体中文](README.zh-CN.md)

`github.com/tokenbeat-lab/barness/ai` is barness's model protocol module: one trusted, concurrency-safe `Client` that
runs **one model operation** against a model provider for **many tenants**. It is a Go port of
[pi-ai](https://github.com/earendil-works/pi), used as an architecture reference, with tenant isolation, explicit resource limits
and observability added on top.

This README covers the public API and host responsibilities. Coding agents should also read [the invariants, layout and
testing guide](#for-coding-agents).

Design requirements live in the [operation spec](../.scratch/barness-ai-pi-1.0/spec.md), [thinking-level sampling
spec](../.scratch/barness-ai-thinking-sampling/spec.md), [Decisions
spec](../.scratch/barness-ai-openai-decisions/spec.md) and [ADRs](../docs/adr/). The API below follows the current
implementation (`go doc ./ai`); support claims follow the catalog and each route's own evidence. Terms are defined in
[GLOSSARY.md](../GLOSSARY.md).

## Model operations

| Operation | Entry points | Request → result | Delivery |
| --- | --- | --- | --- |
| `chat` | `Complete`, `CompleteSimple`, `Stream`, `StreamSimple` | `Request` → `Result` / `Stream` | Final assistant message or pi-ai events |
| `image` | `GenerateImages` | `ImagesRequest` → `ImagesResult` | Synchronous, ordered text/image blocks |
| `classifier` | `Classify` | `ClassifierRequest` → `ClassifierResult` | Synchronous, typed answers to named questions |

`Client` and `HookedClient` expose all six entry points. Image and classifier calls honor context cancellation, emit no
delta events, and accept protocol options or `nil` for defaults; `SimpleOptions` is chat-only.

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

Non-goals: no agent loop, no built-in memory or history trimming, no tool execution, no logging, no reading of
`OPENAI_API_KEY`-style environment variables, no default resource limits.

## Supported combinations

| Operation | Provider × API | `ProviderID` / `API` | Own live report history |
| --- | --- | --- | --- |
| chat | OpenAI × Responses | `openai` / `openai-responses` | PASS |
| chat | Anthropic × Messages | `anthropic` / `anthropic-messages` | PASS |
| chat | Google × Gemini Developer API | `google` / `google-generative-ai` | PASS |
| chat | OpenAI × Chat Completions | `openai` / `openai-completions` | PASS (no replayable reasoning by protocol) |
| chat | DeepSeek × Responses (extension route, not in pi) | `deepseek` / `openai-responses` | PASS |
| chat | DeepSeek × Chat Completions | `deepseek` / `openai-completions` | PASS |
| classifier | TypeSafe × System One, jev-1.13.0 | `typesafe` / `typesafe-system-one` | PASS (context probe returned 400; 422 shape unconfirmed) |
| classifier | OpenAI × Decisions, gpt-6-luna | `openai` / `openai-decisions` | PASS (current source qualified) |
| classifier | TypeSafe × System One, jev-1.14.0 | `typesafe` / `typesafe-system-one` | PASS (current source qualified) |
| chat | China Zhipu × Chat, glm-5.3 / glm-5.2 | `zhipu` / `openai-completions` | PASS (current source qualified) |
| chat | China Zhipu × Responses, glm-5.3 / glm-5.2 | `zhipu` / `openai-responses` | PASS (current source qualified) |
| chat | China Zhipu × Messages, glm-5.3 / glm-5.2 | `zhipu` / `anthropic-messages` | PASS (current full 22-case report; old cancel failure retained) |
| chat | International MiniMax-M3 × Chat | `minimax` / `openai-completions` | PASS (current own formal evidence) |
| chat | International MiniMax-M3 × Responses | `minimax` / `openai-responses` | PASS (current own formal evidence) |
| chat | International MiniMax-M3 × Messages | `minimax` / `anthropic-messages` | PASS (current own formal evidence) |
| image | OpenAI × Images, gpt-image-2.5-sunburst-2026-09-08 | `openai` / `openai-images` | PASS (generation, JSON edit and mask) |
| image | Google × Interactions v1beta, gemini-nano-banana-2.1 | `google` / `google-interactions` | PASS (generation and reference edit, 1K / 1:1) |

The matrix records each identity's own source/catalog-scoped evidence. Jev1.13 and1.14 share one protocol combination and retain separate required identities. GLM routes are Extension Routes with their own evidence, alongside DeepSeek Responses and both native image routes; none claims frozen pi parity. Callable models remain the intersection of the full `(Operation, Provider, API, ModelID)` catalog identity and the binding allowlist.

### Historical verification and release baselines

The [recorded MiniMax delivery](../.scratch/barness-ai-minimax/delivery-status.md) is COMPLETE: at measured source `78843eb`, both vet modes, ordinary and race commands exited 0; each main suite has 12,495 PASS, 0 FAIL, 0 NOT_RUN and audit0. All 17 own support inputs and four effects pass. Eleven source groups qualify; GLM Messages cancel passes in the full report at 1146ms. Later documentation/evidence commits preserve all 1,078 selected source files and are not new Go executions. The following baselines retain their original historical results.

The previous complete baseline is [ticket14's accepted release](../.scratch/jev-1.14-multimodal-box/14-remaining-real/verification.md):11 own live identities, four independent effects, normal/race9098 each,861 frozen pi cases and all gates PASS. The original [pressure issue07](../.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md) is resolved with retained causal evidence; earlier failing releases remain history.

[Historical ticket13 acceptance](../.scratch/barness-ai-glm/evidence/ticket-13/README.md) was PASS at `3d3cd0f`: normal 11,049 and race 11,049 cases, all ten gates, both vet modes and 20 clean input audits. That catalog `2026-10-17.2` retained fourteen own live identities and four independent effects. This ticket used 282 actual HTTP calls: 66 GLM, 180 in the closed original legacy round, and 36 necessary corrective calls. Original failures/captures remain historical. The [verified index](../.scratch/barness-ai-glm/evidence/ticket-13/index.json) records exact source applicability, finite requirement/story traces and full baseline rollback.

## Quick start (local)

Requires **Go 1.26.2 or later** (see [go.mod](../go.mod)). Examples below are function-body snippets using `context`,
`fmt`, `encoding/json`, `ai`, and `ai/examples/localassembly`; callers supply `ctx` and handle the returned error.

A local program uses the same `Client` as a cloud host; it just states explicitly the one tenant, binding and key a
cloud host would resolve. [`examples/localassembly`](examples/localassembly) does that assembly:

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

`OpenOperations` assembles several bindings on one Client. It requires an explicit policy and preserves each binding's
`Enabled` choice. The program reads only the key sources it names, indexed by `CredentialRef`:

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
			ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIDecisions,
			Endpoint: "https://api.openai.com/v1", AccountScopeID: "my-openai-account",
			CredentialRef: "openai-key", AllowedModels: []string{"gpt-6-luna"},
		},
	},
	Keys: map[string]localassembly.KeySource{
		"openai-key": {EnvVar: "BARNESS_OPENAI_KEY"},
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
	ai.Target{BindingID: "classifier", ModelID: "gpt-6-luna"},
	ai.ClassifierRequest{
		State: json.RawMessage(`{"message":"I cannot sign in."}`),
		Questions: map[string]ai.ClassifierQuestion{
			"needs_help": ai.BoolQuestion{Instructions: json.RawMessage(`"Does the user need support?"`)},
		},
	}, ai.OpenAIDecisionsOptions{}) // nil also selects protocol defaults
if err != nil { return err } // classified retains reported usage on answer validation failure
switch answer := classified.Answers["needs_help"].(type) {
case ai.BoolAnswer:
	fmt.Println(answer.Probability)
case ai.RefusalAnswer:
	fmt.Println("needs_help: refused") // the host decides whether to ask again or seek human review
}
```

For OpenAI, an empty `ReferenceImages` generates; inline `ai.Image` references select JSON editing.
`OpenAIImagesOptions` covers quantity, size, quality, background, format, compression, mask, input fidelity and
moderation, subject to model capabilities. The built-in snapshot allows one reference/output, 1024×1024, low/medium
quality, mask and transparency; input fidelity is omitted.

For Google, use `OperationImage`, `ProviderGoogle`, `APIGoogleInteractions`, endpoint
`https://generativelanguage.googleapis.com/v1beta`, and model `gemini-nano-banana-2.1`. `GoogleImagesOptions` exposes
`AspectRatio` and `ImageSize`; the built-in model allows one reference/output, `1:1` and `1K`. Calls are synchronous
with `store=false`, no continuation, and inline output; the request omits `delivery` (ADR-0023). Exact output count is
not guaranteed. Neither route reads files, downloads URLs or accepts provider file IDs.

### Classifier inputs and answers

| Built-in classifier | State / descriptions | Questions and options |
| --- | --- | --- |
| TypeSafe `jev-1.13.0` | Text-only; JSON string/object/array state and descriptions | Choice: 1–255; score: 2–10 ordered levels; bool; `TypeSafeOptions{}` |
| TypeSafe `jev-1.14.0` | Text and explicitly experimental PNG/JPEG at context/State/Instructions/Criteria | Choice, score, bool and Box; `TypeSafeOptions{ExperimentalImages: true}` for images/Box |
| OpenAI `gpt-6-luna` | Text and authorized inline images; JSON string/object/array state, string descriptions | Choice: 2–255; score: 2–10 ordered levels; bool; `OpenAIDecisionsOptions{}` |

Both use `OperationClassifier`; `nil` selects their protocol defaults. Typed nil or the wrong protocol's options fail
before credential access. Decisions converts string state to raw text and object/array state to compact JSON without
rounding numbers. Its catalog context window is 922,000 (frozen pi metadata); the undated model alias's weights are not
pinned by the catalog hash.

`ChoiceAnswer` retains choice, distribution and confidence; `ScoreAnswer` retains expected score, distribution and any
verified legend; `BoolAnswer` is the probability of yes. Score levels start at zero. Decisions may return
`RefusalAnswer{}` (`{"type":"refusal"}`) with `stop` and nil error; mixed or all-refusal results retain every question
name. Hosts handle refusal explicitly and own thresholds, retrying as a new logical call, and human review. Answers are
validated atomically against the final frozen hook request; invalid answers clear the collection but preserve valid
usage.

#### Inline classifier images

`ClassifierRequest.Images` requires three independent opt-ins:

1. `ResourcePolicy.Classifier.Images` with positive `MaxImages` / `MaxTotalImageBytes`.
2. Model image input capability, `MaxInputImages` and allowed MIME types. Built-in Luna allows
   at most 128 PNG/JPEG/WebP images; the host policy may be smaller. Jev 1.13 rejects images; fixed Jev 1.14 additionally requires the experimental image option.
3. `Binding.AllowClassifierImages: true`, granted by trusted host configuration.

The example binding above remains text-only until the third opt-in is set. Supply ordered `ai.Image{Data: base64Data,
MimeType: "image/png"}` values and a required JSON `State` (`{}` works). Strict base64 and matching image headers are
checked without allocating a raster. `MaxImageBytes` bounds each decoded image, the child policy bounds total decoded
bytes, and `MaxRequestBytes` bounds the complete native JSON including base64/data URLs. Hooks obey the same pinned
permission and budgets. Empty images retain the text-only wire; no URLs, paths, file IDs or automatic resize.

See [the Decisions contract](../docs/barness-ai/decisions.md) and [executable Jev/Luna migration example](examples/classifiermigration/README.md) for final hook validation, errors and rollback. The [Chinese evaluation host](examples/chineseeval/README.md) reports each route's business results separately; 24/24 on its fixed synthetic task is not a production accuracy guarantee.

### Unified image Host example

Run `go run ./ai/examples/classifierimages` for a credential-free, controlled HTTP replay through both public Classify entries. The [complete Host example](examples/classifierimages/README.md) uses the same PNG/JPEG context request with only Target/Options changed, then shows explicit reuse, scoped images, visual legends, Box, refusal, revocation and replayable audited evidence. Fixed built-in `jev-1.14.0` supports experimental image locations and Box; finite image policies, model/Binding authorization and `TypeSafeOptions{ExperimentalImages: true}` remain required. Decisions rejects local images and Box before credential access. Controlled PASS is not real-service or quality PASS; ADR-0030 remains proposed; fixed 1.14 support requires its own independently verified formal evidence.

`ImageReferences` declares complete ordered uses; `value` pointers address existing null slots. Source and expanded occurrence budgets are independent (`ClassifierImagePolicy.References`), and final Hooks obey the same authorization and limits. `ClassifierResult.ImageInputs` identifies final images by MIME/decoded-file SHA-256; never use entry indices. `ScoreAnswer.LegendImageReferences` pointers are relative to Legend and index those final identities. `BoxAnswer` retains required nullable Confidence and normalized edges independently of a Bool existence answer. Hosts explicitly associate the search image, apply actual orientation/crop/pixel transforms and choose business thresholds. A failed call publishes neither Answers nor ImageInputs while retaining independently valid usage. Revoke Binding/image permission in the Host store for new calls; no automatic protocol fallback. Live example mode requires explicit `-live -route -config -key-env` Host injection and labels its result OBSERVED, not a support verdict.

## Core concepts

| Concept | Go type | Notes |
| --- | --- | --- |
| Call scope | `CallScope` | `TenantID` and `RequestID` required. The entry sets `Operation`. `RequestID` is globally unique per logical call; retries reuse it, each new call needs a new one. It is attribution, not deduplication — see [`examples/requestid`](examples/requestid). |
| Target | `Target` | `BindingID` + `ModelID`. Protocol, endpoint, account and key come from the binding, never from the request. |
| Binding | `Binding` | Fixes one operation, provider, API, endpoint and account, plus model/tool allowlists and `Retry` (zero = never retry). `Operation` zero means **chat only**; unknown nonempty values fail. `Enabled` and `Version` remain required. |
| Credential | `Credential` | Versioned snapshot; `BindingVersion` must equal the resolved binding's version, otherwise the call fails (`credential_unavailable`/`consistency`, ADR-0003). Keys are `ai.Secret` and never appear in output. |
| Requests | `Request`, `ImagesRequest`, `ClassifierRequest` | Chat history/tools; image prompt/references; classifier state/questions/optional images. Mutable input is independently copied; unary checks known bounds before copying. |
| Options | `ResponsesOptions`, `AnthropicOptions`, `GeminiOptions`, `ChatOptions` / `SimpleOptions`; `OpenAIImagesOptions`, `GoogleImagesOptions`; `TypeSafeOptions`, `OpenAIDecisionsOptions` | Full options must match the binding's API. `Nullable[T]` distinguishes unset / `null` / zero; image options reject explicit null. |
| Results | `Result`, `ImagesResult`, `ClassifierResult` | Each carries `Metadata`: operation, attribution, versions, catalog hash, attempts and usage reporting. Chat retains a valid message on failure; unary content/answers are empty while reported usage survives. |
| Catalog | `Model`, `ImageModel`, `ClassifierModel` | `Catalog.Lookup` / `ModelsOf` are chat-only. Use `LookupImage` / `ImageModelsOf` and `LookupClassifier` / `ClassifierModelsOf` for other operations. `Client.Catalog()` returns an independent snapshot. Discovery grants no authorization. |

Using a binding with the wrong entry fails with `tenant_denied/capability` before reading credentials or making an
attempt. Bindings of the same provider and account may share a credential reference while authorizing distinct
operations.

### Hooks and streaming

`client.WithHooks(hooks)` returns a `HookedClient` with all six entry points. Header/payload hooks run once per logical
call; retries reuse the frozen request. Unary response hooks run once after a successful HTTP response, before reading
its body. Gemini chat continues to omit the response hook.

Use `Complete*` when you only need the final message. A `Stream` queues every event; one whose events are never read
ends with `resource_limit` (`PhaseEventQueue`) once its output outgrows `MaxQueuedEvents` / `MaxQueuedEventBytes`.

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

Events are pi's: `start`; interleaved `text_*`, `thinking_*`, `toolcall_*` (`start`/`delta`/`end`) with stable content
indexes; exactly one terminal `done` or `error`.

### Tool round trip

barness-ai never runs a tool. A turn ending with `StopReasonToolUse` hands the calls to the host, which validates each
with `ValidateToolCall`, runs it, answers with `ToolResultText`, and starts the next turn as a new logical call.
`ToolCall.Arguments` is a best-effort parse for display only; never execute it. Turns ending in `length`, `error` or
`aborted` are never executed. See [`examples/toolloop`](examples/toolloop).

### Native state

Messages produced in this process carry trusted provenance and replay as is. Messages decoded from storage (JSON) are
untrusted until the host calls `TrustNativeState(scope, envelope)` after its own checks. Untrusted or mismatched
(tenant, account, model) native state is downgraded by pi's cross-model rules, never refused; downgrades are counted in
`Metadata`.

### Errors

A failed call returns its operation-specific result plus `*ai.Error` with a stable `Code` and `Phase`:

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

`Error.Message` quotes the provider's error body as pi does (key-shaped text redacted) and may echo request content:
return it only to the calling tenant, never to shared logs. Use the `Observer` for logging. The full table is in
[contract.md §4](../docs/barness-ai/contract.md).

Unary reading, decoding and output validation fail in `PhaseResponse`. Successful HTTP responses are never replayed
after those failures; a started chat stream is never replayed either. Retries apply only to the initial request when
`Binding.Retry` explicitly allows them.

### Resource policy, admission, time limits

Every `Client` requires an explicit, finite `ResourcePolicy` — there are no defaults and no "unlimited" mode (ADR-0002).
Byte limits are enforced while encoding and reading. Each attempt, retries included, takes a per-tenant and per-process
permit, then the host's `Config.Admission` if set (ADR-0008). A call ends at the earliest of the context deadline,
`CallTimeout`, and the per-attempt connect / header / read-idle / protocol timeouts.

Annotated, pressure-tested starting points: `localassembly.LocalPolicy`, `hostintegration.CloudInteractivePolicy`,
`hostintegration.CloudBatchPolicy`. Every value's rationale is in their source comments; measurements are in
[docs/barness-ai/README.md](../docs/barness-ai/README.md).

`ResourcePolicy.Image == nil` disables image generation; `Classifier == nil` disables classification. `Classifier.Images
== nil` separately disables classifier image inputs. Enabled child capacities must be positive and are deep-copied:
image input/output counts and per-image/total output bytes; classifier question count, state bytes and question bytes.
Global request/output bytes, concurrency and timeouts apply to all operations. Unary JSON uses `MaxOutputBytes`, rather
than the SSE-only `MaxFrameBytes`; effective limits also respect protocol/model caps.

For concurrent chat, image and classifier calls on one Client, choose `localassembly.MixedPolicy()` or
`hostintegration.CloudMixedPolicy()`. They explicitly enable finite subpolicies at 4/8 concurrent attempts, sized for
complete unary JSON and base64 buffers. `OpenOperations` assembles independent bindings and named key sources. Cloud
`Host.GenerateImages` and `Host.Classify` build scope from Principal and preserve downstream cancellation. The host
handles typed results/errors, confidence thresholds and human review. Mixed and classifier-image design loads run only in an explicitly selected pressure stage. Select race separately; pressure cases remain NOT_RUN without opt-in.

### Observability, usage and cost

`Config.Observer` receives `call_started`, `attempt_started`, `attempt_finished`, `call_finished` asynchronously through
a bounded queue; a slow observer never blocks a call, drops are counted in `Client.ObserverStats()`. Records hold no
keys, content, tool arguments or error text (ADR-0009). `Usage` keeps pi's numbers; `Usage.Cost` is an estimate from
catalog prices, not billing. Whether usage was reported at all is `Attempt.UsageReporting` — a zero `Usage` never means
free (ADR-0010).

Image calls add optional `Usage.Modalities` token counts and estimate cost by text/image rates; missing details remain
partial or unreported. Google thought tokens count once within output and separately as reasoning. TypeSafe charges only
input tokens. Observations include `Operation` and usage, never prompts, images, classifier state or answers.

Decisions keeps inclusive `input_tokens` and reported output, with canonical cache buckets at zero. Reporting is
complete only when input, output, total, cached, cache-write and reasoning counts are all explicit, including zero;
missing/null counts remain partial, and absent/null usage is unreported. Reasoning is unset unless explicitly reported.
Invalid or ambiguous usage fails the response with unreported usage; a damaged answer retains independently valid usage
and its input-only estimate. The authorized classifier catalog supplies the input rate and full-request tiers. Built-in
Luna uses $0.10 per million through 272000 inclusive inputs, then $0.20 for all inputs. Output/cache add no charge.
Regional premiums and actual bills are not reconciled by this estimate; synthetic cache fixtures prove conversion only.
See the [usage contract](../docs/barness-ai/contract.md#decisions-用量与估价工单-04).

### Chat behavior at the pi-ai 1.1.0 baseline

- Unparseable/nonfinite retry delays use bounded exponential backoff with jitter.
- Responses cost uses the reported service tier, falling back to options: flex
  is 0.5×; priority/fast is 2× (2.5× for gpt-5.5).
- Anthropic's later 1-hour cache-write detail replaces the earlier value.
- Responses/Chat merge named options → model defaults → canonical thinking profile → call overrides
  → `OnPayload` once. Profiles use `off/minimal/low/medium/high/xhigh/max`; native Anthropic/Gemini
  ignore both sampling maps. Full wire effort remains unchanged, while Simple clamps it separately;
  profile selection never runs again after a sampling/hook override. Built-in models declare no
  sampling defaults. See [the complete Host example](../docs/barness-ai/thinking-sampling.md).
- Final chat messages and terminal events include nonnegative integer `DurationMs` (`durationMs`),
  measured from call setup through settlement, including hooks and retries.
- Responses rejects unfinished tool calls, including blocks replaced by duplicate
  output indexes. The host must still validate tool arguments before execution.

## Cloud host integration

[`examples/hostintegration`](examples/hostintegration) is the reference. Checklist:

1. Authentication result → `CallScope` (new `RequestID`, `ActorID`). Refuse a self-claimed tenant or another tenant's session.
2. Read history by `(tenant, session)` and vouch for stored native state with `TrustNativeState`.
3. Implement `BindingResolver` / `CredentialResolver`. Key caches by tenant, credential ID and version; document TTL and revocation propagation.
4. Pick a cloud policy; answer `admission_denied` with 429 + `Retry-After`. Queueing stays outside barness-ai.
5. Inject `Admission` for account/distributed quotas, and `Observer`; watch `ObserverStats`.
6. Set `Binding.Retry` (ADR-0006). Retry only initial requests; never replay started streams.
7. Cancel the context or `Close` the stream on downstream disconnect. Route merged streams by `EventEnvelope.Call` (`hostintegration.Merge`).

Never accept keys, credential refs, endpoints, `Authorization` headers, proxies or callbacks from request JSON — they
come only from trusted assembly.

---

## For coding agents

Read this section, [GLOSSARY.md](../GLOSSARY.md) and the ADRs touching your area before editing. The repo-wide rules in
[AGENTS.md](../AGENTS.md) apply (500-line file limit, delete rather than keep compatibility shims, E2E-first testing).

### Invariants — do not break

1. **Trusted identity.** Never derive `CallScope` from request content, env or context values; every tenant-scoped resolver lookup receives the scope.
2. **Finite limits.** No default policy, limit bypass or unbounded buffers/queues/retries/waits. Every buffer has a policy limit and fails with `CodeResourceLimit`.
3. **Pinned snapshots.** All attempts use the resolved binding/credential/catalog snapshot; never re-resolve within a call (D2, ADR-0003).
4. **Initial-request retries only.** `Binding.Retry` governs them; never retry resource limits, started streams or unary 2xx failures. Disable SDK retries.
5. **Secrets never leave.** `Secret` cannot enter messages, events, errors, observations, evidence or logs. barness-ai writes no logs.
6. **Explicit assembly.** No ambient provider/proxy env, credential files or default endpoints. Require https; `AllowLoopbackHTTP` is test-only.
7. **Own contracts and evidence.** Use pi as an architecture and design reference. Verify barness behavior through its public contracts, E2E and own live evidence; see [ADR-0051](../docs/adr/0051-scoped-verification-receipts.md).
8. **Provider capabilities.** Shared adapters derive fields from the provider, never another provider's protocol assumptions.
9. **Bounded hooks.** Observers never affect calls; hooks cannot widen operation/auth/target/model/native references/hosted tools/image permissions. Revalidate final unary input (ADR-0005).
10. **Validated tools.** Execute only finished tool calls that pass `ValidateToolCall`.
11. **Separate operations.** A zero binding operation authorizes chat only; typed catalog lookup precedes credentials (ADR-0020).
12. **Atomic unary output.** Publish validated images/answers together; preserve reported usage and release permits on every failure.

### Package layout

Production code lives in the single package `ai`; files are grouped by responsibility and protocol prefix.

| Area | Files |
| --- | --- |
| Shared call runtime | `client.go`, `call_runtime.go`, `unary*.go` (lifetime, pinned snapshot, unary HTTP/JSON) |
| Chat entry & stream | `entry.go`, `call.go`, `stream.go`, `event.go`, `assembler.go`, `partial*.go` |
| Domain types | `operation.go`, `message.go`, `request.go`, `result.go`, `binding.go`, `options.go`, `nullable.go`, `reasoning.go`, `errors.go` |
| Resolution & security | `preflight.go`, `policy.go`, `admission.go`, `limits.go`, `timeouts.go`, `retry.go`, `transport.go` |
| History & native state | `transcript.go`, `history.go`, `replay.go`, `native.go`, `cache_key.go` |
| Tools | `tool.go`, `tool_validate.go`, `tool_coerce.go`, `tool_json.go`, `argument_text.go` |
| Catalog | `catalog*.go` (typed models, queries, validation, built-in data, snapshots and prices) |
| Usage, cost | `usage.go`, `estimate.go`, protocol `*_usage.go`; `anthropic_effort.go` handles managed effort |
| Hooks & observability | `hooks.go`, `hooks_run.go`, `observer.go` |
| Chat adapter contract | `adapter.go` (`adapter` interface + fixed `registry()` keyed by `API`); unary routes dispatch explicitly by operation/protocol in `images.go` / `classifier.go` |
| OpenAI Responses | `responses*.go`, `openai_sdk.go` |
| Anthropic Messages | `anthropic*.go` |
| Gemini Developer API | `gemini*.go` (REST, no SDK) |
| Chat Completions | `chat*.go` (`chat_compat.go` holds per-provider differences) |
| Native images | `images.go`, `image*.go`, `openai_images*.go`, `google_images*.go` |
| Classifier protocols | `classifier*.go`, independent `typesafe*.go` and `openai_decisions*.go` |
| Shared HTTP/SDK glue | `sdk_middleware.go`, `http_failures.go` |

Other directories:

| Directory | Purpose |
| --- | --- |
| `examples/` | `localassembly`, `hostintegration` (chat/mixed operations), `toolloop`, `requestid`, `classifiermigration`, `chineseeval` (independent business evaluation); exercised by offline E2E |
| `e2e/` | Offline acceptance tests against the public API, a local controlled provider and a host double; fixtures in `e2e/testdata/<protocol>/` |
| `live/` | Real-API smoke (build tag `live`), feeds `support-matrix.json` |
| `release/` | Read-only release receipt verifier and catalog snapshot |
| `internal/testkit/` | Test-only: controlled provider, host double, evidence, audit and read-only release verification |
| `internal/probe`, `internal/clock` | Test hooks injected through `Config` (hosts cannot construct them) |

### Common changes

- **Add a model:** update its `catalog_*.go` data (including `catalog_decisions.go`) under ADR-0018; review the updated `release/catalog-snapshot.json` content and public `BuiltinCatalog().Hash()` together; `TestBuiltinCatalogReleaseSnapshot` checks both.
- **Add a provider:** add ID/capabilities, fixtures, live combo and ADR; qualify the route using its own evidence.
- **Add a protocol:** define operation/API/options/results, E2E and ADR. Chat uses `adapter` / `registry()`; unary uses shared runtime and explicit dispatch. Bound encoding, reading and validation.
- **Change observable behavior:** write a public failing scenario first, then verify affected E2E; identify required live routes and authorization when vendor behavior changes.

### Testing

Follow the [testing SOP](../docs/agents/verification.md) before changing tests, selecting verification, evaluating release evidence or maintaining artifacts. It defines the ordered steps, command templates, completion criteria and failure handling. Public Client/CLI failure scenarios precede implementation; each attempt retains verifiable native evidence.

The [tool contract](../tools/verification/README.md) specifies receipt shapes and the read-only release CLI. Paid live, race and pressure retain their separate explicit selections; offline/controlled replay does not establish real support qualification. Current rollout facts and deferred execution are recorded in the [process status](../.scratch/barness-ai-test-process-review/status.md).

## Further reading

| Document | Content |
| --- | --- |
| [docs/barness-ai/contract.md](../docs/barness-ai/contract.md) | Public contract, full error table, assembly notes |
| [docs/barness-ai/differences.md](../docs/barness-ai/differences.md) | Historical pi-ai difference research |
| [docs/barness-ai/decisions.md](../docs/barness-ai/decisions.md) | Classifier inputs, refusals, image authorization, usage and rollback |
| [docs/barness-ai/thinking-sampling.md](../docs/barness-ai/thinking-sampling.md) | Thinking profiles, merge order and full/simple boundaries |
| [docs/barness-ai/README.md](../docs/barness-ai/README.md) | Release deliverables, gate status, policy measurements |
| [docs/adr/](../docs/adr/) | Current and historical ADRs |
| [GLOSSARY.md](../GLOSSARY.md) | Domain terms (Tenant, Binding, Logical Call, Native State, …) |
| `go doc -all ./ai` | Package and type documentation |

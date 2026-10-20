# Early GLM Messages contract probe plan

Authorization: parent delegation after the user supplied root `.env` GLM_KEY for requested China-standard API verification. This is a bounded exploration required by spec R13/R14, not formal six-entry acceptance.

## Fixed scope and budgets

- Only public `ai.Client.Stream` via `WithHooks`, with an explicit Host candidate Catalog, Binding and Credential snapshot. Provider ID `zhipu`; API `anthropic-messages`; one fixed base `https://open.bigmodel.cn/api/anthropic`; exact generated target `POST https://open.bigmodel.cn/api/anthropic/v1/messages`; `x-api-key`.
- Two sequential logical calls total, one each for `glm-5.3` and `glm-5.2`; at most two HTTP attempts in the capturing transport, no retries, no redirects, no alternate endpoint/protocol/model fallback. Each request has `max_tokens:4096`, `stream:true`.
- Each call: 120 seconds including setup; connection 15 seconds; response headers 90 seconds; read idle 90 seconds. Outer child process: 300 seconds. Maximum aggregate requested output budget 8192 tokens; cost is not guaranteed by this cap if upstream accounting differs.
- Pure synthetic prompt: `Calculate 17 + 25. Reply with only the integer answer.` No user content, tools, history, external files, hosted state, or advanced parameters.
- Account alias `local-glm`, region alias `cn-standard-endpoint`; aliases do not verify the key's actual account/region. Key purpose label `host-supplied-for-standard-purpose` is Host attestation, not provider entitlement verification.
- Initial budget counters: Messages 0 logical calls / 0 attempts; this exploration consumes at most 2 / 2 toward the spec's aggregate 24 / 32 limit. All future probes must add this consumed count.

## Source and reviewability before execution

- `/tmp/barness-glm-messages-probe/source/main.go`: public Client harness, explicit resource limits, immutable candidate metadata, callback replacing the generated body with the documented basic shape, header transform removing generated browser/beta fields, final transport route/body/header guards, response capture, event/result/attempt recording, raw-shape summary, secret audit and SHA-256 inventory.
- `/tmp/barness-glm-messages-probe/source/run.py`: narrow `.env` parser that reads exactly GLM_KEY without sourcing/evaluating shell, starts the prebuilt harness in an isolated environment containing only that credential, enforces 300-second process bound, never prints child output or credentials.
- Source is written and locally compiled before the single live invocation. The invocation is protected by an exclusive execution lock; it cannot silently rerun.
- Production repo source is untouched. Harness is built from `/Users/cyber/RestoX/barness` so it imports the current public package. The candidate does not claim built-in catalog support.

## Expected wire body

```json
{"model":"<one fixed model>","max_tokens":4096,"stream":true,"messages":[{"role":"user","content":"Calculate 17 + 25. Reply with only the integer answer."}]}
```

`thinking`, effort, budget, display, betas, cache markers/TTL, tools, eager streaming, managed effort, sampling fields and server-side references are omitted. Default SDK `anthropic-version` is retained and captured; this is not a claim of a new GLM requirement.

## Before-code failure scenarios

1. Secret missing/ambiguous/malformed: stop before child or network; never print file or value.
2. Client candidate resolution/validation or callback authorization failure: record local classified error; do not bypass Client or raw-call the endpoint.
3. Any third attempt, wrong method/host/path/query, unsupported body field, model drift, token budget drift, proxy/redirect, extra auth, unsupported beta/browser header: transport rejects before network.
4. Request or response capture exceeds finite byte limit: mark truncation and stop; never accept it as complete.
5. HTTP auth/permission/quota/rate/service failure: record one attempt only, continue only to the other already-authorized fixed model; no retry or variant exploration.
6. Invalid/missing SSE terminal, unknown block/usage shape, absent signature or parser failure: preserve bounded raw evidence and public failure; do not turn uncertainty into success.
7. Timeout/cancel: one terminal result, close resource, no replay.
8. Credential appears anywhere in an output: overwrite its exact bytes with redaction, mark audit fail, expose only safe artifact paths/status until reviewed.

## Artifacts and completion criterion

Each model gets redacted final request/response headers, exact synthetic request JSON, bounded consumed SSE, parsed raw SSE frame JSON, public events, Result, error and attempts, usage, truncation/read state and a structural summary. Bundle contains candidate catalog/config, observations, budget manifest, secret audit, index and SHA-256 inventory. Headers/cookies and exact secret bytes are redacted before persistence; only metadata summaries/paths are reported. Every persisted file is audited for exact secret bytes without persisting secret hashes.

Success for this exploration means two bounded recorded outcomes with honest raw shape findings, not necessarily successful API responses and not completed thinking/tool/replay acceptance. No advanced probe is authorized in this run.

## Second bounded invocation: SDK query acceptance hypothesis

Parent authorization after review of the first local failure: exactly two additional basic public Client logical calls (one each for 5.3 and 5.2), at most two additional HTTP attempts, otherwise unchanged model/body/header/token/time/resource restrictions. Cumulative early reservation is 4 logical calls / 4 library attempts; first invocation sent zero external HTTP requests.

Observed SDK fact: locked `betamessage.go:73,113` generates `v1/messages?beta=true` through `NewBetaMessageService` even with no beta header. First invocation was stopped by the local strict URL guard, proving no GLM behavior. Hypothesis: GLM may accept this ordinary SDK-generated query, because its official guide directs use of the Anthropic SDK. This is not established acceptance and not a claim the query is required.

Only outgoing-request difference allowed: final target guard now permits that exact `?beta=true` query alongside the same scheme/host/path. The harness does not mutate or strip the query and sends no beta header. It keeps all other body/header guards. It performs no advanced probe/retry or production edit. If the service rejects that query, retain the failure; do not try alternate query/service forms.

Source for this second invocation is `/tmp/barness-glm-messages-probe-sdk-query/source/main.go` and `run.py`, separately compiled before the single invocation. Old `/tmp/barness-glm-messages-probe/` artifacts and lock remain untouched. New capture directory `/tmp/barness-glm-messages-probe-sdk-query/` has its own exclusive lock and checksum inventory. Actual external HTTP counts will be reported separately from reserved and library-recorded attempts.

## Third bounded invocation: function tools, native replay, exact enabled/budget hypothesis

Parent authorization now reserves at most6 additional public Client calls/attempts, no retries. Planned order: 5.3 tool request+public tool-result continuation;5.2 tool request+public tool-result continuation;5.2 explicit enabled/budget request+public native-history continuation. If a model fails to call the tool, mark its tool scenario FAIL and skip its continuation; never try forced choice. If explicit thinking is rejected or produces no observable thinking, record that result without alternate parameter trials.

Each call retains max_tokens4096, call timeout120s, connection15s, headers/read90s, transport capture2MiB and event capture16MiB. Child process bound900s. Third invocation maximum requested output24576 tokens. Cumulative reservation10 logical calls/10 local attempts; actual external requests no more than8 including the prior2. Actual library and network counts remain distinct.

Evidence/hypotheses (unverified GLM fields before execution): China's official Claude compatibility page https://docs.bigmodel.cn/cn/guide/develop/claude/introduction.md directs Anthropic SDK use but lacks per-field guarantees. Authorized basic sample already produced a GLM5.3 thinking/signature block and both Messages model responses. Anthropic's own documented tool schema (https://platform.claude.com/docs/en/agents-and-tools/tool-use/implement-tool-use, now Define tools) provides name/description/input_schema and ordinary auto selection. Its extended-thinking legacy docs (https://platform.claude.com/docs/en/build-with-claude/extended-thinking) describe one exact candidate `thinking:{type:enabled,budget_tokens:1024}` with max_tokens4096. These support bounded hypotheses, not a GLM acceptance claim. No Chat reasoning_effort inference or adaptive/effort/beta experiment is permitted.

Tool task uses only a synthetic local lookup: public ai.Tool lookup_probe_code with label enum alpha; first user explicitly asks to call that tool and use its returned code. ValidateToolCall must succeed before host returns synthetic BRN42 through public ToolResultText. The host does no file/network/external tool operation. Continuation history is the original public Request user message, the exact in-memory public Result.Message (including trusted native state) and public ToolResultText. It is never reconstructed from handcrafted wire blocks or JSON-loaded AssistantMessage.

Explicit-thinking task is synthetic modular arithmetic; same-model continuation uses the exact returned public Message and an added public UserText. The payload hook records before/after bodies and strips only unproven generated cache_control,eager_input_streaming,display,block_binding,betas/output_config. It keeps public messages and all thinking/signature/tool blocks; no signature fabrication. Before actual replay send, a guard proves every returned nonempty signed thinking block appears unchanged in the generated assistant history. A local failure is recorded if production conversion changes/drops it.

Final transport guard retains exact https host/path and previously accepted beta=true query, x-api-key only, token4096,streamtrue, allowlisted top-level request keys and exact callback→wire bytes-equivalent JSON; no extra parameters or new headers. The before/after snapshots establish what was pruned. The source is `/tmp/barness-glm-messages-probe-tools-thinking/source/main.go` plus run.py; both prior bundles/locks remain untouched.

Additional failure scenarios before source: wrong tool/argument/malformed/truncated result is not executed; no tool call means FAIL not forced-choice retry; failed first turn means no replay; missing thinking/signature means no false replay success; invalid provenance or changed signed history fails locally before network; unsupported enabled/budget shape is not replaced by another trial. Captured usage remains partial when fields are absent, never known-free. All outputs continue exact-key audits, header/cookie redaction and inventory hashes.

Third source is split into main.go (wire/audit) and host.go (scenarios) so each file stays below500 lines; build both explicit files together.

## Fourth and final bounded wire exploration: equal-prompt default and unsigned replay

Parent authorizes at most3 additional public calls/attempts; no retries/alternatives. Cumulative early ceiling13 Logical Calls/13 conservative local attempts, at most11 external requests. Each max4096 tokens,120s; process420s; same isolated GLM_KEY/base/SDK query/auth/header/resource guards and audits.

A: GLM5.2 same modular-arithmetic prompt and full public options as passed enabled/budget1024 first turn, omitting ONLY its thinking parameter. Compare final callback wire JSON to the pinned prior request minus thinking. Hypothesis: default mode may emit observable thinking for a challenging prompt; absence is only sample evidence, not proof of no internal reasoning.

B: GLM5.3 repeat its successful tool-result continuation using the actual prior signed public tool Message, restored from hash-pinned Host storage with TrustNativeState. After normal public Client history conversion and unchanged-history check, OnPayload removes ONLY signature from that one thinking block. Thinking text/type/block order, tool call/result and tool schema/choice remain identical to pinned prior signed continuation JSON. Hypothesis: endpoint may accept unsigned thinking replay. This is intentional signature omission, never invented signature and never evidence the provider emits unsigned blocks.

C: GLM5.2 repeat its successful enabled/budget1024 history continuation from its actual prior signed public Message. Restore trusted state from hash-pinned Host storage; remove ONLY thinking signature after public conversion. Same prompt,thinking parameter,text,type,order unchanged relative to pinned prior signed continuation. Hypothesis and limits same as B.

Host restoration checks source file SHA256, trusted recorded attribution tenant/account/provider/API/model, then decodes the stored public content blocks without changing values and calls public TrustNativeState for the same scope/identity. Missing/inconsistent provenance or local guard refusal records failure before network with no replacement/retry. The source retains both before/after snapshots and verifies structural JSON equality to the exact signed baseline after the one allowed omission.

Before-code failure scenarios: bad source hash/envelope/block decoding; public conversion loses native state; more/fewer thinking blocks than the one target; alteration beyond signature removal; changed body field/model/tool/auth/path; network/service rejection; missing terminal or truncation. All are recorded and stop that fixed case without any further variable trials. These3 calls do not authorize any effort/other budget/unsigned-output/sig-spoofing probes.

Source: `/tmp/barness-glm-messages-probe-default-unsigned/source/{main.go,host.go,restore.go,run.py}`; all prior bundles/locks retained separately. Compile and inspect guards before one isolated invocation.

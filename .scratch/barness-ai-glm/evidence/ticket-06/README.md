# Ticket 06 — GLM Messages basic public Client evidence

Date: 2026-10-17 (Asia/Shanghai). Confirmed seam: spec T01 public Client,
controlled Provider and trusted Host. Implementation source is committed at
`c0efafa3559ebab36a56b45dbe873db8fd57e899` on `codex/glm-06`, based on integration
`2acb83f`. That integration tip was merged before reporting (already ancestor).
The [index](index.json) pins source hashes, each manifest, archive, case count
and outcome. Raw audited bundles and ZIPs are durable at repo-root
`.evidence/barness-ai-glm/ticket-06/`; no temporary worktree path is required to
inspect them. [Failure modes](failure-modes.md) preceded implementation/tests.

| Slice/check | Result |
| --- | --- |
| Basic request/default/header red → green | 24 FAIL → 24 PASS; both models independently, six nil/full/simple entry variants with/without hooks |
| Explicit option/final callback guard red → green | 32 FAIL plus 2 existing sampling PASS → 58 PASS including basic slice |
| Final complete GLM slice and committed-source replay | **66 PASS** each |
| GLM race replay | **66 PASS** |
| Native Anthropic requests/events/history/usage/cache/beta/managed effort | **104 PASS** |
| Native Anthropic runtime/callback/environment/resource checks | **168 PASS** |
| Frozen pi Messages differential | **70 PASS**, zero pending; no GLM pi support claim |
| Vet, staged/unstaged whitespace checks | PASS |

The first terminal-script iteration had 6 PASS / 2 FAIL because its text
replacement did not match spaces in the real captured JSON, so the intended
missing-stop-reason fault was never injected. It is retained as a harness
iteration, not evidence of an implementation defect. Corrected terminal
mutations all 8 PASS through existing lifecycle code. Initial compile-only
iterations produced no runtime bundle and are not counted as behavior red.

The basic red bundle exposed generated cache markers, browser headers and
lost X-Request-Id. The guard red bundle demonstrated that unsupported named
controls and callback fields reached a provider attempt. Green now rejects
these before OnPayload (named controls) or before any Attempt (final payload).
Neither rejection claims the upstream service rejects a field; it bounds the
contract exposed by this implementation.

## Own-protocol provenance

[Messages exploration index](../messages-exploration/index.json) pins four
separately audited immutable bundles, cumulatively 13 Logical Calls / 13 local
Attempt reservations / 11 actual external HTTP requests. This ticket made
**zero** paid calls and did not inspect `.env`.

- `02-basic-sdk-query`: one actual HTTP200 per model; raw `response.sse` copied
  unchanged into `ai/e2e/testdata/glm-messages/`, SHA-256 pinned by `text.json`.
  Initial input/output zero, actual message_delta usage and absent cache write
  are independently reflected in the fixture expectations. The SDK query and
  `anthropic-version:2023-06-01` were accepted; no beta/browser header was sent.
- `03-tools-thinking`: both models' auto tool schema, JSON deltas and signed
  history; GLM-5.2 exact enabled/budget1024; existing content-block arrays were
  accepted. Raw evidence remains available for ticket 07. Basic implementation
  does not infer budget ranges or canonical mappings from this single shape.
- `04-default-unsigned`: 5.2 same-prompt omission yields text only; both models
  accepted deliberate absence of signature on trusted thinking history. This
  is not empty-string signature evidence or observed unsigned provider output.

No Chat/Responses schema or behavior supplies Messages support conclusions.
The candidates use text-only input and conservative independently documented
model capacities; rates are synthetic test values. See the
[Messages Host notes](../../../../ai/internal/testkit/glm/messages.md) and
[ADR-0035](../../../../docs/adr/0035-barness-ai-zhipu-messages-basic.md).

## Replay

```sh
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-06/replay" \
  go test ./ai/e2e -run '^TestGLMMessages' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-06/replay-native" \
  go test ./ai/e2e -run '^Test(Anthropic|ManagedEffort)' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-06/replay-runtime" \
  go test ./ai/e2e -run '^TestProtocol/^anthropic$' -count=1
BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-glm/ticket-06/replay-race" \
  go test -race ./ai/e2e -run '^TestGLMMessages' -count=1
# After the frozen oracle's own locked dependencies are installed:
BARNESS_AI_PIDIFF=1 go test ./ai/e2e \
  -run '^TestPiDifferential$/^anthropic-messages$' -count=1
go vet ./ai/...
```

Every valid green command's TestMain completes its redaction audit; all had
zero findings. Oracle dependencies were reused from the existing root install
via a temporary symlink, now removed. SSE fixtures keep their required final
empty delimiter under a narrowly scoped whitespace attribute.

Only ticket 06 basic text criteria are resolved. Ticket 07 owns enhanced
thinking/tool/native replay and stream ordering, ticket 11 owns formal
production Client live acceptance, and ticket 12 owns built-in registration.
No support-matrix, built-in catalog, pricing snapshot or full Messages support
claim was changed. Disabling the Host Binding removes this candidate route;
there is no endpoint/protocol/account fallback.

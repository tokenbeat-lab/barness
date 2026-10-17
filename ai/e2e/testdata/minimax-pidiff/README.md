# MiniMax frozen differential and explicit extensions

Frozen pi-ai remains 1.1.0, commit
`abe508e1b89912adde45528136c3221eb69acdd7`, 43-file model data
`8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`,
827-file source `39a9951108d0e7faf60886302ce4bc953d8c51d5ebad22370301f23b29aa2504`
and unchanged SDK lock. The wrapper exposes the existing MINIMAX_CN_MODELS
export. No package data, compiled source, version or dependency is modified.

## Actually compared native M3 Messages

`TestMiniMaxPiMessages` executes public Client Stream/StreamSimple and frozen
pi stream/streamSimple against the same synthetic text and fragmented signed
thinking/two-tool scripts. The Host validates both tool inputs, appends the
complete generated assistant and paired results, and executes a second turn.
The six inference cases are `PIDIFF-MM06-messages-{text,tools,tool-replay}-`
`{stream,streamSimple}`; neither a successful own replay nor wrapper lookup
alone is counted as this paired differential.

Every complete request, event sequence and Result participates in Compare.
Existing shared-view/SDK header decisions are reused without alteration.
New decisions in the central ledger have those exact finite case IDs and
single paths/kinds. Wires and token/cost values are independently asserted
before classification; unknown fields, cases and protocols stay pending.

| Exact difference | Why / independent constraint |
| --- | --- |
| `request.body.max_tokens` | Frozen M3 default512000; current spec conservative131072. Both exact integers are asserted. |
| `request.body.thinking` on text simple only | Frozen explicitly disabled; current omission preserves M3 Messages off default. |
| Thinking `type`, `budget_tokens`, `display` on tools/replay only | Frozen enabled with full1024/simple16384 and summarized; current documented adaptive binary switch, no budget/display. |
| Browser/beta headers and tool eager flag | Frozen browser true and interleaved-thinking beta; current subset sends none and no eager extension. Values independently asserted. |
| Exact final user/tool-result and function `cache_control` | Frozen ephemeral; current subset does not claim active writes/TTL. Automatic token counts remain strict. |
| `request.headers.Content-Length` | Must equal each complete body's actual byte length; all unlisted body fields remain strict. |
| Exact cost input/output/cacheRead/total leaves | Candidate mechanism rates1/2/.1 vs historical frozen .3/1.2/.06; exact usage and worked calculations asserted. No current fee claim. |
| `usage.cacheWrite1h` presence | Frozen native pi materializes0; MiniMax omits unsupported one-hour creation semantics. Both cache-write counts must be0. |
| `durationMs` value/partial presence | Same previously approved elapsed/shared-view contract; nonnegative final integer, one terminal and terminal/Result equality are required. |

Both SDKs append `?beta=true` to the same controlled
`/anthropic/v1/messages` resource. This is captured and asserted on both sides,
not a new exception. Native thinking text/signature, tool IDs/arguments and
complete continuation are strictly compared. Synthetic fixture acceptance is
neither a paid-wire capture nor evidence of the real endpoint's acceptance.

## Full catalog comparison

`PIDIFF-MM06-frozen-catalog` verifies literal frozen facts and confirms that
M3.1, MiniMax Chat and Responses entries are absent. The separate complete
`PIDIFF-MM06-catalog-differences` compares full public model metadata with
Binding-owned endpoint/type added explicitly; it retains both endpoints.
Only exact fields differ: baseUrl, candidate display name, maxTokens, text
input vs image/resize policy, binary/default compatibility metadata and
synthetic cost fields/historical tier. Literal frozen/current metadata is
asserted. The old `api.minimaxi.com` endpoint,512000 budget,image policy and
historical USD rates are recorded honestly, not silently updated.

Current official `api.minimax.cn` endpoint and text subset are spec R02/R06.
Official source hashes and conservative capacity origin are in
`ai/internal/testkit/minimax/sources.json`; candidate prices are deliberately
synthetic and must stay outside formal cost/support claims. Ticket07 owns
current CNY→USD/tier evidence and each real M3 protocol's formal release.

## Extension acceptance, never fabricated parity

`extensions.json` contains two M3 route extensions and three M3.1 model
extensions. Each has a source, independent synthetic script, explicit
`pidiffSkip`, and own public four-entry acceptance. Their evidence IDs begin
`MM06-extension-`, never PIDIFF, and record `piParity=NOT_RUN` plus the reason.
The central route registry adds only minimax-cn Chat/Responses; Messages is
not exempted. M3.1 Messages does not borrow frozen M3's identity. Existing
protocol-specific text/tools/thinking/native-state/error suites retain their
own fixtures and remain separate from this bounded text extension inventory.

No old decisions, route registrations, fixture selection or native subtest is
removed. No whole request/result/protocol exception is added. This work does
not promote a candidate catalog, support-matrix row or live claim.

Replay and immutable captures:
`.scratch/barness-ai-minimax/evidence/ticket-06/README.md`.

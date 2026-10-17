# Ticket 07 — isolated MiniMax live evidence tools

No `.env`, vendor key, or paid API was read/run for this ticket. The six controlled identities use synthetic credentials and an HTTP peer with no network path. Missing key/entitlement is NOT_RUN; actual service errors remain FAIL. M3.1 remains candidate-only. Formal catalog/matrix rows/release-contract changes belong to ticket10.

## Fixed plan and budget

One process selects exactly `BARNESS_MINIMAX_ROUTE=chat|responses|messages`, `BARNESS_MINIMAX_MODEL=MiniMax-M3|MiniMax-M3.1-Flash-Preview`, and `BARNESS_MINIMAX_PURPOSE=standard|mplan` (M3 standard, M3.1 M Plan). A selected process rejects foreign credentials/routes/endpoints and narrowed `-run`. `BARNESS_MINIMAX_KEY` is the sole credential variable; the harness never loads dotenv. Live requires `BARNESS_AI_LIVE=1`, `BARNESS_MINIMAX_ENTITLEMENT=confirmed`, and nonsecret account/region aliases. Only an explicit validated loopback `BARNESS_MINIMAX_PROXY` is used; ambient proxies are ignored.

The immutable plan has13 logical cases /12 actual baseline attempts /17410 requested output tokens. It covers all four Client entries, thinking controls, an own reasoning+automatic-tool turn, Host tool validation/execution, persisted trusted native continuation, truncation, repeated-prefix raw usage, cancellation after visible content, and a zero-attempt preflight error. Only tool-call can rerun once, on the same stage/identity ledger, retaining the unsuccessful first capture. It adds one call/attempt and4096 requested tokens. No missing tool/reasoning is certified by assertion alone.

Optional pre-convergence research uses `BARNESS_MINIMAX_RESEARCH=1` and exact `^TestMiniMaxResearch$`: three fixed probes (text1024 + tool4096 + own native continuation2048) before the full current-source candidate plan. Its report cannot become complete/formal. Full reports preserve earlier research source/capture bytes, failures and receipts. Research + full + permitted rerun is17 calls /16 baseline actual attempts /28674 requested output tokens. Source changes after research do not relabel research as current evidence.

Every actual live process uses the canonical shared `.evidence/minimax-stages/{candidate,formal}.jsonl`; choosing a fresh path is refused. A lock plus fsynced chained receipt is required before each Client call and again before transport. Identity cumulative ceilings survive restart:20 calls,24 attempts,40000 requested output,120000 conservative input; stage ceilings are120/144 for six candidate identities and60/72 for three formal M3 identities. Each call <=4096 output /120 seconds; process <=1800 seconds with finalization reserve. Input reservation is complete synthetic public request bytes +1024, minimum2048; known actual input including caches above reservation stops later calls. Missing/cancelled usage stays explicitly unreported/partial, never assumed free. These are Host sampling bounds, not vendor billing guarantees.

## Evidence and executable seams

`minimax-report.json` plus `minimax-integrity.json` seal an exact finite artifact set. The validator independently recomputes the full runtime/helper/dependency source set (including additions/deletions), immutable official snapshots, plan, catalog, SDK, all receipt/capture/Attempt/Observer counts, raw actual model/usage, USD estimate, endpoint, stream terminal/Result, original native reasoning/signatures and clean audit. Messages' SDK `?beta=true` is recorded transport behavior anticipated by R02/R15; it grants no Anthropic beta feature header/body support.

Public commands:

```sh
go test -tags live -count=1 ./ai/live -run '^TestMiniMax'
go test -race -tags live -count=1 ./ai/live -run '^TestMiniMax'
go test ./ai/internal/testkit/minimaxlive ./ai/internal/testkit/supportmatrix
go run ./ai/live/cmd/minimaxverify <bundle-directory>
go run ./ai/live/cmd/minimaxverify -formal <bundle-directory>
```

Actual candidate entry is exact `^TestMiniMaxCandidate$`; actual formal entry is `^TestMiniMaxFormal$`. Formal selection reads the current BuiltinCatalog and refuses an absent M3 identity or M3.1; ticket10 must promote sourced M3 prices before formal execution. A complete actual-live M3 formal process publishes `live-report.json` through `supportmatrix.NewMiniMaxReport`; matrix import independently runs `VerifyMiniMaxReport`. Candidate, controlled, and replay cannot publish formal PASS. No existing route or release requirement was removed.

Offline recorded replay uses `BARNESS_MINIMAX_REPLAY=<sealed-bundle>`, `BARNESS_AI_LIVE=0`, no credential, and exact `^TestMiniMaxReplay$`; keep the same identity and aliases, and set a synthetic replay ledger path. It verifies the source first, replays exact request/response order without network, preserves response IDs/usage/native signatures, compares Results except Host clock/local IDs, and checks event sequence. It can never become live evidence.

Process E2Es retain bounded `child-capture.tar.gz` artifacts (clean audit and exact seal included) before temporary-child cleanup; outer Fixture SHA256 binds each archive. Every generated suite bundle remains under git-ignored `.evidence`; only curated red/green logs and one checksummed controlled archive are retained here. Original red logs prove missing process report, re-sealed semantic tampering, cancellation replay and optional-rerun replay failures before their fixes.

## Remaining actual evidence

This ticket establishes tools and controlled public E2E behavior. M3 three-protocol actual native/tool/reasoning shapes, actual diagnostic IDs/usage/cache behavior, priced formal catalog, current source applicability and actual own formal reports still require ticket10. M3.1 online entitlement remains independently unknown; all three offline identities were exercised here.

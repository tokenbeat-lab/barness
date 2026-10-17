# Ticket 03 — Responses candidate evidence

Mode: offline, controlled TLS Provider; source: the MiniMax Responses OpenAPI
snapshot recorded by research (2026-10-18). No request reached MiniMax. Catalog
rates are explicitly synthetic. This report does not confer formal support.

[Red evidence](red.md) preserves failing public Client observations from each
implementation slice, including raw reasoning precision. Each E2E run writes its
own immutable `.evidence/barness-ai/<run>/` bundle with requests/SSE/events/Result,
Attempt metadata, assertions, and redaction audit. Replay requires no real key:

```sh
go test ./ai/e2e -run '^TestMiniMaxResponses' -count=1
go test ./ai/e2e -run 'Responses|DeepSeek|MiniMax' -count=1
go test -race ./ai/e2e -run '^TestMiniMaxResponses' -count=1
go vet ./ai/...
```

Cases cover four entries and two models; omitted defaults; full/simple effort
sets and sampling/hook final guards; exact 1..131072 output budgets; service tier,
tools, format, sampling/null constraints; tenant/account cache routing without
TTL/affinity; array content, done-only and terminal-only output, multiple parts,
authoritative corrections, malformed/duplicate/missing terminal and associations;
actual model and retained existing failed-response usage contract; opaque call
IDs and trustworthy/restored/untrusted/cross-model/cross-account tool round trips.
Raw native reasoning is retained including unknown fields/precise numbers;
unsupported unknown native fields refuse replay rather than disappearing.

All targeted commands above passed on the ticket branch. A broader `go test
./ai/...` was interrupted without a conclusion because root owns the complete
release suite; it is not reported PASS here.

[ADR-0039](../../../../docs/adr/0039-barness-ai-minimax-responses.md) records the
lossless content.reasoning_text → summary.summary_text candidate projection.
Actual M3 reasoning/tool continuation, accepted ID constraints and diagnostic
headers remain ticket07 live evidence obligations; M3.1 live status is independent.
No synthetic case is presented as proof that the server accepts the projection.

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
headers remain ticket10 live evidence obligations; M3.1 live status is independent.
No synthetic case is presented as proof that the server accepts the projection.

## Preserved historical observations

[historical-controlled-captures.tar.gz](historical-controlled-captures.tar.gz)
contains the original 33 controlled runs, including red and green observations.
[historical-index.json](historical-index.json) preserves each run's original
source commit, status set, completion time and audit, plus every file hash.
The source commits are `603127aa00a973282a584f3c71f3ae593e6c0e49` and
`69bf852754fb11990c5c8c2626a1a3cae7720ef0`. Twenty runs have only own PASS
cases, nine have only own FAIL cases, and four contain both. None is a current
formal or actual live report.

The archive and index were copied without rewriting their contents. Ticket09
streamed all 29,012 regular archive members and checked their original hashes
and total 118,969,270 bytes; all 33 original audits have zero findings. See the
[preservation receipt](../ticket-09/ticket03-preservation.json). The original
[SHA256SUMS](SHA256SUMS) pins the archive and index:

```sh
(cd .scratch/barness-ai-minimax/evidence/ticket-03 && shasum -a 256 -c SHA256SUMS)
```

The test commands above create new controlled observations. They do not change
these historical statuses or establish that current MiniMax accepts a request.

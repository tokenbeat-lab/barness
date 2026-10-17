# Ticket 06: immutable frozen differential / extension evidence

Mode: synthetic offline public Client plus actual frozen pi. No credentials
were read and no MiniMax network request was made. This is not live support,
current price evidence or a candidate/formal promotion.

Final focused run: **29 PASS**, comprising6 actual paired M3 Messages text/
signed two-tool/Host continuation comparisons,2 full catalog/absence checks,
1 ledger scope guard and20 own extension four-entry acceptances. The extension
cases record `piParity=NOT_RUN`, not native differential PASS.

| Verification | Result |
| --- | --- |
| final-focused |29/29 PASS with BARNESS_AI_PIDIFF=1; no pending findings |
| race |29/29 PASS with race detector and the same actual frozen comparisons |
| native-regression |70/70 existing Anthropic native differential cases PASS |
| oracle-tests |Existing pioracle comparison/classification/record tests PASS |
| runner-guards |9 provenance/empty-env/loopback/isolation tests PASS |
| vet |`go vet ./ai/e2e ./ai/internal/testkit/pioracle` PASS |
| coverage-and-pins |All317 existing decisions,6 old route entries and every frozen pin unchanged;41 finite decisions +2 new route entries |

The wrapper only adds the already-pinned MINIMAX_CN_MODELS export. All old
fixture selections and existing native projections remain intact. New cases
capture the complete two sides, original request/script/events/Result/Attempt/
Usage and metadata-only Observer; source, SDK, catalog and response/request/
result hashes remain attributable. Complete offline source hashes are retained
in execution-sources.json; coverage-and-pins.json independently fixes the old
pin/ledger identities. This is an offline source superset, not a live report.

Each tar.gz keeps its original manifest, assertions, original fixtures and
redacted captured artifacts. SHA256SUMS binds the archived runs and curated
logs/source proofs. Extract with `tar -xzf <archive>` to inspect a historical
run and its recorded per-case replay commands. Existing70-case regression
captures remain local under .evidence; their exact case list/manifest hash is
in native-regression-summary.json.

Failure-first history is preserved:

- red-catalog: frozen M3 Messages wrapper lookup returned null.
- red-inference: catalog mismatches were genuinely pending; initial Observer
  assembly also lacked its mandatory bounded queue. The test assembly was
  corrected, without changing runtime policy defaults.
- red-inference-complete/red-reviewed-bounds: all native turns executed and
  unreviewed budget/default/cache/header/cost/duration differences failed.
- red-extensions: own fixtures ran but the two required route registrations
  were absent. No native parity was fabricated for absent models/routes.
- green-inference/green-complete: scoped classifications worked, but an
  independent test expectation wrongly assumed empty query; both SDKs
  actually append beta=true. The capture-backed expectation was corrected.
- green-accepted/final-focused: all own cases pass with unchanged full fields
  and independently checked wire/token/cost expectations. The final run adds
  literal frozen and candidate metadata guards; no runtime inference change.

Replay from repository root, with the existing frozen dependency installed:

```sh
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^TestMiniMax(Pi|ExtensionRoutes)' -count=1
BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^TestMiniMax(Pi|ExtensionRoutes)' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^TestPiDifferential$/^anthropic-messages$' -count=1
go test ./ai/internal/testkit/pioracle -count=1
env -i node --test ai/internal/testkit/pioracle/node/runner.test.mjs
go vet ./ai/e2e ./ai/internal/testkit/pioracle
```

Exact difference rationale and source/extension boundaries:
`ai/e2e/testdata/minimax-pidiff/README.md`. Candidate USD rates are explicitly
synthetic; ticket10 must independently source current M3 standard prices,
execute the three real protocols and produce their formal acceptance.

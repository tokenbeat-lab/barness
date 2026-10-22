# Kimi own qualification

Kimi uses six independent native routes: `kimi-cn-chat`, `kimi-cn-responses`,
`kimi-cn-messages`, `kimi-intl-chat`, `kimi-intl-responses`, `kimi-intl-messages`.
Their actual Providers remain `moonshot-cn` and `moonshot`. Candidates, controlled
captures, real candidate qualification, catalog promotion and real formal
qualification are separate facts. No Kimi identity is currently promoted.

Follow `docs/agents/verification.md`. Ordinary E2E calls only controlled loopback
services through the real public Client, SDK and native process. It does not
spend money or authorize live, race or pressure stages.

```sh
python3 .scratch/barness-ai-test-process-review/tools/run.py \
  --feature barness-ai-kimi-qualification --stage e2e -- \
  go test -json ./ai/e2e -run '^TestKimiQualification' -count=1
```

The native lifecycle and read-only consumers reuse `REGIONAL.md`; Kimi owns a
different contract namespace, six-record candidate snapshot, source originals,
pricing provenance, fixed nine-case order and separate one-use ledger.

Each route has exactly nine logical calls and nine HTTP attempts, zero retries,
4096 requested output tokens per call, at most 36864 requested output tokens and
131072 input reservation tokens per process. Each call is bounded by120 seconds,
the process by1800 seconds. Input reserves `max(2048, encoded wire bytes+1024)`
before transport. Dollar reservation adds requested output cost and charges the
entire reserved input at the most expensive input/read/5m-write/1h-write rate.
It never releases budget based on a partial, canceled or lower-point estimate.

The fixed scenarios are: Full Stream with omitted effort/default max, Full
Complete high, Simple Stream low, Simple Complete medium clamped high, exact
L01 native-history continuation, tool call, Host tool-result continuation,
cancellation after actual content and a new recovery call. Chat tool selection
is required, Messages any, Responses auto. L01 must supply actual native
reasoning; another successful turn cannot replace it. Responses auto without a
tool call is a genuine qualification failure.

## Future paid execution

No command below is permission to run it. First obtain an explicit human grant
for one exact route, test account, source, binary, expiry and finite USD cap.
Do not infer authority from plan approval, candidate availability or an old grant.

1. Use the keyless `regionalverify -source` to freeze the complete execution
   source. It includes Kimi candidate JSON, price extraction and compressed
   original sources. Build the native binary with that digest:

   ```sh
   go test -c -tags live \
     -ldflags '-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=SOURCE_SHA256' \
     -o /absolute/path/kimi-live.test ./ai/live
   ```

2. `regionalverify -grant-template kimi-intl-messages` prints current machine
   facts. Fill the new ID, actual own account/region aliases, UTC expiry within24
   hours, strictly positive numeric `maxUSD` and actual frozen executable
   `binarySha256` only after the exact paid scope is approved. Templates contain
   no valid human grant or dollar allowance. Use `-stage formal` only after
   separate reviewed promotion; it refuses unpromoted identities.

3. Forward only `BARNESS_REGIONAL_ROUTE`, `BARNESS_REGIONAL_SERVICE`,
   `BARNESS_REGIONAL_GRANT`, `BARNESS_REGIONAL_KEY`, optional explicitly approved
   proxy and stage. Set `BARNESS_AI_LIVE=1`, use SOP `run.py --stage live --route
   <exact route>`, and run only `^TestRegionalCandidate$` or the separately
   authorized `^TestRegionalFormal$`. Controlled mode and live credentials may
   never coexist. Exact endpoint, SDK, source and binary must match the grant.

4. Before any HTTP the grant is burned in the trusted Git common directory's
   `barness-kimi-grants/` ledger using independent ID and normalized-scope keys,
   exclusive creation, locking and durable sync. All linked worktrees share it.
   Copying a path or changing only an ID cannot restore authority. Failure,
   interruption and cancellation consume the grant. Do not delete or edit this
   trusted Host authorization store or restore it from an earlier backup.

5. `regionalverify -qualify <native bundle>` consumes own real candidate/formal
   evidence; `-formal` additionally requires own promoted metadata and fresh
   formal execution. `-controlled` only verifies controlled artifacts. Consumers
   are read-only: no vendor request, build, secret read, promotion or scheduler.
   They replay captured responses independently and compare real counters,
   CostEstimate, exact native source, tool pairing, resources and provenance.
   Bounded costs are valid; missing principal charges are not qualified.

Native originals, failed attempts, grants, ledger receipts, binary/source
bindings, manifest, audit and consumer outputs must be retained. SOP's outer
`consume_live.py` receipt is still required before a real formal bundle can
enter a release record. Current readiness lives only in the feature spec.

## Proof-gated promotion

`regionalpromote -kimi` is a separate six-Kimi contract and targets only
`ai/catalog_kimi.json`. It cannot use old regional candidates. After all six
own real candidate consumers pass against the same reviewed current source,
binary and catalog, separately review the metadata/version/provenance and run:

```sh
regionalpromote -kimi -receipt /absolute/repo/.evidence/barness-ai-kimi/promotion/NEW/lineage.json \
  <six exact native candidate directories>
```

The new receipt parent must already exist. The command retains immutable
candidate/source/hash lineage before an atomic data-only replacement. Any
missing, partial, controlled, stale, substituted or changed proof rejects with
no activation. Today the independent payload is inactive and contains no models.
`-kimi -controlled -proposal <new output> <six controlled dirs>` exercises only
an offline proposal; its inactive payload contains proposed records and is
intentionally rejected as production activation data. Never copy it into the
production catalog.

Actual promotion changes the execution source and catalog hash. Freeze/build
again and obtain a fresh own formal grant for every route. Candidate capture
bytes and pins are never refreshed to simulate a formal run. Roll back by
first disabling Kimi Bindings, then removing Kimi from the Host catalog and
rolling back the reviewed payload/code version; preserve historical envelopes
and every original artifact.

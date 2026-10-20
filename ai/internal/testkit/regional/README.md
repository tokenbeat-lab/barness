# Regional qualification candidates

Current implementation/acceptance status lives only in
[the feature spec](../../../../.scratch/barness-provider-sites/spec.md).
`CandidateCatalog()` provides exactly nine independently declared new targets:
Z.AI GLM-5.3/5.2 and China MiniMax-M3, each through Chat, Responses and Messages.
Each literal complete identity owns its model metadata; no sibling Provider or
protocol generates it at runtime. `Candidates()` returns separately owned status,
price and source records. The package never modifies BuiltinCatalog.

All nine retain `INCOMPLETE`, own live `NOT_RUN`, and explicit qualification gaps.
A usable candidate is not formal support. Credentials, budgets, live qualification
and race require their respective explicit authorization. No real key was read or
model service called while preparing these records.

## Frozen own sources and price meaning

`source-files.json` records official URL, final URL, read time, original SHA-256
and size, plus the losslessly gzip-compressed file's SHA-256/size. The `.gz` files
contain complete retrieved originals, not rephrased summaries. Failed exploratory
URLs remain explicit `FETCH_FAILED`; they cannot substitute for required sources.
`snapshot.json` and the code pin bind the complete source manifest and both literal
candidate datasets. `VerifySources` verifies original and retained bytes. Price or
capability changes require a new version, data hash and reviewed pin.

On 2026-10-20 the own Z.AI pricing source declares GLM-5.3 and GLM-5.2 USD per
million token input1.4, output4.4, cache-read0.26. Its temporary cache-storage
promotion does not establish a cache-write token rate. Both models' own pages
state1M context/128K output and text inputs; Chat's schema independently specifies
131072 output. Candidate Responses/Messages conservatively cap at128000 from the
model pages pending own route qualification. Chat/Responses GLM-5.3 uses the own
low/high/max mapping and default max; GLM-5.2 keeps its own documented remaps.
Messages does not invent an international1024 thinking budget from China proof.
Protocol-specific wire/default/native-state acceptance remains a gap; public
sources do not establish account entitlement or actual captured behavior.

China MiniMax's own standard table supplies CNY per million token2.10 input,
8.40 output,0.42 cache-read up to512k input; above512k these are4.20/16.80/0.84.
512000 is a conservative decimal project threshold, not a measured vendor integer
boundary. Priority is excluded. The independently fetched CFETS2026-10-16 original
states USD1=CNY6.7220. Divide each CNY rate by6.7220, round half-even to ten decimal
places: USD0.3124070217/1.2496280869/0.0624814043, with long input double. This
fixed historical FX snapshot is deliberate, never a claim of today's live rate.
No runtime network exchange or derivation from international prices occurs.

MiniMax China's current Chat schema description states a2048 output ceiling,
while its Messages examples request4096 and model-specific protocol maxima are
absent from the retained sources. The three independent candidates conservatively
support2048 until own current per-protocol limit qualification resolves the
conflicting source statements. This is a documented catalog support restriction,
not a vendor capacity measurement or a budget for paid verification. Test calls
separately request64. Text-only candidate input narrows the vendor's published
multimodal capability to this feature's implemented scope. Neither the old
international131072 record nor another protocol's success fills these gaps.

Cache-write rates remain null in price provenance; ModelCost's zero is a missing
rate placeholder. Positive unknown writes must retain usage and fail with a
protocol/stream error, including Z.AI; they cannot certify a free successful call.
Known zero writes, absent usage, partial usage, estimates and actual billing stay
distinct. Candidate prices describe standard API estimates, never subscription
balance or final debit. Own live is still required before any promotion.

## Ordinary proof

```sh
go test ./ai/e2e -count=1 -run '^TestRegionalProvider'
```

Public Client/Host/Catalog cases retain synthetic scripts, request captures,
events, Result/Attempt/usage, metadata checks and release gauges in the existing
native E2E bundles. The implementation ticket links actual red/green receipts,
exact argv/exits/times, complete source fingerprints and actual case names.
Offline scripts are explicitly synthetic; they never rename historical China or
international live evidence or update the frozen pi oracle.

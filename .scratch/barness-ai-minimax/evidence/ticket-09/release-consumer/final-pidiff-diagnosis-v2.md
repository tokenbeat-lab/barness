# Two retained MiniMax differential admission failures (read-only)

Original source `3d3a1fb`, ordinary bundle `20261018T045647.697909000Z`;
Go toolchain recorded as `go1.26.2`. Both cases remain literal PASS. All six
artifact references verified byte-exact (see companion JSON); no files changed.
The adopted-input diagnostic at7754 remains pass=false/commands=null.

## Exact facts and authorities

* `PIDIFF-MM06-catalog-differences` (`TestMiniMaxPiCatalogDifferences`): four
  artifacts `assertions.json`, `pi-observation.json`, `barness-observation.json`,
  `pidiff.json`; pidiff SHA256
  `0973c228ef6e5484a6f1612ca8646c6dd1e69079f2a0424e064e0a7bba3cebd1`.
  Recorded `sdk_versions` has `@earendil-works/pi-ai=1.1.0`,
  `openai=7.19.0`, `@anthropic-ai/sdk=0.129.0`, but no `go` field.
  These npm versions come from the frozen oracle package-lock through
  pioracle.Open (`oracle.go:86–97`), copied by Oracle.Record
  (`record.go:80–87`). They are not mismatched upstream SDK versions.
  `release/differential_record.go:35` requires pinned pi1.1.0 AND nonempty
  Go toolchain. The missing Go value is supplied only by Sides.BarnessVersions;
  this catalog producer omits that map (`minimax_pidiff_test.go:151`).
  The adjacent wire producer atline118 already supplies runtime.Version() and
  the pinned Go module version. Do not relax this consumer requirement or edit
  the old record to fill it retrospectively.
* `PIDIFF-MM06-frozen-catalog` (`TestMiniMaxPiCatalog`): exactly two artifacts,
  `assertions.json` SHA256
  `f638a32a7550d7c9d4d114c56ecd9f341f01ef1c245d8ad79f19fa7f02d0aa21`
  and `frozen-models.json` SHA256
  `1de3a8af7fe1c1a645658d7f76a29854f6a1ad1876c37c759f895f8613bf8ee7`.
  The producer intentionally only looks up four frozen model keys and asserts
  the one present M3 Messages model and three absent new identities
  (`minimax_pidiff_catalog_test.go:9–30`). It does not call Compare, Record,
  a Go model entry or any HTTP provider. It is a valid catalog/provenance
  control but is incorrectly published in the PIDIFF namespace.

The first case is a genuine catalog differential (whole frozen model vs whole
synthetic bounded Go catalog metadata), not a wire differential: observations
contain Result only, request hashes cover the zero-valued structured Request (empty method/path/query/auth and null headers/body), Frames is the model-data
array. Its classified differences are retained and pending=0/gate=PASS. The
second is only metadata availability/provenance and has no differential verdict.
`release.kindOf` deliberately treats every PIDIFF prefix as a differential;
`differentialGate` consequently requires a pidiff. There is no evidence of
adapter loss or wrong frozen oracle. These are producer record/classification
faults exposed by the strict consumer; the consumer should remain strict.

## Smallest prospective repair (not implemented)

1. Populate actual Go toolchain/dependency provenance in the catalog
   differential's Sides, using the same runtime.Version/evidence.ModuleVersion
   seam already used by the adjacent MiniMax wire producer.
2. Classify the future frozen catalog control as a non-PIDIFF MM06 metadata
   case (retain every assertion/artifact and the same case count). Update only
   its finite current trace references; preserve the old ID and all old bytes
   as literal historical evidence. Do not add a consumer whitelist, delete a
   case, fabricate a differential for a fixture-only probe, or rewrite A.
3. Failure-first public producer subprocess for exactly these two tests with
   the frozen Node runtime, followed by public LoadBundle/consumer admission:
   two cases retained, one true fully valid differential, fixture still present,
   original frozen pi/model/lock pins unchanged. Ordinary/race targeted only.

Only two ai/e2e Go producers and necessary release trace JSON would change.
They are outside all three M3×263 and Q×217 selected maps; no runtime/capture
selector/producer library/pioracle source change is needed. The broad L/Jev
snapshots are already historical after consumer changes. No paid run, full
suite, new oracle, relaxed SDK pin or original report mutation is needed.
Current trace has finite requiredCases references to the old control ID, so
those must migrate for future source rather than silently drop the control.
Old diagnostic A still correctly fails these old records after a future
producer correction; it cannot be turned into a fresh full certificate.

## Explicit v2 correction

The first note incorrectly described the metadata Request as JSON null. Raw observations contain the zero-valued Request object. This does not represent HTTP execution. The first admission RED had an additional incorrect null-hash assertion; that genuine run is retained and superseded by the corrected literal empty-Request assertion. The primary missing-Go/namespace defects were independently reproduced before producer changes.

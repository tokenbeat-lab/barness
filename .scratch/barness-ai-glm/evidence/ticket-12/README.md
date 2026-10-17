# Ticket 12 formal GLM catalog and support

BuiltinCatalog Version `2026-10-17.2`, complete hash `sha256:81d6d9f3b69e9231985967e1ec8e13640aa7f4fdfd9e7ccef72f7d05320a6942`. Six independent literals match the accepted source facts, with fresh query copies and no changes to prior model data. Source/price/capacity provenance: ADR0036, immutable `ai/live/testdata/glm-candidate/sources.json` pin837a83c4711dc4914b8110341403c0d6e43cac254b3101f601d6d538b209c907.

All three own formal paid runs passed at frozen measured source `93c510669cbab6eec6021481fa8f53340400b836`, binary SHA256 `e6b917981689956346a0fc248bd7a0e8b0e3e40290d18af28526922a3df90007`. Each route: both models11/11,22 Logical Calls/22 reservations/22 actual HTTP, zero retries or reruns,359 outer inventory payload hashes verified and exact selected-secret audit passed. Total66 actual calls. Chat130.149s, Responses114.568s, Messages89.279s. All fixed24/32/4096/120s/1800s caps held. One China standard key through the isolated Host, direct fixed endpoints, no proxy/endpoint fallback or paid error probes. Account/region/permission/billing remain unknown.

The actual raw ZIPs and report hashes are pinned in `ai/live/testdata/glm-formal/live-captures.json`; its trusted separate versioned pin is `live-captures.sha256`. Originals remain immutable in root `.evidence/barness-ai-glm/ticket-12/`. Exact no-key formal replay uses the original raw captures and compares final requests, Result/Observer metadata, upstream IDs, signatures, tool pairing and known usage. Only exact Host IDs/clock/duration fields normalize; it never becomes live support.

A post-seal public importer RED exposed generic support Budget token fields set to0 despite correct captured/model-detail known usage. The consumer now derives input-side Input+CacheRead+CacheWrite once and Output once from captured Attempts. Independent known lower bounds: Chat1725/8312, Responses1648/7839, Messages2660/3879; incomplete/unknown attempts2/2/22. Reasoning is already in Output. These are estimates/known counters, not total billing. Original false-zero support bytes are retained and refused as corrected support.

Current no-paid support views are pinned in `ai/live/testdata/glm-formal/support-views.json`. They preserve byte-identical originals and bind the trusted raw index/ZIP/original formal+support report hashes, original measured commit and current importer commit/source hash. The claimed importer commit's complete Go source is independently checked against Git; each view has its own no-paid manifest, exact inventory and audit, published by atomic directory rename. Consumer source drift requires a fresh no-paid view. Actual model execution/catalog drift invalidates the underlying paid evidence. Older derived views remain history; the final index selects the current consumer.

The supportmatrix CLI atomically merged only the three own22-capability GLM rows. Prior11 rows are unchanged. Release registration requires14 live identities/13 protocol combinations, all old P0 and seven native differentials, three old extensions and three GLM extensions. The11 prior live identities need current-hash renewal in ticket13; this ticket does not claim the complete release passed.

Reproduce without any credential:

```sh
go test ./ai/e2e -run '^TestGLMFormal(Catalog|KnownUsage|Reports)$' -count=1 -timeout 180s
go test -tags live ./ai/live -run '^TestGLMFormal(Process|CaptureValidation|CaptureReplayProcess)$' -count=1 -timeout 180s
go run ./ai/live/cmd/glmformal -route chat -out /tmp/new-glm-chat-support-view
```

The original raw archives plus current support-view archives listed by the two indices must be present at their repository-relative ignored paths. Existing raw original report and support-view bytes are never rewritten. `index.json` records all durable raw/controlled/replay/support archives and SHA256 values, including intentional RED/failed diagnostic history.

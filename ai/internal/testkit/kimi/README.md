# Kimi K3 candidates

`CandidateCatalog()` returns six independently declared text-only K3 model
identities. It never activates `BuiltinCatalog` or grants actual support.
`Prices()` preserves native-currency facts and fixed historical conversion;
`SnapshotHash()` binds that extraction, the candidate catalog and the full
source index. Any price/provenance revision requires a new version.

`source-files.json` records URL, retrieval time, original SHA-256 and compressed
SHA-256. Full official originals are retained losslessly under `sources/` and
verified before loading candidates. The China rate uses the existing retained
2026-10-16 CFETS snapshot (1 USD = 6.7220 CNY), ten-decimal half-even rounding;
1h estimates use rounded Input × 2. It is not a current FX quote. The original
China 1h native write rate remains CNY40 per million.

Each protocol/market needs its own authorized candidate live, catalog review,
promotion and own formal qualification. Offline tests exercise behavior only.
See `.scratch/barness-ai-kimi/spec.md` and ADR-0052 for the delivery contract.

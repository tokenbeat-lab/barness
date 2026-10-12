# Decisions candidate sources

Fixed source snapshots retrieved on 2026-10-13 (Asia/Shanghai). `sources.json`
records UTC retrieval time, each SHA-256 and source URL. Its own SHA-256 is pinned
in `decisions_sources_test.go`; changing both a snapshot and this manifest still
fails until that reviewed pin changes. `candidate.json` is a separate host
Catalog, version `decisions-candidate-2026-10-13.1`; it does not alter BuiltinCatalog.

- `guide.md`: complete official Decisions Markdown, downloaded through the local
  proxy. Establishes Decisions input-only $0.10/1M; cache-read/write/output free.
- `model.md`: complete official general Luna Markdown. Used only for the strict
  >272000 full-request 2x input multiplier. Independent Decisions base × this
  multiplier yields $0.20. General chat output/cache prices are never imported.
- `reference.txt`: text nodes of the official reference's main element, with
  script/style omitted. Its extraction and original HTML SHA-256 are recorded.
  Establishes 128 images, text/image input and native string question shapes.
- `pi-classifier.json`: complete classifier entry from the repository's frozen
  official pi-ai 1.1.0 package. Original OpenAI data file SHA-256, complete data
  hash and source commit are recorded; 922000 belongs to this classifier snapshot.
- `gopher.webp` / `LICENSE`: unchanged x/image v0.45.0 fixture and BSD license.
  PNG/JPEG are deterministic 32×32 synthetic red squares from Go stdlib encoders.

PNG/JPEG/WebP and score 2–10 describe the barness supported subset. Own public
fixtures and live captures prove the subset separately. No snapshot establishes
that this harness exercised the 922K context, 128-image maximum or service score
maximum. Alias weights, final bills, cache billing and regional premium remain
unobserved. Live completion requires the candidate's own full pipeline and audit;
controlled process evidence is not real support evidence.

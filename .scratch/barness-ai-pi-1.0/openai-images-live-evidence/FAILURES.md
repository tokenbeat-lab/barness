# Issue 11 failure modes and agreed seams

Before implementation, verify through the issue's public Client.GenerateImages and
supportmatrix.Merge / supportmatrix CLI boundaries, with controlled HTTPS providers
for offline replay and the fixed official endpoint for live evidence.

- Smoke false positives: generation runs before JSON edit; narrowed execution bypasses
  the required edit; multipart/Responses/proxy/model alias replaces the fixed route;
  missing credentials become PASS; a vendor refusal becomes UNSUPPORTED for mandatory
  generation/edit; a mask refusal is accepted without an explicit mask-specific reason.
- Image validation: corrupt/truncated base64, MIME mismatch, unreadable raster, wrong
  dimensions/count, external URL or file ID, captured output too large, unbounded copies,
  pixel equality used as the success criterion, transparent output claimed without alpha.
- Consumption: failed requests and environment retries disappear; calls/images exceed
  4; requested or returned resolution exceeds 1024 square; retries bypass reservations;
  binding-level retries silently double spend; unreported/partial usage becomes zero-cost
  complete reporting; known modality price is replaced by per-image/cached charges.
- Matrix: matching but incorrect operation/provider/API, alias or non-official endpoint,
  incomplete/reduced/duplicate expected capabilities, mandatory UNSUPPORTED, blank request
  IDs, invented PASS without HTTP evidence, inconsistent counters, out-of-order evidence,
  a failed edit followed by claims of generation/mask support, failed merge mutates a file
  or another route, shared format changes break the TypeSafe route.
- Release evidence: a model enters BuiltinCatalog before this route's edit passes; stale
  catalog/price hash; source/fetch date missing; key/credential header/account identifiers
  leak into artifacts; Observer contains prompt/base64/body; live acceptance reported using
  synthetic evidence; real output cannot be decoded or its checksum reproduced.

Architecture: retain one public Client and existing isolated live/report/audit/atomic
matrix pipeline. Add image-specific scenario execution and validation at that boundary;
reserve calls and images before each public call, retain actual HTTP attempts and all
reported usage, and use a sourced host probe catalog until own live acceptance is known.
No new runtime transport/protocol fallback or parallel matrix.

# International M3 complete candidate capsule audit

PASS: keyless, read-only integrity curation of source `760533299b71736998e15cdd027a53ae069aff9e`.

All 15669 integrity checks passed; 2278 exact archive files; archive 5,816,813 bytes. Archive roundtrip verified every file byte. The archive includes all 1026 committed generation-source files, the exact sealed captures, source/binary/plan/approval/check/command/origin receipts, all four historical research captures, and the immutable candidate ledger snapshot. Executable binaries remain in the original prepared collection with SHA-256 and size recorded in the index.

| Route | Base endpoint | Cases | Calls / actual attempts | Known usage / unknown final consumption | Known USD |
| --- | --- | --- | --- | --- | --- |
| chat | `https://api.minimax.io/v1` | 13 PASS | 13 / 12 | 11 / 1 | 0.00143718 |
| responses | `https://api.minimax.io/v1` | 13 PASS | 13 / 12 | 11 / 1 | 0.00092730 |
| messages | `https://api.minimax.io/anthropic` | 13 PASS | 13 / 12 | 11 / 1 | 0.00091434 |

Every actual request returned HTTP 200 and reported actual model `MiniMax-M3`; each route passed persisted trusted native tool-result continuation. The preflight case failed locally with zero HTTP attempts as required. Cancellation after visible output has unknown final consumption: Chat and Responses label it unreported; Messages labels the initial zero usage frame partial. Those zero DTO fields do not establish a zero bill. The index preserves every exact Observer label (Chat: 11 complete / 1 unreported; Responses: 3 complete / 8 partial / 1 unreported; Messages: 12 partial). The table rounds the known USD sum to eight decimal places; the JSON index keeps the exact sum of captured numeric values. Known usage is not an invoice. Cache-repeat checks retain observed usage without promising a cache hit. All bodies and Observers settled; selected-key launcher audits and capture audits have no findings.

The cumulative candidate snapshot is 49 / 120 Logical Calls, 45 / 144 actual Attempts, 49 reservations, and 94 linked receipts. This retains historical CN 2 calls / 2 attempts and international research 8 calls / 7 attempts. Slot totals are Chat 17 calls / 16 attempts, Responses 16 / 14, and Messages 16 / 15; all per-slot output and input caps hold. The original four CN receipts and old launch intent match their previously committed hashes. Historical Chat and Responses research failures remain FAIL; Messages research remains incomplete research PASS. They were neither relabelled nor treated as current candidate proof.

The captured source and 1026 generation pins were checked against Git source 7605332, not the later catalog-activation working tree. All report execution-source pins, official source snapshots, source-review references, sealed file and manifest hashes, canonical origin bindings, exact own receipts, embedded predecessor bytes, and canonical ledger prefixes match. Prior launcher public-verification and keyless promotion receipts are retained; this curator did not rerun verification, read `.env`, execute tests, call a vendor, or modify the repository.

Archive SHA-256: `06fca74279c09495cf817c377311718f66290b63c43103285c26cfd87753ca08`. Full portable file and source inventory: `international-candidates-7605332.json`. Repeatable keyless verification: `python3 verify_capsule.py international-candidates-7605332.json`.

Scope: complete international M3 standard candidates and data-only activation provenance. Formal reports, old-route renewal, business effects, and the final release gate are separate evidence.

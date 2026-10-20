# Current portable verifier gzip boundary

Final review found two inventory readers called eager gzip.decompress before
checking the32MiB expanded limit. A compressed inventory below8MiB could therefore
allocate far beyond the documented boundary. Q also decoded before checking its
compressed hash. This evidence-only fix changes exactly those two readers.

Both now read at most8MiB+1 compressed bytes, refuse overflow, hash those exact
bytes before decoding, then use GzipFile.read(32MiB+1) and reject expanded overflow
before JSON parsing. This bounds actual decompression, including concatenated
members. No Go, ai JSON, launch tools, selectors, canonical records, capture,
report, catalog or historical source pins were changed. All1,034 generation
files match frozen3d3a1fb; M3/Jev source applicability is unchanged. No credentials,
vendor calls or Go suites were used for this correction.

[Current verifier index](current-verifier-index.json) authenticates the corrected
programs. **Existing parent retention indices and provenance remain unchanged
historical records.** Their old verifier hashes identify original bytes in the
[historical metadata proof](historical-verifier-metadata.json) and its explicit
tar.gz archive. Nine original verifier/metadata files roundtrip byte-exact;
original paid/failed/raw archives were never rewritten. Apply the new index to
the current verifier paths, not the historical parent hash entries.

[Failure inventory](failure-inventory.md), [source proof](source-equivalence.json)
and [verification summary](verification-summary.json) retain the bounded public
CLI red/green results. The32,104-byte raw artifact archive has71 original test
files: historical/current program hashes, exact diagnostic output, source
snapshots, fixture-generation parameters and genuine positive command/output
records. The generator derives explicitly synthetic negative copies from the
unchanged original archives; it never executes a captured command, qualifies a
synthetic capture or mutates the originals.

The final public negative suite has6 passing refusal cases: hash-before-decode,
Q96MiB expansion, actual96MiB expansion,48MiB concatenated expansion, compressed
8MiB+1 overflow, and truncated gzip. Final resource peaks were about88MiB for Q
and135MiB for actual. The preserved original programs fail the96MiB resource
probes at about216MiB/260MiB and fail Q's hash ordering.

The initial128MiB process ceiling for both programs was a **test calibration**,
not the repository contract. It was too low for actual's already-verified capsule
baseline. Parent preregistered Q128MiB/actual192MiB process ceilings and96MiB
finite fixtures before the final old/new reruns. All earlier128MiB failures,
intermediate source and the unused-map cleanup attempt are retained; that cleanup
was removed. The production8MiB/32MiB limits were never raised. An earlier positive
harness comparison mistakenly compared Q CLI output with a parent summary schema;
the CLI itself passed. Fresh old/current public CLI runs then confirmed their
outputs are identical except one additional bound check.

Genuine portable positives both PASS: Q152,448 checks (original preflight stays
FAIL, audit2, no paid authorization); actual109,799 checks (qualification stays
PARTIAL, failed Responses remains ineligible,132/132 known renewal spend).
All five external lineage archives were included. Historical counters were
152,447/109,798 and remain unchanged in their original records.

Repeat the negative public executable checks with fresh output:

```sh
python3 -B test_portable.py /tmp/portable-negative --archives /path/to/common/.evidence/minimax-retained
python3 -B test_portable.py /tmp/portable-historical --archives /path/to/common/.evidence/minimax-retained --historical
```

The historical run should exit1 for the reproduced defects. Positive exact argv
and outputs are in verification-summary.json and the raw artifact archive; they
read only the original Q/source-proof, actual corrective and external lineage
archives. No live-support PASS, full acceptance or further paid authorization is
created. A search of all evidence Python found no other eager gzip inventory
decode site: the two earlier candidate/formal capsule verifiers authenticate
archive hashes before reading tar members; their code is unchanged. This is a
narrow inventory-decode audit, not a blanket certification of unrelated readers.

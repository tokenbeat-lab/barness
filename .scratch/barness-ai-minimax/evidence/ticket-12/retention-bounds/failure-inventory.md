# Portable inventory boundary — failure-first scope

The parent approved the two public portable verifier CLIs as the E2E seam.
No Go, ai JSON, launch tools, selectors, capture, canonical intent or paid
process changes are permitted. Current source is integration3d3a1fb.

Before implementation, the required refusal cases are: compressed metadata
larger than8MiB; a mismatched compressed inventory hash before gzip decoding;
a valid-hash malicious gzip whose expanded content exceeds32MiB; concatenated
members whose combined expansion exceeds that same bound; invalid/truncated
compression; and original valid portable evidence. The expanded limit must
bound actual reads before JSON parsing. Hash the exact bounded compressed
bytes that are decoded, rather than rereading a replaceable path.

The resource reproduction uses a finite64MiB expanded zero payload (about65KiB
compressed), never an unbounded/GB payload. A public verifier subprocess records
peak RSS; the preregistered negative process ceiling is128MiB. Darwin refuses
lowering RLIMIT_AS/DATA here, so no enforcement claim is made for those APIs.
The original eager decompressor should fail the resource assertion even though
it eventually rejects length. Current fixtures are synthetic negative copies;
none is an actual origin or qualification. A real positive run must still
validate original source/report/seal/usage/failure/accounting bytes.

Original verifier bytes and parent indices/provenance must survive exactly.
Historical records keep their original meaning; a separate current verifier
index identifies the corrected programs. No actual evidence is relabelled.

Calibration correction before final reruns:128MiB was an initial test-only
process ceiling, not the production inventory contract. The actual verifier's
already-verified capsule baseline makes its bounded reader peak about134MiB.
The attempted unused-map cleanup is removed; it did not solve high-water RSS
and unnecessarily broadened the fix. All initial red and intermediate failed
128MiB observations remain raw history. Parent preregisters Q128MiB and actual
192MiB process ceilings with a finite96MiB expanded single-member bomb, before
running both byte-exact historical and corrected programs again. This gives
ample separation without changing the8MiB/32MiB production limits. A separate
48MiB two-member case tests the combined expansion limit. No host OOM probe.

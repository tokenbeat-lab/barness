# Release bundle consumer correction — targeted PASS, delivery PARTIAL

Consumer source: `7754c7fe9c71c1203a0c9f860da35dbf5973c3d9`; future catalog producer source: `ad20469d0a0d854ebda35c15b9c4ed49beebfda5`.
This closes the local consumer defects that aborted the original public gate.
It does not close the spec, renew paid evidence, or certify a new whole suite.

## Cause and boundary

The producer deliberately preserves an empty effort leaf in eight raw IDs:
`MM03-effort-TestMiniMaxResponsesEffort/{MiniMax-M3,MiniMax-M3.1-Flash-Preview}/{full,hook,sampling,simple}/`.
Its filesystem directory omits the final separator, and `Case.Artifacts` is
recursive. The consumer previously rejected the separator and required a flat
inventory. Every original ID, status, root assertion and artifact byte remains
unchanged. Nested artifacts cannot become a parent's differential/support proof.

The consumer accepts only one benign trailing separator and canonical physical
directories; aliases, traversal, dot/repeated separators and symlinks fail.
Streaming enumeration verifies exact unique entry/file sets. Per case ceilings
are 512 files, 2,048 entries, depth 24, 64 MiB/file and 256 MiB aggregate.
Actual maxima were 133 files, depth 2, 36,351,001 bytes/file and 63,697,858
bytes/case. Reads are descriptor-rooted and bounded before allocation/hash.

After load, a private verified-case reader reopens the original bundle inode,
refuses post-load symlink replacements and authenticates the expected artifact
hash on every consumed read. Existing 4/16 MiB semantic-reader budgets remain.
No long-lived FD or public lifecycle contract was added. The initial selected
bundle must be a real directory; its before/opened/after inode identities agree
before the bounded manifest read. Parent path aliases such as the system `/tmp`
remain valid. Manifest members never receive general path cleaning.

## Exact-source checks

| Check at 7754c7f | Recorded result |
| --- | --- |
| Targeted ordinary public consumers | 93 evidence cases PASS, audit 0, exit 0 |
| Targeted race public consumers | 93 evidence cases PASS, audit 0, exit 0 |
| Read-only original-bundle diagnostic | 1 PASS, audit 0, exit 0; imports both original bundles |
| `go vet ./...` | exit 0 |
| `go vet -tags live ./...` | exit 0 |
| Independent Standards / Spec | PASS 0 / PASS 0 at 7754c7f |

Both original bundles still have **12,391 cases = 12,385 PASS + six GLM FAIL**.
Their literal manifest hashes are unchanged. The six current-source GLM failures
were neither waived nor converted into passes. Public negative cases cover
canonical aliases, nested missing/extra/hashbad artifacts, finite bounds,
post-load bundle/case/ancestor/file replacements and same-semantics JSON tamper.
The preregistered isolated iterator exception tests a finite duplicate-name
stream; all other new regressions use public producer/import/consumer seams.

Independent original c680 findings, the 09fea initial-open finding and their
final closure are retained byte-exact under [reviews](reviews/), with
[provenance](review-provenance.json). Prior c680/09fea passing checks remain
historical intermediate checks. The [failure inventory](../../ticket-12/release-consumer/failure-inventory.md)
was written before the corresponding implementations.

## Source applicability and original execution

[Source proof](source-map-equivalence.json) recomputes complete current selected
sets, including additions/deletions: **three M3 × 263 files and three corrective
GLM × 217 files, zero drift**. The consumer and future-only producer changes are outside every selected set.
The later authorized change affects only two future catalog E2E producers,
their public admission regression and seven finite trace rows. No shared
recorder, runtime, catalog, source selector, paid intent, ledger or actual
capture was changed.

The broad legacy 1,034-file preparation guard and Jev protocol/text/visual
Go-wide snapshots are historical/stale after the consumer Go edits; all five
original corresponding jobs already failed. No source pins were refreshed and
no approval or paid budget was reused. Narrow-map equality preserves existing
qualification; it does not turn the failed GLM Responses capture into PASS.

The [original A index](../international-final-actual/A-final-keyless-3d3.json)
retains the actual source `3d3a1fb` whole ordinary/race executions, their six
failures each, zero audit findings and the outer failed gate attempt. Its four internal
numeric command exits were lost before report publication and remain UNKNOWN.
The [original L index](../../ticket-10/international-legacy-final/legacy-paid-3d3.json)
retains all closed actual jobs and their source/outcomes. Neither is replaced
by these targeted checks.

## Adopted-input public gate diagnostic

The public CLI ran once at `7754c7f` with explicit original `-bundle` and
`-race-bundle` plus all original 17 live and four effect input arguments.
The complete original output is in the compact archive under
`diagnostic-releasegate/`: **outer exit 1, report pass=false, commands=null**.
It took 176.607 seconds and executed no suites or paid calls. Original
numeric command exits remain UNKNOWN; this diagnostic cannot reconstruct them
or certify a current full suite.

The corrected consumer loads the original 12,391 cases per bundle, keeping
all six GLM failures. Revocation, multimodal and six MiniMax load checks pass.
The report still rejects missing Jev support, failed/stale live/effect inputs,
the deliberately unmerged matrix and two original differential metadata defects.
The original 869 differential PASS statuses are retained. New local producer
checks below do not change that diagnostic or its adopted historical inputs.

The initial `auditbundle` command refused this output because it has no evidence
manifest (exit 1, no audit result). A separately retained auxiliary Go overlay
then applied the existing exported `audit.File` + `audit.Environment` rules to
all eight regular output files, with finite 16-file/16 MiB-file/64 MiB-total
bounds: **eight files, zero findings, exit 0**. It adds no manifest, changes no
report bytes and runs no gate. Its source, argv, result and first refused
command remain distinct in the archive.

## Future-only catalog producer correction

The original catalog comparison omitted the actual Go version from
`sdk_versions`. The pure frozen-model lookup control also used a `PIDIFF-`
namespace although it had no two-sided differential record. Their original
PASS statuses, six artifact hashes and exact SDK facts are recorded in
[the diagnosis](final-pidiff-diagnosis.json). Neither strict importer semantics
nor frozen pi/oracle pins were relaxed.

The future comparison now records `runtime.Version()` and the actual Go
Anthropic module version using the same seam as the adjacent wire comparator.
The lookup control becomes `MM06-frozen-catalog-metadata`. It keeps all prior
assertions/artifacts and the two-case producer inventory. Exactly seven trace
rows change two identity leaves each; every other trace field/count is intact.
No HTTP request is fabricated: the comparison retains the explicit zero-valued
Request object. [Diagnosis v2](final-pidiff-diagnosis-v2.md) explicitly corrects
[v1](final-pidiff-diagnosis-v1.md)'s inaccurate JSON-null description.

The [preregistered inventory](catalog-producer-failure-inventory.md) and raw
history distinguish the initial compile typo, missing ignored node runtime,
incorrect test expectation, corrected behavioral RED and later GREEN. None
is relabelled as a provider failure or a complete source-frozen RED checkout.
The frozen runtime was an existing byte-pinned local installation, linked
only for these tests; no dependency installation or network request occurred.

| Exact ad20469 check | Result |
| --- | --- |
| Public producer → LoadBundle → differential consumer, ordinary | 1 recorded parent PASS; two child cases/all six artifacts retained; audit 0, exit 0 |
| Same public chain, parent and child race enabled | 1 recorded parent PASS; audit 0, exit 0 |
| Coherently rehashed missing-Go and forged-pi records | inventory valid, strict SDK semantics reject both |
| `go vet ./...` / `go vet -tags live ./...` | exit 0 / exit 0 |
| Independent Standards / Spec at ad20469 | PASS 0 / PASS 0 |

These targeted checks grant no new full-suite certificate. The original A,
its 869 differential PASS statuses/six GLM FAIL statuses, diagnostic false
report, failed live renewal and all paid one-shots remain literal history.

## Retention and replay

[index.json](index.json) pins the complete compact raw archive and bounded
member inventory; [commands-summary.json](commands-summary.json) keeps every
actual red/green exit and argv distinct. The archive includes full original
test bundles/assertions/artifacts/audits, logs, source slices and the complete
diagnostic output. Every member and every recorded `Case.Artifacts` reference
was compared byte-for-byte; `Fixture` raw-source attribution is not substituted
for the retained artifact hash. Early RED tests were uncommitted when run;
their complete working tree was not separately frozen, which is explicitly
recorded rather than retrospectively assigning a later source snapshot.

Authenticate with `shasum -a 256 -c SHA256SUMS` before extracting the archive
into a new directory. Re-run the recorded targeted argv from the tested source,
using a fresh absolute `BARNESS_AI_EVIDENCE_DIR`. Recorded-bundle diagnostics
add `BARNESS_RELEASE_DIAGNOSTIC_BUNDLES` as a JSON array containing the original
ordinary and race directories. They read these inputs only. Original paid
one-shots stay closed; there is no automatic rerun or further model call.

The archive is also copied byte-exact to the common repository's ignored
`.evidence/minimax-retained/release-consumer-ad20469.tar.gz`; tracked source,
review reports, index and checksums provide its portable front door. Original
A/L large archives are linked above rather than duplicated. No cleanup may
remove their only retained copies. All targeted raw bundles and command/source
metadata are inside this compact archive before temporary-worktree cleanup.

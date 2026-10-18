# International catalog activation preflight

Code checkpoint: `de051931b8d4e30fb6adf5cc5b2a3672f243ba35`.
This work reads no credential or `.env`, invokes no vendor and grants no paid
approval. Integration remains inactive; the separate Git fixture activates only
existing sourced M3 data for offline verification, without promotion receipts or
actual support claims. [ADR0044](../../../../../docs/adr/0044-barness-ai-historical-catalog-replay.md)
records the historical/current authority boundary. The prior research at source
38b7098 remains immutable, with 10 canonical reservations/9 actual Attempts
including the historical CN 2/2 prefix.

The initial one-file compiler overlay exposed the GLM replay blocker and current
version literals, but also made Jev's compiled/disk snapshots inconsistent. A
separate shared-object clone then used complete catalog/release/route data: Jev
passed; the GLM, Decisions and six MiniMax replay failures persisted. The old
helper additionally left legacy Chinese default and Responses usage pins stale.
No overlay-only Jev failure is called a production regression. Both red states
are retained separately.

Historical formal GLM replay now loads the original catalog from the pinned
original Git revision with bounded reads, checks its canonical hash and exact
original report version/hash/source, and preserves Result/Observer metadata.
Missing Git history, wrong hash and oversized snapshot refuse before replay.
Replay cannot publish/import current support; actual formal and all six stale
current GLM source/catalog checks remain strict. MiniMax replay selects between
the two existing finite catalogs by complete identity, avoiding their shared
version-label collision.

Provider source facts and unchanged old model content remain exact. Current
attribution follows the independently committed release snapshot; the separate
complete usage content pin is preserved. Only the controlled Host fixture's
MiniMax provider subset is replaced with its six synthetic candidates, retaining
M3.1 mechanisms. The renewal CLI synchronizes only catalog version/hash across
four current pin documents plus the generated snapshot, retaining every other
field and legacy 180s versus route 600s limits.

Both independent source axes pass with zero findings for
[de05193](../international-preflight-review-de05193.json) and
[df342b9](../international-business-boundary-review-df342b9.json).

| Check | Result | Complete archive/index |
| --- | --- | --- |
| MiniMax ordinary / race at de05193, frozen pi + pressure | 1,340 PASS each; audit0; both guards unchanged | [Ordinary](preactivation-ordinary-de05193.json), [race](preactivation-race-de05193.json) |
| Focused live ordinary / race | 67 PASS each, audit0 | [Index](preactivation-focused-de05193.json) |
| Inactive affected E2E at de05193 | 156 PASS, audit0 | Same focused index |
| Complete activated live | Go exit0; 331 PASS / 62 expected unselected vendor NOT_RUN / 0 FAIL; audit0 | [Index](preactivation-activated-live-de05193.json) |
| Complete activated ordinary E2E | Go exit1; 12,378 PASS / 11 FAIL; audit0; guard unchanged; 570.53s | [Full immutable index](preactivation-activated-whole-de05193.json), [triage](whole-failure-triage.json) |
| Fixed business task, active ordinary / race | 24 PASS each, audit0; race guard unchanged | [Index](preactivation-business-active-df342b9.json) |
| Same current source inactive ordinary / race | 23 PASS each, audit0; both guards unchanged | [Index](preactivation-business-inactive-df342b9.json) |
| Renewal public CLI / ordinary and live vets | 11/11 PASS / both PASS; repeated vets PASS on df342b9 | Retained command receipts in focused/business indexes |

The whole ordinary result remains 11 FAIL: six stale GLM current-source/catalog
imports, three release-task leaf failures plus their aggregate, and the private
runtime-symlink setup failure. The latter five are proven by the focused fixes;
the six GLM requirements remain unqualified until genuine current formal renewal.
The full 12k suite was not repeated or relabeled. No formal/live release PASS
follows from these controlled checks.

[Red index](preactivation-red-de05193.json) retains all original overlay,
consistent-activation, affected E2E and snapshot-boundary failures.
[Business red](preactivation-business-red-df342b9.json) retains the old importer
with the new public tamper tests. Its original positive failures preceded the
minimal source change. Raw Fixture attribution hashes are kept separate from
Artifact identity; all retained outer Artifact references match exact bytes.

All nine full archives remain in `$COMMON/.evidence/minimax-retained/` with
SHA256 and full compressed file inventories in their indexes. Archives below
8MiB are additionally tracked beside this README. Exact MiniMax command/log,
source-at-start and both source guards are retained as the `mm-*` files linked
and pinned in the ordinary/race indexes. The original complete failed
whole archive is 111,557,747 bytes and retains all 131,964 regular files, original
requests/responses/events/results/Observer/resources/replays, command/log,
manifest/assertions/audit and source guard. Forensic snapshots also preserve
explicit deliberately invalid symlink fixtures as inert TAR link metadata;
curation never follows them and they are not formal import bundles.

The [original activation Git bundle](activation-fixture.bundle) preserves fixture
688b1e6; [follow-up bundle](activation-followup.bundle) preserves 2f0ce8f.
Raw execution commits remain in every index result; the fix checkpoint is not
a replacement for the original test source. Initial TDD probes retain their
new-test/original-implementation state and original metadata. Both bundles verify against repository prerequisites 09d0c68/7c28174. The
[six-document patch](activation-data.patch), [fixture metadata](activation-fixture.json),
[follow-up source equality](business-source-provenance.json) and private
[exclude setup](fixture-setup.json) retain exact reproduction inputs. Code is
identical to the reviewed checkpoints except the explicitly controlled catalog,
generated snapshot and four current pin documents. All other provider catalog
content and all noncatalog config fields remain exact; frozen runtime is a
read-only prerequisite, not copied dependency upgrades.

[Preservation](preservation.json) verifies old14 rows, 317 original decisions,
six original routes, frozen pi pins, the historical CN prefix/intent/34 files,
canonical10/9, inactive real catalogs and all 1,026 final generation-source files
unchanged after the review merges. [Curation](curate.py) retains bounded scans,
artifact-only checks and independent complete archive byte verification.
`checksums.sha256` includes the exact parent review reports as well.

Current real candidates/formal/old-route/effect renewals and final release
acceptance remain open. The parent will fresh-prepare/check exact merged caller
paths and approval before actual work.

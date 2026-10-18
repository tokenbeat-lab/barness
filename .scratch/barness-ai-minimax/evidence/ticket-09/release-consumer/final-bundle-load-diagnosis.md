# Final A bundle-load failure — causal diagnosis and approved scope

Read-only diagnosis against root source3d3a1fb and unchanged A artifacts. Own
branch e3b90a3 contains only the separately reviewed evidence-Python fix. No
test, paid call, credential read, source/root/capture rewrite, rename, omission
or retry was performed to establish this cause. Parent accepted this causal
analysis and explicitly approved the consumer-only scope described below,
including the global L/Jev-source caveat, before implementation.

## Exact offending records

Both actual main manifests have12,391 cases:12,385 PASS and6 FAIL. Their
identical eight violating records are PASS cases from the empty effort-string
subtests of TestMiniMaxResponsesEffort. They are at zero-based indices4247,
4254,4262,4270,4277,4284,4292,4300.

| Raw case ID and recorded dir (unchanged) | Verified artifact references per bundle |
| --- | ---: |
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3.1-Flash-Preview/full/|49|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3.1-Flash-Preview/hook/|56|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3.1-Flash-Preview/sampling/|56|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3.1-Flash-Preview/simple/|49|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3/full/|49|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3/hook/|56|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3/sampling/|56|
|MM03-effort-TestMiniMaxResponsesEffort/MiniMax-M3/simple/|49|

Main manifests:

- offline/20261018T045647.697909000Z/manifest.json:15,550,847 bytes,
  SHA256b4dcf3e8772285e0cd06e03dc627be0dad33315199d8e9965c734187407af21e.
- race/20261018T050644.512932000Z/manifest.json:15,550,847 bytes,
  SHA25614b1f2b38e69e62ac628554c8db436390aeeae682086f0dd9291ab3865a2c907.

Their base is `$COMMON/.evidence/minimax-global-final-keyless/gate/`.
Every offending raw dir ends in one slash; removing that slash yields its
actual canonical contained directory. None is empty, absolute, dotdot-prefixed,
a duplicate ID, a symlink component or outside the bundle. All12,391 physical
canonical paths are distinct, not merely the8 offending paths. Each affected
case's complete artifact map matches its original file bytes:420 hashes per
bundle,840 across both. Both recorded audits are clean. Full safe case/hash
metadata: /tmp/barness-minimax-final-load-offending-cases.json.

## Producer/consumer boundary mismatch

Producer ai/e2e/minimax_responses_effort_test.go:13–20 includes the empty
level string in t.Run(model/source/level) and uses MM03-effort-+t.Name() as
the opaque case identity. Empty effort omission has real tested meaning;
removing/renaming those cases would lose that meaning.

Evidence Run.Case uses filepath.Join for the actual physical directory
(ai/internal/testkit/evidence/evidence.go:101–107), which cleans a trailing
separator. Run.Finish records dir=c.id verbatim instead of the normalized
physical path (:126–137). The produced manifest therefore retains the slash.

LoadBundle rejects filepath.Clean(c.Dir)!=c.Dir
(ai/internal/testkit/release/load.go:50). That comparison alone causes the
error for these8 entries; every other part of its combined condition is false.
The first entry aborts loading, so the gate writes no final release report.
This is benign producer/consumer canonicalization disagreement, not evidence
of traversal, duplicate aliasing, redaction failure or capture tamper. The
manifest is noncanonical under the current consumer rule; accepting only this
specific contained producer representation is narrower than accepting arbitrary
Clean(input). The original security rejection must remain literal in A.

## Proposed minimal consumer correction

Keep Case.ID and original manifest bytes verbatim. Permit only a canonical
relative directory optionally followed by exactly one trailing slash. Use its
canonical physical path for reads and a separate seen-directory map; reject
'a' plus 'a/' under different IDs as an alias collision. Preserve duplicate-ID
checks. Reject empty/dot/absolute/traversal/internal dot/repeated separators,
backslash/escape forms, symlink path components and symlink artifacts. Do not
broadly normalize arbitrary unsafe inputs or change evidence producer, shared
recorder, source selectors, captures or assertions. Existing inventory/hash
checks remain in force; missing evidence stays a problem rather than a PASS.

The current inventory reader rejects nonregular artifact entries but follows a
symlink case directory via ReadDir. The same consumer-boundary tests should
therefore explicitly reject case-path and ancestor symlinks; no relaxation of
that boundary is justified by the benign trailing slash.

## Exact source applicability assessment

Every map below was compared against both current root and own worktree bytes,
with zero pre-change drift. Full metadata:
/tmp/barness-minimax-final-load-source-map-assessment.json.

- Each actual M3 formal map:263 files, all3 exclude release/load.go and all
  ai/e2e test files. Consumer-only load.go plus a new E2E test leaves these
  generation-source sets/bytes unchanged.
- Each Q corrective GLM map:217 files, all3 exclude those paths. They retain
  current selected execution-source applicability; authentic failed Responses
  does not acquire qualification. GLMExecutionSource explicitly excludes
  release plumbing (supportmatrix/glm_source.go:9–12).
- The10 current generic own reports do not carry an own runtime hash map;
  their original3d3 GitCommit and original canonical preparation must remain
  literal. This is not permission to relabel them as newly executed on a new
  whole-source HEAD.
- L's global prepared source set1034 DOES include load.go and allGo tests. A
  consumer Go edit makes that old whole-source guard historical; no reuse of
  its approval/paid launcher is permitted.
- Jev1.14 protocol map852 includes allGo including tests; text effect maps851
  do likewise; visual maps321 include non-testGo including load.go. All5
  existing producer jobs already FAIL. They become historical/stale after
  this edit and must never be refreshed, relabelled current PASS or reused
  as fresh certification.

Parent explicitly accepted these caveats before authorizing implementation.
No statement that all1,034 generation bytes remain unchanged after a Go fix
is valid. Only the measured M3/Q selected execution maps can remain unchanged.

## Minimal public E2E red/green plan

Use the public evidence producer plus release.LoadBundle boundary, generating
8 controlled cases with the exact offending opaque IDs. A first failing test
must prove the legitimate one-slash form is refused today. Add public negative
imports for canonical aliases, duplicate IDs, traversal/dot/absolute/repeated
separator and symlink/escape forms, with exact artifact hashes. No forged
provider or live qualification follows from these controlled cases. Retain
repeatable assertions/manifests/import results. Then apply the smallest loader
fix and run targeted ordinary/race plus both vets; no automatic47-minute whole
suite repeat and no paid call.

A positive readonly diagnostic load of each original main A bundle must preserve
all12,391 original case IDs/statuses and raw manifest/artifact hashes. Any
existing artifact problems and all6 GLM source failures remain visible; loader
acceptance is not a full release PASS. Do not edit/adopt the old raw evidence.

The original full ordinary/race executions stay source3d3,12,385 PASS/6 FAIL,
audit0. Their four internal command receipts were lost when the old gate aborted
before writing its report; numeric command exits must remain unknown, not
reconstructed as passed. The original outer exit1/terminal exitstatus2 and
loader error are retained by parent.

A later corrected consumer may diagnostically reevaluate the originals through
the public -bundle/-race-bundle mode. It must clearly say adopted historical
inputs, command gates NOT_RUN, not newly executed suites, and not fresh-source
full certification. Missing Jev support and all real failed live/effects inputs
stay failed. Existing scalar paid headroom grants no R2, process rerun, scene
splice or source-hash repair.

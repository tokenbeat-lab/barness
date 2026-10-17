# Ticket 14 failure inventory — before implementation

Agreed seam: Client.Classify / HookedClient.Classify against controlled HTTP,
then the existing public releasegate command. Source baseline: bf9e8570cc4d3bff97bf37076f0e8321335c6692.
Dependencies 09/13 are resolved. Existing user edits and historical evidence stay intact.

- Release contract still selects ten routes and omits independent fixed Jev 1.14;
  old 1.13 or another model could stand in for its current evidence.
- New R01–R17, T01–T16 or user stories are unmapped, mapped too broadly, or
  mapped without current passing assertions. Missing plain/hooked runs could pass.
- Mixed images are absent from normal/race imports, or a saved PASS lacks actual
  load, bounded memory, input/response/result/Attempt/Observer and zero-release proof.
- Finite Pi case/path extensions could widen or be reported as differential PASS;
  the frozen baseline, original projections or original design load could drift.
- Catalog/price/source/config/SDK changes could borrow old live/effect results;
  controlled, candidate, wrong model, partial budget or UNSUPPORTED could become PASS.
- Missing credential/authorization must remain NOT_RUN, with zero real calls;
  no reading another task's credentials or silent paid retries.
- Package/inventory/audit failures, Observer content or non-atomic results could
  be hidden by process success or an unrelated expected gate failure.
- The historical pressure-stream-integrity #07 delivery blocker could disappear
  after a successful rerun. Its outstanding causal evidence must still prevent PASS;
  neither its status nor the parent spec/ADR may be silently closed.
- Migration, experimental opt-in, format/location refusals, legend/index and
  versioned revocation docs could disagree with the executable Host example.
- Existing public Assets/obsolete internal branches/shims, or accidental staging
  of user changes, could survive integration unnoticed.
- A failed/incomplete run could overwrite earlier evidence, lose audit/finalization
  reserve, or be represented as current release completion.

Vertical slice 1: execute image+Box through each public entry at the built-in
catalog, then require eleven independent live identities in the CLI report and
frozen trace contract. Confirm RED before changing release implementation.
Subsequent slices add missing trace and mixed-load/blocker contracts with their
own RED, without adding unit tests after implementation.

Final contract audit (before the visual-integration slice): ticket 12 requires
both independent visual tasks at the new formal Hash. Text effects or Decisions
protocol smoke cannot replace its visual task; a Jev-only -formal producer leaves
this requirement impossible even with credentials. Missing/candidate/wrong-route/
wrong-model/stale-source/bad-budget/incomplete or forged-live visual packages
must fail closed. First add formal Decisions public process E2E and required
visual CLI artifact regressions, then extend producer/consumer together.

Review follow-up, before corrections: the visual verifier launched through `go run`
can leave its executable alive after the parent deadline; exercise the public
release CLI with an isolated fake toolchain and a sleeping verifier. The receipt
collector can skip visual-specific manifests; sealing and then deleting an actual
controlled visual capture must make the public receipt verifier fail.

The retained pre-visual full run hit Go's 20-minute race process timeout while
its final artifact audit was still running: all 9,081 cases had terminal PASS,
but audit.json was never written and the command failed. Preserve that failure;
allow 30 minutes for the same full command without changing any scenario load,
deadline, budget or assertion. Bound the outer all-command collector at 60 minutes.

Collector preflight against actual retained recorder output exposed captured
manifest.json files being mistaken for independent recorder roots. Some are
intentional corrupt fixtures. Before changing collection, retain the failed
precheck; distinguish inventory-owned captured manifests and preserve invalid
nested fixtures byte-for-byte without accepting their asserted PASS.

The actual live-tag harness intentionally terminates two child recorders before
audit finalization. Its parent starts first and verifies those crash boundaries.
Collector discovery must keep parent validation strict while recording child
MISSING audits explicitly; it must not reject the passing parent harness or
pretend those killed children completed their audits.

Spec review found the live-tracker E2E hardcoded unresolved=false/Current=false:
legitimate original resolution would fail the future release. Before correcting
that assertion, exercise it in a sanitized temporary repository source copy with
only the copied original tracker changed to resolved. Preserve that failing public
CLI E2E, then make the assertion reflect the real tracker and prove both states.
The working original issue is never edited or closed.

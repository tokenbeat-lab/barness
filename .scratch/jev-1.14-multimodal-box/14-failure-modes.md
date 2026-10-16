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

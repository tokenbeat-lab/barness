# GLM replay diagnostics inside HOME

The original actual-root keyless preparation at `9dcf896` completed all public
assertions (335 PASS, 62 NOT_RUN, zero case FAIL) but exited1 because the
unchanged audit found two HOME paths. Nested synthetic replay stdout included
the checkout-based temporary evidence directory in both its exit record and
assertion record. The original Q, checks=false, prepared source/binary/plan
and command receipts remain immutable; this is not a paid or capability failure.

The agreed public executable/audit seam must work with the checkout beneath
HOME, retain diagnostic session/path context, and reject unrelated environment
leaks. The new controlled-HOME process case reproduces these exact two audit
findings before the fix. Its inner HOME is isolated; no user's environment or
shared audit rule is changed. A small socket-free Responses capture exercises
new-current replay, exact Result/Observer/wire/cancel matching and refusal of
actual support publication.

`ecce867` registers only the generated synthetic replay root through the
existing evidence redaction API, with alias `$GLM_REPLAY_ROOT`. The diagnostic
keeps its `/replayed/<session>` suffix. Registering the entire HOME would mask
unrelated leaks and was not used. The only modified Go helper and the new
regression file are GLM-owned; shared runtime, audit, source selectors, importer
acceptance, catalog and original capture bytes are unchanged.

The initial failure-first command exits1, with the outer expected-failure
assertion and two inner environment findings retained. The same public case
then exits0 with audit0 and verifies the alias and suffix. Focused ordinary/race
checks, both vets and all three existing M3 public formal verifiers are recorded
in the accompanying diagnostic archive. Each M3 execution map remains263
byte-identical sources, with unchanged original report hash. This is offline
verification, with no credential acquisition or vendor request. Corrective-1
remains PENDING HUMAN APPROVAL; a new actual-root full no-key check must pass
on the integrated source before any such approval can qualify.

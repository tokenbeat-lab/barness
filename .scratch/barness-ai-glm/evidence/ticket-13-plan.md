# Ticket 13 acceptance plan (before implementation)

Public seams: `ai.Client` through the executable `ai/examples/glmhost` Host example, actual releasegate CLI subprocess, actual baseline public Client after complete Git-tree restoration. These are the spec R04/R21/R22 and issue13 authorized seams; no new private-module seam or user approval is required.

Failure hypotheses, before source:

- Example absent, wrong protocol endpoint/auth/options, candidate catalog, synthetic price, missing versioned tenant/account snapshots, mutable shared models, or missing stream cleanup makes a Host integration unusable.
- A disabled GLM binding is silently enabled, reads credentials, sends a fallback request, or changes an active call/prior reported cost. Another route/tenant becomes unavailable. Restored authorization uses a stale version.
- A Host executes an incomplete/failed tool call instead of waiting for StopReasonToolUse, or reuses the same RequestID.
- Final release command has the old implicit10m/30m package envelope rather than preregistered20m/45m outer bounds; scenario/runtime/paid limits must stay unchanged.
- Stale catalog/old evidence, missing one of14 identities, missing old P0/7 native differentials/6 extensions/race/pressure/effects/audits or forged replay/source can incorrectly pass (existing12 public negative proofs retained;13 executes full gate).
- A partial rollback leaves new runtime/capabilities/catalog/matrix/traceability or changes root/user-untracked bytes. Full baseline tree identity plus public no-GLM refusal and old-provider controlled test must prove restoration.

Vertical slices: first failing Host example E2E, implement example, prove both models/three APIs/four entries/hooks/tool loops/cleanup; then disable E2E; then outer-envelope public command failure; then no-key source validation and prerelease review before any paid calls. Current-hash collection is separately reserved once at302 Calls/Attempts maximum, with original component limits, selected named keys/direct transport and no automatic rerun. Actual full normal/race release commands run afterward; no -bundle substitution.

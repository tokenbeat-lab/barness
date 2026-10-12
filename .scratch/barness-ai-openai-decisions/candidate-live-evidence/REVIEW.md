# Issue 09 two-axis code review

Fixed point: `3f97f1992434dc1d4766628fca900fa42b1b1e78`.
Initial review: `git diff 3f97f1992434dc1d4766628fca900fa42b1b1e78...85ea917`.
Focused review: `git diff 85ea917...a7c15ec`.
Two independent read-only agents; no real API calls or source edits by reviewers.

## Standards

Initial actionable finding: **P2, successful report survived integrity failure**.
`decisions_report_test.go:74` returned exit 1 on inventory read/write failure after
persisting Complete=true/Outcome=PASS. Violates AGENTS principle 9, requiring
verifiable evidence and observable failures. No other actionable hard breaches
or judgment smells found.

Final: **zero remaining actionable Standards findings**. Fallible reads keep
Complete=false; finalization failures persist FAIL/evidence and revoke acceptance.
Public process cases cover inventory read and write faults. Canonical permission
fields and execution finalization reserve preserve configuration/evidence boundaries.

## Spec

Initial actionable findings:

- **P1, no time before hard process timeout**. `decisions_candidate_test.go:161`
  used started+180s while replay used Go -timeout 180s; remaining failure summaries
  and cleanup could precede a panic that discarded report/manifest/audit. Issue 09:
  “180 秒总 … 失败和预留均入账”; acceptance: “成功与失败都保留”.
- **P2, noncanonical permission aliases accepted**. Exact duplicate checks were
  followed by Go's case-insensitive decoder. `dedicatedTestKey:false` plus
  `DedicatedTestKey:true` yielded PASS/Complete=true and nine controlled calls in
  a temporary overlay reproduction. Issue 09: “错误配置拒绝”.

Final: **zero remaining actionable Spec findings**. Canonical field names are
required before decoding. Runtime chooses the earlier process/test deadline and
reserves 10 seconds for finalization. Four RED-first public process cases passed:
permission aliases, real Go deadline, integrity read and integrity write faults.

Dedicated-key/account/region/low-permission evidence remains an external acceptance
blocker. Controlled PASS is not live acceptance. Candidate separation, source
pins, typed refusals, usage presence/pricing and unchanged formal catalog/matrix
otherwise match the reviewed issue scope.

Final counts: Standards 0, Spec 0. Initial findings were Standards 1 (P2), Spec 2
(P1/P2); each axis was independently re-reviewed after a7c15ec.

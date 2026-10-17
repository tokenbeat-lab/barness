# Ticket 14 independent two-axis review

Fixed point: `bf9e8570cc4d3bff97bf37076f0e8321335c6692`.
Command: `git diff bf9e8570cc4d3bff97bf37076f0e8321335c6692...HEAD`.
Implementation: `ceb4625`; semantic/trace fixes: `1ded61e`; runner fix: `0f5d99a`; independent visual integration: `823cc4a`; tracker-state regression: `4310a53`.
The code-review skill explicitly requested independent parallel Standards/Spec agents.
Both agents reviewed read-only; preexisting user documentation/planning was excluded.

## Standards

Final independent report: **no remaining findings**. The two P1 evidence-validation
findings are resolved by `1ded61e`. The P2 subprocess-timeout finding is resolved by
`0f5d99a`: output collection no longer blocks the deadline; timed-out process groups
receive TERM/KILL, the child is reaped, and exit 124 plus `timedOut: true` is recorded.
The retained timeout check completed in 0.237 seconds. Baseline smell judgment calls:
none worth reporting.

Initial P1s: native wire/response files were only length/count checked, and answer
values were only counted. Four public CLI tamper regressions first reproduced PASS
on bad model/null answers/renamed answers/empty responses, then passed only after
complete fixed-workload input/wire/response/answer/usage reconciliation was added.
The immutable controlled fixture also pins each tenant's synthetic image hashes.

Initial P2: the old runner drained stdout before invoking `wait(timeout)`. The
reviewer reproduced the old loop in memory: a child sleeping 0.4 seconds, followed
by `wait(timeout=0.05)` after draining stdout, returned exit 0 after 0.43 seconds.
The manual script CLI regression uses fake `go` sleeping ten seconds and
`--command-timeout 0.1`; its retained result is
`.evidence/ticket14/timeout-check/result.json` (expected exit 124, process reaped).
No post-implementation unit test was added.

## Spec

Final independent report: **0 remaining code findings**. Mixed-load imports now
reconcile public inputs, native requests/responses, named answers, usage and image
hashes against the fixed controlled workload. Trace requirements pin complete case
inventories; missing any required case gives NO_EVIDENCE. The original one-case
reproduction independently now yields NO_EVIDENCE for UC-R01/R03/T03/T04/US01;
T04 reports 182 missing cases. Its selectors include JSON, invalid-plan, value and
scope cases. The finite extension registry is pinned, consumed, retained in accepted
release outputs, and covered by corruption regressions. No scope creep identified.

Initial P2s: invalid rehashed response/answer content could pass the capacity
consumer; one successful escaped-key call could label entire acceptance groups PASS.
Both were reproduced through existing public Classify/CLI boundaries before fixes.

Final code review is separate from execution acceptance. Missing real-service
evidence and the historical pressure blocker still prevent ticket/whole-release
completion. No reviewer promoted controlled HTTP or a rerun into real-service PASS.

The final contract audit additionally found missing Decisions current-formal visual
effect integration (ticket 12 / ticket 14 AC4). `823cc4a` adds both routes' fixed
formal producer selection, independent required visual imports/audits and the
`UC-visual-effects` trace. Its 12 public process E2Es were RED before implementation;
all 32 UC14 focused cases pass afterward. Only verified live COMPLETE can PASS;
controlled COMPLETE remains NOT_RUN. Data, scoring and 12-call budgets are unchanged.
Final independent Spec rereview confirms the committed visual/trace path. The
collector's Go revision metadata is now derived from the execution report rather
than a hardcoded earlier commit.

The final Standards rereview found two more P2 issues, both fixed and independently
cleared. The visual verifier now executes a built binary directly under its shared
60-second deadline; its public CLI fake-toolchain check records the old surviving
child and the corrected reaped process in `visual-timeout-before/after/result.json`.
The receipt collector recognizes visual manifests and verifies complete inventories,
hashes and audits; deleting an actual controlled visual response now makes the
public verifier fail (`visual-seal-before/after/result.json`). Missing historical
audits remain MISSING. No post-implementation unit tests were added.

A further collector preflight found captured manifests being misread as recorder
roots. The first ownership-based correction had one P2 fail-open edge: an invalid
fixture index could claim a real recorder manifest. The final design removes that
ownership mechanism. It validates known process-output locations and retains all
remaining nested files through filesystem-derived complete tree indexes. No nested
manifest can suppress primary verification. The independent final Standards report
confirms no remaining findings; final public deletion regression is retained at
`visual-seal-current/result.json`.

Standards: 0 remaining (2 P1 + 4 P2 resolved). Spec: 0 remaining code findings
(2 P2 plus the missing independent formal visual integration resolved).
Execution acceptance remains separate and BLOCKED.

The final Spec audit found a latent E2E defect: it required original #07 to remain
unresolved forever. `4310a53` first reproduced rejection of a legitimate resolved
state in an isolated source copy, then changed the assertion to follow the actual
tracker. Both synthetic states pass; the original tracker was never modified.
The complete focused suite now contains 33 passing UC14 cases. Both independent
axes cleared the slice. The receipt also preserves the actual tracker snapshot
and verifies its recorded hash rather than demanding perpetual unresolved status.

The attempted full run at `final-823cc4a` was deliberately interrupted before
changing source; it is retained as partial history, never a complete release run.
Final acceptance runs at `final-4310a53` with stable Go sources.

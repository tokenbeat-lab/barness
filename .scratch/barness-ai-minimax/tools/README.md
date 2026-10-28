> 固定轮次控制器已按 [ADR-0055](../../../docs/adr/0055-route-scoped-evidence-applicability.md) 从活动实现退役。原暂停、授权、账本和失败事实保留；本页旧接口仅按固定 Git 版本追溯，不再作为执行入口。

# MiniMax task-10 bounded collection launcher

The executable boundary is `python3 .scratch/barness-ai-minimax/tools/paid.py`.
No command approves its own source. All actual vendor runs require the parent
reviewer's exact-source `APPROVED` receipt, the prepared binaries, the committed
source set and successful credential-free checks.

```sh
python3 .scratch/barness-ai-minimax/tools/paid.py prepare .evidence/minimax-ticket10-candidate
python3 .scratch/barness-ai-minimax/tools/paid.py check .evidence/minimax-ticket10-candidate
python3 .scratch/barness-ai-minimax/tools/paid.py preview .evidence/minimax-ticket10-candidate mm-global-research-m3-chat
```

`prepare` refuses dirty/untracked execution source, freezes all recursively found
ai Go/JSON plus the actual MiniMax source snapshots, module files and every
launcher Python script, builds live/text/visual binaries without credentials and
writes `prerelease-source-review.template.json` with `PENDING`. After reviewing,
the parent writes `prerelease-source-review.json` using the exact hashes from
that template, `reviewedBy`, `decision: APPROVED`, explicit `authorizedRoutes` and
`credentialScopes: {"MINIMAX_KEY":"global-standard"}`. M3.1 is independently opt-in:
its route requires `MINIMAX_MPLAN_KEY: global-mplan` in credentialScopes and that
separate selected variable in the common repository's `.env`.

```sh
python3 .scratch/barness-ai-minimax/tools/paid.py run-selected .evidence/minimax-ticket10-candidate mm-global-research-m3-chat
python3 .scratch/barness-ai-minimax/tools/paid.py run-selected .evidence/minimax-ticket10-candidate mm-global-candidate-m3-chat
python3 .scratch/barness-ai-minimax/tools/paid.py summarize .evidence/minimax-ticket10-candidate
```

Repeat only the explicitly reviewed protocol names `chat`, `responses`, and
`messages`; there is no batch paid command or process retry. The fixed full
inventory has13 cases plus one possible tool-call rerun and at most3 initial
transient REQUEST scene reruns (once per scene, no started-stream replay); optional research has3
cases. Both share the same per-route/model/purpose budget-slot20calls/24attempts/40000output/120000input
harness ledger. Candidate stage120/144 and formal60/72 survive new collection
roots and worktrees. Actual ledgers always use the common Git repository root
`.evidence/minimax-stages/{candidate,formal}.jsonl`, with no override. Launcher
exclusive one-shot intents use `.evidence/minimax-launches/<route>.json` in that
same common root. Missing/invalid/interrupted outcomes conservatively consume
the original whole budget-slot/component allowance. Never delete these receipts
to rerun; a budget/inventory change requires a separate approved specification.

After own candidate acceptance and the parent-owned data-only catalog promotion,
prepare a fresh collection root and review its new source/catalog/binary pins
before `mm-global-formal-m3-<protocol>` or any old-source renewal. This is required even
when the Go execution bytes stay the same. Legacy route names are the ten
generic formal identities, `protocol114`, `glm-chat/responses/messages`,
`text-jev/luna`, and `visual-jev/decisions`. Each route retains its existing
selector, models, direct endpoint and component cap:302/302 generic/effects plus
72/96 GLM =374calls/398attempts. No cost/dollar ceiling is inferred.

Every paid child receives only one selected credential, a small environment
allowlist, fixed route switches and finite timeout. Ambient keys, proxy and
endpoint switches are excluded. Command receipts and process output redact the
selected value, repository/common root and HOME. Child artifacts are audited for
the selected key without rewriting immutable capture bytes; failed audits stay
failed. A pre-existing selected output directory is refused before credentials
or budget intent so stale reports cannot masquerade as current evidence.

`keyless-plan <collection>` prints verification, atomic merge and actual gate
argv after all own bundles and the derived Jev support bundle exist. It verifies
the existing Jev support bundle and does not recreate it. The printed `$REPO`
placeholder expands to this checkout when executing the displayed argv. It preserves
exactly17 own formal inputs (10 generic + Jev1.14 + GLM3 + M3three) and all four
text/visual effects. The gate argv has neither supplied suite bundle override;
its public command runs the actual normal/race/vet/load/differential requirements.
The launcher only plans these commands; the parent executes them without keys
and checks the child exit, fresh report presence and `pass: true`.

Controlled executable checks:

```sh
python3 .scratch/barness-ai-minimax/tools/test_paid.py
```

These use new synthetic Git repositories, fake Go builds and Python producer
processes, including a directory at `.env` proving keyless/pre-review paths do
not open it. They never read the real credential file or contact a vendor.

Optional scene reruns exclude every429 and any HTTP error with a nonempty
`error.type`; typed permanent errors never become eligible merely because the
HTTP status is transient. Untyped408/409/5xx and initial transport/deadline
failures remain bounded by three scene reruns, once per scene, with no started
stream replay. Raw response fields, phase and event sequence are rechecked by
the independent verifier. Transport replay uses the recorded failed exchange;
a captured deadline retains its original bounded call timeout.

Research captures are copied byte for byte only after the fixed child/source
and credential audit checks into common-root
`.evidence/minimax-research-captures/mm-global-research-<model>-<protocol>/<session>`.
A streaming bounded scan checks the original seal before exclusive retention.
Candidate plans use this stable root across new prepared source collections or
worktrees; all earlier failures/calls/attempts remain in the same stage ledger.

A successful MiniMax child self-check is structural. Actual promotion/import
also requires the trusted external launcher origin pins, canonical ledger
prefix, or the curated repository verifier index. The keyless formal publisher
runs after the launcher records original report/seal/ledger/source/binary/plan
pins. GLM renewal likewise requires the original four-file capture pins in its
canonical launcher receipt. Metadata settlement preserves real producer DTOs:
per-scenario support outcomes (including documented UNSUPPORTED) and COMPLETE
business-effect status; it does not replace public certification. Original
producer report hashes prevent later edits from lowering settled counters.

Known producer settlement reads the original GLM `formal-glm-report.json` even
when public support projection fails. Other support rows retain their real
scenario outcomes; documented UNSUPPORTED scenes can coexist with passed
mandatory scenes. Effect COMPLETE status is retained as successful execution
metadata. The original report hash is pinned before later summaries; changed
reports become UNKNOWN and consume the original ceiling. Both producer
reservation spellings, `attemptsReserved` and `reservedAttempts`, are retained.
This settlement never changes the public certification/import rules.

If the bounded research tool probe fails, the dependent native continuation is
explicitly NOT_RUN and receives no synthetic Call or reservation. Original
failed probe captures, source applicability and receipts remain byte for byte
in the canonical research archive. The complete candidate still executes the
unchanged thirteen-case inventory and accounts for those actual prior calls.
Current global predecessor roots never select archived CN captures.

## Ticket 11 failure inventory (before implementation)

The current executable inventory must select only global job names, explicit
global account/region aliases and global-standard/global-mplan reviewed scopes.
An old CN scope, job name, predecessor root or canonical origin must fail before
any credential acquisition or new intent. New global jobs must preserve the old
CN exclusive intent byte for byte and use the same finite stage/budget slots.

Historical accounting must include the immutable original CN Chat/M3/standard
2 calls, 2 attempts, 5120 requested output and 4096 reserved input exactly once,
including when a current report repeats cumulative ledger totals. Missing,
mutated or incomplete historical receipts/ledger prefixes must consume the full
20/24 slot; unknown global work must do the same without adding a second cap.
Cross-region sums exceeding 20/24, 40000 output or 120000 input must refuse.
Current research/full-candidate reports remain distinct from archived accounting
and cannot reuse CN native/proof data. Preparing must describe carried balances
and still require a fresh exact-source review; old prepared collections expire.
The first14 old own support selectors, four effects and 374/398 old budgets
remain unchanged; only the three inactive MiniMax combination labels change.

The current MiniMax executable inventory uses international `.io` identities,
`minimax-global-standard` / `minimax-global-mplan` account aliases and `global`
region aliases. Global proof identities stay exact; shared accounting slots are
only route/model/purpose and stage. Current summary rows carry the exact
`identity` dict and a separate `budgetSlot`; stage summaries group `budgetSlots`.
Archived accounting rows carry no current proof identity. All current job names begin `mm-global-`.
The current M Plan scope is independent from standard scope; an old CN source
review cannot authorize either global route.

`historicalAccounting` is a single fixed accounting inventory, never a runtime
or proof compatibility layer. It pins only the original common-root CN intent
SHA cfac130565e33b804d22eb2bd7feb58e66d8d6baa12e2094edc4cdcf9f6dfff0
and the first1804 canonical candidate-ledger bytes SHA
222c68c07ce06819b6c84f6aa739b0a08824f1b70893bae7accd014bcae2b4f8.
Those bytes contribute2 calls,2 attempts,5120 output reservation and4096 input
reservation once to the candidate Chat/M3/standard slot. The candidate stage
therefore begins at2/120 calls and2/144 attempts; that slot begins at2/20 and
2/24. Current component reports add their own counts; cumulative ledger totals
are displayed as `cumulativeSlotBudget` separately and never added again.
The immutable historical intent retains its original historical field name;
current global reports do not accept that old metadata label. Historical missing/changed
bytes produce UNKNOWN and a full slot reservation, and refuse paid launch
before a new intent or credential acquisition. Preparing always records carried
history and still requires a new committed-source/binary/review freeze.

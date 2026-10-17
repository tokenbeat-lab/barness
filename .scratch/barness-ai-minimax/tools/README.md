# MiniMax task-10 bounded collection launcher

The executable boundary is `python3 .scratch/barness-ai-minimax/tools/paid.py`.
No command approves its own source. All actual vendor runs require the parent
reviewer's exact-source `APPROVED` receipt, the prepared binaries, the committed
source set and successful credential-free checks.

```sh
python3 .scratch/barness-ai-minimax/tools/paid.py prepare .evidence/minimax-ticket10-candidate
python3 .scratch/barness-ai-minimax/tools/paid.py check .evidence/minimax-ticket10-candidate
python3 .scratch/barness-ai-minimax/tools/paid.py preview .evidence/minimax-ticket10-candidate mm-research-m3-chat
```

`prepare` refuses dirty/untracked execution source, freezes all recursively found
ai Go/JSON plus the actual MiniMax source snapshots, module files and every
launcher Python script, builds live/text/visual binaries without credentials and
writes `prerelease-source-review.template.json` with `PENDING`. After reviewing,
the parent writes `prerelease-source-review.json` using the exact hashes from
that template, `reviewedBy`, `decision: APPROVED`, explicit `authorizedRoutes` and
`credentialScopes: {"MINIMAX_KEY":"cn-standard"}`. M3.1 is independently opt-in:
its route requires `MINIMAX_MPLAN_KEY: cn-mplan` in credentialScopes and that
separate selected variable in the common repository's `.env`.

```sh
python3 .scratch/barness-ai-minimax/tools/paid.py run-selected .evidence/minimax-ticket10-candidate mm-research-m3-chat
python3 .scratch/barness-ai-minimax/tools/paid.py run-selected .evidence/minimax-ticket10-candidate mm-candidate-m3-chat
python3 .scratch/barness-ai-minimax/tools/paid.py summarize .evidence/minimax-ticket10-candidate
```

Repeat only the explicitly reviewed protocol names `chat`, `responses`, and
`messages`; there is no batch paid command or process retry. The fixed full
inventory has13 cases plus one possible tool-call rerun; optional research has3
cases. Both share the same per-identity20calls/24attempts/40000output/120000input
harness ledger. Candidate stage120/144 and formal60/72 survive new collection
roots and worktrees. Actual ledgers always use the common Git repository root
`.evidence/minimax-stages/{candidate,formal}.jsonl`, with no override. Launcher
exclusive one-shot intents use `.evidence/minimax-launches/<route>.json` in that
same common root. Missing/invalid/interrupted outcomes conservatively consume
the original whole identity/component allowance. Never delete these receipts
to rerun; a budget/inventory change requires a separate approved specification.

After own candidate acceptance and the parent-owned data-only catalog promotion,
prepare a fresh collection root and review its new source/catalog/binary pins
before `mm-formal-m3-<protocol>` or any old-source renewal. This is required even
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
argv after all own bundles and the derived Jev support bundle exist. It preserves
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
